package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SanitizationMiddleware validates and cleans query parameters and common headers
func SanitizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Validate Query Parameters length and characters
		queryParams := c.Request.URL.Query()
		for key, values := range queryParams {
			// Limit key length
			if len(key) > 64 {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid parameter key length"})
				return
			}

			for _, value := range values {
				// Limit value length (e.g., 1024 chars for general params)
				if len(value) > 2048 {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Parameter value too long"})
					return
				}

				// Check for null bytes (SQL injection and other exploits often use these)
				if strings.Contains(value, "\x00") {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid characters in parameters"})
					return
				}

				// Basic character validation for specific highly sensitive params if needed
				// For now, general sanitization is enough as handlers should use prepared statements
			}
		}

		// 2. Normalize and check headers
		// Ensure Content-Type is as expected for JSON APIs
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "" && !strings.Contains(strings.ToLower(contentType), "application/json") && !strings.Contains(strings.ToLower(contentType), "multipart/form-data") {
				c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{"error": "Unsupported media type. Use application/json"})
				return
			}
		}

		c.Next()
	}
}
