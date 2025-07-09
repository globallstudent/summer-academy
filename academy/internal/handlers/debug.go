package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DumpRequest logs all details of a request for debugging
func DumpRequest(c *gin.Context) {
	log.Printf("DEBUG Request: %s %s", c.Request.Method, c.Request.URL.Path)
	log.Printf("DEBUG Headers: %v", c.Request.Header)

	if err := c.Request.ParseForm(); err == nil {
		log.Printf("DEBUG Form: %v", c.Request.Form)
		log.Printf("DEBUG PostForm: %v", c.Request.PostForm)
	}

	c.String(http.StatusOK, fmt.Sprintf("Request received: %s %s\n", c.Request.Method, c.Request.URL.Path))
}
