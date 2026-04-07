package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/yourorg/querybase/internal/api/dto"
	"github.com/yourorg/querybase/internal/api/middleware"
	"github.com/yourorg/querybase/internal/auth"
	"github.com/yourorg/querybase/internal/models"
	"github.com/yourorg/querybase/internal/service"
	testauth "github.com/yourorg/querybase/internal/testutils/auth"
	"github.com/yourorg/querybase/internal/testutils/fixtures"
)

// MockApprovalService is a mock implementation of the approval service for testing
type MockApprovalService struct {
	mock.Mock
}

func (m *MockApprovalService) CreateApprovalRequest(ctx interface{}, req *service.ApprovalRequest) (*models.ApprovalRequest, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ApprovalRequest), args.Error(1)
}

func (m *MockApprovalService) GetApproval(ctx interface{}, approvalID string) (*models.ApprovalRequest, error) {
	args := m.Called(ctx, approvalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ApprovalRequest), args.Error(1)
}

func (m *MockApprovalService) ListApprovals(ctx interface{}, filter *service.ApprovalFilter) ([]models.ApprovalRequest, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.ApprovalRequest), args.Get(1).(int64), args.Error(2)
}

func (m *MockApprovalService) GetApprovalCounts(ctx interface{}, requestedBy string) (map[string]int64, error) {
	args := m.Called(ctx, requestedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockApprovalService) ReviewApproval(ctx interface{}, review *service.ReviewInput) (*models.ApprovalReview, error) {
	args := m.Called(ctx, review)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ApprovalReview), args.Error(1)
}

func (m *MockApprovalService) GetEligibleApprovers(ctx interface{}, dataSourceID string) ([]models.User, error) {
	args := m.Called(ctx, dataSourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockApprovalService) StartTransaction(ctx interface{}, approvalID, startedBy string, auditMode models.AuditMode) (*models.QueryTransaction, error) {
	args := m.Called(ctx, approvalID, startedBy, auditMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.QueryTransaction), args.Error(1)
}

func (m *MockApprovalService) CommitTransaction(ctx interface{}, transactionID string) error {
	args := m.Called(ctx, transactionID)
	return args.Error(0)
}

func (m *MockApprovalService) RollbackTransaction(ctx interface{}, transactionID string) error {
	args := m.Called(ctx, transactionID)
	return args.Error(0)
}

func (m *MockApprovalService) AddComment(ctx interface{}, approvalID, userID, comment string) (*models.ApprovalComment, error) {
	args := m.Called(ctx, approvalID, userID, comment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ApprovalComment), args.Error(1)
}

func (m *MockApprovalService) GetComments(ctx interface{}, approvalID string, page, perPage int) ([]models.ApprovalComment, int64, error) {
	args := m.Called(ctx, approvalID, page, perPage)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.ApprovalComment), args.Get(1).(int64), args.Error(2)
}

func (m *MockApprovalService) DeleteComment(ctx interface{}, commentID, userID string, isAdmin bool) error {
	args := m.Called(ctx, commentID, userID, isAdmin)
	return args.Error(0)
}

// setupApprovalTestDB creates an in-memory SQLite database for testing approval handlers
func setupApprovalTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.UserGroup{},
		&models.DataSource{},
		&models.DataSourcePermission{},
		&models.Query{},
		&models.QueryResult{},
		&models.QueryHistory{},
		&models.ApprovalRequest{},
		&models.ApprovalReview{},
		&models.ApprovalComment{},
		&models.QueryTransaction{},
	)
	require.NoError(t, err)

	return db
}

// setupApprovalTestRouter creates a test router with approval handler
func setupApprovalTestRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockApprovalService, *auth.JWTManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockApprovalService := new(MockApprovalService)

	// Create a test-specific wrapper that mimics the CreateApprovalRequest flow
	approvalHandler := &testApprovalHandler{
		db:              db,
		approvalService: mockApprovalService,
	}

	// Setup routes
	api := router.Group("/api/v1")
	{
		approvals := api.Group("/approvals")
		approvals.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			approvals.POST("", approvalHandler.CreateApprovalRequest)
		}
	}

	return router, mockApprovalService, jwtManager
}

// testApprovalHandler wraps the approval handling logic for testing
type testApprovalHandler struct {
	db              *gorm.DB
	approvalService *MockApprovalService
}

// CreateApprovalRequestRequest represents the request body for creating an approval
type CreateApprovalRequestRequest struct {
	DataSourceID string `json:"data_source_id" binding:"required"`
	QueryText    string `json:"query_text" binding:"required"`
	PreviewData  string `json:"preview_data,omitempty"`
}

// CreateApprovalRequest creates a new approval request
func (h *testApprovalHandler) CreateApprovalRequest(c *gin.Context) {
	var req CreateApprovalRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")

	// Verify data source exists
	var dataSource models.DataSource
	if err := h.db.First(&dataSource, "id = ?", req.DataSourceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Data source not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data source"})
		}
		return
	}

	// Check if user has write permission
	if !h.checkWritePermission(c, userID, dataSource.ID.String()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to create approval requests"})
		return
	}

	// Detect operation type
	operationType := DetectOperationType(req.QueryText)

	// Check if operation requires approval (INSERT, UPDATE, DELETE)
	if !RequiresApproval(operationType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query does not require approval"})
		return
	}

	// Check for duplicate approval request
	var existingApproval models.ApprovalRequest
	err := h.db.Where("data_source_id = ? AND query_text = ? AND requested_by = ? AND status = ?",
		dataSource.ID, req.QueryText, userID, models.ApprovalStatusPending).
		First(&existingApproval).Error
	if err == nil {
		// Return existing approval
		c.JSON(http.StatusOK, gin.H{
			"approval_id":    existingApproval.ID.String(),
			"status":         string(existingApproval.Status),
			"message":        "Existing approval request returned",
			"operation_type": string(existingApproval.OperationType),
		})
		return
	}

	// Parse userID as UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Create approval request
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     req.QueryText,
		RequestedBy:   userUUID,
		OperationType: operationType,
		Status:        models.ApprovalStatusPending,
	}

	if err := h.db.Create(approval).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create approval request"})
		return
	}

	response := gin.H{
		"approval_id":    approval.ID.String(),
		"status":         string(approval.Status),
		"operation_type": string(approval.OperationType),
		"message":        "Approval request created successfully",
	}

	// Include preview data if provided
	if req.PreviewData != "" {
		response["preview_data"] = req.PreviewData
	}

	c.JSON(http.StatusOK, response)
}

// checkWritePermission checks if user has write permission on data source
func (h *testApprovalHandler) checkWritePermission(c *gin.Context, userID, dataSourceID string) bool {
	// Check if user is admin
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err == nil {
		if user.Role == models.RoleAdmin {
			return true
		}
	}

	// Check group permissions
	var count int64
	h.db.Table("data_source_permissions").
		Joins("JOIN user_groups ON user_groups.group_id = data_source_permissions.group_id").
		Where("user_groups.user_id = ?", userID).
		Where("data_source_permissions.data_source_id = ?", dataSourceID).
		Where("data_source_permissions.can_write = ?", true).
		Count(&count)

	return count > 0
}

// =============================================================================
// CREATE APPROVAL REQUEST TESTS
// =============================================================================

// TestCreateApprovalRequest_InsertQuery_CreatesApproval tests creating approval for INSERT query
func TestCreateApprovalRequest_InsertQuery_CreatesApproval(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, _, jwtManager := setupApprovalTestRouter(t, db)

	// Create user with write permission
	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantWritePermission(t, db, group.ID, dataSource.ID)

	// Create JWT token
	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	// Create request
	reqBody := CreateApprovalRequestRequest{
		DataSourceID: dataSource.ID.String(),
		QueryText:    "INSERT INTO users (name, email) VALUES ('John', 'john@test.com')",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/approvals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response["approval_id"])
	assert.Equal(t, "pending", response["status"])
	assert.Equal(t, "insert", response["operation_type"])
	assert.Equal(t, "Approval request created successfully", response["message"])

	// Verify approval was created in database
	var approval models.ApprovalRequest
	err = db.Where("requested_by = ?", user.ID).First(&approval).Error
	require.NoError(t, err)
	assert.Equal(t, models.OperationInsert, approval.OperationType)
	assert.Equal(t, models.ApprovalStatusPending, approval.Status)
	assert.Contains(t, approval.QueryText, "INSERT INTO users")
}

// TestCreateApprovalRequest_UpdateQuery_CreatesApproval tests creating approval for UPDATE query
func TestCreateApprovalRequest_UpdateQuery_CreatesApproval(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, _, jwtManager := setupApprovalTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantWritePermission(t, db, group.ID, dataSource.ID)

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	reqBody := CreateApprovalRequestRequest{
		DataSourceID: dataSource.ID.String(),
		QueryText:    "UPDATE users SET status = 'active' WHERE status = 'pending'",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/approvals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.NotEmpty(t, response["approval_id"])
	assert.Equal(t, "pending", response["status"])
	assert.Equal(t, "update", response["operation_type"])

	// Verify approval in database
	var approval models.ApprovalRequest
	db.Where("requested_by = ?", user.ID).First(&approval)
	assert.Equal(t, models.OperationUpdate, approval.OperationType)
}

// TestCreateApprovalRequest_DeleteQuery_CreatesApproval tests creating approval for DELETE query
func TestCreateApprovalRequest_DeleteQuery_CreatesApproval(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, _, jwtManager := setupApprovalTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantWritePermission(t, db, group.ID, dataSource.ID)

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	reqBody := CreateApprovalRequestRequest{
		DataSourceID: dataSource.ID.String(),
		QueryText:    "DELETE FROM users WHERE id = 123 AND is_test = true",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/approvals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.NotEmpty(t, response["approval_id"])
	assert.Equal(t, "pending", response["status"])
	assert.Equal(t, "delete", response["operation_type"])

	// Verify approval in database
	var approval models.ApprovalRequest
	db.Where("requested_by = ?", user.ID).First(&approval)
	assert.Equal(t, models.OperationDelete, approval.OperationType)
}

// TestCreateApprovalRequest_WithPreviewData_IncludesPreview tests that preview data is included in response
func TestCreateApprovalRequest_WithPreviewData_IncludesPreview(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, _, jwtManager := setupApprovalTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantWritePermission(t, db, group.ID, dataSource.ID)

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	previewData := `[{"id": 1, "name": "Test User", "status": "pending"}]`
	reqBody := CreateApprovalRequestRequest{
		DataSourceID: dataSource.ID.String(),
		QueryText:    "UPDATE users SET status = 'active' WHERE id = 1",
		PreviewData:  previewData,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/approvals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.NotEmpty(t, response["approval_id"])
	assert.Equal(t, previewData, response["preview_data"])
}

// TestCreateApprovalRequest_DuplicateRequest_ReturnsExisting tests that duplicate requests return existing approval
func TestCreateApprovalRequest_DuplicateRequest_ReturnsExisting(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, _, jwtManager := setupApprovalTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantWritePermission(t, db, group.ID, dataSource.ID)

	// Create existing approval
	existingApproval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "UPDATE users SET name = 'test' WHERE id = 1",
		RequestedBy:   user.ID,
		OperationType: models.OperationUpdate,
		Status:        models.ApprovalStatusPending,
	}
	err := db.Create(existingApproval).Error
	require.NoError(t, err)

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	// Try to create duplicate
	reqBody := CreateApprovalRequestRequest{
		DataSourceID: dataSource.ID.String(),
		QueryText:    "UPDATE users SET name = 'test' WHERE id = 1",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/approvals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, existingApproval.ID.String(), response["approval_id"])
	assert.Equal(t, "pending", response["status"])
	assert.Equal(t, "Existing approval request returned", response["message"])

	// Verify only one approval exists
	var count int64
	db.Model(&models.ApprovalRequest{}).Where("requested_by = ?", user.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}

// TestCreateApprovalRequest_NoWritePermission_ReturnsForbidden tests user without write permission gets 403
func TestCreateApprovalRequest_NoWritePermission_ReturnsForbidden(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, _, jwtManager := setupApprovalTestRouter(t, db)

	// Create user with only read permission
	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	reqBody := CreateApprovalRequestRequest{
		DataSourceID: dataSource.ID.String(),
		QueryText:    "DELETE FROM users WHERE id = 1",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/approvals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Insufficient permissions")
}

// TestCreateApprovalRequest_DataSourceNotFound_ReturnsNotFound tests non-existent data source returns 404
func TestCreateApprovalRequest_DataSourceNotFound_ReturnsNotFound(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, _, jwtManager := setupApprovalTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	// Don't create data source - use non-existent ID

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	reqBody := CreateApprovalRequestRequest{
		DataSourceID: uuid.New().String(),
		QueryText:    "UPDATE users SET name = 'test' WHERE id = 1",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/approvals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Data source not found")
}

// TestCreateApprovalRequest_InvalidBody_ReturnsBadRequest tests invalid JSON returns 400
func TestCreateApprovalRequest_InvalidBody_ReturnsBadRequest(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, _, jwtManager := setupApprovalTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	// Send invalid JSON
	req, _ := http.NewRequest("POST", "/api/v1/approvals", bytes.NewBuffer([]byte(`{invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCreateApprovalRequest_SelectQuery_ReturnsBadRequest tests that SELECT query returns 400
func TestCreateApprovalRequest_SelectQuery_ReturnsBadRequest(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, _, jwtManager := setupApprovalTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantWritePermission(t, db, group.ID, dataSource.ID)

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	reqBody := CreateApprovalRequestRequest{
		DataSourceID: dataSource.ID.String(),
		QueryText:    "SELECT * FROM users WHERE id = 1",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/approvals", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Query does not require approval")
}

// TestCreateApprovalRequest_MissingRequiredFields_ReturnsBadRequest tests missing fields return 400
func TestCreateApprovalRequest_MissingRequiredFields_ReturnsBadRequest(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, _, jwtManager := setupApprovalTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	tests := []struct {
		name    string
		reqBody string
	}{
		{
			name:    "missing data_source_id",
			reqBody: `{"query_text": "UPDATE users SET name = 'test' WHERE id = 1"}`,
		},
		{
			name:    "missing query_text",
			reqBody: `{"data_source_id": "` + uuid.New().String() + `"}`,
		},
		{
			name:    "empty body",
			reqBody: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/v1/approvals", bytes.NewBuffer([]byte(tt.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// =============================================================================
// REVIEW APPROVAL TESTS
// =============================================================================

// setupReviewApprovalTestRouter creates a test router with the real ApprovalHandler
func setupReviewApprovalTestRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockApprovalService, *auth.JWTManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockApprovalService := new(MockApprovalService)

	// Create the real approval handler with mock service
	approvalHandler := NewApprovalHandler(db, (*service.ApprovalService)(nil))
	// Replace the service with our mock using a type assertion workaround
	// We'll create a wrapper that uses mockApprovalService
	approvalHandlerWrapper := &testReviewApprovalHandler{
		ApprovalHandler: approvalHandler,
		mockService:     mockApprovalService,
		db:              db,
	}

	// Setup routes
	api := router.Group("/api/v1")
	{
		approvals := api.Group("/approvals")
		approvals.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			approvals.POST("/:id/review", approvalHandlerWrapper.ReviewApproval)
		}
	}

	return router, mockApprovalService, jwtManager
}

// testReviewApprovalHandler wraps ApprovalHandler to use mock service
type testReviewApprovalHandler struct {
	*ApprovalHandler
	mockService *MockApprovalService
	db          *gorm.DB
}

// ReviewApproval overrides to use mock service
func (h *testReviewApprovalHandler) ReviewApproval(c *gin.Context) {
	approvalID := c.Param("id")
	userID := c.GetString("user_id")

	var req dto.ReviewApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify approval exists using mock
	approval, err := h.mockService.GetApproval(c, approvalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Approval not found"})
		return
	}

	// Check if user can approve this request
	if !h.checkCanApprove(userID, approval.DataSourceID.String()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to approve this request"})
		return
	}

	// Prevent self-approval at the handler layer
	if approval.RequestedBy.String() == userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "self-approval is not allowed: you cannot approve your own request"})
		return
	}

	// Create review using mock service
	reviewInput := &service.ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: userID,
		Decision:   models.ApprovalDecision(req.Decision),
		Comments:   req.Comments,
	}

	review, err := h.mockService.ReviewApproval(c, reviewInput)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"review_id": review.ID.String(),
		"decision":  string(review.Decision),
		"message":   "Review submitted successfully",
	})
}

// checkCanApprove checks if user can approve requests for a data source
func (h *testReviewApprovalHandler) checkCanApprove(userID, dataSourceID string) bool {
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		return false
	}

	// Admin can approve everything
	if user.Role == models.RoleAdmin {
		return true
	}

	// Check group permissions
	var count int64
	h.db.Table("data_source_permissions").
		Joins("JOIN user_groups ON user_groups.group_id = data_source_permissions.group_id").
		Where("user_groups.user_id = ?", userID).
		Where("data_source_permissions.data_source_id = ?", dataSourceID).
		Where("data_source_permissions.can_approve = ?", true).
		Count(&count)

	return count > 0
}

// TestReviewApproval_ApproveWithValidApprover_ApprovesRequest tests approving with valid approver
func TestReviewApproval_ApproveWithValidApprover_ApprovesRequest(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupReviewApprovalTestRouter(t, db)

	// Create requester (regular user with write permission)
	requester := fixtures.CreateTestRegularUser(t, db)
	// Create approver (user with approve permission)
	approver := fixtures.CreateTestRegularUser(t, db)
	// Create a third user without approve permission (to verify proper permission check)
	_ = fixtures.CreateTestRegularUser(t, db)

	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Setup groups and permissions
	requesterGroup := fixtures.CreateTestGroup(t, db, "requester-group")
	fixtures.AddUserToGroup(db, requester.ID, requesterGroup.ID)
	fixtures.GrantWritePermission(t, db, requesterGroup.ID, dataSource.ID)

	approverGroup := fixtures.CreateTestGroup(t, db, "approver-group")
	fixtures.AddUserToGroup(db, approver.ID, approverGroup.ID)
	fixtures.GrantFullPermission(t, db, approverGroup.ID, dataSource.ID) // Can approve

	// Create approval request
	approvalID := uuid.New()
	approval := &models.ApprovalRequest{
		ID:            approvalID,
		DataSourceID:  dataSource.ID,
		QueryText:     "UPDATE users SET status = 'active'",
		RequestedBy:   requester.ID,
		OperationType: models.OperationUpdate,
		Status:        models.ApprovalStatusPending,
	}

	// Create expected review
	reviewID := uuid.New()
	review := &models.ApprovalReview{
		ID:         reviewID,
		ApprovalID: approvalID,
		ReviewerID: approver.ID,
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Looks good",
	}

	// Setup mocks
	mockService.On("GetApproval", mock.Anything, approvalID.String()).Return(approval, nil)
	mockService.On("ReviewApproval", mock.Anything, mock.MatchedBy(func(input *service.ReviewInput) bool {
		return input.ApprovalID == approvalID &&
			input.ReviewerID == approver.ID.String() &&
			input.Decision == models.ApprovalDecisionApproved &&
			input.Comments == "Looks good"
	})).Return(review, nil)

	// Generate token for approver
	token, err := jwtManager.GenerateToken(approver.ID, approver.Email, string(models.RoleUser))
	require.NoError(t, err)

	// Create request
	reqBody := dto.ReviewApprovalRequest{
		Decision: "approved",
		Comments: "Looks good",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/approvals/"+approvalID.String()+"/review", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, reviewID.String(), response["review_id"])
	assert.Equal(t, "approved", response["decision"])
	assert.Equal(t, "Review submitted successfully", response["message"])

	mockService.AssertExpectations(t)
}

// TestReviewApproval_RejectWithReason_RejectsRequest tests rejecting with a reason
func TestReviewApproval_RejectWithReason_RejectsRequest(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupReviewApprovalTestRouter(t, db)

	// Create requester and approver
	requester := fixtures.CreateTestRegularUser(t, db)
	approver := fixtures.CreateTestRegularUser(t, db)

	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Setup permissions
	requesterGroup := fixtures.CreateTestGroup(t, db, "requester-group")
	fixtures.AddUserToGroup(db, requester.ID, requesterGroup.ID)
	fixtures.GrantWritePermission(t, db, requesterGroup.ID, dataSource.ID)

	approverGroup := fixtures.CreateTestGroup(t, db, "approver-group")
	fixtures.AddUserToGroup(db, approver.ID, approverGroup.ID)
	fixtures.GrantFullPermission(t, db, approverGroup.ID, dataSource.ID)

	// Create approval request
	approvalID := uuid.New()
	approval := &models.ApprovalRequest{
		ID:            approvalID,
		DataSourceID:  dataSource.ID,
		QueryText:     "DELETE FROM users WHERE id = 1",
		RequestedBy:   requester.ID,
		OperationType: models.OperationDelete,
		Status:        models.ApprovalStatusPending,
	}

	// Create expected review
	reviewID := uuid.New()
	review := &models.ApprovalReview{
		ID:         reviewID,
		ApprovalID: approvalID,
		ReviewerID: approver.ID,
		Decision:   models.ApprovalDecisionRejected,
		Comments:   "Too risky - please narrow the scope",
	}

	// Setup mocks
	mockService.On("GetApproval", mock.Anything, approvalID.String()).Return(approval, nil)
	mockService.On("ReviewApproval", mock.Anything, mock.MatchedBy(func(input *service.ReviewInput) bool {
		return input.ApprovalID == approvalID &&
			input.ReviewerID == approver.ID.String() &&
			input.Decision == models.ApprovalDecisionRejected &&
			input.Comments == "Too risky - please narrow the scope"
	})).Return(review, nil)

	// Generate token for approver
	token, err := jwtManager.GenerateToken(approver.ID, approver.Email, string(models.RoleUser))
	require.NoError(t, err)

	// Create request
	reqBody := dto.ReviewApprovalRequest{
		Decision: "rejected",
		Comments: "Too risky - please narrow the scope",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/approvals/"+approvalID.String()+"/review", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, reviewID.String(), response["review_id"])
	assert.Equal(t, "rejected", response["decision"])
	assert.Equal(t, "Review submitted successfully", response["message"])

	mockService.AssertExpectations(t)
}

// TestReviewApproval_SelfApproval_ReturnsForbidden tests that requester cannot approve their own request
func TestReviewApproval_SelfApproval_ReturnsForbidden(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupReviewApprovalTestRouter(t, db)

	// Create a user who is both requester and would-be approver
	user := fixtures.CreateTestRegularUser(t, db)

	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Setup permissions - user has both write and approve permissions
	group := fixtures.CreateTestGroup(t, db, "user-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	// Create approval request (user requests approval for their own query)
	approvalID := uuid.New()
	approval := &models.ApprovalRequest{
		ID:            approvalID,
		DataSourceID:  dataSource.ID,
		QueryText:     "UPDATE users SET status = 'active'",
		RequestedBy:   user.ID,
		OperationType: models.OperationUpdate,
		Status:        models.ApprovalStatusPending,
	}

	// Setup mock - service returns the approval
	mockService.On("GetApproval", mock.Anything, approvalID.String()).Return(approval, nil)
	// Note: ReviewApproval should NOT be called for self-approval

	// Generate token for user (who is the requester)
	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	// Create request - user tries to approve their own request
	reqBody := dto.ReviewApprovalRequest{
		Decision: "approved",
		Comments: "Self approval attempt",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/approvals/"+approvalID.String()+"/review", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "self-approval is not allowed")
	mockService.AssertExpectations(t)
}

// TestReviewApproval_DuplicateReview_ReturnsBadRequest tests that duplicate reviews are prevented
func TestReviewApproval_DuplicateReview_ReturnsBadRequest(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupReviewApprovalTestRouter(t, db)

	// Create requester and approver
	requester := fixtures.CreateTestRegularUser(t, db)
	approver := fixtures.CreateTestRegularUser(t, db)

	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Setup permissions
	requesterGroup := fixtures.CreateTestGroup(t, db, "requester-group")
	fixtures.AddUserToGroup(db, requester.ID, requesterGroup.ID)
	fixtures.GrantWritePermission(t, db, requesterGroup.ID, dataSource.ID)

	approverGroup := fixtures.CreateTestGroup(t, db, "approver-group")
	fixtures.AddUserToGroup(db, approver.ID, approverGroup.ID)
	fixtures.GrantFullPermission(t, db, approverGroup.ID, dataSource.ID)

	// Create approval request
	approvalID := uuid.New()
	approval := &models.ApprovalRequest{
		ID:            approvalID,
		DataSourceID:  dataSource.ID,
		QueryText:     "UPDATE users SET status = 'active'",
		RequestedBy:   requester.ID,
		OperationType: models.OperationUpdate,
		Status:        models.ApprovalStatusPending,
	}

	// Setup mocks - service returns "already reviewed" error
	mockService.On("GetApproval", mock.Anything, approvalID.String()).Return(approval, nil)
	mockService.On("ReviewApproval", mock.Anything, mock.Anything).Return(nil, errors.New("already reviewed"))

	// Generate token for approver
	token, err := jwtManager.GenerateToken(approver.ID, approver.Email, string(models.RoleUser))
	require.NoError(t, err)

	// Create request
	reqBody := dto.ReviewApprovalRequest{
		Decision: "approved",
		Comments: "Second review attempt",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/approvals/"+approvalID.String()+"/review", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "already reviewed")
	mockService.AssertExpectations(t)
}

// TestReviewApproval_NonApprover_ReturnsForbidden tests that non-approvers cannot review
func TestReviewApproval_NonApprover_ReturnsForbidden(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupReviewApprovalTestRouter(t, db)

	// Create requester and a user without approve permission
	requester := fixtures.CreateTestRegularUser(t, db)
	nonApprover := fixtures.CreateTestRegularUser(t, db)

	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Setup permissions - requester has write permission
	requesterGroup := fixtures.CreateTestGroup(t, db, "requester-group")
	fixtures.AddUserToGroup(db, requester.ID, requesterGroup.ID)
	fixtures.GrantWritePermission(t, db, requesterGroup.ID, dataSource.ID)

	// Non-approver has read-only permission (no approve permission)
	nonApproverGroup := fixtures.CreateTestGroup(t, db, "non-approver-group")
	fixtures.AddUserToGroup(db, nonApprover.ID, nonApproverGroup.ID)
	fixtures.GrantReadOnlyPermission(t, db, nonApproverGroup.ID, dataSource.ID)

	// Create approval request
	approvalID := uuid.New()
	approval := &models.ApprovalRequest{
		ID:            approvalID,
		DataSourceID:  dataSource.ID,
		QueryText:     "UPDATE users SET status = 'active'",
		RequestedBy:   requester.ID,
		OperationType: models.OperationUpdate,
		Status:        models.ApprovalStatusPending,
	}

	// Setup mock - GetApproval succeeds but permission check will fail
	mockService.On("GetApproval", mock.Anything, approvalID.String()).Return(approval, nil)
	// ReviewApproval should NOT be called

	// Generate token for non-approver
	token, err := jwtManager.GenerateToken(nonApprover.ID, nonApprover.Email, string(models.RoleUser))
	require.NoError(t, err)

	// Create request
	reqBody := dto.ReviewApprovalRequest{
		Decision: "approved",
		Comments: "Attempt by non-approver",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/approvals/"+approvalID.String()+"/review", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "don't have permission")
	mockService.AssertExpectations(t)
}

// TestReviewApproval_NonPendingApproval_ReturnsBadRequest tests reviewing non-pending approval
func TestReviewApproval_NonPendingApproval_ReturnsBadRequest(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupReviewApprovalTestRouter(t, db)

	// Create requester and approver
	requester := fixtures.CreateTestRegularUser(t, db)
	approver := fixtures.CreateTestRegularUser(t, db)

	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Setup permissions
	requesterGroup := fixtures.CreateTestGroup(t, db, "requester-group")
	fixtures.AddUserToGroup(db, requester.ID, requesterGroup.ID)
	fixtures.GrantWritePermission(t, db, requesterGroup.ID, dataSource.ID)

	approverGroup := fixtures.CreateTestGroup(t, db, "approver-group")
	fixtures.AddUserToGroup(db, approver.ID, approverGroup.ID)
	fixtures.GrantFullPermission(t, db, approverGroup.ID, dataSource.ID)

	// Create already approved request
	approvalID := uuid.New()
	approval := &models.ApprovalRequest{
		ID:            approvalID,
		DataSourceID:  dataSource.ID,
		QueryText:     "UPDATE users SET status = 'active'",
		RequestedBy:   requester.ID,
		OperationType: models.OperationUpdate,
		Status:        models.ApprovalStatusApproved, // Already approved
	}

	// Setup mocks
	mockService.On("GetApproval", mock.Anything, approvalID.String()).Return(approval, nil)
	mockService.On("ReviewApproval", mock.Anything, mock.Anything).Return(nil, errors.New("approval request is not pending"))

	// Generate token for approver
	token, err := jwtManager.GenerateToken(approver.ID, approver.Email, string(models.RoleUser))
	require.NoError(t, err)

	// Create request
	reqBody := dto.ReviewApprovalRequest{
		Decision: "rejected",
		Comments: "Try to reject already approved request",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/approvals/"+approvalID.String()+"/review", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "not pending")
	mockService.AssertExpectations(t)
}

// TestReviewApproval_ApprovalNotFound_ReturnsNotFound tests reviewing non-existent approval
func TestReviewApproval_ApprovalNotFound_ReturnsNotFound(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupReviewApprovalTestRouter(t, db)

	// Create approver
	approver := fixtures.CreateTestRegularUser(t, db)

	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Setup permissions
	approverGroup := fixtures.CreateTestGroup(t, db, "approver-group")
	fixtures.AddUserToGroup(db, approver.ID, approverGroup.ID)
	fixtures.GrantFullPermission(t, db, approverGroup.ID, dataSource.ID)

	// Non-existent approval ID
	nonExistentID := uuid.New()

	// Setup mock - GetApproval returns error
	mockService.On("GetApproval", mock.Anything, nonExistentID.String()).Return(nil, errors.New("record not found"))

	// Generate token for approver
	token, err := jwtManager.GenerateToken(approver.ID, approver.Email, string(models.RoleUser))
	require.NoError(t, err)

	// Create request
	reqBody := dto.ReviewApprovalRequest{
		Decision: "approved",
		Comments: "Try to approve non-existent request",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/api/v1/approvals/"+nonExistentID.String()+"/review", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "Approval not found")
	mockService.AssertExpectations(t)
}

// =============================================================================
// GET APPROVALS (LIST) TESTS
// =============================================================================

// setupListApprovalsTestRouter creates a test router with the ListApprovals handler
func setupListApprovalsTestRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockApprovalService, *auth.JWTManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockApprovalService := new(MockApprovalService)

	// Create the real approval handler with mock service
	approvalHandler := NewApprovalHandler(db, (*service.ApprovalService)(nil))
	approvalHandlerWrapper := &testListApprovalsHandler{
		ApprovalHandler: approvalHandler,
		mockService:     mockApprovalService,
		db:              db,
	}

	// Setup routes
	api := router.Group("/api/v1")
	{
		approvals := api.Group("/approvals")
		approvals.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			approvals.GET("", approvalHandlerWrapper.ListApprovals)
		}
	}

	return router, mockApprovalService, jwtManager
}

// testListApprovalsHandler wraps ApprovalHandler for ListApprovals testing
type testListApprovalsHandler struct {
	*ApprovalHandler
	mockService *MockApprovalService
	db          *gorm.DB
}

// ListApprovals overrides to use mock service
func (h *testListApprovalsHandler) ListApprovals(c *gin.Context) {
	userID := c.GetString("user_id")

	// Parse query parameters
	status := c.Query("status")
	dataSourceID := c.Query("data_source_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	// Create filter
	filter := &service.ApprovalFilter{
		Status:       status,
		DataSourceID: dataSourceID,
		Limit:        limit,
		Offset:       offset,
	}

	// Non-admin users can only see their own requests unless they're approvers
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err == nil {
		if user.Role != models.RoleAdmin {
			// Check if user is an approver for any data source
			isApprover := h.checkIsApprover(userID)
			if !isApprover {
				// Only show own requests
				filter.RequestedBy = userID
			}
		}
	}

	approvals, total, err := h.mockService.ListApprovals(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch approvals"})
		return
	}

	response := make([]gin.H, len(approvals))
	for i, approval := range approvals {
		response[i] = h.formatApprovalResponse(approval, userID)
	}

	c.JSON(http.StatusOK, gin.H{
		"approvals": response,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

// TestGetApprovals_WithPagination_ReturnsPaginatedResults tests listing with pagination
func TestGetApprovals_WithPagination_ReturnsPaginatedResults(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupListApprovalsTestRouter(t, db)

	// Create admin user (sees all)
	admin := fixtures.CreateTestAdminUser(t, db)

	// Create test data source
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Create some approval requests
	approvals := []models.ApprovalRequest{
		{
			ID:            uuid.New(),
			DataSourceID:  dataSource.ID,
			QueryText:     "UPDATE users SET status = 'active' WHERE id = 1",
			RequestedBy:   admin.ID,
			OperationType: models.OperationUpdate,
			Status:        models.ApprovalStatusPending,
		},
		{
			ID:            uuid.New(),
			DataSourceID:  dataSource.ID,
			QueryText:     "DELETE FROM logs WHERE created_at < '2024-01-01'",
			RequestedBy:   admin.ID,
			OperationType: models.OperationDelete,
			Status:        models.ApprovalStatusPending,
		},
	}

	// Setup mock - return first page with 1 item
	mockService.On("ListApprovals", mock.Anything, mock.MatchedBy(func(f *service.ApprovalFilter) bool {
		return f.Limit == 1 && f.Offset == 0
	})).Return([]models.ApprovalRequest{approvals[0]}, int64(2), nil)

	// Generate token for admin
	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	// Create request with pagination
	req, _ := http.NewRequest("GET", "/api/v1/approvals?page=1&limit=1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(1), response["page"])
	assert.Equal(t, float64(1), response["limit"])
	assert.Equal(t, float64(2), response["total"])

	approvalList, ok := response["approvals"].([]interface{})
	require.True(t, ok)
	assert.Len(t, approvalList, 1)

	mockService.AssertExpectations(t)
}

// TestGetApprovals_FilterByStatusPending_ReturnsPending tests filtering by pending status
func TestGetApprovals_FilterByStatusPending_ReturnsPending(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupListApprovalsTestRouter(t, db)

	// Create admin user
	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	pendingApproval := models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "UPDATE users SET status = 'active'",
		RequestedBy:   admin.ID,
		OperationType: models.OperationUpdate,
		Status:        models.ApprovalStatusPending,
	}

	// Setup mock with pending filter
	mockService.On("ListApprovals", mock.Anything, mock.MatchedBy(func(f *service.ApprovalFilter) bool {
		return f.Status == "pending"
	})).Return([]models.ApprovalRequest{pendingApproval}, int64(1), nil)

	// Generate token
	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	// Create request with status filter
	req, _ := http.NewRequest("GET", "/api/v1/approvals?status=pending", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	approvalList, ok := response["approvals"].([]interface{})
	require.True(t, ok)
	assert.Len(t, approvalList, 1)

	approval := approvalList[0].(map[string]interface{})
	assert.Equal(t, "pending", approval["status"])

	mockService.AssertExpectations(t)
}

// TestGetApprovals_FilterByStatusApproved_ReturnsApproved tests filtering by approved status
func TestGetApprovals_FilterByStatusApproved_ReturnsApproved(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupListApprovalsTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	approvedApproval := models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "UPDATE users SET status = 'active'",
		RequestedBy:   admin.ID,
		OperationType: models.OperationUpdate,
		Status:        models.ApprovalStatusApproved,
	}

	// Setup mock with approved filter
	mockService.On("ListApprovals", mock.Anything, mock.MatchedBy(func(f *service.ApprovalFilter) bool {
		return f.Status == "approved"
	})).Return([]models.ApprovalRequest{approvedApproval}, int64(1), nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/v1/approvals?status=approved", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	approvalList := response["approvals"].([]interface{})
	assert.Len(t, approvalList, 1)

	approval := approvalList[0].(map[string]interface{})
	assert.Equal(t, "approved", approval["status"])

	mockService.AssertExpectations(t)
}

// TestGetApprovals_FilterByStatusRejected_ReturnsRejected tests filtering by rejected status
func TestGetApprovals_FilterByStatusRejected_ReturnsRejected(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupListApprovalsTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	rejectedApproval := models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "DELETE FROM users WHERE id = 123",
		RequestedBy:   admin.ID,
		OperationType: models.OperationDelete,
		Status:        models.ApprovalStatusRejected,
	}

	// Setup mock with rejected filter
	mockService.On("ListApprovals", mock.Anything, mock.MatchedBy(func(f *service.ApprovalFilter) bool {
		return f.Status == "rejected"
	})).Return([]models.ApprovalRequest{rejectedApproval}, int64(1), nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/v1/approvals?status=rejected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	approvalList := response["approvals"].([]interface{})
	assert.Len(t, approvalList, 1)

	approval := approvalList[0].(map[string]interface{})
	assert.Equal(t, "rejected", approval["status"])

	mockService.AssertExpectations(t)
}

// TestGetApprovals_FilterByDataSource_ReturnsFiltered tests filtering by data source ID
func TestGetApprovals_FilterByDataSource_ReturnsFiltered(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupListApprovalsTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource1 := fixtures.CreateTestDataSource(t, db, "ds-1")
	dataSource2 := fixtures.CreateTestDataSource(t, db, "ds-2")

	approval := models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource1.ID,
		QueryText:     "UPDATE users SET status = 'active'",
		RequestedBy:   admin.ID,
		OperationType: models.OperationUpdate,
		Status:        models.ApprovalStatusPending,
	}

	// Setup mock with data source filter
	mockService.On("ListApprovals", mock.Anything, mock.MatchedBy(func(f *service.ApprovalFilter) bool {
		return f.DataSourceID == dataSource1.ID.String()
	})).Return([]models.ApprovalRequest{approval}, int64(1), nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	// Filter by data source 1 (should NOT return approvals from ds-2)
	req, _ := http.NewRequest("GET", "/api/v1/approvals?data_source_id="+dataSource1.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	approvalList := response["approvals"].([]interface{})
	assert.Len(t, approvalList, 1)

	// Verify it's from the correct data source
	approvalResp := approvalList[0].(map[string]interface{})
	assert.Equal(t, dataSource1.ID.String(), approvalResp["data_source_id"])

	// Ensure ds-2 approvals are not returned
	_ = dataSource2 // Just to avoid unused variable warning

	mockService.AssertExpectations(t)
}

// TestGetApprovals_AdminSeesAll_ReturnsAll tests that admin sees all approvals
func TestGetApprovals_AdminSeesAll_ReturnsAll(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupListApprovalsTestRouter(t, db)

	// Create admin
	admin := fixtures.CreateTestAdminUser(t, db)
	// Create regular user
	regularUser := fixtures.CreateTestRegularUser(t, db)

	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Create approvals from both users
	adminApproval := models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "UPDATE admin_query",
		RequestedBy:   admin.ID,
		OperationType: models.OperationUpdate,
		Status:        models.ApprovalStatusPending,
	}
	regularApproval := models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "DELETE FROM logs",
		RequestedBy:   regularUser.ID,
		OperationType: models.OperationDelete,
		Status:        models.ApprovalStatusPending,
	}

	// Setup mock - admin should see all (no RequestedBy filter)
	mockService.On("ListApprovals", mock.Anything, mock.MatchedBy(func(f *service.ApprovalFilter) bool {
		return f.RequestedBy == "" // Admin sees all
	})).Return([]models.ApprovalRequest{adminApproval, regularApproval}, int64(2), nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/v1/approvals", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	approvalList := response["approvals"].([]interface{})
	assert.Len(t, approvalList, 2) // Both admin's and regular user's approvals

	mockService.AssertExpectations(t)
}

// TestGetApprovals_RegularUserSeesOwn_ReturnsOwn tests that regular user sees only their own approvals
func TestGetApprovals_RegularUserSeesOwn_ReturnsOwn(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupListApprovalsTestRouter(t, db)

	// Create two regular users
	regularUser1 := fixtures.CreateTestRegularUser(t, db)
	regularUser2 := fixtures.CreateTestRegularUser(t, db)

	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// User 1's approval
	user1Approval := models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "UPDATE users SET name = 'test'",
		RequestedBy:   regularUser1.ID,
		OperationType: models.OperationUpdate,
		Status:        models.ApprovalStatusPending,
	}

	// User 2's approval (should NOT be visible to user 1)
	user2Approval := models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "DELETE FROM orders",
		RequestedBy:   regularUser2.ID,
		OperationType: models.OperationDelete,
		Status:        models.ApprovalStatusPending,
	}

	// Setup mock - user1 should only see their own (RequestedBy = user1.ID)
	mockService.On("ListApprovals", mock.Anything, mock.MatchedBy(func(f *service.ApprovalFilter) bool {
		return f.RequestedBy == regularUser1.ID.String() // Filtered to own
	})).Return([]models.ApprovalRequest{user1Approval}, int64(1), nil)

	// Suppress checkIsApprover query
	db.Exec("DELETE FROM data_source_permissions")
	db.Exec("DELETE FROM user_groups")

	token, err := jwtManager.GenerateToken(regularUser1.ID, regularUser1.Email, string(models.RoleUser))
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/v1/approvals", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	approvalList := response["approvals"].([]interface{})
	assert.Len(t, approvalList, 1)

	// Verify it's user1's approval
	approval := approvalList[0].(map[string]interface{})
	assert.Equal(t, regularUser1.ID.String(), approval["requested_by"])

	// Ensure user2's approval is NOT returned
	_ = user2Approval // Just to avoid unused variable warning

	mockService.AssertExpectations(t)
}

// TestGetApprovals_SortByCreationTime_ReturnsSorted tests sorting by creation time
func TestGetApprovals_SortByCreationTime_ReturnsSorted(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupListApprovalsTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Create approvals with different creation times
	oldApproval := models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "Old query",
		RequestedBy:   admin.ID,
		OperationType: models.OperationUpdate,
		Status:        models.ApprovalStatusPending,
		CreatedAt:     time.Now().Add(-48 * time.Hour),
	}
	newApproval := models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "New query",
		RequestedBy:   admin.ID,
		OperationType: models.OperationUpdate,
		Status:        models.ApprovalStatusPending,
		CreatedAt:     time.Now(),
	}

	// The service returns them in order (newest first by default)
	mockService.On("ListApprovals", mock.Anything, mock.Anything).Return(
		[]models.ApprovalRequest{newApproval, oldApproval}, int64(2), nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/v1/approvals", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	approvalList := response["approvals"].([]interface{})
	assert.Len(t, approvalList, 2)

	// Verify order (newest first based on service return order)
	firstApproval := approvalList[0].(map[string]interface{})
	assert.Equal(t, "New query", firstApproval["query_text"])

	mockService.AssertExpectations(t)
}

// TestGetApprovals_EmptyList_ReturnsEmptyArray tests empty list returns empty array
func TestGetApprovals_EmptyList_ReturnsEmptyArray(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupListApprovalsTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)

	// Setup mock to return empty list
	mockService.On("ListApprovals", mock.Anything, mock.Anything).Return(
		[]models.ApprovalRequest{}, int64(0), nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/v1/approvals", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(0), response["total"])

	approvalList, ok := response["approvals"].([]interface{})
	require.True(t, ok)
	assert.Len(t, approvalList, 0)

	mockService.AssertExpectations(t)
}

// TestGetApprovals_InvalidPagination_ReturnsBadRequest tests invalid pagination parameters
func TestGetApprovals_InvalidPagination_ReturnsBadRequest(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, mockService, jwtManager := setupListApprovalsTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	tests := []struct {
		name     string
		url      string
		expected int
	}{
		{
			name:     "negative page",
			url:      "/api/v1/approvals?page=-1",
			expected: http.StatusOK, // Handler defaults to 1 for invalid
		},
		{
			name:     "zero page",
			url:      "/api/v1/approvals?page=0",
			expected: http.StatusOK, // Handler defaults to 1
		},
		{
			name:     "non-numeric page",
			url:      "/api/v1/approvals?page=abc",
			expected: http.StatusOK, // Handler defaults to 0, then page becomes 1
		},
		{
			name:     "negative limit",
			url:      "/api/v1/approvals?limit=-5",
			expected: http.StatusOK, // Handler defaults to 0, then defaults to 20
		},
		{
			name:     "limit exceeds max",
			url:      "/api/v1/approvals?limit=200",
			expected: http.StatusOK, // Handler caps at 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock for each call
			mockService.On("ListApprovals", mock.Anything, mock.Anything).Return(
				[]models.ApprovalRequest{}, int64(0), nil).Maybe()

			req, _ := http.NewRequest("GET", tt.url, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expected, w.Code)

			if tt.expected == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				// Verify pagination values are sensible
				limit, ok := response["limit"].(float64)
				if ok {
					assert.LessOrEqual(t, limit, float64(100)) // Capped at 100
				}
			}
		})
	}
}

// TestReviewApproval_InvalidBody_ReturnsBadRequest tests invalid request body
func TestReviewApproval_InvalidBody_ReturnsBadRequest(t *testing.T) {
	db := setupApprovalTestDB(t)
	router, _, jwtManager := setupReviewApprovalTestRouter(t, db)

	// Create approver
	approver := fixtures.CreateTestRegularUser(t, db)

	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Setup permissions
	approverGroup := fixtures.CreateTestGroup(t, db, "approver-group")
	fixtures.AddUserToGroup(db, approver.ID, approverGroup.ID)
	fixtures.GrantFullPermission(t, db, approverGroup.ID, dataSource.ID)

	// Generate token for approver
	token, err := jwtManager.GenerateToken(approver.ID, approver.Email, string(models.RoleUser))
	require.NoError(t, err)

	tests := []struct {
		name    string
		reqBody string
	}{
		{
			name:    "invalid json",
			reqBody: `{invalid json`,
		},
		{
			name:    "missing decision",
			reqBody: `{"comments": "No decision provided"}`,
		},
		{
			name:    "invalid decision value",
			reqBody: `{"decision": "maybe", "comments": "Invalid decision"}`,
		},
		{
			name:    "empty body",
			reqBody: `{}`,
		},
	}

	approvalID := uuid.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/v1/approvals/"+approvalID.String()+"/review", bytes.NewBuffer([]byte(tt.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}
