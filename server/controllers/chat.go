package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"

	"diplom/config"
	"diplom/models"
)

/*────────────────────────── globals ──────────────────────────*/

var (
	rooms   = make(map[uint]map[*websocket.Conn]uint) // rooms[familyID] = map[conn]userID
	roomsMu sync.Mutex
)

/*────────────────────────── helpers ──────────────────────────*/

// безопасная запись в сокет (клиент мог уже закрыть вкладку)
func safeWrite(conn *websocket.Conn, typ int, payload []byte) {
	if err := conn.WriteMessage(typ, payload); err != nil && !websocket.IsCloseError(err) {
		log.Printf("WS write error: %v\n", err)
	}
}

// сохраняем base64-картинку, возвращаем относительный URL
func saveBase64Image(dataURL string, userID uint) (*string, error) {
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("bad dataURL")
	}

	raw, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll("./public/uploads", 0755); err != nil {
		return nil, err
	}
	name := fmt.Sprintf("%d_%d.jpg", userID, time.Now().UnixNano())
	path := filepath.Join("public", "uploads", name)

	if err := os.WriteFile(path, raw, 0644); err != nil {
		return nil, err
	}
	url := "/uploads/" + name
	return &url, nil
}

/*────────────────────────── HTTP: история ───────────────────*/

func ChatHistory(c *fiber.Ctx) error {
	claims := c.Locals("user").(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var u models.User
	if err := config.DB.First(&u, userID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	var msgs []models.ChatMessage
	if err := config.DB.
		Where("family_id = ?", u.FamilyID).
		Order("created_at").
		Find(&msgs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot load history"})
	}
	return c.JSON(msgs)
}

/*────────────────────────── WebSocket ────────────────────────*/

func ChatWebSocket(c *websocket.Conn) {
	/* 1. ─── валидация токена ───────────────────────────────*/
	tokStr := c.Query("token")
	if tokStr == "" {
		c.Close(); return
	}

	secret := os.Getenv("JWT_SECRET")
	tok, err := jwt.Parse(tokStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !tok.Valid {
		c.Close(); return
	}
	claims := tok.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.Close(); return
	}
	familyID := user.FamilyID

	/* 2. ─── регистрируем соединение ────────────────────────*/
	roomsMu.Lock()
	if rooms[familyID] == nil {
		rooms[familyID] = make(map[*websocket.Conn]uint)
	}
	rooms[familyID][c] = userID
	roomsMu.Unlock()
	broadcastPresence(familyID)

	/* 3. ─── цикл чтения ────────────────────────────────────*/
	for {
		var raw json.RawMessage
		if err := c.ReadJSON(&raw); err != nil {
			if err == io.EOF || websocket.IsCloseError(err) {
				break
			}
			log.Println("WS read error:", err)
			break
		}

		// --- разбор типа входящего сообщения ---
		var envelope struct {
			Content   *string `json:"content"`
			ReplyTo   *uint   `json:"reply_to"`
			MediaB64  *string `json:"media"`
			DeleteID  *uint   `json:"delete_id"`
		}
		_ = json.Unmarshal(raw, &envelope)

		/* 3a. Удаление своего сообщения */
		if envelope.DeleteID != nil {
			var m models.ChatMessage
			if err := config.DB.First(&m, *envelope.DeleteID).Error; err == nil && m.UserID == userID {
				config.DB.Delete(&m) // soft-delete
				broadcastDelete(familyID, m.ID)
			}
			continue
		}

		/* 3b. Новое сообщение/ответ */
		if envelope.Content == nil && envelope.MediaB64 == nil {
			continue // пустое
		}

		var mediaURL *string
		if envelope.MediaB64 != nil {
			if url, err := saveBase64Image(*envelope.MediaB64, userID); err == nil {
				mediaURL = url
			} else {
				log.Println("save image:", err)
			}
		}

		msg := models.ChatMessage{
			FamilyID:  familyID,
			UserID:    userID,
			Content:   coalesce(envelope.Content),
			MediaURL:  mediaURL,
			ReplyToID: envelope.ReplyTo,
			CreatedAt: time.Now(),
		}
		if err := config.DB.Create(&msg).Error; err != nil {
			log.Println("db create:", err)
			continue
		}
		broadcastMessage(familyID, msg)
	}

	/* 4. ─── отключение клиента ────────────────────────────*/
	roomsMu.Lock()
	delete(rooms[familyID], c)
	roomsMu.Unlock()
	broadcastPresence(familyID)
}

/*────────────────────────── helpers ──────────────────────────*/

func coalesce(ptr *string) string {
	if ptr == nil { return "" }
	return *ptr
}

/*────────────────────────── broadcast ───────────────────────*/

func broadcastMessage(fam uint, msg models.ChatMessage) {
	roomsMu.Lock(); defer roomsMu.Unlock()

	payload, _ := json.Marshal(struct {
		Type string             `json:"type"`
		Data models.ChatMessage `json:"data"`
	}{"message", msg})

	for conn := range rooms[fam] {
		safeWrite(conn, websocket.TextMessage, payload)
	}
}

func broadcastDelete(fam uint, id uint) {
	roomsMu.Lock(); defer roomsMu.Unlock()

	payload, _ := json.Marshal(struct {
		Type string `json:"type"`
		Data uint   `json:"data"`
	}{"delete", id})

	for conn := range rooms[fam] {
		safeWrite(conn, websocket.TextMessage, payload)
	}
}

func broadcastPresence(fam uint) {
	roomsMu.Lock(); defer roomsMu.Unlock()

	online := make([]uint, 0, len(rooms[fam]))
	for _, uid := range rooms[fam] {
		online = append(online, uid)
	}

	payload, _ := json.Marshal(struct {
		Type string `json:"type"`
		Data []uint `json:"data"`
	}{"presence", online})

	for conn := range rooms[fam] {
		safeWrite(conn, websocket.TextMessage, payload)
	}
}
