package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/querybase/internal/auth"
)

func TestCreateTestJWTManager(t *testing.T) {
	manager := CreateTestJWTManager()
	require.NotNil(t, manager)

	token, err := manager.GenerateToken(
		auth.MustParseUUID(TestAdminUserID),
		TestAdminEmail,
		"admin",
	)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestCreateTestJWTToken(t *testing.T) {
	token, err := CreateTestJWTToken(TestAdminUserID, TestAdminEmail, "admin")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	manager := CreateTestJWTManager()
	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, TestAdminUserID, claims.UserID.String())
	assert.Equal(t, TestAdminEmail, claims.Email)
	assert.Equal(t, "admin", claims.Role)
}

func TestCreateTestJWTTokenWithManager(t *testing.T) {
	manager := CreateTestJWTManager()
	token, err := CreateTestJWTTokenWithManager(manager, TestUserID, TestUserEmail, "user")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, TestUserID, claims.UserID.String())
	assert.Equal(t, TestUserEmail, claims.Email)
	assert.Equal(t, "user", claims.Role)
}

func TestCreateAdminToken(t *testing.T) {
	token, err := CreateAdminToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	manager := CreateTestJWTManager()
	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, TestAdminUserID, claims.UserID.String())
	assert.Equal(t, TestAdminEmail, claims.Email)
	assert.Equal(t, "admin", claims.Role)
}

func TestCreateUserToken(t *testing.T) {
	token, err := CreateUserToken()
	require.NoError(t, err)

	manager := CreateTestJWTManager()
	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, TestUserID, claims.UserID.String())
	assert.Equal(t, TestUserEmail, claims.Email)
	assert.Equal(t, "user", claims.Role)
}

func TestCreateViewerToken(t *testing.T) {
	token, err := CreateViewerToken()
	require.NoError(t, err)

	manager := CreateTestJWTManager()
	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, TestViewerUserID, claims.UserID.String())
	assert.Equal(t, TestViewerEmail, claims.Email)
	assert.Equal(t, "viewer", claims.Role)
}

func TestCreateExpiredToken(t *testing.T) {
	token, err := CreateExpiredToken()
	require.NoError(t, err)

	manager := CreateTestJWTManager()
	_, err = manager.ValidateToken(token)
	assert.Error(t, err)
}

func TestSetMockUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	SetMockUser(c, TestAdminUserID, TestAdminEmail, "admin")

	assert.Equal(t, TestAdminUserID, c.GetString("user_id"))
	assert.Equal(t, TestAdminEmail, c.GetString("email"))
	assert.Equal(t, "admin", c.GetString("role"))
}

func TestMockAdminContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	MockAdminContext(c)

	assert.Equal(t, TestAdminUserID, c.GetString("user_id"))
	assert.Equal(t, TestAdminEmail, c.GetString("email"))
	assert.Equal(t, "admin", c.GetString("role"))
}

func TestMockUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	MockUserContext(c)

	assert.Equal(t, TestUserID, c.GetString("user_id"))
	assert.Equal(t, TestUserEmail, c.GetString("email"))
	assert.Equal(t, "user", c.GetString("role"))
}

func TestMockViewerContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	MockViewerContext(c)

	assert.Equal(t, TestViewerUserID, c.GetString("user_id"))
	assert.Equal(t, TestViewerEmail, c.GetString("email"))
	assert.Equal(t, "viewer", c.GetString("role"))
}

func TestMockAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(MockAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": c.GetString("user_id"),
			"email":   c.GetString("email"),
			"role":    c.GetString("role"),
		})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, TestUserID, response["user_id"])
	assert.Equal(t, TestUserEmail, response["email"])
	assert.Equal(t, "user", response["role"])
}

func TestMockAuthMiddlewareWithUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(MockAuthMiddlewareWithUser(TestAdminUserID, TestAdminEmail, "admin"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": c.GetString("user_id"),
			"email":   c.GetString("email"),
			"role":    c.GetString("role"),
		})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, TestAdminUserID, response["user_id"])
	assert.Equal(t, TestAdminEmail, response["email"])
	assert.Equal(t, "admin", response["role"])
}

func TestSetupTestRouterWithAuth(t *testing.T) {
	manager := CreateTestJWTManager()
	router := SetupTestRouterWithAuth(manager)

	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": c.GetString("user_id"),
			"email":   c.GetString("email"),
			"role":    c.GetString("role"),
		})
	})

	token, err := CreateAdminToken()
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, TestAdminUserID, response["user_id"])
	assert.Equal(t, TestAdminEmail, response["email"])
	assert.Equal(t, "admin", response["role"])
}

func TestSetupTestRouterWithAuth_NoToken(t *testing.T) {
	manager := CreateTestJWTManager()
	router := SetupTestRouterWithAuth(manager)

	router.GET("/protected", func(c *gin.Context) {
		_, exists := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"authenticated": exists})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]bool
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["authenticated"])
}

func TestSetupTestRouterWithAuth_InvalidToken(t *testing.T) {
	manager := CreateTestJWTManager()
	router := SetupTestRouterWithAuth(manager)

	router.GET("/protected", func(c *gin.Context) {
		_, exists := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"authenticated": exists})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]bool
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["authenticated"])
}
