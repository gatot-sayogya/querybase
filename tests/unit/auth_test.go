package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/querybase/internal/auth"
)

// TestAuthMiddleware_ValidToken tests auth middleware with valid token
func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager("test-secret", 24*time.Hour, "querybase")

	// Generate test token
	token, err := jwtManager.GenerateToken(
		auth.MustParseUUID("00000000-0000-0000-0000-000000000001"),
		"test@example.com",
		"admin",
	)
	require.NoError(t, err)

	// Setup middleware
	router.Use(AuthMiddleware(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		userID := c.GetString("user_id")
		email := c.GetString("email")
		role := c.GetString("role")

		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"email":   email,
			"role":    role,
		})
	})

	// Test with valid token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify context values were set
	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-0000-0000-000000000001", response["user_id"])
	assert.Equal(t, "test@example.com", response["email"])
	assert.Equal(t, "admin", response["role"])
}

// TestAuthMiddleware_NoToken tests auth middleware without token
func TestAuthMiddleware_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager("test-secret", 24*time.Hour, "querybase")
	router.Use(AuthMiddleware(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Authorization header required", response["error"])
}

// TestAuthMiddleware_InvalidTokenFormat tests auth middleware with invalid token format
func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager("test-secret", 24*time.Hour, "querybase")
	router.Use(AuthMiddleware(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Missing Bearer prefix",
			authHeader:     "invalid.token.format",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid authorization header format",
		},
		{
			name:           "Empty header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header required",
		},
		{
			name:           "Only Bearer",
			authHeader:     "Bearer",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid authorization header format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedError, response["error"])
		})
	}
}

// TestAuthMiddleware_InvalidToken tests auth middleware with invalid token
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager("test-secret", 24*time.Hour, "querybase")
	router.Use(AuthMiddleware(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Invalid token", response["error"])
}

// TestRequireAdmin_AdminRole tests RequireAdmin middleware with admin user
func TestRequireAdmin_AdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager("test-secret", 24*time.Hour, "querybase")

	// Generate admin token
	token, err := jwtManager.GenerateToken(
		auth.MustParseUUID("00000000-0000-0000-0000-000000000001"),
		"admin@example.com",
		"admin",
	)
	require.NoError(t, err)

	router.Use(RequireAdmin())
	router.Use(func(c *gin.Context) {
		// Simulate auth middleware setting user info
		c.Set("user_id", "00000000-0000-0000-0000-000000000001")
		c.Set("email", "admin@example.com")
		c.Set("role", "admin")
		c.Next()
	})
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access"})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestRequireAdmin_NonAdminRole tests RequireAdmin middleware with non-admin user
func TestRequireAdmin_NonAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager("test-secret", 24*time.Hour, "querybase")

	// Generate user token (not admin)
	token, err := jwtManager.GenerateToken(
		auth.MustParseUUID("00000000-0000-0000-0000-000000000002"),
		"user@example.com",
		"user",
	)
	require.NoError(t, err)

	router.Use(RequireAdmin())
	router.Use(func(c *gin.Context) {
		// Simulate auth middleware setting user info for non-admin
		c.Set("user_id", "00000000-0000-0000-0000-000000000002")
		c.Set("email", "user@example.com")
		c.Set("role", "user")
		c.Next()
	})
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access"})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Admin access required", response["error"])
}

// TestRequireAdmin_NoRole tests RequireAdmin middleware without role set
func TestRequireAdmin_NoRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(RequireAdmin())
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access"})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Admin access required", response["error"])
}

// TestRequireAdmin_ViewerRole tests RequireAdmin middleware with viewer role
func TestRequireAdmin_ViewerRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(RequireAdmin())
	router.Use(func(c *gin.Context) {
		// Simulate auth middleware setting user info for viewer
		c.Set("user_id", "00000000-0000-0000-0000-000000000003")
		c.Set("email", "viewer@example.com")
		c.Set("role", "viewer")
		c.Next()
	})
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access"})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestMiddleware_Chain tests middleware chain execution order
func TestMiddleware_Chain(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager("test-secret", 24*time.Hour, "querybase")

	// Generate admin token
	token, err := jwtManager.GenerateToken(
		auth.MustParseUUID("00000000-0000-0000-0000-000000000001"),
		"admin@example.com",
		"admin",
	)
	require.NoError(t, err)

	executionOrder := []string{}

	// Custom middleware to track execution
	middleware1 := func(c *gin.Context) {
		executionOrder = append(executionOrder, "middleware1-before")
		c.Next()
		executionOrder = append(executionOrder, "middleware1-after")
	}

	middleware2 := func(c *gin.Context) {
		executionOrder = append(executionOrder, "middleware2-before")
		c.Next()
		executionOrder = append(executionOrder, "middleware2-after")
	}

	router.Use(middleware1)
	router.Use(middleware2)
	router.Use(AuthMiddleware(jwtManager))
	router.Use(RequireAdmin())
	router.GET("/admin", func(c *gin.Context) {
		executionOrder = append(executionOrder, "handler")
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{
		"middleware1-before",
		"middleware2-before",
		"handler",
		"middleware2-after",
		"middleware1-after",
	}, executionOrder)
}
