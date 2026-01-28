package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/querybase/internal/api/response"
)

// LoggingMiddleware logs all HTTP requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Log request details
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path

		// Get user ID if available
		userID := c.GetString("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		// Log format
		log.Printf("[%s] %s %s %s %d %v %s",
			time.Now().Format("2006-01-02 15:04:05"),
			method,
			path,
			userID,
			statusCode,
			duration,
			clientIP,
		)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.Printf("Error: %v", e.Error())
			}
		}

		// Log slow requests (> 1 second)
		if duration > time.Second {
			log.Printf("SLOW REQUEST: %s %s took %v", method, path, duration)
		}
	}
}

// ErrorRecoveryMiddleware recovers from panics and returns a 500 error
func ErrorRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC recovered: %v", err)
				response.SendInternalError(c, "An internal server error occurred")
				c.Abort()
			}
		}()
		c.Next()
	}
}
