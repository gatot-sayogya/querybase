package middleware

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

// TestAuthMiddleware_Simple tests basic middleware functionality
func TestAuthMiddleware_Simple(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	jwtManager := auth.NewJWTManager("test-secret", 24*time.Hour, "querybase")

	// Create a protected route
	router.GET("/protected", AuthMiddleware(jwtManager), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("No token returns 401", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Authorization header required", response["error"])
	})

	t.Run("Invalid token format returns 401", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid token returns 401", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Invalid token", response["error"])
	})

	t.Run("Valid token returns 200", func(t *testing.T) {
		// Generate a valid token
		userID := auth.MustParseUUID("00000000-0000-0000-0000-000000000001")
		token, err := jwtManager.GenerateToken(userID, "test@example.com", "admin")
		require.NoError(t, err)

		// Make request with valid token
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["message"])
	})

	t.Run("Valid token sets context values", func(t *testing.T) {
		// Generate a valid token
		userID := auth.MustParseUUID("00000000-0000-0000-0000-000000000001")
		token, err := jwtManager.GenerateToken(userID, "test@example.com", "admin")
		require.NoError(t, err)

		// Create a handler that reads context values
		router.GET("/context", AuthMiddleware(jwtManager), func(c *gin.Context) {
			userID := c.GetString("user_id")
			email := c.GetString("email")
			role := c.GetString("role")

			c.JSON(http.StatusOK, gin.H{
				"user_id": userID,
				"email":   email,
				"role":    role,
			})
		})

		// Make request
		req, _ := http.NewRequest("GET", "/context", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "00000000-0000-0000-0000-000000000001", response["user_id"])
		assert.Equal(t, "test@example.com", response["email"])
		assert.Equal(t, "admin", response["role"])
	})
}

// TestRequireAdmin_Simple tests admin middleware
func TestRequireAdmin_Simple(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Admin role passes", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			// Simulate auth middleware
			c.Set("user_id", "00000000-0000-0000-0000-000000000001")
			c.Set("email", "admin@example.com")
			c.Set("role", "admin")
			c.Next()
		})
		router.Use(RequireAdmin())
		router.GET("/admin", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "admin access"})
		})

		req, _ := http.NewRequest("GET", "/admin", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("User role is forbidden", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			// Simulate auth middleware for regular user
			c.Set("user_id", "00000000-0000-0000-0000-000000000002")
			c.Set("email", "user@example.com")
			c.Set("role", "user")
			c.Next()
		})
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
	})

	t.Run("Viewer role is forbidden", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", "00000000-0000-0000-0000-000000000003")
			c.Set("email", "viewer@example.com")
			c.Set("role", "viewer")
			c.Next()
		})
		router.Use(RequireAdmin())
		router.GET("/admin", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "admin access"})
		})

		req, _ := http.NewRequest("GET", "/admin", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("No role is forbidden", func(t *testing.T) {
		router := gin.New()
		router.Use(RequireAdmin())
		router.GET("/admin", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "admin access"})
		})

		req, _ := http.NewRequest("GET", "/admin", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
