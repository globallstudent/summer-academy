package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/globallstudent/academy/internal/config"
	"github.com/globallstudent/academy/internal/database"
	"github.com/globallstudent/academy/internal/models"
)

// SubmissionHandlers contains handlers for submission routes
type SubmissionHandlers struct {
	db  *database.DB
	cfg *config.Config
}

// NewSubmissionHandlers creates a new SubmissionHandlers instance
func NewSubmissionHandlers(db *database.DB, cfg *config.Config) *SubmissionHandlers {
	return &SubmissionHandlers{db: db, cfg: cfg}
}

// SubmitPage godoc
// @Summary      Display submission form
// @Description  Renders the submission form for a specific problem
// @Tags         submission
// @Accept       html
// @Produce      html
// @Security     JWTCookie
// @Param        slug    path      string  true  "Problem slug"
// @Success      200  {object}  nil  "Submission form page"
// @Failure      401  {object}  nil  "Unauthorized"
// @Failure      500  {object}  nil  "Internal server error"
// @Router       /submit/{slug} [get]
func (h *SubmissionHandlers) SubmitPage(c *gin.Context) {
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
		c.HTML(http.StatusInternalServerError, "pages/submit.html", gin.H{
			"Error": "Failed to load problem content",
		})
		return
	}

	// Get supported languages
	languages := getSupportedLanguages(problem.Type)

	c.HTML(http.StatusOK, "pages/submit.html", gin.H{
		"Title":     "Submit: " + problem.Title + " - Summer Academy",
		"Problem":   problem,
		"Content":   content,
		"Languages": languages,
		"WBFY":      h.cfg.WBFY,
	})
}

// TestSubmission godoc
// @Summary      Test code submission
// @Description  Tests the submitted code against non-hidden test cases
// @Tags         submission
// @Accept       multipart/form-data
// @Produce      json
// @Security     JWTCookie
// @Param        slug        path      string  true  "Problem slug"
// @Param        code        formData  string  true  "Submitted code"
// @Param        language    formData  string  true  "Programming language used"
// @Success      200  {object}  map[string]interface{}  "Test results"
// @Failure      400  {object}  map[string]interface{}  "Bad request"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /test/{slug} [post]
func (h *SubmissionHandlers) TestSubmission(c *gin.Context) {
	slug := c.Param("slug")
	code := c.PostForm("code")
	language := c.PostForm("language")

	// Validate input
	if code == "" || language == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Code and language are required",
		})
		return
	}

	// Get problem
	problem, err := getProblemBySlug(h.db, slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get problem details",
		})
		return
	}

	// Get test cases (only non-hidden ones for testing)
	testcases, err := getTestcases(problem.ID, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to load test cases",
		})
		return
	}

	// Run tests
	results, err := runTests(code, language, testcases, problem.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to run tests: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"results": results,
	})
}

// ProcessSubmission godoc
// @Summary      Process final submission
// @Description  Processes a final submission, runs all tests and saves result
// @Tags         submission
// @Accept       multipart/form-data
// @Produce      json
// @Security     JWTCookie
// @Param        slug        path      string  true  "Problem slug"
// @Param        code        formData  string  true  "Submitted code"
// @Param        language    formData  string  true  "Programming language used"
// @Success      200  {object}  map[string]interface{}  "Submission results"
// @Failure      400  {object}  map[string]interface{}  "Bad request"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /submit/{slug} [post]
func (h *SubmissionHandlers) ProcessSubmission(c *gin.Context) {
	slug := c.Param("slug")
	code := c.PostForm("code")
	language := c.PostForm("language")

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "User not authenticated",
		})
		return
	}

	// Validate input
	if code == "" || language == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Code and language are required",
		})
		return
	}

	// Get problem
	problem, err := getProblemBySlug(h.db, slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get problem details",
		})
		return
	}

	// Get all test cases (including hidden ones)
	testcases, err := getTestcases(problem.ID, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to load test cases",
		})
		return
	}

	// Run tests
	results, err := runTests(code, language, testcases, problem.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to run tests: " + err.Error(),
		})
		return
	}

	// Calculate score
	passed := 0
	for _, result := range results {
		if result.Passed {
			passed++
		}
	}
	score := int(float64(passed) / float64(len(testcases)) * float64(problem.Score))

	// Create submission record
	submission := models.Submission{
		ID:          uuid.New(),
		UserID:      userID.(uuid.UUID),
		ProblemID:   problem.ID,
		Language:    language,
		Status:      getSubmissionStatus(passed, len(testcases)),
		Output:      resultsToString(results),
		Score:       score,
		SubmittedAt: time.Now(),
	}

	// In production, save submission to database

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"submission": submission,
		"results":    results,
	})
}

// TestResult represents the result of a single test case
type TestResult struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	ActualOutput   string `json:"actual_output"`
	Passed         bool   `json:"passed"`
	IsHidden       bool   `json:"is_hidden"`
}

// Helper function to get supported languages
func getSupportedLanguages(problemType string) []string {
	switch problemType {
	case "dsa":
		return []string{"python", "go", "javascript", "cpp"}
	case "linux":
		return []string{"bash", "zsh"}
	case "build":
		return []string{"bash", "zsh", "python"}
	default:
		return []string{"python", "go"}
	}
}

// Helper function to run tests
func runTests(code, language string, testcases []models.Testcase, problemType string) ([]TestResult, error) {
	results := make([]TestResult, len(testcases))

	// In a real implementation, this would run code in a container
	// For this example, we'll simulate test results
	for i, testcase := range testcases {
		results[i] = TestResult{
			Input:          testcase.Input,
			ExpectedOutput: testcase.ExpectedOutput,
			ActualOutput:   testcase.ExpectedOutput, // Simulate correct output
			Passed:         true,
			IsHidden:       testcase.IsHidden,
		}
	}

	return results, nil
}

// Helper function to get submission status
func getSubmissionStatus(passed, total int) string {
	if passed == total {
		return "passed"
	} else if passed > 0 {
		return "partial"
	}
	return "failed"
}

// Helper function to convert test results to string
func resultsToString(results []TestResult) string {
	var sb strings.Builder
	for i, result := range results {
		sb.WriteString(fmt.Sprintf("Test Case %d: %s\n", i+1, statusString(result.Passed)))
		if !result.IsHidden {
			sb.WriteString(fmt.Sprintf("Input: %s\n", result.Input))
			sb.WriteString(fmt.Sprintf("Expected: %s\n", result.ExpectedOutput))
			if !result.Passed {
				sb.WriteString(fmt.Sprintf("Actual: %s\n", result.ActualOutput))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// Helper function to get status string
func statusString(passed bool) string {
	if passed {
		return "Passed"
	}
	return "Failed"
}
