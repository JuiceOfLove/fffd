package controllers

import (
	"os"
	"time"

	"diplom/config"
	"diplom/mail"
	"diplom/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// CreateFamilyInput – структура для создания семьи.
type CreateFamilyInput struct {
	Name string `json:"name"`
}

// CreateFamily создает новую семью, одновременно создавая запись в Calendar,
// и обновляет данные пользователя, если он еще не состоит в семье.
func CreateFamily(c *fiber.Ctx) error {
	// Извлекаем данные пользователя из jwt.MapClaims
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка извлечения данных из токена",
		})
	}
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка извлечения user_id из токена",
		})
	}
	userID := uint(userIDFloat)

	// Находим пользователя
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Пользователь не найден",
		})
	}

	// Если уже есть семья
	if user.FamilyID != 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Вы уже состоите в семье",
		})
	}

	// Считываем входные данные
	var input CreateFamilyInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Невозможно разобрать JSON",
		})
	}

	// Создаем семью
	family := models.Family{
		Name:      input.Name,
		OwnerID:   user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := config.DB.Create(&family).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка создания семьи",
		})
	}

	// Создаем календарь, связанный с семьёй
	calendar := models.Calendar{
		FamilyID:  family.ID,
		Title:     "", // Можно сразу "Семейный календарь" или оставить пустым
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := config.DB.Create(&calendar).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка создания календаря для семьи",
		})
	}

	// Привязываем пользователя к семье
	user.FamilyID = family.ID
	config.DB.Save(&user)

	return c.JSON(fiber.Map{
		"family": family,
	})
}

// InviteInput – структура для приглашения члена семьи.
type InviteInput struct {
	Email string `json:"email"`
}

// InviteMember создает приглашение в семью для указанного email и отправляет приглашение.
func InviteMember(c *fiber.Ctx) error {
	var input InviteInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Невозможно разобрать JSON",
		})
	}

	// Извлекаем данные отправителя из JWT-claims.
	claims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка извлечения данных из токена",
		})
	}
	senderIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка извлечения user_id из токена",
		})
	}
	senderID := uint(senderIDFloat)

	var inviter models.User
	if err := config.DB.First(&inviter, senderID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка извлечения отправителя",
		})
	}
	if inviter.FamilyID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Вы не состоите в семье, не можете приглашать",
		})
	}

	// Проверяем, существует ли пользователь с данным email.
	var invitee models.User
	if err := config.DB.Where("email = ?", input.Email).First(&invitee).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Пользователь с таким email не найден",
		})
	}
	if invitee.FamilyID != 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Пользователь уже состоит в семье",
		})
	}

	// Генерируем токен приглашения.
	inviteToken := uuid.New().String()

	// Создаем приглашение в БД.
	invitation := models.FamilyInvitation{
		FamilyID:  inviter.FamilyID,
		Email:     invitee.Email,
		Token:     inviteToken,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := config.DB.Create(&invitation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка создания приглашения",
		})
	}

	// Формируем ссылку для приглашения.
	inviteLink := os.Getenv("CLIENT_URL") + "/dashboard/family/invite/" + inviteToken
	mailService := mail.NewMailService()
	if err := mailService.SendFamilyInviteMail(invitee.Email, inviteLink); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка отправки приглашения",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Приглашение отправлено",
	})
}

// AcceptInvitation принимает приглашение и обновляет FamilyID у приглашенного пользователя.
func AcceptInvitation(c *fiber.Ctx) error {
	token := c.Params("token")
	var invitation models.FamilyInvitation
	if err := config.DB.Where("token = ?", token).First(&invitation).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный токен приглашения",
		})
	}

	// Находим пользователя по email из приглашения.
	var user models.User
	if err := config.DB.Where("email = ?", invitation.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Пользователь не найден",
		})
	}

	// Обновляем FamilyID у пользователя.
	user.FamilyID = invitation.FamilyID
	config.DB.Save(&user)

	// Удаляем приглашение.
	config.DB.Delete(&invitation)

	return c.JSON(fiber.Map{"message": "Приглашение принято. Вы вступили в семью."})
}

func GetFamilyDetails(c *fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Вы не состоите в семье"})
	}

	// Получаем данные семьи
	var family models.Family
	if err := config.DB.First(&family, user.FamilyID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Семья не найдена"})
	}

	// Получаем всех членов семьи
	var members []models.User
	if err := config.DB.Where("family_id = ?", user.FamilyID).Find(&members).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка загрузки членов семьи"})
	}

	return c.JSON(fiber.Map{
		"family":  family,
		"members": members,
	})
}
