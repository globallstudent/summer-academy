package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globallstudent/academy/internal/auth"
	"github.com/globallstudent/academy/internal/config"
	"github.com/globallstudent/academy/internal/database"
	"github.com/globallstudent/academy/internal/models"
	"github.com/google/uuid"
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

// HomePage godoc
// @Summary      Show the home page
// @Description  Renders the home page with today's problems if available
// @Tags         public
// @Accept       html
// @Produce      html
// @Success      200  {object}  nil  "Home page"
// @Failure      500  {object}  nil  "Internal server error"
// @Router       / [get]
func (h *PublicHandlers) HomePage(c *gin.Context) {
	// Check if user is authenticated
	user, isAuthenticated := c.Get("user")

	// Get today's problem if available
	todayProblems, err := getTodaysProblems(h.db)
	if err != nil {
		c.HTML(http.StatusOK, "main", gin.H{
			"Title":           "Summer Academy - Learn DSA and Linux",
			"Error":           "Failed to get today's problems",
			"IsAuthenticated": isAuthenticated,
			"User":            user,
		})
		return
	}

	c.HTML(http.StatusOK, "main", gin.H{
		"Title":           "Summer Academy - Learn DSA and Linux",
		"TodayProblems":   todayProblems,
		"IsAuthenticated": isAuthenticated,
		"User":            user,
	})
}

// LoginPage godoc
// @Summary      Show the login page
// @Description  Renders the login page for user authentication
// @Tags         auth
// @Accept       html
// @Produce      html
// @Success      200  {object}  nil  "Login page"
// @Router       /login [get]
func (h *PublicHandlers) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "main", gin.H{
		"Title": "Login - Summer Academy",
	})
}

// VerifyOTPPage godoc
// @Summary      Show OTP verification page
// @Description  Renders the OTP verification page with phone number and OTP if provided
// @Tags         auth
// @Accept       html
// @Produce      html
// @Param        phone  query     string  false  "Phone number"
// @Param        otp    query     string  false  "OTP code"
// @Success      200    {object}  nil     "Verification page"
// @Router       /verify [get]
func (h *PublicHandlers) VerifyOTPPage(c *gin.Context) {
	phoneNumber := c.Query("phone")
	otp := c.Query("otp")

	// Render the verification page
	c.HTML(http.StatusOK, "main", gin.H{
		"Title": "Verify OTP - Summer Academy",
		"Phone": phoneNumber,
		"OTP":   otp,
	})
}

// ProcessLogin godoc
// @Summary      Process login form submission
// @Description  Validates OTP, creates/updates user, and issues JWT token on successful login
// @Tags         auth
// @Accept       multipart/form-data
// @Produce      html
// @Param        phone  formData  string  true   "Phone number"
// @Param        otp    formData  string  true   "OTP code"
// @Success      302    {object}  nil     "Redirect to days page"
// @Failure      400    {object}  nil     "Bad request"
// @Failure      500    {object}  nil     "Internal server error"
// @Router       /login [post]
func (h *PublicHandlers) ProcessLogin(c *gin.Context) {
	// Create background context
	_ = context.Background()
	phoneNumber := c.PostForm("phone")
	otp := c.PostForm("otp")

	// Validate input
	if otp == "" || len(otp) != 6 {
		c.HTML(http.StatusBadRequest, "main", gin.H{
			"Title": "Verify OTP - Summer Academy",
			"Error": "Invalid verification code",
			"OTP":   otp,
		})
		return
	}

	var isValid bool
	// Verify OTP against Redis if available
	if h.redis != nil && h.redis.Client != nil {
		isValid, _ = h.redis.VerifyOTP(phoneNumber, otp)
	} else if h.cfg.Environment != "production" {
		// For development when Redis isn't available, accept any code
		isValid = true
		log.Printf("Development mode: accepting any OTP code: %s", otp)
	} else {
		c.HTML(http.StatusInternalServerError, "main", gin.H{
			"Title": "Verify OTP - Summer Academy",
			"Error": "OTP service unavailable",
		})
		return
	}

	if !isValid {
		c.HTML(http.StatusBadRequest, "main", gin.H{
			"Title": "Verify OTP - Summer Academy",
			"Error": "Invalid or expired verification code",
			"OTP":   otp,
		})
		return
	}

	// Find user by phone number or create a new user
	var user models.User
	// TODO: Replace with actual database query in production
	// Example query: SELECT * FROM users WHERE phone_number = $1

	// For now, simulate user lookup/creation
	username := "Student"
	if phoneNumber != "" && len(phoneNumber) > 4 {
		username = "Student" + phoneNumber[len(phoneNumber)-4:]
	} else {
		username = "Student" + otp[:4]
	}

	user = models.User{
		ID:           uuid.New(),
		PhoneNumber:  phoneNumber,
		Username:     username,
		RegisteredAt: time.Now(),
		Role:         "user",
	}

	// In production, store the user in database if new
	// INSERT INTO users (...) VALUES (...) ON CONFLICT DO UPDATE ...

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "main", gin.H{
			"Title": "Verify OTP - Summer Academy",
			"Error": "Failed to generate session token",
			"OTP":   otp,
		})
		return
	}

	// Set cookie
	secure := c.Request.TLS != nil
	c.SetCookie(
		h.cfg.Auth.CookieName,
		token,
		h.cfg.Auth.CookieMaxAge,
		"/",
		"",
		secure,
		true,
	)

	// Set the user in the context for consistent behavior
	c.Set("user", user)
	c.Set("IsAuthenticated", true)

	// Redirect to days page
	c.Redirect(http.StatusFound, "/days")
}

// LeaderboardPage godoc
// @Summary      Show leaderboard page
// @Description  Displays a leaderboard with top users and their scores
// @Tags         public
// @Accept       html
// @Produce      html
// @Success      200  {object}  nil  "Leaderboard page"
// @Failure      500  {object}  nil  "Internal server error"
// @Router       /leaderboard [get]
func (h *PublicHandlers) LeaderboardPage(c *gin.Context) {
	// Check if user is authenticated
	user, isAuthenticated := c.Get("user")

	// Get leaderboard entries
	entries, err := getLeaderboard(h.db)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "main", gin.H{
			"Title":           "Leaderboard - Summer Academy",
			"Error":           "Failed to get leaderboard data",
			"IsAuthenticated": isAuthenticated,
			"User":            user,
		})
		return
	}

	c.HTML(http.StatusOK, "main", gin.H{
		"Title":           "Leaderboard - Summer Academy",
		"Users":           entries, // Changed from Entries to Users to match the leaderboard.html template
		"IsAuthenticated": isAuthenticated,
		"User":            user,
	})
}

// LogoutHandler godoc
// @Summary      Logout the current user
// @Description  Clears the session cookie and redirects to home page
// @Tags         auth
// @Accept       html
// @Produce      html
// @Success      302  {object}  nil  "Redirect to home page"
// @Router       /logout [get]
func (h *PublicHandlers) LogoutHandler(c *gin.Context) {
	// Clear the cookie
	secure := c.Request.TLS != nil
	c.SetCookie(
		h.cfg.Auth.CookieName,
		"",
		-1, // Expire immediately
		"/",
		"",
		secure,
		true,
	)

	// Redirect to home page
	c.Redirect(http.StatusFound, "/")
}

// Helper function to get today's problems
func getTodaysProblems(db *database.DB) ([]models.Problem, error) {
	// In production, actually query the database
	return []models.Problem{}, nil
}

// Helper function to get leaderboard entries
func getLeaderboard(db *database.DB) ([]models.User, error) {
	// In production, actually query the database
	// For now, return sample data
	return []models.User{
		{
			ID:           uuid.New(),
			PhoneNumber:  "+1234567890",
			Username:     "student1",
			RegisteredAt: time.Now().Add(-24 * time.Hour),
			Role:         "user",
		},
		{
			ID:           uuid.New(),
			PhoneNumber:  "+1987654321",
			Username:     "student2",
			RegisteredAt: time.Now().Add(-48 * time.Hour),
			Role:         "user",
		},
	}, nil
}
