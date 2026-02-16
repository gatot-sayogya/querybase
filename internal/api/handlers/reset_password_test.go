package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourorg/querybase/internal/api/middleware"
	"github.com/yourorg/querybase/internal/auth"
	"github.com/yourorg/querybase/internal/models"
)

// TestResetUserPassword_Success tests admin resetting user password
func TestResetUserPassword_Success(t *testing.T) {
	db := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	jwtManager := auth.NewJWTManager("test-secret", 24*3600, "querybase")
	authHandler := NewAuthHandler(db, jwtManager)

	// Setup admin route
	admin := router.Group("/")
	admin.Use(middleware.AuthMiddleware(jwtManager))
	admin.Use(middleware.RequireAdmin())
	{
		admin.POST("/users/:id/reset-password", authHandler.ResetUserPassword)
	}

	// Create admin user
	adminPasswordHash, err := auth.HashPassword("adminpass")
	require.NoError(t, err)
	adminUser := models.User{
		ID:           uuid.New(),
		Email:        "admin@example.com",
		Username:     "admin",
		PasswordHash: adminPasswordHash,
		Role:         models.RoleAdmin,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&adminUser).Error)

	// Create regular user
	userPasswordHash, err := auth.HashPassword("oldpassword")
	require.NoError(t, err)
	regularUser := models.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		Username:     "user",
		PasswordHash: userPasswordHash,
		Role:         models.RoleUser,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&regularUser).Error)

	// Generate admin token
	adminToken, err := jwtManager.GenerateToken(adminUser.ID, adminUser.Email, string(adminUser.Role))
	require.NoError(t, err)

	// Reset user password
	resetReq := map[string]string{
		"new_password": "newpassword123",
	}
	body, _ := json.Marshal(resetReq)

	req, _ := http.NewRequest("POST", "/users/"+regularUser.ID.String()+"/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify password was changed
	var updatedUser models.User
	err = db.First(&updatedUser, regularUser.ID).Error
	require.NoError(t, err)

	assert.True(t, auth.CheckPassword("newpassword123", updatedUser.PasswordHash))
	assert.False(t, auth.CheckPassword("oldpassword", updatedUser.PasswordHash))
}

// TestResetUserPassword_NonAdmin tests non-admin cannot reset password
func TestResetUserPassword_NonAdmin(t *testing.T) {
	db := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	jwtManager := auth.NewJWTManager("test-secret", 24*3600, "querybase")
	authHandler := NewAuthHandler(db, jwtManager)

	// Setup route with admin middleware
	admin := router.Group("/")
	admin.Use(middleware.AuthMiddleware(jwtManager))
	admin.Use(middleware.RequireAdmin())
	{
		admin.POST("/users/:id/reset-password", authHandler.ResetUserPassword)
	}

	// Create regular user
	userPasswordHash, err := auth.HashPassword("password")
	require.NoError(t, err)
	regularUser := models.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		Username:     "user",
		PasswordHash: userPasswordHash,
		Role:         models.RoleUser,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&regularUser).Error)

	// Generate regular user token
	userToken, err := jwtManager.GenerateToken(regularUser.ID, regularUser.Email, string(regularUser.Role))
	require.NoError(t, err)

	// Try to reset password
	resetReq := map[string]string{
		"new_password": "newpassword123",
	}
	body, _ := json.Marshal(resetReq)

	req, _ := http.NewRequest("POST", "/users/"+regularUser.ID.String()+"/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should be forbidden (403)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestResetUserPassword_OwnPassword tests admin cannot reset own password
func TestResetUserPassword_OwnPassword(t *testing.T) {
	db := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	jwtManager := auth.NewJWTManager("test-secret", 24*3600, "querybase")
	authHandler := NewAuthHandler(db, jwtManager)

	// Setup admin route
	admin := router.Group("/")
	admin.Use(middleware.AuthMiddleware(jwtManager))
	admin.Use(middleware.RequireAdmin())
	{
		admin.POST("/users/:id/reset-password", authHandler.ResetUserPassword)
	}

	// Create admin user
	adminPasswordHash, err := auth.HashPassword("adminpass")
	require.NoError(t, err)
	adminUser := models.User{
		ID:           uuid.New(),
		Email:        "admin@example.com",
		Username:     "admin",
		PasswordHash: adminPasswordHash,
		Role:         models.RoleAdmin,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&adminUser).Error)

	// Generate admin token
	adminToken, err := jwtManager.GenerateToken(adminUser.ID, adminUser.Email, string(adminUser.Role))
	require.NoError(t, err)

	// Try to reset own password
	resetReq := map[string]string{
		"new_password": "newpassword123",
	}
	body, _ := json.Marshal(resetReq)

	req, _ := http.NewRequest("POST", "/users/"+adminUser.ID.String()+"/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "change password endpoint")
}

// TestResetUserPassword_UserNotFound tests resetting non-existent user
func TestResetUserPassword_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	jwtManager := auth.NewJWTManager("test-secret", 24*3600, "querybase")
	authHandler := NewAuthHandler(db, jwtManager)

	// Setup admin route
	admin := router.Group("/")
	admin.Use(middleware.AuthMiddleware(jwtManager))
	admin.Use(middleware.RequireAdmin())
	{
		admin.POST("/users/:id/reset-password", authHandler.ResetUserPassword)
	}

	// Create admin user
	adminPasswordHash, err := auth.HashPassword("adminpass")
	require.NoError(t, err)
	adminUser := models.User{
		ID:           uuid.New(),
		Email:        "admin@example.com",
		Username:     "admin",
		PasswordHash: adminPasswordHash,
		Role:         models.RoleAdmin,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&adminUser).Error)

	// Generate admin token
	adminToken, err := jwtManager.GenerateToken(adminUser.ID, adminUser.Email, string(adminUser.Role))
	require.NoError(t, err)

	// Try to reset non-existent user
	fakeUserID := uuid.New()
	resetReq := map[string]string{
		"new_password": "newpassword123",
	}
	body, _ := json.Marshal(resetReq)

	req, _ := http.NewRequest("POST", "/users/"+fakeUserID.String()+"/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestResetUserPassword_ShortPassword tests password validation
func TestResetUserPassword_ShortPassword(t *testing.T) {
	db := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	jwtManager := auth.NewJWTManager("test-secret", 24*3600, "querybase")
	authHandler := NewAuthHandler(db, jwtManager)

	// Setup admin route
	admin := router.Group("/")
	admin.Use(middleware.AuthMiddleware(jwtManager))
	admin.Use(middleware.RequireAdmin())
	{
		admin.POST("/users/:id/reset-password", authHandler.ResetUserPassword)
	}

	// Create admin and regular user
	adminPasswordHash, err := auth.HashPassword("adminpass")
	require.NoError(t, err)
	adminUser := models.User{
		ID:           uuid.New(),
		Email:        "admin@example.com",
		Username:     "admin",
		PasswordHash: adminPasswordHash,
		Role:         models.RoleAdmin,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&adminUser).Error)

	userPasswordHash, err := auth.HashPassword("oldpassword")
	require.NoError(t, err)
	regularUser := models.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		Username:     "user",
		PasswordHash: userPasswordHash,
		Role:         models.RoleUser,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&regularUser).Error)

	// Generate admin token
	adminToken, err := jwtManager.GenerateToken(adminUser.ID, adminUser.Email, string(adminUser.Role))
	require.NoError(t, err)

	// Try to reset with short password
	resetReq := map[string]string{
		"new_password": "short",
	}
	body, _ := json.Marshal(resetReq)

	req, _ := http.NewRequest("POST", "/users/"+regularUser.ID.String()+"/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
