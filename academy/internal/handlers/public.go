package handlers

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globallstudent/academy/internal/auth"
	"github.com/globallstudent/academy/internal/config"
	"github.com/globallstudent/academy/internal/database"
	"github.com/globallstudent/academy/internal/models"
	"github.com/google/uuid"
)

// Development mode only: in-memory OTP store with mutex for thread safety
var (
	devOTPStore     = make(map[string]string)
	devOTPStoreLock sync.RWMutex
)

// OTP verification rate limiting
var (
	failedAttempts     = make(map[string]int)
	failedAttemptTimes = make(map[string]time.Time)
	maxFailedAttempts  = 5
	lockoutDuration    = 10 * time.Minute
	failedAttemptsLock sync.RWMutex
)

// StoreDevelopmentOTP stores an OTP for development mode only
func StoreDevelopmentOTP(phoneNumber, otp string) {
	devOTPStoreLock.Lock()
	defer devOTPStoreLock.Unlock()
	devOTPStore[phoneNumber] = otp

	// For security, automatically expire OTPs after 5 minutes
	go func(phone string) {
		time.Sleep(5 * time.Minute)
		devOTPStoreLock.Lock()
		defer devOTPStoreLock.Unlock()
		delete(devOTPStore, phone)
	}(phoneNumber)
}

// checkRateLimit checks if a phone number has exceeded the maximum allowed failed attempts
// Returns true if rate limited, false otherwise
func checkRateLimit(phoneNumber string) bool {
	failedAttemptsLock.RLock()
	attempts, exists := failedAttempts[phoneNumber]
	lastAttemptTime, timeExists := failedAttemptTimes[phoneNumber]
	failedAttemptsLock.RUnlock()

	if !exists {
		return false
	}

	// Reset attempts if lockout period has passed
	if timeExists && time.Since(lastAttemptTime) > lockoutDuration {
		failedAttemptsLock.Lock()
		delete(failedAttempts, phoneNumber)
		delete(failedAttemptTimes, phoneNumber)
		failedAttemptsLock.Unlock()
		return false
	}

	// Check if max attempts reached
	return attempts >= maxFailedAttempts
}

// incrementFailedAttempts increases the count of failed attempts for a phone number
func incrementFailedAttempts(phoneNumber string) {
	failedAttemptsLock.Lock()
	defer failedAttemptsLock.Unlock()

	count, exists := failedAttempts[phoneNumber]
	if !exists {
		failedAttempts[phoneNumber] = 1
	} else {
		failedAttempts[phoneNumber] = count + 1
	}
	failedAttemptTimes[phoneNumber] = time.Now()
}

// resetFailedAttempts resets the failed attempts counter for a phone number
func resetFailedAttempts(phoneNumber string) {
	failedAttemptsLock.Lock()
	defer failedAttemptsLock.Unlock()
	delete(failedAttempts, phoneNumber)
	delete(failedAttemptTimes, phoneNumber)
}

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
	// Get phone and OTP from query params if provided
	phoneNumber := c.Query("phone")
	otp := c.Query("otp")

	c.HTML(http.StatusOK, "main", gin.H{
		"Title": "Login - Summer Academy",
		"Phone": phoneNumber,
		"OTP":   otp,
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

	// Log the request details for debugging
	log.Printf("ProcessLogin: received request path=%s method=%s phone=%s otp=%s",
		c.Request.URL.Path, c.Request.Method, phoneNumber, otp)

	// Dump all form data for debugging
	for key, values := range c.Request.PostForm {
		log.Printf("Form field %s = %v", key, values)
	}

	// Dump all headers for debugging
	for key, values := range c.Request.Header {
		log.Printf("Header %s = %v", key, values)
	}

	// Check if this is an HTMX request
	isHtmx := c.GetHeader("HX-Request") == "true"

	// Function to return error response in appropriate format
	renderError := func(status int, title, errorMsg string) {
		if isHtmx {
			// For HTMX requests, return a partial page with the form and error
			c.Header("HX-Retarget", "#verify-section")
			c.HTML(status, "main", gin.H{
				"Title":   title,
				"Error":   errorMsg,
				"Content": "verify",
				"Phone":   phoneNumber,
			})
		} else {
			// For regular form submissions, render the full page
			c.HTML(status, "main", gin.H{
				"Title":   title,
				"Error":   errorMsg,
				"Content": "verify",
				"Phone":   phoneNumber,
			})
		}
	}

	// Validate input
	if otp == "" || len(otp) != 6 {
		renderError(http.StatusBadRequest, "Verify OTP - Summer Academy",
			"Invalid verification code. Code must be 6 digits.")
		return
	}

	if phoneNumber == "" {
		renderError(http.StatusBadRequest, "Verify OTP - Summer Academy",
			"Phone number is required. Please use the Telegram bot to get a verification code.")
		return
	}

	// Check rate limit for OTP verification attempts
	if checkRateLimit(phoneNumber) {
		renderError(http.StatusTooManyRequests, "Verify OTP - Summer Academy",
			"Too many failed verification attempts. Please try again in 10 minutes.")
		return
	}

	var isValid bool
	var verifyErr error

	// Verify OTP against Redis if available
	if h.redis != nil && h.redis.Client != nil {
		isValid, verifyErr = h.redis.VerifyOTP(phoneNumber, otp)
		if verifyErr != nil {
			log.Printf("Error verifying OTP: %v", verifyErr)

			// Try development store as fallback if Redis fails and we're in development
			if h.cfg != nil && h.cfg.Environment == "development" {
				devOTPStoreLock.RLock()
				actualOTP, exists := devOTPStore[phoneNumber]
				devOTPStoreLock.RUnlock()

				if exists && actualOTP == otp {
					isValid = true
					devOTPStoreLock.Lock()
					delete(devOTPStore, phoneNumber)
					devOTPStoreLock.Unlock()
					log.Printf("Redis failed but development OTP store verification succeeded for %s", phoneNumber)
				} else {
					renderError(http.StatusInternalServerError, "Verify OTP - Summer Academy",
						"Verification service error. Please try again or request a new code.")
					return
				}
			} else {
				renderError(http.StatusInternalServerError, "Verify OTP - Summer Academy",
					"Verification service error. Please try again or request a new code.")
				return
			}
		}
	} else if h.cfg != nil && h.cfg.Environment == "development" {
		// ONLY for development when Redis isn't available

		// Get the stored dev OTPs from temporary in-memory store
		devOTPStoreLock.RLock()
		actualOTP, exists := devOTPStore[phoneNumber]
		devOTPStoreLock.RUnlock()

		if exists && actualOTP == otp {
			isValid = true
			log.Printf("Development mode: OTP verified for %s", phoneNumber)
			// Remove the OTP from store to prevent reuse
			devOTPStoreLock.Lock()
			delete(devOTPStore, phoneNumber)
			devOTPStoreLock.Unlock()
		} else {
			isValid = false
			log.Printf("Development mode: Invalid OTP attempt: %s", otp)
		}
	} else {
		renderError(http.StatusServiceUnavailable, "Verify OTP - Summer Academy",
			"Verification service unavailable. Please try again later or contact support.")
		return
	}

	if !isValid {
		// Increment failed attempts counter
		incrementFailedAttempts(phoneNumber)

		renderError(http.StatusBadRequest, "Verify OTP - Summer Academy",
			"Invalid or expired verification code. Please request a new code.")
		return
	}

	// Reset failed attempts counter on successful verification
	resetFailedAttempts(phoneNumber)

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
		renderError(http.StatusInternalServerError, "Verify OTP - Summer Academy",
			"Failed to generate session token. Please try again.")
		return
	}

	// Set secure cookie
	secure := c.Request.TLS != nil || h.cfg.Environment == "production"
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

	// For HTMX requests, set headers for proper client-side handling
	if isHtmx {
		// Tell HTMX to redirect to the days page
		c.Header("HX-Redirect", "/days")
		// Return a success message that will be shown briefly before redirect
		c.HTML(http.StatusOK, "main", gin.H{
			"Title":   "Login Successful - Summer Academy",
			"Content": "login-success",
		})
	} else {
		// For regular form submissions, redirect to days page
		c.Redirect(http.StatusFound, "/days")
	}
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
