package main

import (
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"SmartExpenseAI/internal/database"
	"SmartExpenseAI/internal/routes"
	"SmartExpenseAI/internal/services"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize database connection
	database.InitDB()

	// Get bot token from environment
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	// Initialize bot for scheduler
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("Failed to create new bot: ", err)
	}

	// Get user ID from environment for the scheduled recap
	allowedUserIDStr := os.Getenv("TELEGRAM_USER_ID")
	if allowedUserIDStr == "" {
		log.Fatal("TELEGRAM_USER_ID environment variable is not set")
	}

	allowedUserID, err := strconv.ParseInt(allowedUserIDStr, 10, 64)
	if err != nil {
		log.Fatal("TELEGRAM_USER_ID must be a valid integer: ", err)
	}

	// Create Fiber app - only for webhook handling
	app := fiber.New()

	// Initialize Telegram bot and user ID
	routes.InitTelegram(bot, allowedUserID)

	// Register Telegram routes
	routes.TelegramRoutes(app)

	// Schedule weekly recap
	go services.ScheduleWeeklyRecap(bot, allowedUserID)

	// Get port from environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		portNum = 8080
	}

	// Start the server
	log.Printf("Server starting on port %d", portNum)
	log.Fatal(app.Listen(":" + strconv.Itoa(portNum)))
}
