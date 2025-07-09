package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/yourusername/academy/internal/config"
	"github.com/yourusername/academy/internal/database"
	"github.com/yourusername/academy/internal/handlers"
	"github.com/yourusername/academy/internal/middleware"
	"github.com/yourusername/academy/internal/telegrambot"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize configuration
	cfg := config.New()

	// Setup database connection
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Setup Redis connection
	redis, err := database.ConnectRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := gin.Default()

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:8080"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	// Set up static file serving
	router.Static("/static", "./web/static")

	// Load HTML templates
	router.LoadHTMLGlob("web/templates/**/*")

	// Apply middlewares
	router.Use(middleware.Logger())

	// Register routes
	handlers.RegisterRoutes(router, db, redis, cfg)

	// Define port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize Telegram bot
	serverURL := "http://localhost:" + port
	if os.Getenv("SERVER_URL") != "" {
		serverURL = os.Getenv("SERVER_URL")
	}

	// Only start the bot if token is provided
	if cfg.Telegram.BotToken != "" {
		bot, err := telegrambot.New(cfg, redis, db, serverURL)
		if err != nil {
			log.Printf("Warning: Failed to initialize Telegram bot: %v", err)
		} else {
			// Start the bot in a goroutine
			go bot.Start()
			log.Println("Telegram bot started successfully")
		}
	} else {
		log.Println("No Telegram bot token provided, skipping bot initialization")
	}

	log.Printf("Server starting on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
