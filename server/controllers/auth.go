package controllers

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"diplom/config"
	"diplom/mail"
	"diplom/models"
	"diplom/utils"
)

// RegisterInput — структура для входящего JSON при регистрации.
type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register обрабатывает регистрацию нового пользователя с отправкой письма активации.
func Register(c *fiber.Ctx) error {
	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Невозможно разобрать JSON"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при хэшировании пароля"})
	}

	activationLink := uuid.New().String()

	user := models.User{
		Name:           input.Name,
		Email:          input.Email,
		Password:       string(hashedPassword),
		IsActivated:    false,
		ActivationLink: activationLink,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Не удалось создать пользователя"})
	}

	// Отправляем письмо с активацией
	mailService := mail.NewMailService()
	activationURL := os.Getenv("CLIENT_URL") + "/auth/activate/" + activationLink
	if err := mailService.SendActivationMail(user.Email, activationURL); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка отправки письма с активацией"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Пользователь зарегистрирован. Проверьте вашу почту для активации аккаунта."})
}

// LoginInput — структура для входящего JSON при логине.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login обрабатывает авторизацию и возвращает access-токен, устанавливая refresh-токен в httpOnly cookie.
func Login(c *fiber.Ctx) error {
	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Невозможно разобрать JSON"})
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		if err == fiber.ErrNotFound || err == (config.DB.Error) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Неверные учетные данные"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка сервера"})
	}

	// Если аккаунт не активирован
	if !user.IsActivated {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Аккаунт не активирован. Проверьте вашу почту."})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Неверные учетные данные"})
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка генерации access токена"})
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка генерации refresh токена"})
	}

	// Сохраняем refresh-токен в БД (опционально)
	tokenEntry := models.Token{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := config.DB.Create(&tokenEntry).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка сохранения refresh токена"})
	}

	// Устанавливаем httpOnly cookie для refresh токена
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
		Domain:   "localhost", // настройка для разработки; в production корректируйте
	})

	return c.JSON(fiber.Map{
		"access_token": accessToken,
		"user":         user,
	})
}

// Activate активирует аккаунт по ссылке.
func Activate(c *fiber.Ctx) error {
	link := c.Params("link")
	var user models.User
	if err := config.DB.Where("activation_link = ?", link).First(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверная ссылка активации"})
	}
	// Если пользователь уже активирован, просто возвращаем сообщение об этом
	if user.IsActivated {
		return c.JSON(fiber.Map{"message": "Ваш аккаунт уже активирован"})
	}

	user.IsActivated = true
	user.ActivationLink = ""
	config.DB.Save(&user)
	return c.JSON(fiber.Map{"message": "Ваш аккаунт успешно активирован"})
}

// Refresh обрабатывает обновление access-токена, читая refresh-токен из httpOnly cookie.
func Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Refresh токен не найден"})
	}

	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(refreshSecret), nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Неверный или просроченный refresh токен"})
	}
	claims := token.Claims.(jwt.MapClaims)
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Неверные данные в токене"})
	}
	userID := uint(userIDFloat)

	var storedToken models.Token
	if err := config.DB.Where("user_id = ? AND token = ?", userID, refreshToken).First(&storedToken).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Refresh токен не найден"})
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Refresh токен истек"})
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Пользователь не найден"})
	}
	newAccessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка генерации нового access токена"})
	}

	return c.JSON(fiber.Map{
		"access_token": newAccessToken,
		"user":         user,
	})
}

// Logout очищает refresh-токен из БД и cookie.
func Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Refresh токен не найден"})
	}

	if err := config.DB.Where("token = ?", refreshToken).Delete(&models.Token{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при выходе из системы"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
		Domain:   "localhost",
	})

	return c.JSON(fiber.Map{"message": "Вы успешно вышли из системы"})
}