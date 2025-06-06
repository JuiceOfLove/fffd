package controllers

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"

	"diplom/config"
	"diplom/models"
	"diplom/utils"
)

var (
	newTicketConns   = make(map[*websocket.Conn]bool)
	newTicketConnsMu = utils.NewMutex()
)

type ticketClient struct {
	UserID uint
	Role   string
}

var (
	ticketRooms   = make(map[uint]map[*websocket.Conn]ticketClient)
	ticketRoomsMu = utils.NewMutex()
)

type CreateTicketInput struct {
	Subject string `json:"subject"`
	Content string `json:"content"`
}

type ITicketInfo struct {
	ID            uint      `json:"id"`
	Subject       string    `json:"subject"`
	Status        string    `json:"status"` // "new", "active" или "closed"
	UserID        uint      `json:"user_id"`
	UserName      string    `json:"user_name"`
	OperatorID    *uint     `json:"operator_id"`     // nil, если ещё не назначен
	OperatorName  *string   `json:"operator_name"`   // nil, если ещё не назначен
	LastMessageAt time.Time `json:"last_message_at"` // время последнего сообщения
}

// CreateTicket — создаёт новый тикет и первое сообщение
func CreateTicket(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	userID := uint(claims["user_id"].(float64))

	var input CreateTicketInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Невозможно разобрать JSON"})
	}

	ticket := models.Ticket{
		UserID:        userID,
		Subject:       input.Subject,
		Status:        "new",
		LastMessageAt: time.Now(),
	}
	if err := config.DB.Create(&ticket).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при создании тикета"})
	}

	msg := models.TicketMessage{
		TicketID:   ticket.ID,
		SenderID:   userID,
		SenderRole: "user",
		Content:    input.Content,
		CreatedAt:  time.Now(),
	}
	if err := config.DB.Create(&msg).Error; err != nil {
		config.DB.Delete(&ticket)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Тикет создан, но не удалось сохранить сообщение"})
	}

	notifyNewTicketWS(ticket)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"ticket_id": ticket.ID})
}

// GetMyTickets — возвращает список тикетов текущего пользователя
func GetMyTickets(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	userID := uint(claims["user_id"].(float64))

	var tickets []models.Ticket
	if err := config.DB.
		Where("user_id = ?", userID).
		Order("last_message_at DESC").
		Find(&tickets).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось получить тикеты"})
	}

	out := make([]fiber.Map, 0, len(tickets))
	for _, t := range tickets {
		statusLabel := ""
		switch t.Status {
		case "new":
			statusLabel = "Новый"
		case "active":
			statusLabel = "В работе"
		case "closed":
			statusLabel = "Закрыт"
		}
		out = append(out, fiber.Map{
			"id":              t.ID,
			"subject":         t.Subject,
			"status":          t.Status,
			"status_label":    statusLabel,
			"last_message_at": t.LastMessageAt,
			"operator_id":     t.OperatorID,
		})
	}
	return c.JSON(out)
}

// GetTicketInfo — возвращает детальную информацию по тикету, включая имена user и operator
func GetTicketInfo(c *fiber.Ctx) error {
	// 1. Получаем сам тикет
	ticketID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверный ID тикета"})
	}
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Тикет не найден"})
	}

	// 2. Достаём JWT-claims
	claimsAny := c.Locals("user")
	if claimsAny == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	claims := claimsAny.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	role := claims["role"].(string)

	// 3. Проверка прав: пользователь — только свои тикеты; оператор — любой.
	if role != "operator" && role != "admin" && ticket.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ запрещён"})
	}

	// 4. Достаём имя создателя тикета (пользователя)
	var user models.User
	if err := config.DB.First(&user, ticket.UserID).Error; err != nil {
		// теоретически такого не бывает, но на всякий случай:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось получить информацию о пользователе"})
	}

	// 5. Достаём информацию об операторе (если назначен)
	var operatorName *string
	if ticket.OperatorID != nil {
		var op models.User
		if err := config.DB.First(&op, *ticket.OperatorID).Error; err == nil {
			operatorName = &op.Name
		}
	}

	// 6. Собираем ответ
	out := ITicketInfo{
		ID:            ticket.ID,
		Subject:       ticket.Subject,
		Status:        ticket.Status,
		UserID:        ticket.UserID,
		UserName:      user.Name,
		OperatorID:    ticket.OperatorID,
		OperatorName:  operatorName,
		LastMessageAt: ticket.LastMessageAt,
	}

	return c.JSON(out)
}

// GetTicketMessages — возвращает историю сообщений тикета (доступ по тому же принципу)
func GetTicketMessages(c *fiber.Ctx) error {
	ticketID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверный ID тикета"})
	}
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Тикет не найден"})
	}

	claims, _ := c.Locals("user").(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	role := claims["role"].(string)

	isOwner := ticket.UserID == userID
	isAssignedOperator := role == "operator" && ticket.OperatorID != nil && *ticket.OperatorID == userID
	isAdmin := role == "admin"

	if !isOwner && !isAssignedOperator && !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ запрещён"})
	}

	var msgs []models.TicketMessage
	if err := config.DB.
		Where("ticket_id = ?", ticket.ID).
		Order("created_at ASC").
		Find(&msgs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось загрузить сообщения"})
	}
	return c.JSON(msgs)
}

// ListOperatorTickets — возвращает список тикетов для оператора по статусу
func ListOperatorTickets(c *fiber.Ctx) error {
	claims, _ := c.Locals("user").(jwt.MapClaims)
	role := claims["role"].(string)
	if role != "operator" && role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ только для операторов"})
	}

	status := c.Query("status", "new")
	if status != "new" && status != "active" && status != "closed" {
		status = "new"
	}

	var tickets []models.Ticket
	if err := config.DB.
		Where("status = ?", status).
		Order("last_message_at DESC").
		Find(&tickets).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось получить тикеты"})
	}

	out := make([]fiber.Map, 0, len(tickets))
	for _, t := range tickets {
		var user models.User
		config.DB.First(&user, t.UserID)
		out = append(out, fiber.Map{
			"id":              t.ID,
			"subject":         t.Subject,
			"user_id":         t.UserID,
			"user_name":       user.Name,
			"last_message_at": t.LastMessageAt,
			"status":          t.Status,
			"operator_id":     t.OperatorID,
		})
	}
	return c.JSON(out)
}

// AssignTicket — оператор берёт тикет в работу
func AssignTicket(c *fiber.Ctx) error {
	claims, _ := c.Locals("user").(jwt.MapClaims)
	role := claims["role"].(string)
	if role != "operator" && role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ только для операторов"})
	}
	operatorID := uint(claims["user_id"].(float64))

	ticketID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверный ID тикета"})
	}
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Тикет не найден"})
	}
	if ticket.Status != "new" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Тикет уже взят или закрыт"})
	}

	ticket.OperatorID = &operatorID
	ticket.Status = "active"
	ticket.LastMessageAt = time.Now()
	if err := config.DB.Save(&ticket).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось присвоить тикет"})
	}

	notifyTicketAssignedWS(ticket)
	return c.JSON(fiber.Map{"message": "Тикет взят в работу"})
}

// CloseTicket — оператор (или admin) закрывает тикет
func CloseTicket(c *fiber.Ctx) error {
	claims, _ := c.Locals("user").(jwt.MapClaims)
	role := claims["role"].(string)
	if role != "operator" && role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ только для операторов"})
	}

	ticketID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверный ID тикета"})
	}
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Тикет не найден"})
	}
	if ticket.Status != "active" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Тикет не в работе"})
	}

	ticket.Status = "closed"
	ticket.LastMessageAt = time.Now()
	if err := config.DB.Save(&ticket).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось закрыть тикет"})
	}

	notifyTicketClosedWS(ticket)
	return c.JSON(fiber.Map{"message": "Тикет закрыт"})
}

type CreateTicketMessageInput struct {
	Content   string  `json:"content"`
	MediaURL  *string `json:"media_url,omitempty"`
	ReplyToID *uint   `json:"reply_to_id,omitempty"`
}

// CreateTicketMessage — создаёт новое сообщение в тикете
func CreateTicketMessage(c *fiber.Ctx) error {
	ticketID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверный ID тикета"})
	}
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Тикет не найден"})
	}

	claims, _ := c.Locals("user").(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	role := claims["role"].(string)

	isOwner := ticket.UserID == userID
	isAssignedOperator := role == "operator" && ticket.OperatorID != nil && *ticket.OperatorID == userID
	isAdmin := role == "admin"
	if !isOwner && !isAssignedOperator && !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Доступ запрещён"})
	}

	var input CreateTicketMessageInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверные данные сообщения"})
	}

	senderRole := "user"
	if role == "operator" || role == "admin" {
		senderRole = "operator"
	}
	msg := models.TicketMessage{
		TicketID:   ticket.ID,
		SenderID:   userID,
		SenderRole: senderRole,
		Content:    input.Content,
		MediaURL:   input.MediaURL,
		ReplyToID:  input.ReplyToID,
		CreatedAt:  time.Now(),
	}
	if err := config.DB.Create(&msg).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось сохранить сообщение"})
	}

	ticket.LastMessageAt = time.Now()
	config.DB.Save(&ticket)

	notifyTicketMessageWS(msg)
	return c.Status(fiber.StatusCreated).JSON(msg)
}

// DeleteTicketMessageHTTP — удаляет (soft-delete) сообщение
func DeleteTicketMessageHTTP(c *fiber.Ctx) error {
	tid, _ := c.ParamsInt("id")
	mid, _ := c.ParamsInt("msgId")
	claims, _ := c.Locals("user").(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	role := claims["role"].(string)

	var msg models.TicketMessage
	if err := config.DB.First(&msg, mid).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Сообщение не найдено"})
	}
	if msg.TicketID != uint(tid) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Сообщение не к этому тикету"})
	}
	if role != "operator" && role != "admin" && msg.SenderID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа"})
	}

	if err := config.DB.Delete(&msg).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Не удалось удалить сообщение"})
	}
	broadcastTicketDelete(uint(tid), msg.ID)
	return c.JSON(fiber.Map{"message": "OK"})
}

// SupportNewTicketsWS — WebSocket для уведомления операторов о новых тикетах
func SupportNewTicketsWS(c *websocket.Conn) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.Close()
		return
	}
	secret := []byte(os.Getenv("JWT_SECRET"))
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !tok.Valid {
		c.Close()
		return
	}
	claims := tok.Claims.(jwt.MapClaims)
	if claims["role"].(string) != "operator" && claims["role"].(string) != "admin" {
		c.Close()
		return
	}

	newTicketConnsMu.Lock()
	newTicketConns[c] = true
	newTicketConnsMu.Unlock()

	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}

	newTicketConnsMu.Lock()
	delete(newTicketConns, c)
	newTicketConnsMu.Unlock()
	c.Close()
}

// SupportTicketChatWS — WebSocket для обмена сообщениями внутри конкретного тикета
func SupportTicketChatWS(c *websocket.Conn) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.Close()
		return
	}
	secret := []byte(os.Getenv("JWT_SECRET"))
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !tok.Valid {
		c.Close()
		return
	}
	claims := tok.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	role := claims["role"].(string)

	ticketIDStr := c.Params("id")
	ticketIDInt, err := strconv.Atoi(ticketIDStr)
	if err != nil {
		c.Close()
		return
	}
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketIDInt).Error; err != nil {
		c.Close()
		return
	}

	isOwner := ticket.UserID == userID
	isAssignedOperator := role == "operator" && ticket.OperatorID != nil && *ticket.OperatorID == userID
	isAdmin := role == "admin"
	if !isOwner && !isAssignedOperator && !isAdmin {
		c.Close()
		return
	}

	tid := ticket.ID
	ticketRoomsMu.Lock()
	if ticketRooms[tid] == nil {
		ticketRooms[tid] = make(map[*websocket.Conn]ticketClient)
	}
	ticketRooms[tid][c] = ticketClient{UserID: userID, Role: role}
	ticketRoomsMu.Unlock()
	broadcastTicketPresence(tid)

	for {
		var raw json.RawMessage
		if err := c.ReadJSON(&raw); err != nil {
			if err == io.EOF || websocket.IsCloseError(err) {
				break
			}
			log.Println("WS read error:", err)
			break
		}

		var envelope struct {
			Content  *string `json:"content"`
			ReplyTo  *uint   `json:"reply_to"`
			MediaB64 *string `json:"media"`
			DeleteID *uint   `json:"delete_id"`
		}
		_ = json.Unmarshal(raw, &envelope)

		if envelope.DeleteID != nil {
			var m models.TicketMessage
			if err := config.DB.First(&m, *envelope.DeleteID).Error; err == nil {
				if role == "operator" || role == "admin" || m.SenderID == userID {
					config.DB.Delete(&m)
					broadcastTicketDelete(tid, m.ID)
				}
			}
			continue
		}

		if envelope.Content == nil && envelope.MediaB64 == nil {
			continue
		}

		var mediaURL *string
		if envelope.MediaB64 != nil {
			if url, err := saveBase64Image(*envelope.MediaB64, userID); err == nil {
				mediaURL = url
			} else {
				log.Println("save image:", err)
			}
		}

		senderRole := "user"
		if role == "operator" || role == "admin" {
			senderRole = "operator"
		}
		msg := models.TicketMessage{
			TicketID:   tid,
			SenderID:   userID,
			SenderRole: senderRole,
			Content:    coalesce(envelope.Content),
			MediaURL:   mediaURL,
			ReplyToID:  envelope.ReplyTo,
			CreatedAt:  time.Now(),
		}
		if err := config.DB.Create(&msg).Error; err != nil {
			log.Println("db create:", err)
			continue
		}
		ticket.LastMessageAt = time.Now()
		config.DB.Save(&ticket)
		notifyTicketMessageWS(msg)
	}

	ticketRoomsMu.Lock()
	delete(ticketRooms[tid], c)
	if len(ticketRooms[tid]) == 0 {
		delete(ticketRooms, tid)
	}
	ticketRoomsMu.Unlock()
	broadcastTicketPresence(tid)
	c.Close()
}

// Ниже идут все вспомогательные функции для рассылки через WebSocket

func notifyNewTicketWS(ticket models.Ticket) {
	payload, _ := json.Marshal(fiber.Map{
		"event": "support:new_ticket",
		"data": fiber.Map{
			"ticket_id":       ticket.ID,
			"subject":         ticket.Subject,
			"user_id":         ticket.UserID,
			"last_message_at": ticket.LastMessageAt,
		},
	})
	newTicketConnsMu.Lock()
	for conn := range newTicketConns {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("WS new_ticket error: %v\n", err)
		}
	}
	newTicketConnsMu.Unlock()
}

func notifyTicketAssignedWS(ticket models.Ticket) {
	payload, _ := json.Marshal(fiber.Map{
		"event": "support:ticket_assigned",
		"data": fiber.Map{
			"ticket_id":   ticket.ID,
			"operator_id": ticket.OperatorID,
		},
	})
	newTicketConnsMu.Lock()
	for conn := range newTicketConns {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("WS ticket_assigned error: %v\n", err)
		}
	}
	newTicketConnsMu.Unlock()
}

func notifyTicketClosedWS(ticket models.Ticket) {
	payload, _ := json.Marshal(fiber.Map{
		"event": "support:ticket_closed",
		"data": fiber.Map{
			"ticket_id": ticket.ID,
		},
	})
	ticketID := ticket.ID

	ticketRoomsMu.Lock()
	for conn := range ticketRooms[ticketID] {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("WS ticket_closed error: %v\n", err)
		}
	}
	ticketRoomsMu.Unlock()
}

func notifyTicketMessageWS(msg models.TicketMessage) {
	payload, _ := json.Marshal(fiber.Map{
		"event": "support:ticket_message",
		"data":  msg,
	})
	ticketID := msg.TicketID

	ticketRoomsMu.Lock()
	for conn := range ticketRooms[ticketID] {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("WS ticket_message error: %v\n", err)
		}
	}
	ticketRoomsMu.Unlock()
}

func broadcastTicketDelete(ticketID, messageID uint) {
	payload, _ := json.Marshal(fiber.Map{
		"event": "support:ticket_delete",
		"data": fiber.Map{
			"message_id": messageID,
		},
	})
	ticketRoomsMu.Lock()
	for conn := range ticketRooms[ticketID] {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("WS ticket_delete error: %v\n", err)
		}
	}
	ticketRoomsMu.Unlock()
}

func broadcastTicketPresence(ticketID uint) {
	ticketRoomsMu.Lock()
	conns := ticketRooms[ticketID]
	ids := make([]uint, 0, len(conns))
	for _, info := range conns {
		if info.Role == "operator" || info.Role == "admin" {
			ids = append(ids, info.UserID)
		}
	}
	payload, _ := json.Marshal(fiber.Map{
		"event": "support:ticket_presence",
		"data":  ids,
	})
	for conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("WS ticket_presence error: %v\n", err)
		}
	}
	ticketRoomsMu.Unlock()
}
