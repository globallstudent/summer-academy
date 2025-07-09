package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globallstudent/academy/internal/config"
	"github.com/globallstudent/academy/internal/database"
	"github.com/globallstudent/academy/internal/models"
	"github.com/google/uuid"
)

// ContestHandlers contains handlers for contest-related routes
type ContestHandlers struct {
	db    *database.DB
	redis *database.Redis
	cfg   *config.Config
}

// NewContestHandlers creates a new ContestHandlers instance
func NewContestHandlers(db *database.DB, redis *database.Redis, cfg *config.Config) *ContestHandlers {
	return &ContestHandlers{db: db, redis: redis, cfg: cfg}
}

// Contest represents a time-bound collection of problems
type Contest struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	DurationDays int       `json:"duration_days"`
	Status       string    `json:"status"` // upcoming, active, ended
	IsJoined     bool      `json:"is_joined"`
}

// ContestDay represents a single day in a contest with its problems
type ContestDay struct {
	Day         int              `json:"day"`
	Title       string           `json:"title"`
	UnlockTime  time.Time        `json:"unlock_time"`
	IsLocked    bool             `json:"is_locked"`
	IsCompleted bool             `json:"is_completed"`
	IsCurrent   bool             `json:"is_current"`
	Problems    []models.Problem `json:"problems"`
}

// ListContests godoc
// @Summary      List available contests
// @Description  Displays all available contests for the user
// @Tags         contests
// @Accept       html
// @Produce      html
// @Security     JWTCookie
// @Success      200  {object}  nil  "Contest list page"
// @Failure      401  {object}  nil  "Unauthorized"
// @Failure      500  {object}  nil  "Internal server error"
// @Router       /contests [get]
func (h *ContestHandlers) ListContests(c *gin.Context) {
	// Get user from context (set by auth middleware)
	user, exists := c.Get("user")
	if !exists {
		c.HTML(http.StatusUnauthorized, "main", gin.H{
			"Title": "Unauthorized - Summer Academy",
			"Error": "You must be logged in to view this page",
		})
		return
	}

	// Get all available contests
	contests, err := h.getAvailableContests(user.(models.User))
	if err != nil {
		c.HTML(http.StatusInternalServerError, "main", gin.H{
			"Title": "Contests - Summer Academy",
			"Error": "Failed to get available contests",
		})
		return
	}

	c.HTML(http.StatusOK, "main", gin.H{
		"Title":           "Available Contests - Summer Academy",
		"Contests":        contests,
		"User":            user,
		"IsAuthenticated": true,
	})
}

// ContestDetail godoc
// @Summary      Show contest detail
// @Description  Displays details of a specific contest and its days
// @Tags         contests
// @Accept       html
// @Produce      html
// @Param        slug  path      string  true  "Contest slug"
// @Security     JWTCookie
// @Success      200  {object}  nil  "Contest detail page"
// @Failure      401  {object}  nil  "Unauthorized"
// @Failure      404  {object}  nil  "Contest not found"
// @Failure      500  {object}  nil  "Internal server error"
// @Router       /contests/{slug} [get]
func (h *ContestHandlers) ContestDetail(c *gin.Context) {
	slug := c.Param("slug")

	// Get user from context (set by auth middleware)
	user, exists := c.Get("user")
	if !exists {
		c.HTML(http.StatusUnauthorized, "main", gin.H{
			"Title": "Unauthorized - Summer Academy",
			"Error": "You must be logged in to view this page",
		})
		return
	}

	// Get contest by slug
	contest, err := h.getContestBySlug(slug)
	if err != nil {
		c.HTML(http.StatusNotFound, "main", gin.H{
			"Title": "Contest Not Found - Summer Academy",
			"Error": "The requested contest could not be found",
		})
		return
	}

	// Check if user has joined this contest
	isJoined, err := h.hasUserJoinedContest(user.(models.User).ID, contest.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "main", gin.H{
			"Title": contest.Title + " - Summer Academy",
			"Error": "Failed to check contest participation",
		})
		return
	}

	if !isJoined {
		// Redirect to join page if not joined
		c.Redirect(http.StatusFound, "/contests/"+slug+"/join")
		return
	}

	// Get contest days with their locked/unlocked status
	days, currentDay, err := h.getContestDays(contest)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "main", gin.H{
			"Title": contest.Title + " - Summer Academy",
			"Error": "Failed to get contest days",
		})
		return
	}

	// Calculate progress percentage
	progressPercent := (float64(currentDay) / float64(contest.DurationDays)) * 100

	// Get user's score for this contest
	userScore, totalPossibleScore, err := h.getUserContestScore(user.(models.User).ID, contest.ID)
	if err != nil {
		userScore = 0
		totalPossibleScore = 0
		// Log error but don't fail the request
	}

	// Check if user is admin
	isAdmin := user.(models.User).Role == "admin"

	c.HTML(http.StatusOK, "main", gin.H{
		"Title":              contest.Title + " - Summer Academy",
		"Contest":            contest,
		"Days":               days,
		"CurrentDay":         currentDay,
		"ProgressPercent":    progressPercent,
		"UserScore":          userScore,
		"TotalPossibleScore": totalPossibleScore,
		"IsAdmin":            isAdmin,
		"User":               user,
		"IsAuthenticated":    true,
	})
}

// JoinContest godoc
// @Summary      Join a contest
// @Description  Allows a user to join a specific contest
// @Tags         contests
// @Accept       html
// @Produce      html
// @Param        slug  path      string  true  "Contest slug"
// @Security     JWTCookie
// @Success      302  {object}  nil  "Redirect to contest detail"
// @Failure      401  {object}  nil  "Unauthorized"
// @Failure      404  {object}  nil  "Contest not found"
// @Failure      500  {object}  nil  "Internal server error"
// @Router       /contests/{slug}/join [get]
func (h *ContestHandlers) JoinContest(c *gin.Context) {
	slug := c.Param("slug")

	// Get user from context (set by auth middleware)
	user, exists := c.Get("user")
	if !exists {
		c.HTML(http.StatusUnauthorized, "main", gin.H{
			"Title": "Unauthorized - Summer Academy",
			"Error": "You must be logged in to join a contest",
		})
		return
	}

	// Get contest by slug
	contest, err := h.getContestBySlug(slug)
	if err != nil {
		c.HTML(http.StatusNotFound, "main", gin.H{
			"Title": "Contest Not Found - Summer Academy",
			"Error": "The requested contest could not be found",
		})
		return
	}

	// Check if contest is active
	if contest.Status != "active" {
		c.HTML(http.StatusBadRequest, "main", gin.H{
			"Title": "Cannot Join - Summer Academy",
			"Error": "This contest is not currently active",
		})
		return
	}

	// Check if user has already joined
	isJoined, _ := h.hasUserJoinedContest(user.(models.User).ID, contest.ID)
	if isJoined {
		// Already joined, redirect to contest detail
		c.Redirect(http.StatusFound, "/contests/"+slug)
		return
	}

	// Join the contest
	err = h.joinUserToContest(user.(models.User).ID, contest.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "main", gin.H{
			"Title": "Error - Summer Academy",
			"Error": "Failed to join the contest",
		})
		return
	}

	// Redirect to contest detail
	c.Redirect(http.StatusFound, "/contests/"+slug)
}

// ContestLeaderboard godoc
// @Summary      Show contest leaderboard
// @Description  Displays the leaderboard for a specific contest
// @Tags         contests
// @Accept       html
// @Produce      html
// @Param        slug  path      string  true  "Contest slug"
// @Security     JWTCookie
// @Success      200  {object}  nil  "Contest leaderboard page"
// @Failure      401  {object}  nil  "Unauthorized"
// @Failure      404  {object}  nil  "Contest not found"
// @Failure      500  {object}  nil  "Internal server error"
// @Router       /contests/{slug}/leaderboard [get]
func (h *ContestHandlers) ContestLeaderboard(c *gin.Context) {
	slug := c.Param("slug")

	// Get user from context (set by auth middleware)
	user, exists := c.Get("user")
	if !exists {
		c.HTML(http.StatusUnauthorized, "main", gin.H{
			"Title": "Unauthorized - Summer Academy",
			"Error": "You must be logged in to view this page",
		})
		return
	}

	// Get contest by slug
	contest, err := h.getContestBySlug(slug)
	if err != nil {
		c.HTML(http.StatusNotFound, "main", gin.H{
			"Title": "Contest Not Found - Summer Academy",
			"Error": "The requested contest could not be found",
		})
		return
	}

	// Get leaderboard entries
	leaderboard, err := h.getContestLeaderboard(contest.ID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "main", gin.H{
			"Title": contest.Title + " Leaderboard - Summer Academy",
			"Error": "Failed to get leaderboard data",
		})
		return
	}

	c.HTML(http.StatusOK, "main", gin.H{
		"Title":           contest.Title + " Leaderboard - Summer Academy",
		"Contest":         contest,
		"Users":           leaderboard,
		"User":            user,
		"IsAuthenticated": true,
	})
}

// Helper function to get available contests
func (h *ContestHandlers) getAvailableContests(user models.User) ([]Contest, error) {
	// In production, query the database
	// For demo, return hardcoded data
	summerChallenge := Contest{
		ID:           uuid.New(),
		Title:        "Summer Coding Challenge",
		Slug:         "summer-challenge",
		Description:  "30-day coding challenge with daily problems",
		StartDate:    time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2025, 7, 30, 23, 59, 59, 0, time.UTC),
		DurationDays: 30,
		Status:       "active",
		IsJoined:     true,
	}

	return []Contest{summerChallenge}, nil
}

// Helper function to get a contest by slug
func (h *ContestHandlers) getContestBySlug(slug string) (Contest, error) {
	// In production, query the database
	// For demo, return hardcoded data if slug matches
	if slug == "summer-challenge" {
		return Contest{
			ID:           uuid.New(),
			Title:        "Summer Coding Challenge",
			Slug:         "summer-challenge",
			Description:  "30-day coding challenge with daily problems",
			StartDate:    time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2025, 7, 30, 23, 59, 59, 0, time.UTC),
			DurationDays: 30,
			Status:       "active",
		}, nil
	}

	return Contest{}, nil
}

// Helper function to check if a user has joined a contest
func (h *ContestHandlers) hasUserJoinedContest(userID uuid.UUID, contestID uuid.UUID) (bool, error) {
	// In production, query the database
	// For demo, return true
	return true, nil
}

// Helper function to join a user to a contest
func (h *ContestHandlers) joinUserToContest(userID uuid.UUID, contestID uuid.UUID) error {
	// In production, insert into database
	// For demo, return nil (success)
	return nil
}

// Helper function to get contest days
func (h *ContestHandlers) getContestDays(contest Contest) ([]ContestDay, int, error) {
	// In production, query the database
	// For demo, create days based on contest duration
	days := make([]ContestDay, contest.DurationDays)
	currentDay := 9 // Mock current day (in a real app, calculate based on current date)

	for i := 0; i < contest.DurationDays; i++ {
		dayNum := i + 1
		isLocked := dayNum > currentDay
		isCompleted := dayNum < currentDay
		isCurrent := dayNum == currentDay

		days[i] = ContestDay{
			Day:         dayNum,
			Title:       "Day " + strconv.Itoa(dayNum) + " Challenge",
			UnlockTime:  contest.StartDate.AddDate(0, 0, i),
			IsLocked:    isLocked,
			IsCompleted: isCompleted,
			IsCurrent:   isCurrent,
		}
	}

	return days, currentDay, nil
}

// Helper function to get user's score in a contest
func (h *ContestHandlers) getUserContestScore(userID uuid.UUID, contestID uuid.UUID) (int, int, error) {
	// In production, query the database
	// For demo, return mock scores
	userScore := 450
	totalPossibleScore := 1500

	return userScore, totalPossibleScore, nil
}

// Helper function to get contest leaderboard
func (h *ContestHandlers) getContestLeaderboard(contestID uuid.UUID) ([]models.User, error) {
	// In production, query the database
	// For demo, return hardcoded data
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
