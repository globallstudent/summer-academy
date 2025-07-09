package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/globallstudent/academy/internal/config"
	"github.com/globallstudent/academy/internal/database"
	"github.com/globallstudent/academy/internal/models"
)

// ProblemHandlers contains handlers for problem routes
type ProblemHandlers struct {
	db  *database.DB
	cfg *config.Config
}

// NewProblemHandlers creates a new ProblemHandlers instance
func NewProblemHandlers(db *database.DB, cfg *config.Config) *ProblemHandlers {
	return &ProblemHandlers{db: db, cfg: cfg}
}

// ListDays godoc
// @Summary      List available days
// @Description  Displays a list of all available days with their problems
// @Tags         problems
// @Accept       html
// @Produce      html
// @Security     JWTCookie
// @Success      200  {object}  nil  "Days list page"
// @Failure      401  {object}  nil  "Unauthorized"
// @Failure      500  {object}  nil  "Internal server error"
// @Router       /days [get]
func (h *ProblemHandlers) ListDays(c *gin.Context) {
	// Get all available days (in production, filter by unlock time)
	days, err := getAvailableDays(h.db)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "pages/days.html", gin.H{
			"Error": "Failed to get available days",
		})
		return
	}

	c.HTML(http.StatusOK, "pages/days.html", gin.H{
		"Title": "Available Days - Summer Academy",
		"Days":  days,
	})
}

// DayDetail godoc
// @Summary      Show day detail page
// @Description  Displays details of a specific day and its problems
// @Tags         problems
// @Accept       html
// @Produce      html
// @Param        day   path      string  true  "Day number"
// @Security     JWTCookie
// @Success      200  {object}  nil  "Day detail page"
// @Failure      400  {object}  nil  "Invalid day parameter"
// @Failure      401  {object}  nil  "Unauthorized"
// @Failure      404  {object}  nil  "Day not found"
// @Failure      500  {object}  nil  "Internal server error"
// @Router       /days/{day} [get]
func (h *ProblemHandlers) DayDetail(c *gin.Context) {
	dayParam := c.Param("day")
	day, err := strconv.Atoi(dayParam)
	if err != nil {
		c.HTML(http.StatusBadRequest, "pages/error.html", gin.H{
			"Error": "Invalid day parameter",
		})
		return
	}

	// Get problems for this day
	problems, err := getProblemsForDay(h.db, day)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "pages/day_detail.html", gin.H{
			"Error": "Failed to get problems for this day",
		})
		return
	}

	c.HTML(http.StatusOK, "pages/day_detail.html", gin.H{
		"Title":    "Day " + dayParam + " - Summer Academy",
		"Day":      day,
		"Problems": problems,
	})
}

// ProblemDetail godoc
// @Summary      Show problem detail page
// @Description  Displays details of a specific problem including description and examples
// @Tags         problems
// @Accept       html
// @Produce      html
// @Param        slug  path      string  true  "Problem slug"
// @Security     JWTCookie
// @Success      200  {object}  nil  "Problem detail page"
// @Failure      401  {object}  nil  "Unauthorized"
// @Failure      404  {object}  nil  "Problem not found"
// @Failure      500  {object}  nil  "Internal server error"
// @Router       /problems/{slug} [get]
func (h *ProblemHandlers) ProblemDetail(c *gin.Context) {
	slug := c.Param("slug")

	// Get problem by slug
	problem, err := getProblemBySlug(h.db, slug)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "pages/error.html", gin.H{
			"Error": "Failed to get problem details",
		})
		return
	}

	// Get problem content
	content, err := getProblemContent(problem.FilePath)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "pages/problem_detail.html", gin.H{
			"Error": "Failed to load problem content",
		})
		return
	}

	// Get test cases (only non-hidden ones)
	testcases, err := getTestcases(problem.ID, false)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "pages/problem_detail.html", gin.H{
			"Error": "Failed to load test cases",
		})
		return
	}

	c.HTML(http.StatusOK, "pages/problem_detail.html", gin.H{
		"Title":     problem.Title + " - Summer Academy",
		"Problem":   problem,
		"Content":   content,
		"Testcases": testcases,
	})
}

// AdminProblemList godoc
// @Summary      List all problems (admin)
// @Description  Lists all problems for admin management
// @Tags         admin
// @Accept       html
// @Produce      html
// @Security     JWTCookie
// @Success      200  {object}  nil  "Admin problem list page"
// @Failure      401  {object}  nil  "Unauthorized"
// @Failure      403  {object}  nil  "Forbidden - Not admin"
// @Failure      500  {object}  nil  "Internal server error"
// @Router       /admin/problems [get]
func (h *ProblemHandlers) AdminProblemList(c *gin.Context) {
	// Get all problems
	problems, err := getAllProblems(h.db)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "pages/admin/problems.html", gin.H{
			"Error": "Failed to get problems",
		})
		return
	}

	c.HTML(http.StatusOK, "pages/admin/problems.html", gin.H{
		"Title":    "Manage Problems - Admin - Summer Academy",
		"Problems": problems,
	})
}

// CreateProblem godoc
// @Summary      Create new problem
// @Description  Creates a new problem in the platform
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     JWTCookie
// @Param        problem  body      object  true  "Problem details"
// @Success      200  {object}  map[string]interface{}  "Problem created successfully"
// @Failure      400  {object}  map[string]interface{}  "Bad request"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Forbidden - Admin access required"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /admin/problems [post]
func (h *ProblemHandlers) CreateProblem(c *gin.Context) {
	// In production, validate and create the problem
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Problem created"})
}

// UpdateProblem godoc
// @Summary      Update existing problem
// @Description  Updates an existing problem in the platform
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     JWTCookie
// @Param        id       path      string  true  "Problem ID"
// @Param        problem  body      object  true  "Problem details"
// @Success      200  {object}  map[string]interface{}  "Problem updated successfully"
// @Failure      400  {object}  map[string]interface{}  "Bad request"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      403  {object}  map[string]interface{}  "Forbidden - Admin access required"
// @Failure      404  {object}  map[string]interface{}  "Problem not found"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /admin/problems/{id} [put]
func (h *ProblemHandlers) UpdateProblem(c *gin.Context) {
	// In production, validate and update the problem
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Problem updated"})
}

// Helper function to get available days
func getAvailableDays(db *database.DB) ([]int, error) {
	// In production, actually query the database
	return []int{1, 2, 3}, nil
}

// Helper function to get problems for a specific day
func getProblemsForDay(db *database.DB, day int) ([]models.Problem, error) {
	// In production, actually query the database
	// Example mock data
	mockProblems := []models.Problem{
		{
			ID:         uuid.New(),
			Day:        day,
			Type:       "dsa",
			Slug:       "two-sum",
			Title:      "Two Sum",
			FilePath:   "problems/day-1/dsa.md",
			Score:      100,
			UnlockTime: time.Now(),
		},
		{
			ID:         uuid.New(),
			Day:        day,
			Type:       "linux",
			Slug:       "file-finder",
			Title:      "File Finder",
			FilePath:   "problems/day-1/linux.md",
			Score:      100,
			UnlockTime: time.Now(),
		},
	}
	return mockProblems, nil
}

// Helper function to get problem by slug
func getProblemBySlug(db *database.DB, slug string) (models.Problem, error) {
	// In production, actually query the database
	return models.Problem{
		ID:         uuid.New(),
		Day:        1,
		Type:       "dsa",
		Slug:       slug,
		Title:      "Two Sum",
		FilePath:   "problems/day-1/dsa.md",
		Score:      100,
		UnlockTime: time.Now(),
	}, nil
}

// Helper function to get problem content
func getProblemContent(filePath string) (string, error) {
	// In production, read from the file
	return "# Two Sum\n\nGiven an array of integers `nums` and an integer `target`, return indices of the two numbers such that they add up to `target`.\n", nil
}

// Helper function to get test cases
func getTestcases(problemID uuid.UUID, includeHidden bool) ([]models.Testcase, error) {
	// In production, read from problem JSON file
	return []models.Testcase{
		{
			Input:          "[2, 7, 11, 15]\n9",
			ExpectedOutput: "[0, 1]",
			IsHidden:       false,
		},
	}, nil
}

// Helper function to get all problems
func getAllProblems(db *database.DB) ([]models.Problem, error) {
	// In production, actually query the database
	return []models.Problem{}, nil
}
