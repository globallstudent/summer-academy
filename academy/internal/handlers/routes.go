package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/globallstudent/academy/internal/config"
	"github.com/globallstudent/academy/internal/database"
	"github.com/globallstudent/academy/internal/middleware"
)

// RegisterRoutes sets up all the routes for the application
func RegisterRoutes(router *gin.Engine, db *database.DB, redis *database.Redis, cfg *config.Config) {
	// Create handler groups
	publicHandlers := NewPublicHandlers(db, redis, cfg)
	problemHandlers := NewProblemHandlers(db, cfg)
	submissionHandlers := NewSubmissionHandlers(db, cfg)
	userHandlers := NewUserHandlers(db, cfg)
	wbfyHandlers := NewWBFYHandlers(db, cfg)
	wbfyHandlers.StartCleanupJob()
	contestHandlers := NewContestHandlers(db, redis, cfg)

	// Public routes
	router.GET("/", publicHandlers.HomePage)
	auth := router.Group("/auth")
	{
		auth.GET("/login", publicHandlers.LoginPage)
		auth.POST("/login", publicHandlers.ProcessLogin)
		auth.GET("/verify", publicHandlers.VerifyOTPPage)
		auth.GET("/logout", publicHandlers.LogoutHandler)
	}

	// Auth required routes
	authenticated := router.Group("/")
	authenticated.Use(middleware.Auth())
	{
		authenticated.GET("/leaderboard", publicHandlers.LeaderboardPage)

		contests := authenticated.Group("/contests")
		{
			contests.GET("", contestHandlers.ListContests)
			contests.GET(":slug", contestHandlers.ContestDetail)
			contests.GET(":slug/join", contestHandlers.JoinContest)
			contests.GET(":slug/leaderboard", contestHandlers.ContestLeaderboard)
		}

		days := authenticated.Group("/days")
		{
			days.GET("", problemHandlers.ListDays)
			days.GET(":day", problemHandlers.DayDetail)
		}

		problems := authenticated.Group("/problems")
		{
			problems.GET(":slug", problemHandlers.ProblemDetail)
		}

		submissions := authenticated.Group("/submissions")
		{
			submissions.GET(":slug", submissionHandlers.SubmitPage)
			submissions.POST(":slug", submissionHandlers.ProcessSubmission)
			submissions.POST("test/:slug", submissionHandlers.TestSubmission)
		}

		users := authenticated.Group("/users")
		{
			users.GET("me", userHandlers.ProfilePage)
			users.POST("me", userHandlers.UpdateProfile)
		}

		terminal := authenticated.Group("/terminal")
		{
			terminal.POST(":slug", wbfyHandlers.CreateTerminal)
			terminal.GET(":id", wbfyHandlers.TerminalPage)
		}
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
