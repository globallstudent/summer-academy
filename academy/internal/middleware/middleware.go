package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globallstudent/academy/internal/auth"
	"github.com/globallstudent/academy/internal/models"
	"github.com/google/uuid"
)

// cookieName is the name of the session cookie. It can be configured at runtime.
var cookieName = "academy_session"

// SetCookieName allows configuring the cookie name for authentication
func SetCookieName(name string) {
	if name != "" {
		cookieName = name
	}
}

// Logger returns a middleware that logs requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Add request ID to the context
		requestID := uuid.New().String()
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)

		// Process request
		c.Next()

		// Log request details
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		statusCode := c.Writer.Status()

		log.Printf("[%s] %s | %3d | %v | %s | %s",
			requestID, method, statusCode, latency, clientIP, path)
	}
}

// Auth returns a middleware that checks if the user is authenticated
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(cookieName)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Unauthorized",
			})
			return
		}

		claims, err := auth.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		// Create a user object and set in context for subsequent handlers
		user := models.User{
			ID:       claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		}

		// Set both individual fields and the complete user object
		c.Set("userID", claims.UserID.String())
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("user", user)
		c.Set("IsAuthenticated", true)

		c.Next()
	}
}

// AdminOnly returns a middleware that checks if the user is an admin
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.AbortWithStatusJSON(403, gin.H{
				"error": "Forbidden: Admin access required",
			})
			return
		}
		c.Next()
	}
}
