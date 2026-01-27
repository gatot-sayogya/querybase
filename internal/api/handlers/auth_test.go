package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/yourorg/querybase/internal/auth"
	"github.com/yourorg/querybase/internal/models"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Enable UUID extension for SQLite (if needed)
	// For SQLite, we'll use a different approach

	err = db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.DataSource{},
		&models.DataSourcePermission{},
		&models.Query{},
		&models.QueryResult{},
		&models.QueryHistory{},
		&models.ApprovalRequest{},
		&models.ApprovalReview{},
		&models.NotificationConfig{},
		&models.Notification{},
	)
	require.NoError(t, err)

	return db
}

// setupTestRouter creates a test router with auth handler
func setupTestRouter(db *gorm.DB) (*gin.Engine, *auth.JWTManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager("test-secret", 24*3600, "querybase")
	authHandler := NewAuthHandler(db, jwtManager)

	// Setup routes
	router.POST("/login", authHandler.Login)
	router.GET("/me", authHandler.GetMe)
	router.POST("/change-password", authHandler.ChangePassword)

	return router, jwtManager
}

// TestLogin_Success tests successful login
func TestLogin_Success(t *testing.T) {
	db := setupTestDB(t)
	router, _ := setupTestRouter(db)

	// Create test user
	passwordHash, err := auth.HashPassword("password123")
	require.NoError(t, err)

	user := models.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: passwordHash,
		FullName:     "Test User",
		Role:         models.RoleAdmin,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&user).Error)

	// Test login
	loginReq := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	body, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response["token"])
	assert.NotEmpty(t, response["user"])

	userData := response["user"].(map[string]interface{})
	assert.Equal(t, "test@example.com", userData["email"])
	assert.Equal(t, "testuser", userData["username"])
	assert.Equal(t, "admin", userData["role"])
}

// TestLogin_InvalidCredentials tests login with invalid credentials
func TestLogin_InvalidCredentials(t *testing.T) {
	db := setupTestDB(t)
	router, _ := setupTestRouter(db)

	// Create test user
	passwordHash, err := auth.HashPassword("password123")
	require.NoError(t, err)

	user := models.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: passwordHash,
		Role:         models.RoleUser,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&user).Error)

	// Test login with wrong password
	loginReq := map[string]string{
		"username": "testuser",
		"password": "wrongpassword",
	}
	body, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Invalid credentials", response["error"])
}

// TestLogin_InactiveUser tests login with inactive user
func TestLogin_InactiveUser(t *testing.T) {
	db := setupTestDB(t)
	router, _ := setupTestRouter(db)

	// Create inactive test user
	passwordHash, err := auth.HashPassword("password123")
	require.NoError(t, err)

	user := models.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: passwordHash,
		Role:         models.RoleUser,
		IsActive:     false,
	}
	require.NoError(t, db.Create(&user).Error)

	// Test login
	loginReq := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	body, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestLogin_MissingFields tests login with missing fields
func TestLogin_MissingFields(t *testing.T) {
	db := setupTestDB(t)
	router, _ := setupTestRouter(db)

	tests := []struct {
		name       string
		username   string
		password   string
		statusCode int
	}{
		{"Missing username", "", "password123", http.StatusBadRequest},
		{"Missing password", "testuser", "", http.StatusBadRequest},
		{"Missing both", "", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loginReq := map[string]string{
				"username": tt.username,
				"password": tt.password,
			}
			body, _ := json.Marshal(loginReq)

			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

// TestGetMe_Success tests getting current user info
func TestGetMe_Success(t *testing.T) {
	db := setupTestDB(t)
	router, jwtManager := setupTestRouter(db)

	// Create test user
	passwordHash, err := auth.HashPassword("password123")
	require.NoError(t, err)

	user := models.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: passwordHash,
		FullName:     "Test User",
		Role:         models.RoleAdmin,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&user).Error)

	// Generate token
	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(user.Role))
	require.NoError(t, err)

	// Test get current user
	req, _ := http.NewRequest("GET", "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, user.ID.String(), response["id"])
	assert.Equal(t, "test@example.com", response["email"])
	assert.Equal(t, "testuser", response["username"])
	assert.Equal(t, "admin", response["role"])
}

// TestGetMe_NoToken tests getting current user without token
func TestGetMe_NoToken(t *testing.T) {
	db := setupTestDB(t)
	router, _ := setupTestRouter(db)

	req, _ := http.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetMe_InvalidToken tests getting current user with invalid token
func TestGetMe_InvalidToken(t *testing.T) {
	db := setupTestDB(t)
	router, _ := setupTestRouter(db)

	req, _ := http.NewRequest("GET", "/me", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestChangePassword_Success tests password change
func TestChangePassword_Success(t *testing.T) {
	db := setupTestDB(t)
	router, jwtManager := setupTestRouter(db)

	// Create test user
	passwordHash, err := auth.HashPassword("oldpassword123")
	require.NoError(t, err)

	user := models.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: passwordHash,
		Role:         models.RoleUser,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&user).Error)

	// Generate token
	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(user.Role))
	require.NoError(t, err)

	// Change password
	changeReq := map[string]string{
		"current_password": "oldpassword123",
		"new_password":     "newpassword456",
	}
	body, _ := json.Marshal(changeReq)

	req, _ := http.NewRequest("POST", "/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify password was changed
	var updatedUser models.User
	err = db.First(&updatedUser, user.ID).Error
	require.NoError(t, err)

	assert.True(t, auth.CheckPassword("newpassword456", updatedUser.PasswordHash))
	assert.False(t, auth.CheckPassword("oldpassword123", updatedUser.PasswordHash))
}

// TestChangePassword_WrongCurrentPassword tests password change with wrong current password
func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	db := setupTestDB(t)
	router, jwtManager := setupTestRouter(db)

	// Create test user
	passwordHash, err := auth.HashPassword("oldpassword123")
	require.NoError(t, err)

	user := models.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: passwordHash,
		Role:         models.RoleUser,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&user).Error)

	// Generate token
	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(user.Role))
	require.NoError(t, err)

	// Try to change with wrong current password
	changeReq := map[string]string{
		"current_password": "wrongpassword",
		"new_password":     "newpassword456",
	}
	body, _ := json.Marshal(changeReq)

	req, _ := http.NewRequest("POST", "/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestChangePassword_ShortNewPassword tests password change with short new password
func TestChangePassword_ShortNewPassword(t *testing.T) {
	db := setupTestDB(t)
	router, jwtManager := setupTestRouter(db)

	// Create test user
	passwordHash, err := auth.HashPassword("oldpassword123")
	require.NoError(t, err)

	user := models.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: passwordHash,
		Role:         models.RoleUser,
		IsActive:     true,
	}
	require.NoError(t, db.Create(&user).Error)

	// Generate token
	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(user.Role))
	require.NoError(t, err)

	// Try to change with short new password
	changeReq := map[string]string{
		"current_password": "oldpassword123",
		"new_password":     "short",
	}
	body, _ := json.Marshal(changeReq)

	req, _ := http.NewRequest("POST", "/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
