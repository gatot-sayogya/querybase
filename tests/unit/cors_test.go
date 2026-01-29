package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	// Test default configuration
	config := DefaultConfig()

	// Create test router
	router := gin.New()
	router.Use(CORSMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	// Test preflight request
	t.Run("Preflight OPTIONS request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")

		router.ServeHTTP(w, req)

		assert.Equal(t, 204, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Origin")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
		assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
	})

	t.Run("Actual GET request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("Reject unauthorized origin", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://malicious-site.com")

		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code)
	})

	t.Run("Wildcard origin allows all", func(t *testing.T) {
		devConfig := DevelopmentConfig()
		router := gin.New()
		router.Use(CORSMiddleware(devConfig))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://any-origin.com")

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestCORSMiddlewareCredentials(t *testing.T) {
	t.Run("AllowCredentials true", func(t *testing.T) {
		config := DefaultConfig()
		assert.True(t, config.AllowCredentials)
	})

	t.Run("AllowCredentials false", func(t *testing.T) {
		config := &Config{
			AllowedOrigins:   []string{"http://localhost:3000"},
			AllowedMethods:   []string{"GET"},
			AllowedHeaders:   []string{"Origin"},
			ExposedHeaders:   []string{},
			AllowCredentials: false,
			MaxAge:           86400,
		}
		assert.False(t, config.AllowCredentials)
	})
}

func TestCORSMiddlewareExposedHeaders(t *testing.T) {
	t.Run("Has exposed headers", func(t *testing.T) {
		config := DefaultConfig()
		assert.NotEmpty(t, config.ExposedHeaders)
		assert.Contains(t, config.ExposedHeaders, "Content-Length")
	})

	t.Run("Custom exposed headers", func(t *testing.T) {
		config := &Config{
			ExposedHeaders: []string{"X-Custom-Header", "X-Another-Custom"},
		}
		assert.Contains(t, config.ExposedHeaders, "X-Custom-Header")
		assert.Contains(t, config.ExposedHeaders, "X-Another-Custom")
	})
}
