package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/academy/internal/auth"
	"github.com/yourusername/academy/internal/config"
	"github.com/yourusername/academy/internal/database"
	"github.com/yourusername/academy/internal/models"
)

// PublicHandlers contains handlers for public routes
type PublicHandlers struct {
	db    *database.DB
	redis *database.Redis
	cfg   *config.Config
}

// NewPublicHandlers creates a new PublicHandlers instance
func NewPublicHandlers(db *database.DB, redis *database.Redis, cfg *config.Config) *PublicHandlers {
	return &PublicHandlers{db: db, redis: redis, cfg: cfg}
}

// HomePage handles the home page
func (h *PublicHandlers) HomePage(c *gin.Context) {
	// Get today's problem if available
	todayProblems, err := getTodaysProblems(h.db)
	if err != nil {
		c.HTML(http.StatusOK, "pages/home.html", gin.H{
			"Error": "Failed to get today's problems",
		})
		return
	}

	c.HTML(http.StatusOK, "pages/home.html", gin.H{
		"Title":         "Summer Academy - Learn DSA and Linux",
		"TodayProblems": todayProblems,
	})
}

// LoginPage handles the login page
func (h *PublicHandlers) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/login.html", gin.H{
		"Title": "Login - Summer Academy",
	})
}

// VerifyOTPPage handles the OTP verification page
func (h *PublicHandlers) VerifyOTPPage(c *gin.Context) {
	phoneNumber := c.Query("phone")
	otp := c.Query("otp")

	// Render the verification page
	c.HTML(http.StatusOK, "pages/verify.html", gin.H{
		"Title": "Verify OTP - Summer Academy",
		"Phone": phoneNumber,
		"OTP":   otp,
	})
}

// ProcessLogin handles login form submission
func (h *PublicHandlers) ProcessLogin(c *gin.Context) {
	// Create background context
	_ = context.Background()
	phoneNumber := c.PostForm("phone")
	otp := c.PostForm("otp")

	// Validate input
	if phoneNumber == "" || otp == "" || len(otp) != 6 {
		c.HTML(http.StatusBadRequest, "pages/verify.html", gin.H{
			"Title": "Verify OTP - Summer Academy",
			"Error": "Invalid phone number or OTP",
			"Phone": phoneNumber,
		})
		return
	}

	// Verify OTP against Redis
	isValid, err := h.redis.VerifyOTP(phoneNumber, otp)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "pages/verify.html", gin.H{
			"Title": "Verify OTP - Summer Academy",
			"Error": "An error occurred while verifying your code",
			"Phone": phoneNumber,
		})
		return
	}

	if !isValid {
		c.HTML(http.StatusBadRequest, "pages/verify.html", gin.H{
			"Title": "Verify OTP - Summer Academy",
			"Error": "Invalid or expired verification code",
			"Phone": phoneNumber,
		})
		return
	}

	// Find user by phone number or create a new user
	var user models.User
	// TODO: Replace with actual database query in production
	// Example query: SELECT * FROM users WHERE phone_number = $1

	// For now, simulate user lookup/creation
	user = models.User{
		ID:           uuid.New(),
		PhoneNumber:  phoneNumber,
		Username:     "Student" + phoneNumber[len(phoneNumber)-4:],
		RegisteredAt: time.Now(),
		Role:         "user",
	}

	// In production, store the user in database if new
	// INSERT INTO users (...) VALUES (...) ON CONFLICT DO UPDATE ...

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "pages/verify.html", gin.H{
			"Title": "Verify OTP - Summer Academy",
			"Error": "Failed to generate session token",
			"Phone": phoneNumber,
		})
		return
	}

	// Set cookie
	c.SetCookie(
		h.cfg.Auth.CookieName,
		token,
		h.cfg.Auth.CookieMaxAge,
		"/",
		"",
		false,
		true,
	)

	// Redirect to days page
	c.Redirect(http.StatusFound, "/days")
}

// LeaderboardPage handles the leaderboard page
func (h *PublicHandlers) LeaderboardPage(c *gin.Context) {
	// Get leaderboard entries
	entries, err := getLeaderboard(h.db)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "pages/leaderboard.html", gin.H{
			"Error": "Failed to get leaderboard data",
		})
		return
	}

	c.HTML(http.StatusOK, "pages/leaderboard.html", gin.H{
		"Title":   "Leaderboard - Summer Academy",
		"Entries": entries,
	})
}

// Helper function to get today's problems
func getTodaysProblems(db *database.DB) ([]models.Problem, error) {
	// In production, actually query the database
	return []models.Problem{}, nil
}

// Helper function to get leaderboard entries
func getLeaderboard(db *database.DB) ([]models.LeaderboardEntry, error) {
	// In production, actually query the database
	return []models.LeaderboardEntry{}, nil
}
