package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateAccessToken создает access токен для пользователя с коротким сроком действия.
func GenerateAccessToken(userID uint, email, role string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["email"] = email
	claims["role"] = role
	claims["exp"] = time.Now().Add(15 * time.Minute).Unix() // Access-токен действует 15 минут
	secret := os.Getenv("JWT_SECRET")
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken создает refresh токен для пользователя с более длительным сроком действия.
func GenerateRefreshToken(userID uint) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(7 * 24 * time.Hour).Unix() // Refresh-токен действует 7 дней
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	return token.SignedString([]byte(refreshSecret))
}
