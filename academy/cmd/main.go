package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/globallstudent/academy/docs"
	"github.com/globallstudent/academy/internal/auth"
	"github.com/globallstudent/academy/internal/config"
	"github.com/globallstudent/academy/internal/database"
	"github.com/globallstudent/academy/internal/handlers"
	"github.com/globallstudent/academy/internal/middleware"
	"github.com/globallstudent/academy/internal/telegrambot"
	"github.com/globallstudent/academy/internal/template"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Summer Academy API
// @version         1.0
// @description     API Server for Summer Academy educational platform
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey  JWT
// @in                          header
// @name                        Authorization
// @description                 Bearer JWT token for authentication

// @securityDefinitions.apikey  JWTCookie
// @in                          cookie
// @name                        academy_session
// @description                 JWT token stored in cookie for authentication

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize configuration
	cfg := config.New()

	// Configure authentication parameters
	auth.SetJWTSecret(cfg.Auth.JWTSecret)
	middleware.SetCookieName(cfg.Auth.CookieName)

	// Setup database connection
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Setup Redis connection (with fallback for development)
	var redis *database.Redis

	// Try connecting to Redis with a timeout
	connectCh := make(chan *database.Redis)
	errCh := make(chan error)

	go func() {
		redisConn, connectErr := database.ConnectRedis(cfg.Redis)
		if connectErr != nil {
			errCh <- connectErr
			return
		}
		connectCh <- redisConn
	}()

	// Use a timeout to prevent hanging if Redis is down
	select {
	case redis = <-connectCh:
		log.Println("Successfully connected to Redis")
		defer redis.Close()
	case err = <-errCh:
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		if cfg.Environment == "production" {
			log.Println("Redis connection required in production. Exiting.")
			os.Exit(1)
		}
		log.Println("Continuing without Redis for development purposes.")
		redis = nil
	case <-time.After(5 * time.Second):
		log.Printf("Warning: Redis connection timed out")
		if cfg.Environment == "production" {
			log.Println("Redis connection required in production. Exiting.")
			os.Exit(1)
		}
		log.Println("Continuing without Redis for development purposes.")
		redis = nil
	}

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := gin.Default()
	// Do not trust all proxies by default
	router.SetTrustedProxies(nil)

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:8080"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	// Set up static file serving
	router.Static("/static", "./web/static")

	// Load HTML templates with functions
	router.SetFuncMap(template.Functions())
	router.LoadHTMLGlob("web/templates/**/*")

	// Apply middlewares
	router.Use(middleware.Logger())

	// Register routes
	handlers.RegisterRoutes(router, db, redis, cfg)

	// Swagger documentation
	docs.SwaggerInfo.BasePath = "/"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
			// Connect development OTP store if Redis is not available
			if redis == nil && cfg.Environment != "production" {
				handlers.ExportedStoreDevelopmentOTP = handlers.StoreDevelopmentOTP
				bot.SetDevOTPStore(handlers.ExportedStoreDevelopmentOTP)
			}

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
