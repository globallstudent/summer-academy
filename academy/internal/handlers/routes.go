package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/academy/internal/config"
	"github.com/yourusername/academy/internal/database"
	"github.com/yourusername/academy/internal/middleware"
)

// RegisterRoutes sets up all the routes for the application
func RegisterRoutes(router *gin.Engine, db *database.DB, redis *database.Redis, cfg *config.Config) {
	// Create handler groups
	publicHandlers := NewPublicHandlers(db, redis, cfg)
	problemHandlers := NewProblemHandlers(db, cfg)
	submissionHandlers := NewSubmissionHandlers(db, cfg)
	userHandlers := NewUserHandlers(db, cfg)
	wbfyHandlers := NewWBFYHandlers(db, cfg)

	// Public routes (no auth required)
	router.GET("/", publicHandlers.HomePage)
	router.GET("/login", publicHandlers.LoginPage)
	router.GET("/verify", publicHandlers.VerifyOTPPage)
	router.POST("/login", publicHandlers.ProcessLogin)
	router.GET("/leaderboard", publicHandlers.LeaderboardPage)

	// Auth required routes
	authenticated := router.Group("/")
	authenticated.Use(middleware.Auth())
	{
		// Days and problems
		authenticated.GET("/days", problemHandlers.ListDays)
		authenticated.GET("/days/:day", problemHandlers.DayDetail)
		authenticated.GET("/problems/:slug", problemHandlers.ProblemDetail)

		// Submissions
		authenticated.GET("/submit/:slug", submissionHandlers.SubmitPage)
		authenticated.POST("/submit/:slug", submissionHandlers.ProcessSubmission)
		authenticated.POST("/test/:slug", submissionHandlers.TestSubmission)

		// User profile
		authenticated.GET("/profile", userHandlers.ProfilePage)
		authenticated.POST("/profile", userHandlers.UpdateProfile)

		// WBFY Terminal integration
		authenticated.POST("/terminal/:slug", wbfyHandlers.CreateTerminal)
		authenticated.GET("/terminal/:id", wbfyHandlers.TerminalPage)
	}

	// Admin routes
	admin := router.Group("/admin")
	admin.Use(middleware.Auth(), middleware.AdminOnly())
	{
		admin.GET("/", userHandlers.AdminDashboard)
		admin.GET("/users", userHandlers.UserList)
		admin.GET("/problems", problemHandlers.AdminProblemList)
		admin.POST("/problems", problemHandlers.CreateProblem)
		admin.PUT("/problems/:id", problemHandlers.UpdateProblem)
	}
}
