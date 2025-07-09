package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/globallstudent/academy/internal/config"
	"github.com/globallstudent/academy/internal/database"
	"github.com/globallstudent/academy/internal/models"
	"github.com/google/uuid"
)

// UserHandlers contains handlers for user routes
type UserHandlers struct {
	db  *database.DB
	cfg *config.Config
}

// NewUserHandlers creates a new UserHandlers instance
func NewUserHandlers(db *database.DB, cfg *config.Config) *UserHandlers {
	return &UserHandlers{db: db, cfg: cfg}
}

// ProfilePage godoc
// @Summary      Show user profile
// @Description  Displays user profile with submission history and statistics
// @Tags         user
// @Accept       html
// @Produce      html
// @Security     JWTCookie
// @Success      200  {object}  nil  "Profile page"
// @Failure      401  {object}  nil  "Unauthorized"
// @Failure      500  {object}  nil  "Internal server error"
// @Router       /profile [get]
// ProfilePage handles the user profile page
func (h *UserHandlers) ProfilePage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.HTML(http.StatusUnauthorized, "pages/error.html", gin.H{
			"Error": "User not authenticated",
		})
		return
	}

	// Get user details - userID could be a string or uuid.UUID
	var userUUID uuid.UUID
	if userIDString, ok := userID.(string); ok {
		// Parse string into UUID
		parsedUUID, err := uuid.Parse(userIDString)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "main", gin.H{
				"Title":           "Profile - Summer Academy",
				"Error":           "Invalid user ID format",
				"IsAuthenticated": true,
			})
			return
		}
		userUUID = parsedUUID
	} else if userIDUUID, ok := userID.(uuid.UUID); ok {
		userUUID = userIDUUID
	} else {
		// Use the user from context directly if available
		if userObj, exists := c.Get("user"); exists {
			if userModel, ok := userObj.(models.User); ok {
				// Render with the user from context
				c.HTML(http.StatusOK, "main", gin.H{
					"Title":           "Profile - Summer Academy",
					"User":            userModel,
					"Submissions":     []models.Submission{},
					"IsAuthenticated": true,
				})
				return
			}
		}

		c.HTML(http.StatusInternalServerError, "main", gin.H{
			"Title":           "Profile - Summer Academy",
			"Error":           "User ID not found or has invalid type",
			"IsAuthenticated": true,
		})
		return
	}

	// Get user details
	user, err := getUserByID(h.db, userUUID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "main", gin.H{
			"Title":           "Error - Summer Academy",
			"Error":           "Failed to get user details",
			"IsAuthenticated": true,
		})
		return
	}

	// Get user submissions
	submissions, err := getUserSubmissions(h.db, userUUID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "main", gin.H{
			"Title":           "Profile - Summer Academy",
			"Error":           "Failed to get user submissions",
			"User":            user,
			"IsAuthenticated": true,
		})
		return
	}

	c.HTML(http.StatusOK, "main", gin.H{
		"Title":           "Profile - Summer Academy",
		"User":            user,
		"Submissions":     submissions,
		"IsAuthenticated": true,
	})
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Update user profile information like username
// @Tags         user
// @Accept       multipart/form-data
// @Produce      json
// @Security     JWTCookie
// @Param        username   formData  string  true  "User's username"
// @Success      200  {object}  map[string]interface{}  "Profile updated successfully"
// @Failure      400  {object}  map[string]interface{}  "Bad request"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /profile [post]
func (h *UserHandlers) UpdateProfile(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "User not authenticated",
		})
		return
	}

	username := c.PostForm("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Username is required",
		})
		return
	}

	// Update user in database (in production)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profile updated successfully",
	})
}

// AdminDashboard godoc
// @Summary      Admin dashboard
// @Description  Displays admin dashboard with platform statistics
// @Tags         admin
// @Accept       html
// @Produce      html
// @Security     JWTCookie
// @Success      200  {object}  nil  "Admin dashboard"
// @Failure      401  {object}  nil  "Unauthorized"
// @Failure      403  {object}  nil  "Forbidden - Admin access required"
// @Router       /admin/ [get]
// AdminDashboard handles the admin dashboard page
func (h *UserHandlers) AdminDashboard(c *gin.Context) {
	// Get stats
	stats := getAdminStats(h.db)

	c.HTML(http.StatusOK, "pages/admin/dashboard.html", gin.H{
		"Title": "Admin Dashboard - Summer Academy",
		"Stats": stats,
	})
}

// UserList godoc
// @Summary      List all users
// @Description  Displays a list of all registered users for admin management
// @Tags         admin
// @Accept       html
// @Produce      html
// @Security     JWTCookie
// @Success      200  {object}  nil  "Users list"
// @Failure      401  {object}  nil  "Unauthorized"
// @Failure      403  {object}  nil  "Forbidden - Admin access required"
// @Failure      500  {object}  nil  "Internal server error"
// @Router       /admin/users [get]
// UserList handles the admin user listing page
func (h *UserHandlers) UserList(c *gin.Context) {
	// Get all users
	users, err := getAllUsers(h.db)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "pages/admin/users.html", gin.H{
			"Error": "Failed to get users",
		})
		return
	}

	c.HTML(http.StatusOK, "pages/admin/users.html", gin.H{
		"Title": "Manage Users - Admin - Summer Academy",
		"Users": users,
	})
}

// Helper function to get user by ID
func getUserByID(db *database.DB, userID uuid.UUID) (models.User, error) {
	// In production, actually query the database
	return models.User{
		ID:          userID,
		PhoneNumber: "+12345678901",
		Username:    "Student1234",
		Role:        "user",
	}, nil
}

// Helper function to get user submissions
func getUserSubmissions(db *database.DB, userID uuid.UUID) ([]models.Submission, error) {
	// In production, actually query the database
	return []models.Submission{}, nil
}

// AdminStats represents statistics for the admin dashboard
type AdminStats struct {
	TotalUsers       int
	TotalProblems    int
	TotalSubmissions int
	PassRate         float64
}

// Helper function to get admin stats
func getAdminStats(db *database.DB) AdminStats {
	// In production, actually query the database
	return AdminStats{
		TotalUsers:       0,
		TotalProblems:    0,
		TotalSubmissions: 0,
		PassRate:         0,
	}
}

// Helper function to get all users
func getAllUsers(db *database.DB) ([]models.User, error) {
	// In production, actually query the database
	return []models.User{}, nil
}
