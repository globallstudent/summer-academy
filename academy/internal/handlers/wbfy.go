package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globallstudent/academy/internal/config"
	"github.com/globallstudent/academy/internal/database"
	"github.com/google/uuid"
)

// WBFYHandlers contains handlers for WBFY terminal integration
type WBFYHandlers struct {
	db           *database.DB
	cfg          *config.Config
	portMutex    sync.Mutex
	portMap      map[string]int
	sessionMutex sync.RWMutex
	sessionMap   map[string]*TerminalSession
}

// NewWBFYHandlers creates a new WBFYHandlers instance
func NewWBFYHandlers(db *database.DB, cfg *config.Config) *WBFYHandlers {
	return &WBFYHandlers{
		db:           db,
		cfg:          cfg,
		portMutex:    sync.Mutex{},
		portMap:      make(map[string]int),
		sessionMutex: sync.RWMutex{},
		sessionMap:   make(map[string]*TerminalSession),
	}
}

// TerminalSession represents a terminal session
type TerminalSession struct {
	ID            string    `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	ProblemID     uuid.UUID `json:"problem_id"`
	Port          int       `json:"port"`
	ContainerID   string    `json:"container_id"`
	ContainerName string    `json:"container_name"`
	Command       string    `json:"command"`
	Language      string    `json:"language"`
	TempDir       string    `json:"temp_dir"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// AllocatePort allocates a port for a terminal session
func (h *WBFYHandlers) AllocatePort() (int, error) {
	h.portMutex.Lock()
	defer h.portMutex.Unlock()

	// Try to allocate a port in range 10000-10999
	for port := 10000; port < 11000; port++ {
		// Check if port is in use
		inUse := false
		for _, p := range h.portMap {
			if p == port {
				inUse = true
				break
			}
		}

		if !inUse {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports")
}

// ReleasePort releases a port
func (h *WBFYHandlers) ReleasePort(sessionID string) {
	h.portMutex.Lock()
	defer h.portMutex.Unlock()
	delete(h.portMap, sessionID)
}

// CreateTerminal godoc
// @Summary      Create a new terminal session
// @Description  Creates a new WBFY terminal session for a problem
// @Tags         terminal
// @Accept       multipart/form-data
// @Produce      json
// @Security     JWTCookie
// @Param        slug        path      string  true  "Problem slug"
// @Param        language    formData  string  false  "Programming language (default: bash)"
// @Success      202  {object}  map[string]interface{}  "Terminal session created"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Failure      503  {object}  map[string]interface{}  "Service unavailable - no ports available"
// @Router       /terminal/{slug} [post]
func (h *WBFYHandlers) CreateTerminal(c *gin.Context) {
	slug := c.Param("slug")
	language := c.DefaultPostForm("language", "bash")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "User not authenticated",
		})
		return
	}

	// Get problem by slug
	problem, err := h.getProblemBySlug(slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get problem details",
		})
		return
	}

	// Create a unique session ID
	sessionID := uuid.New().String()

	// Allocate a port for this session
	port, err := h.AllocatePort()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": "No available ports for terminal session",
		})
		return
	}

	// Register the port as in use
	h.portMap[sessionID] = port

	// Determine the command to run based on problem type and language
	command := getTerminalCommand(problem.Type, language)

	// Determine which Docker image to use
	image := getDockerImage(language)

	// Create the container name
	containerName := fmt.Sprintf("wbfy-%s", sessionID)

	// Start WBFY container in the background
	go func() {
		// Create temporary directory for session
		tempDir := filepath.Join(os.TempDir(), "academy-sessions", sessionID)
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			fmt.Printf("Failed to create temp directory: %v\n", err)
			h.ReleasePort(sessionID)
			return
		}

		// Copy problem files to the temp directory
		problemDir := filepath.Join("problems", fmt.Sprintf("day%d", problem.Day))
		if err := copyProblemFiles(problemDir, tempDir); err != nil {
			fmt.Printf("Failed to copy problem files: %v\n", err)
		}

		// Create a file that will track the session status
		statusFile := filepath.Join(tempDir, "session.json")
		sessionData := map[string]interface{}{
			"id":         sessionID,
			"user_id":    userID.(uuid.UUID).String(),
			"problem_id": problem.ID.String(),
			"language":   language,
			"start_time": time.Now(),
		}

		sessionJSON, _ := json.Marshal(sessionData)
		if err := os.WriteFile(statusFile, sessionJSON, 0644); err != nil {
			fmt.Printf("Failed to write session data: %v\n", err)
		}

		// Start WBFY Docker container
		dockerArgs := []string{
			"run",
			"--name", containerName,
			"-d", // detached mode
			"-p", fmt.Sprintf("%d:8081", port),
			"-v", fmt.Sprintf("%s:/workspace", tempDir),
			"-e", "WBFY_CMD=" + command,
			"-e", "PROBLEM_TYPE=" + problem.Type,
			"-e", "SESSION_ID=" + sessionID,
			"--rm", // remove container when stopped
			image,
		}

		cmd := exec.Command("docker", dockerArgs...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to start WBFY container: %v\nOutput: %s\n", err, out)
			h.ReleasePort(sessionID)
			return
		}

		containerID := string(out)[:12] // Get container ID from docker output

		// Store container info for cleanup
		containerInfo := map[string]string{
			"container_id":   containerID,
			"container_name": containerName,
			"session_id":     sessionID,
		}
		containerJSON, _ := json.Marshal(containerInfo)
		os.WriteFile(filepath.Join(tempDir, "container.json"), containerJSON, 0644)

		// Create a session with 2 hour expiry
		session := TerminalSession{
			ID:            sessionID,
			UserID:        userID.(uuid.UUID),
			ProblemID:     problem.ID,
			Port:          port,
			ContainerID:   containerID,
			ContainerName: containerName,
			Command:       command,
			Language:      language,
			TempDir:       tempDir,
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(2 * time.Hour),
		}

		// Store in memory map
		h.sessionMutex.Lock()
		h.sessionMap[sessionID] = &session
		h.sessionMutex.Unlock()

		// Store session in database (could use Redis in production)
		h.storeTerminalSession(session)
	}()

	// Create session record for immediate return
	session := TerminalSession{
		ID:        sessionID,
		UserID:    userID.(uuid.UUID),
		ProblemID: problem.ID,
		Port:      port,
		Command:   command,
		Language:  language,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}

	// Get the protocol (http/https)
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}

	// Generate the terminal URL
	terminalURL := fmt.Sprintf("%s://%s/terminal/%s",
		protocol,
		c.Request.Host,
		sessionID)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"session": session,
		"url":     terminalURL,
	})
}

// TerminalPage godoc
// @Summary      Display terminal interface
// @Description  Renders the terminal interface for an active terminal session
// @Tags         terminal
// @Accept       html
// @Produce      html
// @Security     JWTCookie
// @Param        id    path      string  true  "Terminal session ID"
// @Success      200  {object}  nil  "Terminal page"
// @Failure      401  {object}  nil  "Unauthorized"
// @Router       /terminal/{id} [get]
func (h *WBFYHandlers) TerminalPage(c *gin.Context) {
	sessionID := c.Param("id")

	// Get session from memory map
	h.sessionMutex.RLock()
	session, exists := h.sessionMap[sessionID]
	h.sessionMutex.RUnlock()
	if !exists {
		// In a production environment, you might want to fetch from database
		// and recreate the container if it doesn't exist

		c.HTML(http.StatusOK, "pages/terminal.html", gin.H{
			"Title":     "Terminal - Summer Academy",
			"SessionID": sessionID,
			"Error":     "Session not found or expired",
		})
		return
	}

	// In production, check if the session has expired
	if time.Now().After(session.ExpiresAt) {
		c.HTML(http.StatusOK, "pages/terminal.html", gin.H{
			"Title":     "Terminal - Summer Academy",
			"SessionID": sessionID,
			"Error":     "Session has expired",
		})
		return
	}

	// Get the WebSocket URL for the terminal
	wsProtocol := "ws"
	if c.Request.TLS != nil {
		wsProtocol = "wss"
	}

	wsURL := fmt.Sprintf("%s://%s/ws/%s",
		wsProtocol,
		c.Request.Host,
		sessionID)

	c.HTML(http.StatusOK, "pages/terminal.html", gin.H{
		"Title":     "Terminal - Summer Academy",
		"SessionID": sessionID,
		"Port":      session.Port,
		"WBFY": map[string]interface{}{
			"BaseURL": h.cfg.WBFY.BaseURL,
			"WSPath":  wsURL,
		},
	})
}

// WebSocketProxy handles WebSocket proxying to the WBFY container
func (h *WBFYHandlers) WebSocketProxy(c *gin.Context) {
	sessionID := c.Param("id")

	// Get session from memory map
	h.sessionMutex.RLock()
	session, exists := h.sessionMap[sessionID]
	h.sessionMutex.RUnlock()
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Terminal session not found",
		})
		return
	}

	// Construct the target WebSocket URL
	targetURL := fmt.Sprintf("ws://localhost:%d/ws", session.Port)

	// TODO: Implement WebSocket proxy to forward WebSocket connection to the container
	// This would require a WebSocket proxy library or implementation

	// For now, just return the URL for frontend to connect directly
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"url":    targetURL,
	})
}

// CleanupTerminal handles cleaning up terminal sessions
func (h *WBFYHandlers) CleanupTerminal(c *gin.Context) {
	sessionID := c.Param("id")

	// Get session
	h.sessionMutex.RLock()
	session, exists := h.sessionMap[sessionID]
	h.sessionMutex.RUnlock()
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Terminal session not found",
		})
		return
	}

	// Stop and remove the container
	if session.ContainerName != "" {
		cmd := exec.Command("docker", "stop", session.ContainerName)
		cmd.Run() // Ignore errors, container might already be stopped
	}

	// Delete the temporary directory
	if session.TempDir != "" {
		os.RemoveAll(session.TempDir)
	}

	// Release the port
	h.ReleasePort(sessionID)

	// Remove from memory map
	h.sessionMutex.Lock()
	delete(h.sessionMap, sessionID)
	h.sessionMutex.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Terminal session cleaned up",
	})
}

// StartCleanupJob starts a background job to clean up expired terminal sessions
func (h *WBFYHandlers) StartCleanupJob() {
	ticker := time.NewTicker(15 * time.Minute)
	go func() {
		for range ticker.C {
			h.cleanupExpiredSessions()
		}
	}()
}

// cleanupExpiredSessions cleans up expired terminal sessions
func (h *WBFYHandlers) cleanupExpiredSessions() {
	now := time.Now()

	// Create a list of sessions to remove to avoid concurrent map iteration
	var sessionsToRemove []string

	// Check all sessions
	h.sessionMutex.RLock()
	for id, session := range h.sessionMap {
		if now.After(session.ExpiresAt) {
			sessionsToRemove = append(sessionsToRemove, id)
		}
	}
	h.sessionMutex.RUnlock()

	// Remove expired sessions
	for _, id := range sessionsToRemove {
		h.sessionMutex.RLock()
		session := h.sessionMap[id]
		h.sessionMutex.RUnlock()

		// Stop and remove the container
		if session.ContainerName != "" {
			cmd := exec.Command("docker", "stop", session.ContainerName)
			cmd.Run() // Ignore errors, container might already be stopped
		}

		// Delete the temporary directory
		if session.TempDir != "" {
			os.RemoveAll(session.TempDir)
		}

		// Release the port
		h.ReleasePort(id)

		// Remove from memory map
		h.sessionMutex.Lock()
		delete(h.sessionMap, id)
		h.sessionMutex.Unlock()

		fmt.Printf("Cleaned up expired session: %s\n", id)
	}
}

// Helper function to get terminal command
func getTerminalCommand(problemType, language string) string {
	switch {
	case language == "python":
		return "python3"
	case language == "go":
		return "go run"
	case language == "javascript":
		return "node"
	case language == "bash" || language == "zsh":
		return language
	default:
		return "bash"
	}
}

// Helper function to get Docker image based on language
func getDockerImage(language string) string {
	switch language {
	case "python":
		return "globalstudent/wbfy-python:latest"
	case "go":
		return "globalstudent/wbfy-golang:latest"
	case "javascript":
		return "globalstudent/wbfy-node:latest"
	default:
		return "globalstudent/wbfy-base:latest" // Base image with common tools
	}
}

// Helper function to copy problem files to the workspace
func copyProblemFiles(src, dst string) error {
	// Walk through the source directory
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		// If it's a directory, create it
		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		// Copy the file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

// storeTerminalSession stores a terminal session in the database
func (h *WBFYHandlers) storeTerminalSession(session TerminalSession) error {
	// In a production environment, this would store the session in the database
	// For now, we'll just log it
	fmt.Printf("Storing terminal session: %+v\n", session)

	// Here's how you would implement the database storage:
	/*
		ctx := context.Background()
		_, err := h.db.Pool.Exec(ctx, `
			INSERT INTO terminal_sessions
			(id, user_id, problem_id, port, container_id, container_name, command, language, temp_dir, created_at, expires_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			session.ID, session.UserID, session.ProblemID, session.Port,
			session.ContainerID, session.ContainerName, session.Command, session.Language,
			session.TempDir, session.CreatedAt, session.ExpiresAt)
		return err
	*/

	return nil
}

// getProblemBySlug gets a problem by its slug
func (h *WBFYHandlers) getProblemBySlug(slug string) (*Problem, error) {
	// In a production environment, fetch from database
	// For now, return a mock problem
	return &Problem{
		ID:       uuid.New(),
		Title:    "Sample Problem",
		Slug:     slug,
		Type:     "linux",
		FilePath: "/app/problems/day1/linux.md",
		Day:      1,
	}, nil
}

// Problem represents a coding problem
type Problem struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Slug     string    `json:"slug"`
	Type     string    `json:"type"`
	FilePath string    `json:"file_path"`
	Day      int       `json:"day"`
}
