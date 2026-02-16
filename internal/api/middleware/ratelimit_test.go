package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiterMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a strict rate limit config for testing
	config := &RateLimiterConfig{
		RequestsPerMinute:      6, // 6 requests per minute = 1 every 10 seconds
		BurstSize:              2, // Allow bursts of 2 requests
		SkipSuccessfulRequests: false,
		SkipPaths:              []string{},
	}

	t.Run("Allows requests within rate limit", func(t *testing.T) {
		router := gin.New()
		router.Use(RateLimiterMiddleware(config))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		// First request should succeed
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		// Second request should succeed (burst allows 2)
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("Blocks requests exceeding rate limit", func(t *testing.T) {
		router := gin.New()
		router.Use(RateLimiterMiddleware(config))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		// First request should succeed
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)

		// Second request should succeed (burst)
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)

		// Third request should be rate limited
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 429, w.Code)
	})

	t.Run("Skips specified paths", func(t *testing.T) {
		skipConfig := &RateLimiterConfig{
			RequestsPerMinute: 1,
			BurstSize:         1,
			SkipPaths:         []string{"/health"},
		}

		router := gin.New()
		router.Use(RateLimiterMiddleware(skipConfig))
		router.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		// Health check should not be rate limited
		for i := 0; i < 10; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/health", nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code, "Health check should not be rate limited")
		}

		// Regular endpoint should be rate limited
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 429, w.Code)
	})

	t.Run("Different IP addresses have separate rate limits", func(t *testing.T) {
		router := gin.New()
		router.Use(RateLimiterMiddleware(config))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		// First IP
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/test", nil)
		req1.RemoteAddr = "192.168.1.1:1234"
		router.ServeHTTP(w1, req1)
		assert.Equal(t, 200, w1.Code)

		w1 = httptest.NewRecorder()
		req1, _ = http.NewRequest("GET", "/test", nil)
		req1.RemoteAddr = "192.168.1.1:1234"
		router.ServeHTTP(w1, req1)
		assert.Equal(t, 200, w1.Code)

		w1 = httptest.NewRecorder()
		req1, _ = http.NewRequest("GET", "/test", nil)
		req1.RemoteAddr = "192.168.1.1:1234"
		router.ServeHTTP(w1, req1)
		assert.Equal(t, 429, w1.Code, "First IP should be rate limited")

		// Second IP should not be affected
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = "192.168.1.2:1234"
		router.ServeHTTP(w2, req2)
		assert.Equal(t, 200, w2.Code, "Second IP should not be rate limited")
	})

	t.Run("Tokens replenish over time", func(t *testing.T) {
		// Use a faster refill rate for testing
		fastConfig := &RateLimiterConfig{
			RequestsPerMinute: 12, // 12 per minute = 1 every 5 seconds
			BurstSize:         1,  // Only 1 burst token
			SkipPaths:         []string{},
		}

		router := gin.New()
		router.Use(RateLimiterMiddleware(fastConfig))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		// First request succeeds
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)

		// Second request is rate limited
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 429, w.Code)

		// Wait for tokens to replenish (5+ seconds)
		time.Sleep(6 * time.Second)

		// Request should succeed again after waiting
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "Request should succeed after token replenishment")
	})
}

func TestRateLimiterByPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Different rate limits for different paths", func(t *testing.T) {
		defaultConfig := &RateLimiterConfig{
			RequestsPerMinute: 6,
			BurstSize:         1,
		}

		strictConfig := &RateLimiterConfig{
			RequestsPerMinute: 1,
			BurstSize:         1,
		}

		limiter := NewRateLimiterByPath(defaultConfig)
		limiter.AddPath("/api/v1/queries", strictConfig)

		router := gin.New()
		router.Use(limiter.Middleware())
		router.GET("/api/v1/queries", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "query"})
		})
		router.GET("/api/v1/users", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "users"})
		})

		// Queries endpoint should be rate limited after 1 request
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/api/v1/queries", nil)
		router.ServeHTTP(w1, req1)
		assert.Equal(t, 200, w1.Code)

		w1 = httptest.NewRecorder()
		req1, _ = http.NewRequest("GET", "/api/v1/queries", nil)
		router.ServeHTTP(w1, req1)
		assert.Equal(t, 429, w1.Code, "Queries endpoint should be rate limited")

		// Users endpoint should still work
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/api/v1/users", nil)
		router.ServeHTTP(w2, req2)
		assert.Equal(t, 200, w2.Code, "Users endpoint should not be rate limited")
	})

	t.Run("Default rate limit applies to paths without specific config", func(t *testing.T) {
		defaultConfig := &RateLimiterConfig{
			RequestsPerMinute: 6,
			BurstSize:         2,
		}

		limiter := NewRateLimiterByPath(defaultConfig)

		router := gin.New()
		router.Use(limiter.Middleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		// Should allow burst of 2
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w1, req1)
		assert.Equal(t, 200, w1.Code)

		w1 = httptest.NewRecorder()
		req1, _ = http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w1, req1)
		assert.Equal(t, 200, w1.Code)

		w1 = httptest.NewRecorder()
		req1, _ = http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w1, req1)
		assert.Equal(t, 429, w1.Code, "Should be rate limited after burst")
	})
}

func TestDefaultRateLimitConfig(t *testing.T) {
	config := DefaultRateLimitConfig()

	assert.Equal(t, 60, config.RequestsPerMinute)
	assert.Equal(t, 10, config.BurstSize)
	assert.False(t, config.SkipSuccessfulRequests)
	assert.Contains(t, config.SkipPaths, "/health")
	assert.Contains(t, config.SkipPaths, "/api/v1/auth")
}
