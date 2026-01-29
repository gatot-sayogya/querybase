package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/querybase/internal/config"
)

// Config represents CORS configuration
type Config struct {
	// AllowedOrigins is a list of origins allowed to make requests
	// Use "*" to allow all origins (not recommended for production)
	AllowedOrigins []string

	// AllowedMethods is the list of HTTP methods allowed
	AllowedMethods []string

	// AllowedHeaders is the list of header keys that will be allowed
	AllowedHeaders []string

	// ExposedHeaders is the list of header keys that clients can read from responses
	ExposedHeaders []string

	// AllowCredentials indicates whether the request can include user credentials
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached
	MaxAge int
}

// DefaultConfig returns the default CORS configuration
func DefaultConfig() *Config {
	return &Config{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"PATCH",
			"OPTIONS",
			"HEAD",
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-HTTP-Method-Override",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// DevelopmentConfig returns a permissive CORS configuration for development
func DevelopmentConfig() *Config {
	return &Config{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           86400,
	}
}

// ProdConfig returns a strict CORS configuration for production
func ProdConfig() *Config {
	return &Config{
		AllowedOrigins: []string{
			// Add your production frontend domains here
			// Example: "https://querybase.example.com"
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           86400,
	}
}

// CORSMiddleware creates a CORS middleware from the given configuration
func CORSMiddleware(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowedOrigin := ""
		if len(config.AllowedOrigins) > 0 {
			for _, allowed := range config.AllowedOrigins {
				if allowed == "*" || allowed == origin {
					allowedOrigin = allowed
					break
				}
			}
		}

		// If origin is not allowed, return error
		if allowedOrigin == "" && origin != "" {
			c.AbortWithStatusJSON(403, gin.H{
				"error": "Origin not allowed",
			})
			return
		}

		// Set CORS headers
		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
		}

		// Expose headers
		if len(config.ExposedHeaders) > 0 {
			for _, header := range config.ExposedHeaders {
				c.Header("Access-Control-Expose-Headers", header)
			}
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			c.Header("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
			c.AbortWithStatus(204)
			return
		}

		// Allow credentials
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Next()
	}
}

// Helper function to create CORS middleware from environment variable
func CORSMiddlewareFromEnv(env string) gin.HandlerFunc {
	var config *Config

	switch env {
	case "development", "dev":
		config = DevelopmentConfig()
	case "production", "prod":
		config = ProdConfig()
	default:
		config = DefaultConfig()
	}

	return CORSMiddleware(config)
}

// CORSMiddlewareFromConfig creates CORS middleware from application config
func CORSMiddlewareFromConfig(corsConfig config.CORSConfig) gin.HandlerFunc {
	config := &Config{
		AllowedOrigins:   corsConfig.GetAllowedOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposedHeaders:   []string{"Content-Length", "Content-Type"},
		AllowCredentials: corsConfig.AllowCredentials,
		MaxAge:           corsConfig.MaxAge,
	}
	return CORSMiddleware(config)
}
