package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"diplom/config"
	"diplom/models"
	"diplom/routes"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Нет файла .env, используем системные переменные")
	}

	db := config.InitDB()
	config.DB = db

	config.DB.AutoMigrate(&models.User{}, &models.Token{}, &models.Family{}, &models.FamilyInvitation{}, &models.Calendar{}, &models.Event{}, &models.FamilySubscription{}, &models.Payment{}, &models.ChatMessage{}, &models.Ticket{}, &models.TicketMessage{},)

	app := fiber.New()

	// CORS с указанием AllowOrigins и AllowCredentials
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
	}))

	app.Static("/uploads", "./public/uploads")

	routes.Setup(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Сервер запускается на порту %s\n", port)
	log.Fatal(app.Listen(":" + port))
}
