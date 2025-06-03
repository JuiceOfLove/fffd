package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"diplom/config"
	"diplom/models"
)

// DTO для запроса к YooKassa
type CreatePaymentRequest struct {
	Amount struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Confirmation struct {
		Type      string `json:"type"`
		ReturnURL string `json:"return_url"`
	} `json:"confirmation"`
	Capture     bool   `json:"capture"`
	Description string `json:"description"`
}

// DTO для ответа от YooKassa
type CreatePaymentResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Amount struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Confirmation struct {
		Type            string `json:"type"`
		ConfirmationURL string `json:"confirmation_url"`
	} `json:"confirmation"`
	Paid        bool   `json:"paid"`
	CreatedAt   string `json:"created_at"`
	Description string `json:"description"`
}

// BuySubscription — инициация платежа
func BuySubscription(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	userID := uint(claims["user_id"].(float64))

	// Находим пользователя
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	if user.FamilyID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Нет семьи"})
	}

	// Сумма
	amountStr := os.Getenv("PAYMENT_AMOUNT")
	if amountStr == "" {
		amountStr = "300"
	}
	valFloat, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		valFloat = 199.0
	}
	finalSum := fmt.Sprintf("%.2f", valFloat)

	// Формируем тело запроса
	reqBody := CreatePaymentRequest{
		Capture:     true,
		Description: fmt.Sprintf("Подписка на 1 месяц для семьи #%d", user.FamilyID),
	}
	reqBody.Amount.Value = finalSum
	reqBody.Amount.Currency = "RUB"
	reqBody.Confirmation.Type = "redirect"
	reqBody.Confirmation.ReturnURL = os.Getenv("CLIENT_URL") + "/dashboard/payment-success"

	jsonData, _ := json.Marshal(reqBody)

	yoykassaURL := "https://api.yookassa.ru/v3/payments"
	httpReq, err := http.NewRequest("POST", yoykassaURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка формирования запроса"})
	}

	shopID := os.Getenv("YOO_KASSA_SHOP_ID")
	secretKey := os.Getenv("YOO_KASSA_SECRET_KEY")
	httpReq.SetBasicAuth(shopID, secretKey)

	httpReq.Header.Set("Content-Type", "application/json")
	// Уникальный ключ для идемпотентности
	idempotenceKey := uuid.New().String()
	httpReq.Header.Set("Idempotence-Key", idempotenceKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка при запросе к YooKassa"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		var failBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&failBody)
		return c.Status(500).JSON(fiber.Map{
			"error":  "YooKassa error",
			"detail": failBody,
		})
	}

	var payResp CreatePaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payResp); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка парсинга ответа YooKassa"})
	}

	confirmationURL := payResp.Confirmation.ConfirmationURL
	if confirmationURL == "" {
		return c.Status(500).JSON(fiber.Map{"error": "Не удалось получить ссылку на оплату"})
	}

	// Сохраняем запись Payment
	payment := models.Payment{
		PaymentID: payResp.ID,
		FamilyID:  user.FamilyID,
		UserID:    user.ID,
		Amount:    finalSum,
		Status:    payResp.Status, // обычно "pending"
	}
	if err := config.DB.Create(&payment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Ошибка сохранения Payment",
		})
	}

	// Возвращаем ссылку фронту
	return c.JSON(fiber.Map{
		"payment_url": confirmationURL,
	})
}

// YooKassaWebhook — публичный endpoint
func YooKassaWebhook(c *fiber.Ctx) error {
	type YooKassaCallback struct {
		Object struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"object"`
	}

	var cb YooKassaCallback
	if err := c.BodyParser(&cb); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad json"})
	}

	paymentID := cb.Object.ID
	status := cb.Object.Status
	log.Printf("Webhook from YooKassa: payment_id=%s, status=%s\n", paymentID, status)

	// Ищем Payment в своей БД
	var payment models.Payment
	if err := config.DB.Where("payment_id = ?", paymentID).First(&payment).Error; err != nil {
		log.Println("Payment not found:", paymentID)
		return c.JSON(fiber.Map{"message": "payment not found"})
	}

	// Обновляем статус
	payment.Status = status
	config.DB.Save(&payment)

	if status == "succeeded" {
		// Активируем подписку
		var sub models.FamilySubscription
		err := config.DB.Where("family_id = ?", payment.FamilyID).First(&sub).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Если записи нет — создаём новую
			sub = models.FamilySubscription{
				FamilyID:  payment.FamilyID,
				IsActive:  true,
				StartDate: time.Now(),
				EndDate:   time.Now().AddDate(0, 1, 0),
			}
			config.DB.Create(&sub)
		} else if err != nil {
			// Если другая ошибка БД — логируем, но не возвращаем 500 клиенту
			log.Printf("Ошибка при получении подписки: %v\n", err)
		} else {
			// Обновляем существующую
			sub.IsActive = true
			sub.StartDate = time.Now()
			sub.EndDate = time.Now().AddDate(0, 1, 0)
			config.DB.Save(&sub)
		}

		log.Printf("Подписка активирована для FamilyID=%d\n", payment.FamilyID)
	}

	return c.JSON(fiber.Map{"message": "ok"})
}

// CheckSubscription — смотрим, есть ли активная подписка
func CheckSubscription(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	userID := uint(claims["user_id"].(float64))

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	if user.FamilyID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "У пользователя нет семьи"})
	}

	var sub models.FamilySubscription
	err := config.DB.
		Where("family_id = ?", user.FamilyID).
		First(&sub).
		Error

	// Если записи нет — просто возвращаем isActive: false
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(fiber.Map{"isActive": false})
	}
	// Если другая ошибка БД — возвращаем 500
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка базы данных"})
	}

	now := time.Now()
	isActive := sub.IsActive && now.Before(sub.EndDate)
	return c.JSON(fiber.Map{"isActive": isActive})
}