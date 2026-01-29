package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiterConfig defines rate limiting configuration
type RateLimiterConfig struct {
	// RequestsPerMinute is the maximum number of requests allowed per minute
	RequestsPerMinute int

	// BurstSize is the maximum number of requests allowed in a short burst
	BurstSize int

	// SkipSuccessfulRequests doesn't count successful requests towards the rate limit
	SkipSuccessfulRequests bool

	// SkipPaths is a list of paths to skip rate limiting for
	SkipPaths []string
}

// DefaultRateLimitConfig returns the default rate limiting configuration
func DefaultRateLimitConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		RequestsPerMinute:      60, // 60 requests per minute = 1 request per second
		BurstSize:              10, // Allow bursts of up to 10 requests
		SkipSuccessfulRequests: false,
		SkipPaths: []string{
			"/health",
			"/favicon.ico",
			"/static",
			"/api/v1/auth",        // Don't rate limit auth endpoints
			"/ws",                 // Don't rate limit WebSocket endpoint
			"/api/v1/datasources", // Don't rate limit data source/schema endpoints
			"/api/v1/approvals",    // Don't rate limit approval endpoints
			"/api/v1/groups",       // Don't rate limit group endpoints
			// Note: /api/v1/queries is NOT skipped - query execution IS rate limited
		},
	}
}

// tokenBucket represents a token bucket for rate limiting
type tokenBucket struct {
	tokens     int
	lastUpdate time.Time
}

// inMemoryRateLimiter is an in-memory rate limiter using token bucket algorithm
type inMemoryRateLimiter struct {
	buckets map[string]*tokenBucket
	mu      sync.RWMutex
	config  *RateLimiterConfig
}

// newInMemoryRateLimiter creates a new in-memory rate limiter
func newInMemoryRateLimiter(config *RateLimiterConfig) *inMemoryRateLimiter {
	limiter := &inMemoryRateLimiter{
		buckets: make(map[string]*tokenBucket),
		config:  config,
	}

	// Start cleanup goroutine to remove old entries
	go limiter.cleanup()

	return limiter
}

// allow checks if a request from the given key is allowed
func (rl *inMemoryRateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.buckets[key]

	if !exists {
		// First request from this key
		rl.buckets[key] = &tokenBucket{
			tokens:     rl.config.BurstSize - 1,
			lastUpdate: now,
		}
		return true
	}

	// Calculate elapsed time and add tokens
	elapsed := now.Sub(bucket.lastUpdate)
	tokensToAdd := int(elapsed.Minutes() * float64(rl.config.RequestsPerMinute))

	bucket.tokens += tokensToAdd
	if bucket.tokens > rl.config.BurstSize {
		bucket.tokens = rl.config.BurstSize
	}
	bucket.lastUpdate = now

	// Check if we have tokens available
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// cleanup removes old entries from the rate limiter
func (rl *inMemoryRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			// Remove entries that haven't been used in 10 minutes
			if now.Sub(bucket.lastUpdate) > 10*time.Minute {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimiterMiddleware creates a rate limiting middleware
func RateLimiterMiddleware(config *RateLimiterConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	limiter := newInMemoryRateLimiter(config)

	return func(c *gin.Context) {
		// Check if path should be skipped
		for _, skipPath := range config.SkipPaths {
			if c.Request.URL.Path == skipPath {
				c.Next()
				return
			}
		}

		// Use IP address as the rate limit key
		// In production, you might want to use user ID if authenticated
		key := c.ClientIP()

		// Check if request is allowed
		if !limiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
				"code":  http.StatusTooManyRequests,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimiterByPath creates rate limiters for different paths with different configurations
type RateLimiterByPath struct {
	limiters map[string]*inMemoryRateLimiter
	defaultLimiter *inMemoryRateLimiter
	mu sync.RWMutex
}

// NewRateLimiterByPath creates a new rate limiter that can have different limits per path
func NewRateLimiterByPath(defaultConfig *RateLimiterConfig) *RateLimiterByPath {
	if defaultConfig == nil {
		defaultConfig = DefaultRateLimitConfig()
	}

	return &RateLimiterByPath{
		limiters: make(map[string]*inMemoryRateLimiter),
		defaultLimiter: newInMemoryRateLimiter(defaultConfig),
	}
}

// AddPath adds a rate limiter for a specific path
func (rlbp *RateLimiterByPath) AddPath(path string, config *RateLimiterConfig) {
	rlbp.mu.Lock()
	defer rlbp.mu.Unlock()
	rlbp.limiters[path] = newInMemoryRateLimiter(config)
}

// Middleware returns the gin middleware for rate limiting by path
func (rlbp *RateLimiterByPath) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Find appropriate limiter for this path
		rlbp.mu.RLock()
		limiter, exists := rlbp.limiters[path]
		if !exists {
			limiter = rlbp.defaultLimiter
		}
		rlbp.mu.RUnlock()

		// Use IP address as the rate limit key
		key := fmt.Sprintf("%s:%s", path, c.ClientIP())

		// Check if request is allowed
		if !limiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
				"code":  http.StatusTooManyRequests,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
