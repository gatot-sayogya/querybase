package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SecurityHeadersMiddleware adds security-related headers to the response
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate or get Request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)

		// Add Request ID to response headers
		c.Header("X-Request-ID", requestID)

		// HSTS - Force HTTPS (use max-age of 1 year)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// XSS Protection for older browsers
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (Basic default)
		// Note: For a production build, this should be more specific
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'")

		c.Next()
	}
}
