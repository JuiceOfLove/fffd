package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"diplom/config"
	"diplom/models"
)

// GetPaymentHistory возвращает историю транзакций
// Доступно только для администраторов (role == "admin")
func GetPaymentHistory(c *fiber.Ctx) error {
	// Проверяем JWT и извлекаем claims
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Нет JWT claims"})
	}
	// Проверяем роль
	role, ok := claims["role"].(string)
	if !ok || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа"})
	}

	// Загружаем все платежи, сортировка по дате создания (от новых к старым)
	var payments []models.Payment
	if err := config.DB.Order("created_at DESC").Find(&payments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка загрузки транзакций"})
	}

	return c.JSON(payments)
}
