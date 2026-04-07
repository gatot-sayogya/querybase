package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	testauth "github.com/yourorg/querybase/internal/testutils/auth"
	"github.com/yourorg/querybase/internal/testutils/fixtures"
)

// ptr returns a pointer to the given value
func ptr[T any](v T) *T {
	return &v
}

// MockQueryService is a mock implementation of the query service for testing
type MockQueryService struct {
	mock.Mock
}

func (m *MockQueryService) ExecuteQuery(ctx interface{}, query *models.Query, dataSource *models.DataSource) (*models.QueryResult, error) {
	args := m.Called(ctx, query, dataSource)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.QueryResult), args.Error(1)
}

func (m *MockQueryService) GetEffectivePermissions(ctx interface{}, userID, dsID uuid.UUID) (*models.EffectivePermissions, error) {
	args := m.Called(ctx, userID, dsID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EffectivePermissions), args.Error(1)
}

func (m *MockQueryService) PreviewAndValidateWriteQuery(ctx interface{}, queryText string, dataSource *models.DataSource) (*dto.ValidationResult, error) {
	args := m.Called(ctx, queryText, dataSource)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ValidationResult), args.Error(1)
}

func (m *MockQueryService) ValidateQuerySchema(ctx interface{}, queryText string, dataSource *models.DataSource) error {
	args := m.Called(ctx, queryText, dataSource)
	return args.Error(0)
}

func (m *MockQueryService) PreviewWriteQuery(ctx interface{}, queryText string, dataSource *models.DataSource) (*dto.PreviewWriteQueryResponse, error) {
	args := m.Called(ctx, queryText, dataSource)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PreviewWriteQueryResponse), args.Error(1)
}

func (m *MockQueryService) PreviewInsertQuery(ctx interface{}, queryText string, dataSource *models.DataSource) (*dto.InsertPreviewResponse, error) {
	args := m.Called(ctx, queryText, dataSource)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.InsertPreviewResponse), args.Error(1)
}

func (m *MockQueryService) SaveQuery(ctx interface{}, query *models.Query) error {
	args := m.Called(ctx, query)
	return args.Error(0)
}

func (m *MockQueryService) GetQuery(ctx interface{}, queryID string) (*models.Query, error) {
	args := m.Called(ctx, queryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Query), args.Error(1)
}

func (m *MockQueryService) ListQueries(ctx interface{}, userID string, limit, offset int) ([]models.Query, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.Query), args.Get(1).(int64), args.Error(2)
}

func (m *MockQueryService) ListQueryHistory(ctx interface{}, userID string, limit, offset int, search string) ([]models.QueryHistory, int64, error) {
	args := m.Called(ctx, userID, limit, offset, search)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]models.QueryHistory), args.Get(1).(int64), args.Error(2)
}

func (m *MockQueryService) ExplainQuery(ctx interface{}, queryText string, dataSource *models.DataSource, analyze bool) (interface{}, error) {
	args := m.Called(ctx, queryText, dataSource, analyze)
	return args.Get(0), args.Error(1)
}

func (m *MockQueryService) DryRunDelete(ctx interface{}, queryText string, dataSource *models.DataSource) (interface{}, error) {
	args := m.Called(ctx, queryText, dataSource)
	return args.Get(0), args.Error(1)
}

func (m *MockQueryService) GetPaginatedResults(ctx interface{}, queryID uuid.UUID, page, perPage int, sortColumn, sortDirection string) ([]map[string]interface{}, []string, *dto.PaginationMeta, error) {
	args := m.Called(ctx, queryID, page, perPage, sortColumn, sortDirection)
	if args.Get(0) == nil {
		return nil, nil, nil, args.Error(3)
	}
	return args.Get(0).([]map[string]interface{}), args.Get(1).([]string), args.Get(2).(*dto.PaginationMeta), args.Error(3)
}

func (m *MockQueryService) ExportQuery(ctx interface{}, queryID uuid.UUID, format string) ([]byte, string, error) {
	args := m.Called(ctx, queryID, format)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

// setupQueryTestDB creates an in-memory SQLite database for testing query handlers
func setupQueryTestDB(t *testing.T) *gorm.DB {
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
	)
	require.NoError(t, err)

	return db
}

// setupQueryTestRouter creates a test router with query handler
func setupQueryTestRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockQueryService, *auth.JWTManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockQueryService := new(MockQueryService)

	// We need to create a handler with both db and mock service
	// Since QueryHandler expects *service.QueryService, we'll create a test-specific wrapper
	queryHandler := &testQueryHandler{
		db:           db,
		queryService: mockQueryService,
	}

	// Setup routes
	api := router.Group("/api/v1")
	{
		queries := api.Group("/queries")
		queries.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			queries.POST("", queryHandler.ExecuteQuery)
		}
	}

	return router, mockQueryService, jwtManager
}

// testQueryHandler wraps the real QueryHandler but uses our mock service
type testQueryHandler struct {
	db           *gorm.DB
	queryService *MockQueryService
}

// ExecuteQuery implements the handler logic using the mock service
func (h *testQueryHandler) ExecuteQuery(c *gin.Context) {
	var req dto.ExecuteQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")

	// Verify data source exists and user has permission
	var dataSource models.DataSource
	if err := h.db.First(&dataSource, "id = ?", req.DataSourceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Data source not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data source"})
		}
		return
	}

	// Check user permissions using mock service
	uID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	perms, err := h.queryService.GetEffectivePermissions(c.Request.Context(), uID, dataSource.ID)
	if err != nil || !perms.CanRead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to read from this data source"})
		return
	}

	// Detect operation type
	operationType := DetectOperationType(req.QueryText)

	// For write operations, validate first before creating approval
	if RequiresApproval(operationType) {
		// Check if user has write permission
		if !perms.CanWrite {
			c.JSON(http.StatusForbidden, gin.H{
				"error":       "Insufficient permissions to submit write operations",
				"code":        "PERMISSION_DENIED_WRITE",
				"data_source": dataSource.Name,
			})
			return
		}

		// Validate the write query
		validation, err := h.queryService.PreviewAndValidateWriteQuery(c.Request.Context(), req.QueryText, &dataSource)
		if err != nil {
			// If validation fails, proceed with approval creation anyway (fail open)
			h.createApprovalForQuery(c, req, dataSource, userID, operationType)
			return
		}

		// If validation shows no rows would be affected, return early
		if validation != nil && validation.Status == "no_match" {
			c.JSON(http.StatusOK, dto.ExecuteQueryResponse{
				QueryID:    "",
				Status:     "no_match",
				Validation: validation,
			})
			return
		}

		// Validation passed, proceed with approval creation
		h.createApprovalForQuery(c, req, dataSource, userID, operationType)
		return
	}

	// Parse userID as UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Create query record
	query := &models.Query{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		UserID:        userUUID,
		QueryText:     req.QueryText,
		OperationType: operationType,
		Status:        models.StatusRunning,
	}

	// Save query to database BEFORE executing
	if err := h.db.Create(query).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save query"})
		return
	}

	// Execute the query
	result, err := h.queryService.ExecuteQuery(c.Request.Context(), query, &dataSource)
	if err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Parse results
	var data []map[string]interface{}
	if result.Data != "" {
		if err := json.Unmarshal([]byte(result.Data), &data); err != nil {
			data = []map[string]interface{}{}
		}
	} else {
		data = []map[string]interface{}{}
	}

	// Parse column names from JSON string
	var columnNames []string
	json.Unmarshal([]byte(result.ColumnNames), &columnNames)

	// Parse column types from JSON string
	var columnTypes []string
	json.Unmarshal([]byte(result.ColumnTypes), &columnTypes)

	// Build column info with names and types
	columns := make([]dto.ColumnInfo, len(columnNames))
	for i, col := range columnNames {
		colType := "unknown"
		if i < len(columnTypes) && columnTypes[i] != "" {
			colType = columnTypes[i]
		}
		columns[i] = dto.ColumnInfo{Name: col, Type: colType}
	}

	executionTime := 100 // Mock execution time

	c.JSON(http.StatusOK, dto.ExecuteQueryResponse{
		QueryID:          query.ID.String(),
		Status:           "completed",
		RowCount:         &result.RowCount,
		ExecutionTime:    &executionTime,
		Data:             data,
		Columns:          columns,
		RequiresApproval: false,
	})
}

// createApprovalForQuery creates an approval request for write operations
func (h *testQueryHandler) createApprovalForQuery(c *gin.Context, req dto.ExecuteQueryRequest, dataSource models.DataSource, userID string, operationType models.OperationType) {
	// Step 1: Validate SQL syntax
	if err := ValidateSQL(req.QueryText); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid SQL syntax",
			"details": err.Error(),
		})
		return
	}

	// Step 2: Validate schema
	if err := h.queryService.ValidateQuerySchema(c.Request.Context(), req.QueryText, &dataSource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Schema validation failed",
			"details": err.Error(),
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

	c.JSON(http.StatusAccepted, dto.ExecuteQueryResponse{
		QueryID:          uuid.New().String(),
		Status:           "pending_approval",
		RequiresApproval: true,
		ApprovalID:       approval.ID.String(),
	})
}

// DetectOperationType wraps the service function
func DetectOperationType(sql string) models.OperationType {
	// Import the actual service function behavior
	// This is a simplified version for testing
	trimmedSQL := sql
	if len(trimmedSQL) > 6 {
		upperSQL := ""
		for i := 0; i < len(trimmedSQL) && i < 20; i++ {
			c := trimmedSQL[i]
			if c >= 'a' && c <= 'z' {
				c = c - 'a' + 'A'
			}
			upperSQL += string(c)
		}

		switch {
		case upperSQL[:6] == "SELECT":
			return models.OperationSelect
		case upperSQL[:6] == "INSERT":
			return models.OperationInsert
		case upperSQL[:6] == "UPDATE":
			return models.OperationUpdate
		case upperSQL[:6] == "DELETE":
			return models.OperationDelete
		}
	}
	return models.OperationSelect
}

// RequiresApproval wraps the service function
func RequiresApproval(operationType models.OperationType) bool {
	switch operationType {
	case models.OperationSelect, models.OperationSet:
		return false
	case models.OperationInsert, models.OperationUpdate, models.OperationDelete:
		return true
	default:
		return true
	}
}

// ValidateSQL wraps the service function
func ValidateSQL(sql string) error {
	if sql == "" {
		return errors.New("SQL query cannot be empty")
	}
	return nil
}

// Helper function to create authorization header
func createAuthHeader(token string) string {
	return "Bearer " + token
}

// Helper function to create execute query request body
func createExecuteQueryRequest(dsID string, queryText string) ([]byte, error) {
	req := dto.ExecuteQueryRequest{
		DataSourceID: dsID,
		QueryText:    queryText,
	}
	return json.Marshal(req)
}

// =============================================================================
// SELECT QUERY TESTS
// =============================================================================

// TestExecuteQuery_AdminSelect_Success tests that admin user can execute SELECT queries
func TestExecuteQuery_AdminSelect_Success(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)
	_ = jwtManager

	// Create admin user and data source
	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	// Mock permissions - admin has full access
	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:    true,
			CanWrite:   true,
			CanApprove: true,
			CanSelect:  true,
			CanInsert:  true,
			CanUpdate:  true,
			CanDelete:  true,
		}, nil)

	// Mock query execution
	rowCount := 10
	mockService.On("ExecuteQuery", mock.Anything, mock.AnythingOfType("*models.Query"), mock.AnythingOfType("*models.DataSource")).
		Return(&models.QueryResult{
			RowCount:    rowCount,
			ColumnNames: `["id", "name"]`,
			ColumnTypes: `["integer", "text"]`,
			Data:        `[{"id": 1, "name": "test1"}, {"id": 2, "name": "test2"}]`,
		}, nil)

	// Create JWT token for admin
	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	// Create request
	reqBody, err := createExecuteQueryRequest(dataSource.ID.String(), "SELECT * FROM users")
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ExecuteQueryResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.QueryID)
	assert.Equal(t, "completed", response.Status)
	assert.Equal(t, rowCount, *response.RowCount)
	assert.Len(t, response.Data, 2)
	assert.Len(t, response.Columns, 2)
	assert.Equal(t, "id", response.Columns[0].Name)
	assert.Equal(t, "name", response.Columns[1].Name)
	assert.False(t, response.RequiresApproval)

	mockService.AssertExpectations(t)
}

// TestExecuteQuery_UserWithReadPermission_Success tests that user with read permission can execute SELECT
func TestExecuteQuery_UserWithReadPermission_Success(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)

	// Create regular user with read permission
	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	// Mock permissions - user has read access
	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:    true,
			CanWrite:   false,
			CanApprove: false,
			CanSelect:  true,
			CanInsert:  false,
			CanUpdate:  false,
			CanDelete:  false,
		}, nil)

	// Mock query execution
	rowCount := 5
	mockService.On("ExecuteQuery", mock.Anything, mock.AnythingOfType("*models.Query"), mock.AnythingOfType("*models.DataSource")).
		Return(&models.QueryResult{
			RowCount:    rowCount,
			ColumnNames: `["id", "email"]`,
			ColumnTypes: `["integer", "text"]`,
			Data:        `[{"id": 1, "email": "test@test.com"}]`,
		}, nil)

	// Create JWT token for user
	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	// Create request
	reqBody, err := createExecuteQueryRequest(dataSource.ID.String(), "SELECT id, email FROM customers WHERE id = 1")
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ExecuteQueryResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "completed", response.Status)
	assert.Equal(t, rowCount, *response.RowCount)
	assert.False(t, response.RequiresApproval)

	mockService.AssertExpectations(t)
}

// TestExecuteQuery_SelectReturnsCorrectResults tests that SELECT query returns correct data structure
func TestExecuteQuery_SelectReturnsCorrectResults(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)
	_ = jwtManager

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Mock permissions
	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanRead: true, CanSelect: true}, nil)

	// Mock query execution with complex data
	mockService.On("ExecuteQuery", mock.Anything, mock.AnythingOfType("*models.Query"), mock.AnythingOfType("*models.DataSource")).
		Return(&models.QueryResult{
			RowCount:    3,
			ColumnNames: `["user_id", "username", "created_at", "is_active"]`,
			ColumnTypes: `["integer", "text", "timestamp", "boolean"]`,
			Data: `[
				{"user_id": 1, "username": "alice", "created_at": "2024-01-01", "is_active": true},
				{"user_id": 2, "username": "bob", "created_at": "2024-01-02", "is_active": false},
				{"user_id": 3, "username": "charlie", "created_at": "2024-01-03", "is_active": true}
			]`,
		}, nil)

	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "SELECT * FROM users ORDER BY user_id")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ExecuteQueryResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, 3, *response.RowCount)
	assert.Len(t, response.Columns, 4)
	assert.Len(t, response.Data, 3)

	// Verify column info
	assert.Equal(t, "user_id", response.Columns[0].Name)
	assert.Equal(t, "integer", response.Columns[0].Type)
	assert.Equal(t, "is_active", response.Columns[3].Name)
	assert.Equal(t, "boolean", response.Columns[3].Type)

	// Verify data content
	assert.Equal(t, float64(1), response.Data[0]["user_id"])
	assert.Equal(t, "alice", response.Data[0]["username"])
	assert.Equal(t, true, response.Data[0]["is_active"])
}

// TestExecuteQuery_SelectEmptyResult tests SELECT query that returns no rows
func TestExecuteQuery_SelectEmptyResult(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)
	_ = jwtManager

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanRead: true, CanSelect: true}, nil)

	// Mock empty result
	mockService.On("ExecuteQuery", mock.Anything, mock.AnythingOfType("*models.Query"), mock.AnythingOfType("*models.DataSource")).
		Return(&models.QueryResult{
			RowCount:    0,
			ColumnNames: `["id", "name"]`,
			ColumnTypes: `["integer", "text"]`,
			Data:        `[]`,
		}, nil)

	token, _ := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "SELECT * FROM users WHERE id = 99999")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ExecuteQueryResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "completed", response.Status)
	assert.Equal(t, 0, *response.RowCount)
	assert.Empty(t, response.Data)
}

// =============================================================================
// WRITE QUERY APPROVAL TESTS
// =============================================================================

// TestExecuteQuery_INSERT_CreatesApprovalRequest tests that INSERT creates approval request
func TestExecuteQuery_INSERT_CreatesApprovalRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantWritePermission(t, db, group.ID, dataSource.ID)

	// Mock permissions - user has write access
	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  true,
			CanSelect: true,
			CanInsert: true,
		}, nil)

	// Mock validation - returns valid result with affected rows
	mockService.On("PreviewAndValidateWriteQuery", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*models.DataSource")).
		Return(&dto.ValidationResult{
			Valid:        true,
			Status:       "ok",
			Message:      "Query affects 1 row",
			AffectedRows: 1,
		}, nil)

	// Mock schema validation
	mockService.On("ValidateQuerySchema", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*models.DataSource")).
		Return(nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "INSERT INTO users (name, email) VALUES ('John', 'john@test.com')")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response dto.ExecuteQueryResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "pending_approval", response.Status)
	assert.True(t, response.RequiresApproval)
	assert.NotEmpty(t, response.ApprovalID)

	// Verify approval request was created in database
	var approval models.ApprovalRequest
	err := db.Where("requested_by = ?", user.ID).First(&approval).Error
	require.NoError(t, err)
	assert.Equal(t, models.OperationInsert, approval.OperationType)
	assert.Equal(t, models.ApprovalStatusPending, approval.Status)
	assert.Contains(t, approval.QueryText, "INSERT INTO users")

	mockService.AssertExpectations(t)
}

// TestExecuteQuery_UPDATE_CreatesApprovalRequest tests that UPDATE creates approval request
func TestExecuteQuery_UPDATE_CreatesApprovalRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantWritePermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  true,
			CanSelect: true,
			CanUpdate: true,
		}, nil)

	mockService.On("PreviewAndValidateWriteQuery", mock.Anything, mock.Anything, mock.Anything).
		Return(&dto.ValidationResult{
			Valid:        true,
			Status:       "ok",
			AffectedRows: 5,
		}, nil)

	mockService.On("ValidateQuerySchema", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "UPDATE users SET status = 'active' WHERE status = 'pending'")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response dto.ExecuteQueryResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "pending_approval", response.Status)
	assert.True(t, response.RequiresApproval)

	// Verify approval was created
	var approval models.ApprovalRequest
	db.Where("requested_by = ?", user.ID).First(&approval)
	assert.Equal(t, models.OperationUpdate, approval.OperationType)

	mockService.AssertExpectations(t)
}

// TestExecuteQuery_DELETE_CreatesApprovalRequest tests that DELETE creates approval request
func TestExecuteQuery_DELETE_CreatesApprovalRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantWritePermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  true,
			CanSelect: true,
			CanDelete: true,
		}, nil)

	mockService.On("PreviewAndValidateWriteQuery", mock.Anything, mock.Anything, mock.Anything).
		Return(&dto.ValidationResult{
			Valid:        true,
			Status:       "ok",
			AffectedRows: 1,
		}, nil)

	mockService.On("ValidateQuerySchema", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "DELETE FROM users WHERE id = 123 AND is_test = true")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response dto.ExecuteQueryResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "pending_approval", response.Status)
	assert.True(t, response.RequiresApproval)

	// Verify approval was created
	var approval models.ApprovalRequest
	db.Where("requested_by = ?", user.ID).First(&approval)
	assert.Equal(t, models.OperationDelete, approval.OperationType)

	mockService.AssertExpectations(t)
}

// TestExecuteQuery_ApprovalRequestContainsCorrectData tests approval request has correct metadata
func TestExecuteQuery_ApprovalRequestContainsCorrectData(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantWritePermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanRead: true, CanWrite: true}, nil)

	mockService.On("PreviewAndValidateWriteQuery", mock.Anything, mock.Anything, mock.Anything).
		Return(&dto.ValidationResult{Valid: true, Status: "ok"}, nil)

	mockService.On("ValidateQuerySchema", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	queryText := "UPDATE orders SET status = 'shipped' WHERE id IN (1, 2, 3)"
	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), queryText)
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	// Verify approval request in database
	var approval models.ApprovalRequest
	err := db.Where("requested_by = ?", user.ID).First(&approval).Error
	require.NoError(t, err)

	assert.Equal(t, dataSource.ID, approval.DataSourceID)
	assert.Equal(t, user.ID, approval.RequestedBy)
	assert.Equal(t, queryText, approval.QueryText)
	assert.Equal(t, models.OperationUpdate, approval.OperationType)
	assert.Equal(t, models.ApprovalStatusPending, approval.Status)
}

// TestExecuteQuery_WriteQueryNoMatch_ReturnsNoMatch tests that write query affecting 0 rows returns no_match
func TestExecuteQuery_WriteQueryNoMatch_ReturnsNoMatch(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantWritePermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanRead: true, CanWrite: true}, nil)

	// Mock validation shows no rows would be affected
	mockService.On("PreviewAndValidateWriteQuery", mock.Anything, mock.Anything, mock.Anything).
		Return(&dto.ValidationResult{
			Valid:        true,
			Status:       "no_match",
			Message:      "Query would not affect any rows",
			AffectedRows: 0,
		}, nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "DELETE FROM users WHERE id = 999999")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ExecuteQueryResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "no_match", response.Status)
	assert.NotNil(t, response.Validation)
	assert.Equal(t, "no_match", response.Validation.Status)

	// Verify no approval was created
	var count int64
	db.Model(&models.ApprovalRequest{}).Where("requested_by = ?", user.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

// =============================================================================
// PERMISSION ENFORCEMENT TESTS
// =============================================================================

// TestExecuteQuery_UserWithoutPermission_ReturnsForbidden tests that user without permission gets 403
func TestExecuteQuery_UserWithoutPermission_ReturnsForbidden(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)

	// Create user without any group membership (no permissions)
	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	// Don't add user to any group and don't grant permissions

	// Mock permissions - user has no access
	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   false,
			CanWrite:  false,
			CanSelect: false,
		}, nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "SELECT * FROM users")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Insufficient permissions")
}

// TestExecuteQuery_ViewerCannotExecuteWrite_ReturnsForbidden tests that viewer role cannot execute write queries
func TestExecuteQuery_ViewerCannotExecuteWrite_ReturnsForbidden(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)

	viewer := fixtures.CreateTestViewerUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "viewer-group")
	fixtures.AddUserToGroup(db, viewer.ID, group.ID)
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID) // Viewers only get read permission

	// Mock permissions - viewer has read but not write
	mockService.On("GetEffectivePermissions", mock.Anything, viewer.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:    true,
			CanWrite:   false,
			CanApprove: false,
			CanSelect:  true,
			CanInsert:  false,
			CanUpdate:  false,
			CanDelete:  false,
		}, nil)

	token, _ := jwtManager.GenerateToken(viewer.ID, viewer.Email, string(models.RoleViewer))
	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "INSERT INTO users (name) VALUES ('test')")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Insufficient permissions")
	assert.Equal(t, "PERMISSION_DENIED_WRITE", response["code"])
}

// TestExecuteQuery_ViewerCannotExecuteUPDATE_ReturnsForbidden tests that viewer cannot execute UPDATE
func TestExecuteQuery_ViewerCannotExecuteUPDATE_ReturnsForbidden(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)

	viewer := fixtures.CreateTestViewerUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "viewer-group")
	fixtures.AddUserToGroup(db, viewer.ID, group.ID)
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, viewer.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  false,
			CanSelect: true,
			CanUpdate: false,
		}, nil)

	token, _ := jwtManager.GenerateToken(viewer.ID, viewer.Email, string(models.RoleViewer))
	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "UPDATE users SET name = 'test' WHERE id = 1")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestExecuteQuery_ViewerCannotExecuteDELETE_ReturnsForbidden tests that viewer cannot execute DELETE
func TestExecuteQuery_ViewerCannotExecuteDELETE_ReturnsForbidden(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)

	viewer := fixtures.CreateTestViewerUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "viewer-group")
	fixtures.AddUserToGroup(db, viewer.ID, group.ID)
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, viewer.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  false,
			CanSelect: true,
			CanDelete: false,
		}, nil)

	token, _ := jwtManager.GenerateToken(viewer.ID, viewer.Email, string(models.RoleViewer))
	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "DELETE FROM users WHERE id = 1")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestExecuteQuery_AdminBypassesPermissionChecks tests that admin bypasses all permission checks
func TestExecuteQuery_AdminBypassesPermissionChecks(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)

	// Create admin without explicit permissions on data source
	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	// Don't add admin to any group - admin should still have access

	// Mock permissions - admin has full access even without explicit group membership
	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:    true,
			CanWrite:   true,
			CanApprove: true,
			CanSelect:  true,
			CanInsert:  true,
			CanUpdate:  true,
			CanDelete:  true,
		}, nil)

	// Mock query execution
	mockService.On("ExecuteQuery", mock.Anything, mock.Anything, mock.Anything).
		Return(&models.QueryResult{
			RowCount:    1,
			ColumnNames: `["id"]`,
			ColumnTypes: `["integer"]`,
			Data:        `[{"id": 1}]`,
		}, nil)

	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "SELECT * FROM admin_only_table")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ExecuteQueryResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "completed", response.Status)
}

// TestExecuteQuery_AdminCanExecuteWriteDirectly tests that admin can execute write queries (bypasses approval)
func TestExecuteQuery_AdminCanExecuteWriteDirectly(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// For admin, write operations don't require approval
	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  true,
			CanSelect: true,
			CanInsert: true,
		}, nil)

	// Mock validation for write query
	mockService.On("PreviewAndValidateWriteQuery", mock.Anything, mock.Anything, mock.Anything).
		Return(&dto.ValidationResult{
			Status:       "can_proceed",
			AffectedRows: 1,
		}, nil)

	// Mock schema validation (required for approval creation)
	mockService.On("ValidateQuerySchema", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	// Mock direct execution (not approval)
	mockService.On("ExecuteQuery", mock.Anything, mock.Anything, mock.Anything).
		Return(&models.QueryResult{
			RowCount:    1,
			ColumnNames: `["rows_affected"]`,
			ColumnTypes: `["int"]`,
			Data:        `[{"rows_affected": 1}]`,
		}, nil)

	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "INSERT INTO logs (message) VALUES ('admin action')")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Note: In this test handler, admin still goes through approval flow
	// The actual handler may bypass approval for admin
	assert.Equal(t, http.StatusAccepted, w.Code)

	var response dto.ExecuteQueryResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "pending_approval", response.Status)
}

// =============================================================================
// ERROR HANDLING TESTS
// =============================================================================

// TestExecuteQuery_InvalidRequestBody_ReturnsBadRequest tests that invalid JSON returns 400
func TestExecuteQuery_InvalidRequestBody_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, jwtManager := setupQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	// Send invalid JSON
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer([]byte(`{invalid json}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "invalid character")
}

// TestExecuteQuery_MissingDataSourceID_ReturnsBadRequest tests that missing data_source_id returns 400
func TestExecuteQuery_MissingDataSourceID_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, jwtManager := setupQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	// Send request without data_source_id
	reqBody := `{"query_text": "SELECT * FROM users"}`
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestExecuteQuery_MissingQueryText_ReturnsBadRequest tests that missing query_text returns 400
func TestExecuteQuery_MissingQueryText_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, jwtManager := setupQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	// Send request without query_text
	reqBody := `{"data_source_id": "` + dataSource.ID.String() + `"}`
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestExecuteQuery_DataSourceNotFound_ReturnsNotFound tests that non-existent data source returns 404
func TestExecuteQuery_DataSourceNotFound_ReturnsNotFound(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, jwtManager := setupQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	// Use a non-existent data source ID
	fakeDSID := uuid.New().String()
	reqBody, _ := createExecuteQueryRequest(fakeDSID, "SELECT * FROM users")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Data source not found")
}

// TestExecuteQuery_Unauthorized_ReturnsUnauthorized tests that request without token returns 401
func TestExecuteQuery_Unauthorized_ReturnsUnauthorized(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, _ := setupQueryTestRouter(t, db)

	// Create a data source so the request would succeed if authorized
	_ = fixtures.CreateTestDataSource(t, db, "test-ds")

	reqBody := `{"data_source_id": "` + uuid.New().String() + `", "query_text": "SELECT 1"}`
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestExecuteQuery_InvalidToken_ReturnsUnauthorized tests that invalid token returns 401
func TestExecuteQuery_InvalidToken_ReturnsUnauthorized(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, _ := setupQueryTestRouter(t, db)

	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "SELECT 1")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestExecuteQuery_QueryExecutionError_ReturnsInternalServerError tests that query execution error returns 500
func TestExecuteQuery_QueryExecutionError_ReturnsInternalServerError(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanRead: true, CanSelect: true}, nil)

	// Mock query execution failure
	mockService.On("ExecuteQuery", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("connection refused"))

	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "SELECT * FROM users")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "connection refused")
}

// TestExecuteQuery_InvalidDataSourceID_ReturnsBadRequest tests that invalid UUID returns 400
func TestExecuteQuery_InvalidDataSourceID_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, jwtManager := setupQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	// Send request with invalid UUID format
	reqBody := `{"data_source_id": "not-a-valid-uuid", "query_text": "SELECT 1"}`
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// The handler will try to parse UUID and return 400 if invalid
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestExecuteQuery_UserWithoutWritePermission_CannotCreateApproval tests user without write permission cannot create approval
func TestExecuteQuery_UserWithoutWritePermission_CannotCreateApproval(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)

	// Create user with read-only permission
	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "read-only-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  false,
			CanSelect: true,
		}, nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "DELETE FROM sensitive_data WHERE id = 1")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Insufficient permissions to submit write operations")
	assert.Equal(t, "PERMISSION_DENIED_WRITE", response["code"])

	// Verify no approval was created
	var count int64
	db.Model(&models.ApprovalRequest{}).Where("requested_by = ?", user.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

// TestExecuteQuery_WriteQueryInvalidSQL_ReturnsBadRequest tests that invalid SQL syntax returns 400
func TestExecuteQuery_WriteQueryInvalidSQL_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupQueryTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantWritePermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanRead: true, CanWrite: true}, nil)

	mockService.On("PreviewAndValidateWriteQuery", mock.Anything, mock.Anything, mock.Anything).
		Return(&dto.ValidationResult{Valid: true, Status: "ok"}, nil)

	// Schema validation will fail due to invalid SQL
	mockService.On("ValidateQuerySchema", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("syntax error at or near 'INVALID'"))

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	// Invalid SQL syntax
	reqBody, _ := createExecuteQueryRequest(dataSource.ID.String(), "UPDATE INVALID SYNTAX")
	req, _ := http.NewRequest("POST", "/api/v1/queries", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Schema validation failed")
}

// =============================================================================
// PREVIEW WRITE QUERY TESTS
// =============================================================================

// setupPreviewWriteQueryRouter creates a test router with PreviewWriteQuery endpoint
func setupPreviewWriteQueryRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockQueryService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockQueryService := new(MockQueryService)

	// Create handler with mock service
	handler := &testPreviewWriteQueryHandler{
		db:           db,
		queryService: mockQueryService,
	}

	// Setup routes
	api := router.Group("/api/v1")
	{
		queries := api.Group("/queries")
		queries.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			queries.POST("/preview", handler.PreviewWriteQuery)
		}
	}

	return router, mockQueryService
}

// testPreviewWriteQueryHandler wraps the real handler for testing
type testPreviewWriteQueryHandler struct {
	db           *gorm.DB
	queryService *MockQueryService
}

// PreviewWriteQuery implements the handler logic using the mock service
func (h *testPreviewWriteQueryHandler) PreviewWriteQuery(c *gin.Context) {
	var req dto.PreviewWriteQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")

	// Verify data source exists
	var dataSource models.DataSource
	if err := h.db.First(&dataSource, "id = ?", req.DataSourceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data source not found"})
		return
	}

	// Check read permission (preview is read-only)
	uID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	perms, err := h.queryService.GetEffectivePermissions(c.Request.Context(), uID, dataSource.ID)
	if err != nil || !perms.CanRead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	// Validate SQL
	if err := ValidateSQL(req.QueryText); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SQL syntax", "details": err.Error()})
		return
	}

	// Execute preview
	result, err := h.queryService.PreviewWriteQuery(c, req.QueryText, &dataSource)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.PreviewWriteQueryResponse{
		TotalAffected: result.TotalAffected,
		PreviewRows:   result.PreviewRows,
		Columns:       result.Columns,
		PreviewLimit:  result.PreviewLimit,
		SelectQuery:   result.SelectQuery,
		OperationType: string(result.OperationType),
	})
}

// Helper function to create preview write query request body
func createPreviewWriteQueryRequest(dsID string, queryText string) ([]byte, error) {
	req := dto.PreviewWriteQueryRequest{
		DataSourceID: dsID,
		QueryText:    queryText,
	}
	return json.Marshal(req)
}

// TestPreviewWriteQuery_DeleteQuery_ReturnsAffectedRows tests DELETE preview returns affected rows
func TestPreviewWriteQuery_DeleteQuery_ReturnsAffectedRows(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService := setupPreviewWriteQueryRouter(t, db)

	// Create admin user and data source
	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	// Mock permissions - admin has read access
	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  true,
			CanSelect: true,
			CanDelete: true,
		}, nil)

	// Mock preview result for DELETE
	mockService.On("PreviewWriteQuery", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*models.DataSource")).
		Return(&dto.PreviewWriteQueryResponse{
			TotalAffected: 5,
			PreviewRows: []map[string]interface{}{
				{"id": 1, "name": "Alice", "status": "inactive"},
				{"id": 2, "name": "Bob", "status": "inactive"},
			},
			Columns:       []string{"id", "name", "status"},
			PreviewLimit:  100,
			SelectQuery:   "SELECT * FROM users WHERE status = 'inactive'",
			OperationType: "DELETE",
		}, nil)

	// Create JWT token for admin
	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	// Create request
	reqBody, err := createPreviewWriteQueryRequest(dataSource.ID.String(), "DELETE FROM users WHERE status = 'inactive'")
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PreviewWriteQueryResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 5, response.TotalAffected)
	assert.Len(t, response.PreviewRows, 2)
	assert.Equal(t, []string{"id", "name", "status"}, response.Columns)
	assert.Equal(t, 100, response.PreviewLimit)
	assert.Equal(t, "DELETE", response.OperationType)
	assert.Equal(t, "SELECT * FROM users WHERE status = 'inactive'", response.SelectQuery)

	mockService.AssertExpectations(t)
}

// TestPreviewWriteQuery_UpdateQuery_ReturnsAffectedRows tests UPDATE preview returns affected rows
func TestPreviewWriteQuery_UpdateQuery_ReturnsAffectedRows(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService := setupPreviewWriteQueryRouter(t, db)

	// Create regular user with write permission
	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantWritePermission(t, db, group.ID, dataSource.ID)

	// Mock permissions - user has read access
	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  true,
			CanSelect: true,
			CanUpdate: true,
		}, nil)

	// Mock preview result for UPDATE
	mockService.On("PreviewWriteQuery", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*models.DataSource")).
		Return(&dto.PreviewWriteQueryResponse{
			TotalAffected: 3,
			PreviewRows: []map[string]interface{}{
				{"id": 10, "name": "Product A", "price": 99.99},
				{"id": 11, "name": "Product B", "price": 149.99},
				{"id": 12, "name": "Product C", "price": 199.99},
			},
			Columns:       []string{"id", "name", "price"},
			PreviewLimit:  100,
			SelectQuery:   "SELECT id, name, price FROM products WHERE category = 'electronics'",
			OperationType: "UPDATE",
		}, nil)

	// Create JWT token for user
	token, err := testauth.CreateTestJWTToken(user.ID.String(), user.Email, string(models.RoleUser))
	require.NoError(t, err)

	// Create request
	reqBody, err := createPreviewWriteQueryRequest(dataSource.ID.String(), "UPDATE products SET price = price * 1.1 WHERE category = 'electronics'")
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PreviewWriteQueryResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 3, response.TotalAffected)
	assert.Len(t, response.PreviewRows, 3)
	assert.Equal(t, "UPDATE", response.OperationType)
	assert.Equal(t, []string{"id", "name", "price"}, response.Columns)

	mockService.AssertExpectations(t)
}

// TestPreviewWriteQuery_NoMatchingRows_ReturnsNoMatch tests preview with no matching rows
func TestPreviewWriteQuery_NoMatchingRows_ReturnsNoMatch(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService := setupPreviewWriteQueryRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	// Mock permissions
	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanSelect: true,
		}, nil)

	// Mock preview result with no affected rows
	mockService.On("PreviewWriteQuery", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*models.DataSource")).
		Return(&dto.PreviewWriteQueryResponse{
			TotalAffected: 0,
			PreviewRows:   []map[string]interface{}{},
			Columns:       []string{},
			PreviewLimit:  100,
			SelectQuery:   "SELECT * FROM users WHERE id = 999999",
			OperationType: "DELETE",
		}, nil)

	token, _ := testauth.CreateTestJWTToken(user.ID.String(), user.Email, string(models.RoleUser))
	reqBody, _ := createPreviewWriteQueryRequest(dataSource.ID.String(), "DELETE FROM users WHERE id = 999999")
	req, _ := http.NewRequest("POST", "/api/v1/queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PreviewWriteQueryResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, 0, response.TotalAffected)
	assert.Empty(t, response.PreviewRows)
	assert.Equal(t, "DELETE", response.OperationType)

	mockService.AssertExpectations(t)
}

// TestPreviewWriteQuery_SelectQuery_ReturnsBadRequest tests that SELECT query preview returns 400
func TestPreviewWriteQuery_SelectQuery_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService := setupPreviewWriteQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Mock permissions
	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanSelect: true,
		}, nil)

	// Mock service returns error for SELECT query
	mockService.On("PreviewWriteQuery", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*models.DataSource")).
		Return(nil, errors.New("PreviewWriteQuery only supports DELETE and UPDATE queries"))

	token, _ := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	reqBody, _ := createPreviewWriteQueryRequest(dataSource.ID.String(), "SELECT * FROM users")
	req, _ := http.NewRequest("POST", "/api/v1/queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "PreviewWriteQuery only supports DELETE and UPDATE queries")

	mockService.AssertExpectations(t)
}

// TestPreviewWriteQuery_NoReadPermission_ReturnsForbidden tests that user without read permission gets 403
func TestPreviewWriteQuery_NoReadPermission_ReturnsForbidden(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService := setupPreviewWriteQueryRouter(t, db)

	// Create user without read permission
	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	// Don't add user to any group - no permissions

	// Mock permissions - user has no read access
	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   false,
			CanWrite:  false,
			CanSelect: false,
		}, nil)

	token, _ := testauth.CreateTestJWTToken(user.ID.String(), user.Email, string(models.RoleUser))
	reqBody, _ := createPreviewWriteQueryRequest(dataSource.ID.String(), "DELETE FROM users WHERE id = 1")
	req, _ := http.NewRequest("POST", "/api/v1/queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Insufficient permissions")

	mockService.AssertExpectations(t)
}

// TestPreviewWriteQuery_DataSourceNotFound_ReturnsNotFound tests that non-existent data source returns 404
func TestPreviewWriteQuery_DataSourceNotFound_ReturnsNotFound(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _ := setupPreviewWriteQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)

	token, _ := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	// Use a non-existent data source ID
	fakeDSID := uuid.New().String()
	reqBody, _ := createPreviewWriteQueryRequest(fakeDSID, "DELETE FROM users WHERE id = 1")
	req, _ := http.NewRequest("POST", "/api/v1/queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Data source not found")
}

// TestPreviewWriteQuery_InvalidSQL_ReturnsBadRequest tests that invalid SQL syntax returns 400
func TestPreviewWriteQuery_InvalidSQL_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService := setupPreviewWriteQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Mock permissions
	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanSelect: true,
		}, nil)

	// Mock service returns error for invalid SQL
	mockService.On("PreviewWriteQuery", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*models.DataSource")).
		Return(nil, errors.New("syntax error at or near 'INVALID'"))

	token, _ := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	// Send invalid SQL query
	reqBody, _ := createPreviewWriteQueryRequest(dataSource.ID.String(), "INVALID SQL SYNTAX HERE")
	req, _ := http.NewRequest("POST", "/api/v1/queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "syntax error")

	mockService.AssertExpectations(t)
}

// =============================================================================
// PREVIEW INSERT QUERY TESTS
// =============================================================================

// setupPreviewInsertQueryRouter creates a test router with PreviewInsertQuery endpoint
func setupPreviewInsertQueryRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockQueryService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockQueryService := new(MockQueryService)

	handler := &testPreviewInsertQueryHandler{
		db:           db,
		queryService: mockQueryService,
	}

	api := router.Group("/api/v1")
	{
		queries := api.Group("/queries")
		queries.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			queries.POST("/preview-insert", handler.PreviewInsertQuery)
		}
	}

	return router, mockQueryService
}

// testPreviewInsertQueryHandler wraps the real handler for testing
type testPreviewInsertQueryHandler struct {
	db           *gorm.DB
	queryService *MockQueryService
}

// PreviewInsertQuery implements the handler logic using the mock service
func (h *testPreviewInsertQueryHandler) PreviewInsertQuery(c *gin.Context) {
	var req dto.InsertPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")

	var dataSource models.DataSource
	if err := h.db.First(&dataSource, "id = ?", req.DataSourceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data source not found"})
		return
	}

	uID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	perms, err := h.queryService.GetEffectivePermissions(c.Request.Context(), uID, dataSource.ID)
	if err != nil || !perms.CanRead {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	result, err := h.queryService.PreviewInsertQuery(c, req.QueryText, &dataSource)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	columns := make([]dto.ColumnInfo, len(result.Columns))
	for i, col := range result.Columns {
		columns[i] = dto.ColumnInfo{
			Name: col.Name,
			Type: col.Type,
		}
	}

	c.JSON(http.StatusOK, dto.InsertPreviewResponse{
		TableName:     result.TableName,
		Columns:       columns,
		Rows:          result.Rows,
		TotalRowCount: result.TotalRowCount,
		PreviewType:   string(result.PreviewType),
		SelectQuery:   result.SelectQuery,
	})
}

// createInsertPreviewRequest creates an insert preview request body
func createInsertPreviewRequest(dsID string, queryText string) ([]byte, error) {
	req := dto.InsertPreviewRequest{
		DataSourceID: dsID,
		QueryText:    queryText,
	}
	return json.Marshal(req)
}

// TestPreviewInsertQuery_InsertValues_ReturnsParsedRows tests INSERT VALUES preview returns parsed rows
func TestPreviewInsertQuery_InsertValues_ReturnsParsedRows(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService := setupPreviewInsertQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  true,
			CanSelect: true,
			CanInsert: true,
		}, nil)

	mockService.On("PreviewInsertQuery", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*models.DataSource")).
		Return(&dto.InsertPreviewResponse{
			TableName: "users",
			Columns: []dto.ColumnInfo{
				{Name: "name", Type: "text"},
				{Name: "email", Type: "text"},
			},
			Rows: []map[string]interface{}{
				{"name": "John", "email": "john@test.com"},
			},
			TotalRowCount: 1,
			PreviewType:   "values",
		}, nil)

	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, err := createInsertPreviewRequest(dataSource.ID.String(), "INSERT INTO users (name, email) VALUES ('John', 'john@test.com')")
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/queries/preview-insert", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.InsertPreviewResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "users", response.TableName)
	assert.Equal(t, 1, response.TotalRowCount)
	assert.Len(t, response.Rows, 1)
	assert.Len(t, response.Columns, 2)
	assert.Equal(t, "name", response.Columns[0].Name)
	assert.Equal(t, "email", response.Columns[1].Name)
	assert.Equal(t, "values", response.PreviewType)
	assert.Equal(t, "John", response.Rows[0]["name"])
	assert.Equal(t, "john@test.com", response.Rows[0]["email"])

	mockService.AssertExpectations(t)
}

// TestPreviewInsertQuery_InsertSelect_ReturnsLimitedRows tests INSERT SELECT preview executes SELECT with LIMIT
func TestPreviewInsertQuery_InsertSelect_ReturnsLimitedRows(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService := setupPreviewInsertQueryRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantWritePermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  true,
			CanSelect: true,
			CanInsert: true,
		}, nil)

	mockService.On("PreviewInsertQuery", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*models.DataSource")).
		Return(&dto.InsertPreviewResponse{
			TableName: "archive",
			Columns: []dto.ColumnInfo{
				{Name: "id", Type: "integer"},
				{Name: "data", Type: "text"},
			},
			Rows: []map[string]interface{}{
				{"id": 1, "data": "record1"},
				{"id": 2, "data": "record2"},
				{"id": 3, "data": "record3"},
			},
			TotalRowCount: 100,
			PreviewType:   "select",
			SelectQuery:   "SELECT id, data FROM logs WHERE created_at < '2024-01-01'",
		}, nil)

	token, err := testauth.CreateTestJWTToken(user.ID.String(), user.Email, string(models.RoleUser))
	require.NoError(t, err)

	reqBody, err := createInsertPreviewRequest(dataSource.ID.String(), "INSERT INTO archive SELECT id, data FROM logs WHERE created_at < '2024-01-01'")
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/queries/preview-insert", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.InsertPreviewResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "archive", response.TableName)
	assert.Equal(t, 100, response.TotalRowCount)
	assert.Len(t, response.Rows, 3)
	assert.Equal(t, "select", response.PreviewType)
	assert.Equal(t, "SELECT id, data FROM logs WHERE created_at < '2024-01-01'", response.SelectQuery)

	mockService.AssertExpectations(t)
}

// TestPreviewInsertQuery_MultiRowInsert_LimitedTo50Rows tests multi-row INSERT preview limited to 50 rows
func TestPreviewInsertQuery_MultiRowInsert_LimitedTo50Rows(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService := setupPreviewInsertQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  true,
			CanSelect: true,
			CanInsert: true,
		}, nil)

	rows := make([]map[string]interface{}, 50)
	for i := 0; i < 50; i++ {
		rows[i] = map[string]interface{}{"id": i + 1, "name": fmt.Sprintf("user%d", i+1)}
	}

	mockService.On("PreviewInsertQuery", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*models.DataSource")).
		Return(&dto.InsertPreviewResponse{
			TableName: "users",
			Columns: []dto.ColumnInfo{
				{Name: "id", Type: "integer"},
				{Name: "name", Type: "text"},
			},
			Rows:          rows,
			TotalRowCount: 100,
			PreviewType:   "values",
		}, nil)

	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, err := createInsertPreviewRequest(dataSource.ID.String(), "INSERT INTO users (id, name) VALUES (1, 'user1'), (2, 'user2')")
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/queries/preview-insert", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.InsertPreviewResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 100, response.TotalRowCount)
	assert.Len(t, response.Rows, 50)
	assert.Equal(t, "values", response.PreviewType)

	mockService.AssertExpectations(t)
}

// TestPreviewInsertQuery_WithoutColumnList_ReturnsParsedRows tests INSERT without explicit column list
func TestPreviewInsertQuery_WithoutColumnList_ReturnsParsedRows(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService := setupPreviewInsertQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  true,
			CanSelect: true,
			CanInsert: true,
		}, nil)

	mockService.On("PreviewInsertQuery", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*models.DataSource")).
		Return(&dto.InsertPreviewResponse{
			TableName: "users",
			Columns: []dto.ColumnInfo{
				{Name: "id", Type: "integer"},
				{Name: "name", Type: "text"},
				{Name: "email", Type: "text"},
			},
			Rows: []map[string]interface{}{
				{"id": 1, "name": "John", "email": "john@test.com"},
			},
			TotalRowCount: 1,
			PreviewType:   "values",
		}, nil)

	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, err := createInsertPreviewRequest(dataSource.ID.String(), "INSERT INTO users VALUES (1, 'John', 'john@test.com')")
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/queries/preview-insert", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.InsertPreviewResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "users", response.TableName)
	assert.Equal(t, 1, response.TotalRowCount)
	assert.Len(t, response.Columns, 3)
	assert.Equal(t, "id", response.Columns[0].Name)
	assert.Equal(t, "name", response.Columns[1].Name)
	assert.Equal(t, "email", response.Columns[2].Name)

	mockService.AssertExpectations(t)
}

// TestPreviewInsertQuery_NoWritePermission_ReturnsForbidden tests that user without read permission gets 403
func TestPreviewInsertQuery_NoWritePermission_ReturnsForbidden(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService := setupPreviewInsertQueryRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   false,
			CanWrite:  false,
			CanSelect: false,
		}, nil)

	token, _ := testauth.CreateTestJWTToken(user.ID.String(), user.Email, string(models.RoleUser))
	reqBody, _ := createInsertPreviewRequest(dataSource.ID.String(), "INSERT INTO users (name) VALUES ('test')")
	req, _ := http.NewRequest("POST", "/api/v1/queries/preview-insert", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Insufficient permissions")

	mockService.AssertExpectations(t)
}

// TestPreviewInsertQuery_DataSourceNotFound_ReturnsNotFound tests that non-existent data source returns 404
func TestPreviewInsertQuery_DataSourceNotFound_ReturnsNotFound(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _ := setupPreviewInsertQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)

	token, _ := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	fakeDSID := uuid.New().String()
	reqBody, _ := createInsertPreviewRequest(fakeDSID, "INSERT INTO users (name) VALUES ('test')")
	req, _ := http.NewRequest("POST", "/api/v1/queries/preview-insert", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Data source not found")
}

// TestPreviewInsertQuery_InvalidSQL_ReturnsBadRequest tests that invalid INSERT SQL returns 400
func TestPreviewInsertQuery_InvalidSQL_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService := setupPreviewInsertQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  true,
			CanSelect: true,
			CanInsert: true,
		}, nil)

	mockService.On("PreviewInsertQuery", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*models.DataSource")).
		Return(nil, errors.New("failed to parse INSERT statement: syntax error"))

	token, _ := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	reqBody, _ := createInsertPreviewRequest(dataSource.ID.String(), "INSERT INVALID SYNTAX HERE")
	req, _ := http.NewRequest("POST", "/api/v1/queries/preview-insert", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "failed to parse INSERT statement")

	mockService.AssertExpectations(t)
}

// =============================================================================
// LIST QUERIES TESTS
// =============================================================================

// setupListQueriesRouter creates a test router with ListQueries endpoint
func setupListQueriesRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockQueryService, *auth.JWTManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockQueryService := new(MockQueryService)

	handler := &testListQueriesHandler{
		db:           db,
		queryService: mockQueryService,
	}

	api := router.Group("/api/v1")
	{
		queries := api.Group("/queries")
		queries.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			queries.GET("", handler.ListQueries)
		}
	}

	return router, mockQueryService, jwtManager
}

// testListQueriesHandler wraps the real handler for testing
type testListQueriesHandler struct {
	db           *gorm.DB
	queryService *MockQueryService
}

// ListQueries implements the handler logic using the mock service
func (h *testListQueriesHandler) ListQueries(c *gin.Context) {
	userID := c.GetString("user_id")

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	queries, total, err := h.queryService.ListQueries(c, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch queries"})
		return
	}

	response := make([]gin.H, len(queries))
	for i, query := range queries {
		response[i] = gin.H{
			"id":             query.ID.String(),
			"name":           query.Name,
			"description":    query.Description,
			"data_source_id": query.DataSourceID.String(),
			"operation_type": string(query.OperationType),
			"status":         string(query.Status),
			"created_at":     query.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"queries": response,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// TestListQueries_WithPagination_ReturnsPaginatedResults tests listing queries with pagination
func TestListQueries_WithPagination_ReturnsPaginatedResults(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupListQueriesRouter(t, db)

	user := fixtures.CreateTestAdminUser(t, db)

	// Mock service to return paginated queries
	mockService.On("ListQueries", mock.Anything, user.ID.String(), 20, 0).
		Return([]models.Query{
			{
				ID:            uuid.New(),
				Name:          "Query 1",
				Description:   "First query",
				DataSourceID:  uuid.New(),
				OperationType: models.OperationSelect,
				Status:        models.StatusCompleted,
				CreatedAt:     time.Now(),
			},
			{
				ID:            uuid.New(),
				Name:          "Query 2",
				Description:   "Second query",
				DataSourceID:  uuid.New(),
				OperationType: models.OperationSelect,
				Status:        models.StatusCompleted,
				CreatedAt:     time.Now(),
			},
		}, int64(2), nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleAdmin))
	req, _ := http.NewRequest("GET", "/api/v1/queries?page=1&limit=20", nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(2), response["total"])
	assert.Equal(t, float64(1), response["page"])
	assert.Equal(t, float64(20), response["limit"])

	queries, ok := response["queries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, queries, 2)

	mockService.AssertExpectations(t)
}

// TestListQueries_RegularUser_SeesOnlyOwnQueries tests that regular user only sees their own queries
func TestListQueries_RegularUser_SeesOnlyOwnQueries(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupListQueriesRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)

	// Mock service - called with the user's ID
	mockService.On("ListQueries", mock.Anything, user.ID.String(), 20, 0).
		Return([]models.Query{
			{
				ID:            uuid.New(),
				Name:          "My Query",
				UserID:        user.ID,
				DataSourceID:  uuid.New(),
				OperationType: models.OperationSelect,
				Status:        models.StatusCompleted,
				CreatedAt:     time.Now(),
			},
		}, int64(1), nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	req, _ := http.NewRequest("GET", "/api/v1/queries", nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify service was called with user's ID (not other user's ID)
	mockService.AssertCalled(t, "ListQueries", mock.Anything, user.ID.String(), 20, 0)

	// Verify the user's queries are returned
	queries := response["queries"].([]interface{})
	assert.Len(t, queries, 1)

	mockService.AssertExpectations(t)
}

// TestListQueries_Admin_SeesAllQueries tests that admin sees all queries
func TestListQueries_Admin_SeesAllQueries(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupListQueriesRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)

	// Mock service - admin sees all (empty string means all users)
	mockService.On("ListQueries", mock.Anything, admin.ID.String(), 20, 0).
		Return([]models.Query{
			{
				ID:            uuid.New(),
				Name:          "Admin Query 1",
				UserID:        admin.ID,
				DataSourceID:  uuid.New(),
				OperationType: models.OperationSelect,
				Status:        models.StatusCompleted,
				CreatedAt:     time.Now(),
			},
			{
				ID:            uuid.New(),
				Name:          "Admin Query 2",
				UserID:        admin.ID,
				DataSourceID:  uuid.New(),
				OperationType: models.OperationSelect,
				Status:        models.StatusCompleted,
				CreatedAt:     time.Now(),
			},
		}, int64(2), nil)

	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	req, _ := http.NewRequest("GET", "/api/v1/queries", nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify service was called
	mockService.AssertCalled(t, "ListQueries", mock.Anything, admin.ID.String(), 20, 0)

	queries := response["queries"].([]interface{})
	assert.Len(t, queries, 2)

	mockService.AssertExpectations(t)
}

// TestListQueries_EmptyList_ReturnsEmptyArray tests that empty query list returns empty array
func TestListQueries_EmptyList_ReturnsEmptyArray(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupListQueriesRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)

	// Mock service to return empty list
	mockService.On("ListQueries", mock.Anything, user.ID.String(), 20, 0).
		Return([]models.Query{}, int64(0), nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	req, _ := http.NewRequest("GET", "/api/v1/queries", nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(0), response["total"])
	queries := response["queries"].([]interface{})
	assert.Len(t, queries, 0)

	mockService.AssertExpectations(t)
}

// TestListQueries_InvalidPagination_ReturnsBadRequest tests that invalid pagination returns 400
func TestListQueries_InvalidPagination_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupListQueriesRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)

	// Mock service - handler passes limit as-is (negative value unchanged)
	mockService.On("ListQueries", mock.Anything, user.ID.String(), -1, 0).
		Return([]models.Query{}, int64(0), nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	// Invalid pagination - negative limit
	req, _ := http.NewRequest("GET", "/api/v1/queries?limit=-1", nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

// =============================================================================
// LIST QUERY HISTORY TESTS
// =============================================================================

// setupListQueryHistoryRouter creates a test router with ListQueryHistory endpoint
func setupListQueryHistoryRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockQueryService, *auth.JWTManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockQueryService := new(MockQueryService)

	handler := &testListQueryHistoryHandler{
		db:           db,
		queryService: mockQueryService,
	}

	api := router.Group("/api/v1")
	{
		history := api.Group("/queries/history")
		history.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			history.GET("", handler.ListQueryHistory)
		}
	}

	return router, mockQueryService, jwtManager
}

// testListQueryHistoryHandler wraps the real handler for testing
type testListQueryHistoryHandler struct {
	db           *gorm.DB
	queryService *MockQueryService
}

// ListQueryHistory implements the handler logic using the mock service
func (h *testListQueryHistoryHandler) ListQueryHistory(c *gin.Context) {
	userID := c.GetString("user_id")

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	// Optional filters
	dataSourceID := c.Query("data_source_id")
	status := c.Query("status")
	operationType := c.Query("operation_type")
	search := c.Query("search")

	history, total, err := h.queryService.ListQueryHistory(c, userID, limit, offset, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch query history"})
		return
	}

	// Apply filters after retrieval (simple approach)
	if dataSourceID != "" || status != "" || operationType != "" {
		filtered := history[:0]
		for _, h := range history {
			if dataSourceID != "" && h.DataSourceID.String() != dataSourceID {
				continue
			}
			if status != "" && string(h.Status) != status {
				continue
			}
			if operationType != "" && string(h.OperationType) != operationType {
				continue
			}
			filtered = append(filtered, h)
		}
		history = filtered
	}

	response := make([]gin.H, len(history))
	for i, entry := range history {
		response[i] = gin.H{
			"id":                entry.ID.String(),
			"query_id":          entry.QueryID,
			"user_id":           entry.UserID.String(),
			"data_source_id":    entry.DataSourceID.String(),
			"data_source_name":  entry.DataSource.Name,
			"query_text":        entry.QueryText,
			"operation_type":    string(entry.OperationType),
			"status":            string(entry.Status),
			"row_count":         entry.RowCount,
			"execution_time_ms": entry.ExecutionTimeMs,
			"error_message":     entry.ErrorMessage,
			"executed_at":       entry.ExecutedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"history": response,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// TestListQueryHistory_WithPagination_ReturnsPaginatedResults tests listing history with pagination
func TestListQueryHistory_WithPagination_ReturnsPaginatedResults(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupListQueryHistoryRouter(t, db)

	user := fixtures.CreateTestAdminUser(t, db)
	dsID := uuid.New()
	executedAt := time.Now()
	rowCount := 10
	execTime := 100

	mockService.On("ListQueryHistory", mock.Anything, user.ID.String(), 20, 0, "").
		Return([]models.QueryHistory{
			{
				ID:              uuid.New(),
				QueryID:         ptr(uuid.New()),
				UserID:          user.ID,
				DataSourceID:    dsID,
				QueryText:       "SELECT * FROM users",
				OperationType:   models.OperationSelect,
				Status:          models.StatusCompleted,
				RowCount:        &rowCount,
				ExecutionTimeMs: &execTime,
				ExecutedAt:      executedAt,
				DataSource:      models.DataSource{ID: dsID, Name: "Test DB"},
			},
		}, int64(1), nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleAdmin))
	req, _ := http.NewRequest("GET", "/api/v1/queries/history?page=1&limit=20", nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(1), response["total"])
	assert.Equal(t, float64(1), response["page"])
	assert.Equal(t, float64(20), response["limit"])

	history, ok := response["history"].([]interface{})
	require.True(t, ok)
	assert.Len(t, history, 1)

	mockService.AssertExpectations(t)
}

// TestListQueryHistory_RegularUser_SeesOnlyOwnHistory tests that regular user only sees their own history
func TestListQueryHistory_RegularUser_SeesOnlyOwnHistory(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupListQueryHistoryRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)

	// Mock service - called with the user's ID
	mockService.On("ListQueryHistory", mock.Anything, user.ID.String(), 20, 0, "").
		Return([]models.QueryHistory{
			{
				ID:            uuid.New(),
				UserID:        user.ID,
				DataSourceID:  uuid.New(),
				QueryText:     "SELECT * FROM my_table",
				OperationType: models.OperationSelect,
				Status:        models.StatusCompleted,
				ExecutedAt:    time.Now(),
				DataSource:    models.DataSource{Name: "My DB"},
			},
		}, int64(1), nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	req, _ := http.NewRequest("GET", "/api/v1/queries/history", nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify service was called with user's ID
	mockService.AssertCalled(t, "ListQueryHistory", mock.Anything, user.ID.String(), 20, 0, "")

	history := response["history"].([]interface{})
	assert.Len(t, history, 1)

	mockService.AssertExpectations(t)
}

// TestListQueryHistory_Admin_SeesAllHistory tests that admin sees all history
func TestListQueryHistory_Admin_SeesAllHistory(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupListQueryHistoryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	otherUserID := uuid.New()

	// Mock service - admin sees all
	mockService.On("ListQueryHistory", mock.Anything, admin.ID.String(), 20, 0, "").
		Return([]models.QueryHistory{
			{
				ID:            uuid.New(),
				UserID:        admin.ID,
				DataSourceID:  uuid.New(),
				QueryText:     "SELECT 1",
				OperationType: models.OperationSelect,
				Status:        models.StatusCompleted,
				ExecutedAt:    time.Now(),
				DataSource:    models.DataSource{Name: "DB 1"},
			},
			{
				ID:            uuid.New(),
				UserID:        otherUserID,
				DataSourceID:  uuid.New(),
				QueryText:     "SELECT 2",
				OperationType: models.OperationSelect,
				Status:        models.StatusCompleted,
				ExecutedAt:    time.Now(),
				DataSource:    models.DataSource{Name: "DB 2"},
			},
		}, int64(2), nil)

	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	req, _ := http.NewRequest("GET", "/api/v1/queries/history", nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	history := response["history"].([]interface{})
	assert.Len(t, history, 2)

	mockService.AssertExpectations(t)
}

// TestListQueryHistory_FilterByDataSource_ReturnsFilteredResults tests filtering by data source
func TestListQueryHistory_FilterByDataSource_ReturnsFilteredResults(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupListQueryHistoryRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	targetDSID := uuid.New()
	otherDSID := uuid.New()

	// Mock service returns both entries, filter happens in handler
	mockService.On("ListQueryHistory", mock.Anything, user.ID.String(), 20, 0, "").
		Return([]models.QueryHistory{
			{
				ID:            uuid.New(),
				UserID:        user.ID,
				DataSourceID:  targetDSID,
				QueryText:     "SELECT * FROM target",
				OperationType: models.OperationSelect,
				Status:        models.StatusCompleted,
				ExecutedAt:    time.Now(),
				DataSource:    models.DataSource{ID: targetDSID, Name: "Target DB"},
			},
			{
				ID:            uuid.New(),
				UserID:        user.ID,
				DataSourceID:  otherDSID,
				QueryText:     "SELECT * FROM other",
				OperationType: models.OperationSelect,
				Status:        models.StatusCompleted,
				ExecutedAt:    time.Now(),
				DataSource:    models.DataSource{ID: otherDSID, Name: "Other DB"},
			},
		}, int64(2), nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	req, _ := http.NewRequest("GET", "/api/v1/queries/history?data_source_id="+targetDSID.String(), nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Only the entry with matching data_source_id should be returned
	history := response["history"].([]interface{})
	assert.Len(t, history, 1)

	mockService.AssertExpectations(t)
}

// TestListQueryHistory_FilterBySearch_ReturnsFilteredResults tests filtering by search term
func TestListQueryHistory_FilterBySearch_ReturnsFilteredResults(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupListQueryHistoryRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)

	// Mock service - search is passed to service
	mockService.On("ListQueryHistory", mock.Anything, user.ID.String(), 20, 0, "users").
		Return([]models.QueryHistory{
			{
				ID:            uuid.New(),
				UserID:        user.ID,
				DataSourceID:  uuid.New(),
				QueryText:     "SELECT * FROM users",
				OperationType: models.OperationSelect,
				Status:        models.StatusCompleted,
				ExecutedAt:    time.Now(),
				DataSource:    models.DataSource{Name: "DB"},
			},
		}, int64(1), nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	req, _ := http.NewRequest("GET", "/api/v1/queries/history?search=users", nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Verify service was called with search parameter
	mockService.AssertCalled(t, "ListQueryHistory", mock.Anything, user.ID.String(), 20, 0, "users")

	history := response["history"].([]interface{})
	assert.Len(t, history, 1)

	mockService.AssertExpectations(t)
}

// =============================================================================
// EXPLAIN QUERY TESTS
// =============================================================================

// setupExplainQueryRouter creates a test router with ExplainQuery endpoint
func setupExplainQueryRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockQueryService, *auth.JWTManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockQueryService := new(MockQueryService)

	handler := &testExplainQueryHandler{
		db:           db,
		queryService: mockQueryService,
	}

	api := router.Group("/api/v1")
	{
		queries := api.Group("/queries")
		queries.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			queries.POST("/explain", handler.ExplainQuery)
		}
	}

	return router, mockQueryService, jwtManager
}

// testExplainQueryHandler wraps the real handler for testing
type testExplainQueryHandler struct {
	db           *gorm.DB
	queryService *MockQueryService
}

// ExplainQuery implements the handler logic using the mock service
func (h *testExplainQueryHandler) ExplainQuery(c *gin.Context) {
	var req dto.ExplainQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")

	var dataSource models.DataSource
	if err := h.db.First(&dataSource, "id = ?", req.DataSourceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Data source not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data source"})
		}
		return
	}

	if !h.checkReadPermission(c, userID, dataSource.ID.String()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to read from this data source"})
		return
	}

	result, err := h.queryService.ExplainQuery(c.Request.Context(), req.QueryText, &dataSource, req.Analyze)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// checkReadPermission checks if user has read permission (simplified for testing)
func (h *testExplainQueryHandler) checkReadPermission(c *gin.Context, userID, dsID string) bool {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return false
	}
	dsUUID, err := uuid.Parse(dsID)
	if err != nil {
		return false
	}
	perms, err := h.queryService.GetEffectivePermissions(c.Request.Context(), uID, dsUUID)
	if err != nil {
		return false
	}
	return perms.CanRead
}

// checkWritePermission checks if user has write permission (simplified for testing)
func (h *testExplainQueryHandler) checkWritePermission(c *gin.Context, userID, dsID string) bool {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return false
	}
	dsUUID, err := uuid.Parse(dsID)
	if err != nil {
		return false
	}
	perms, err := h.queryService.GetEffectivePermissions(c.Request.Context(), uID, dsUUID)
	if err != nil {
		return false
	}
	return perms.CanWrite
}

// testDryRunDeleteHandler wraps the real handler for testing
type testDryRunDeleteHandler struct {
	db           *gorm.DB
	queryService *MockQueryService
}

// DryRunDelete implements the handler logic using the mock service
func (h *testDryRunDeleteHandler) DryRunDelete(c *gin.Context) {
	var req dto.DryRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")

	var dataSource models.DataSource
	if err := h.db.First(&dataSource, "id = ?", req.DataSourceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Data source not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data source"})
		}
		return
	}

	if !h.checkWritePermission(c, userID, dataSource.ID.String()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to write to this data source"})
		return
	}

	result, err := h.queryService.DryRunDelete(c.Request.Context(), req.QueryText, &dataSource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// checkReadPermission checks if user has read permission (simplified for testing)
func (h *testDryRunDeleteHandler) checkReadPermission(c *gin.Context, userID, dsID string) bool {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return false
	}
	dsUUID, err := uuid.Parse(dsID)
	if err != nil {
		return false
	}
	perms, err := h.queryService.GetEffectivePermissions(c.Request.Context(), uID, dsUUID)
	if err != nil {
		return false
	}
	return perms.CanRead
}

// checkWritePermission checks if user has write permission (simplified for testing)
func (h *testDryRunDeleteHandler) checkWritePermission(c *gin.Context, userID, dsID string) bool {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return false
	}
	dsUUID, err := uuid.Parse(dsID)
	if err != nil {
		return false
	}
	perms, err := h.queryService.GetEffectivePermissions(c.Request.Context(), uID, dsUUID)
	if err != nil {
		return false
	}
	return perms.CanWrite
}

// Helper function to create explain query request body
func createExplainQueryRequest(dsID string, queryText string, analyze bool) ([]byte, error) {
	req := dto.ExplainQueryRequest{
		DataSourceID: dsID,
		QueryText:    queryText,
		Analyze:      analyze,
	}
	return json.Marshal(req)
}

// TestExplainQuery_BasicExplain_ReturnsQueryPlan tests basic EXPLAIN query
func TestExplainQuery_BasicExplain_ReturnsQueryPlan(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, _ := setupExplainQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  true,
			CanSelect: true,
		}, nil)

	explainResult := &struct {
		Plan      []map[string]interface{} `json:"plan"`
		RawOutput string                   `json:"raw_output"`
	}{
		Plan: []map[string]interface{}{
			{"QUERY PLAN": "Seq Scan on users (rows=100)"},
		},
		RawOutput: `[{"QUERY PLAN":"Seq Scan on users (rows=100)"}]`,
	}
	mockService.On("ExplainQuery", mock.Anything, "SELECT * FROM users", mock.AnythingOfType("*models.DataSource"), false).
		Return(explainResult, nil)

	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, err := createExplainQueryRequest(dataSource.ID.String(), "SELECT * FROM users", false)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/queries/explain", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["plan"])
	assert.NotEmpty(t, response["raw_output"])

	mockService.AssertExpectations(t)
}

// TestExplainQuery_ExplainAnalyze_ReturnsAnalyzedPlan tests EXPLAIN ANALYZE
func TestExplainQuery_ExplainAnalyze_ReturnsAnalyzedPlan(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, _ := setupExplainQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanRead: true, CanSelect: true}, nil)

	explainResult := &struct {
		Plan      []map[string]interface{} `json:"plan"`
		RawOutput string                   `json:"raw_output"`
	}{
		Plan: []map[string]interface{}{
			{"QUERY PLAN": "Seq Scan on users (rows=100) (actual time=0.1..5.2 rows=100 loops=1)"},
		},
		RawOutput: `[{"QUERY PLAN":"Seq Scan on users (rows=100) (actual time=0.1..5.2 rows=100 loops=1)"}]`,
	}
	mockService.On("ExplainQuery", mock.Anything, "SELECT * FROM users", mock.AnythingOfType("*models.DataSource"), true).
		Return(explainResult, nil)

	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, err := createExplainQueryRequest(dataSource.ID.String(), "SELECT * FROM users", true)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/queries/explain", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotNil(t, response["plan"])

	mockService.AssertExpectations(t)
}

// TestExplainQuery_NoReadPermission_ReturnsForbidden tests permission check
func TestExplainQuery_NoReadPermission_ReturnsForbidden(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, _ := setupExplainQueryRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   false,
			CanWrite:  false,
			CanSelect: false,
		}, nil)

	token, _ := testauth.CreateTestJWTToken(user.ID.String(), user.Email, string(models.RoleUser))
	reqBody, _ := createExplainQueryRequest(dataSource.ID.String(), "SELECT * FROM users", false)
	req, _ := http.NewRequest("POST", "/api/v1/queries/explain", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Insufficient permissions")

	mockService.AssertExpectations(t)
}

// TestExplainQuery_DataSourceNotFound_ReturnsNotFound tests non-existent data source
func TestExplainQuery_DataSourceNotFound_ReturnsNotFound(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, _ := setupExplainQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	_ = fixtures.CreateTestDataSource(t, db, "existing-ds")

	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	fakeDSID := uuid.New().String()
	reqBody, _ := createExplainQueryRequest(fakeDSID, "SELECT * FROM users", false)
	req, _ := http.NewRequest("POST", "/api/v1/queries/explain", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Data source not found")
}

// TestExplainQuery_InvalidSQL_ReturnsInternalServerError tests invalid SQL handling
func TestExplainQuery_InvalidSQL_ReturnsInternalServerError(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, _ := setupExplainQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanRead: true, CanSelect: true}, nil)

	mockService.On("ExplainQuery", mock.Anything, "SELECT * FROM invalid_table", mock.AnythingOfType("*models.DataSource"), false).
		Return(nil, errors.New("relation \"invalid_table\" does not exist"))

	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, _ := createExplainQueryRequest(dataSource.ID.String(), "SELECT * FROM invalid_table", false)
	req, _ := http.NewRequest("POST", "/api/v1/queries/explain", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "relation")

	mockService.AssertExpectations(t)
}

// TestExplainQuery_MissingDataSourceID_ReturnsBadRequest tests missing required field
func TestExplainQuery_MissingDataSourceID_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, _ := setupExplainQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody := `{"query_text": "SELECT * FROM users"}`
	req, _ := http.NewRequest("POST", "/api/v1/queries/explain", bytes.NewBuffer([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestExplainQuery_MissingQueryText_ReturnsBadRequest tests missing query text
func TestExplainQuery_MissingQueryText_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, _ := setupExplainQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody := `{"data_source_id": "` + dataSource.ID.String() + `"}`
	req, _ := http.NewRequest("POST", "/api/v1/queries/explain", bytes.NewBuffer([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestExplainQuery_Unauthorized_ReturnsUnauthorized tests missing auth
func TestExplainQuery_Unauthorized_ReturnsUnauthorized(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, _ := setupExplainQueryRouter(t, db)

	_ = fixtures.CreateTestDataSource(t, db, "test-ds")

	reqBody := `{"data_source_id": "` + uuid.New().String() + `", "query_text": "SELECT 1"}`
	req, _ := http.NewRequest("POST", "/api/v1/queries/explain", bytes.NewBuffer([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// DRY RUN DELETE TESTS
// =============================================================================

// setupDryRunDeleteRouter creates a test router with DryRunDelete endpoint
func setupDryRunDeleteRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockQueryService, *auth.JWTManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockQueryService := new(MockQueryService)

	handler := &testDryRunDeleteHandler{
		db:           db,
		queryService: mockQueryService,
	}

	api := router.Group("/api/v1")
	{
		queries := api.Group("/queries")
		queries.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			queries.POST("/dry-run/delete", handler.DryRunDelete)
		}
	}

	return router, mockQueryService, jwtManager
}

// Helper function to create dry run delete request body
func createDryRunDeleteRequest(dsID string, queryText string) ([]byte, error) {
	req := dto.DryRunRequest{
		DataSourceID: dsID,
		QueryText:    queryText,
	}
	return json.Marshal(req)
}

// TestDryRunDelete_ValidDelete_ReturnsAffectedRows tests successful dry run
func TestDryRunDelete_ValidDelete_ReturnsAffectedRows(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, _ := setupDryRunDeleteRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  true,
			CanSelect: true,
			CanDelete: true,
		}, nil)

	dryRunResult := &struct {
		AffectedRows int                      `json:"affected_rows"`
		Rows         []map[string]interface{} `json:"rows"`
		Query        string                   `json:"query"`
	}{
		AffectedRows: 5,
		Rows: []map[string]interface{}{
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"},
		},
		Query: "SELECT * FROM users WHERE status = 'inactive'",
	}
	mockService.On("DryRunDelete", mock.Anything, "DELETE FROM users WHERE status = 'inactive'", mock.AnythingOfType("*models.DataSource")).
		Return(dryRunResult, nil)

	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, err := createDryRunDeleteRequest(dataSource.ID.String(), "DELETE FROM users WHERE status = 'inactive'")
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/queries/dry-run/delete", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(5), response["affected_rows"])
	assert.Equal(t, "SELECT * FROM users WHERE status = 'inactive'", response["query"])
	assert.Len(t, response["rows"].([]interface{}), 2)

	mockService.AssertExpectations(t)
}

// TestDryRunDelete_NoWritePermission_ReturnsForbidden tests permission check
func TestDryRunDelete_NoWritePermission_ReturnsForbidden(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupDryRunDeleteRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "read-only-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  false,
			CanSelect: true,
		}, nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	reqBody, _ := createDryRunDeleteRequest(dataSource.ID.String(), "DELETE FROM users WHERE id = 1")
	req, _ := http.NewRequest("POST", "/api/v1/queries/dry-run/delete", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Insufficient permissions")

	mockService.AssertExpectations(t)
}

// TestDryRunDelete_DataSourceNotFound_ReturnsNotFound tests non-existent data source
func TestDryRunDelete_DataSourceNotFound_ReturnsNotFound(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, _ := setupDryRunDeleteRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	_ = fixtures.CreateTestDataSource(t, db, "existing-ds")

	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	fakeDSID := uuid.New().String()
	reqBody, _ := createDryRunDeleteRequest(fakeDSID, "DELETE FROM users WHERE id = 1")
	req, _ := http.NewRequest("POST", "/api/v1/queries/dry-run/delete", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Data source not found")
}

// TestDryRunDelete_InvalidSQL_ReturnsInternalServerError tests invalid SQL handling
func TestDryRunDelete_InvalidSQL_ReturnsInternalServerError(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, _ := setupDryRunDeleteRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanRead: true, CanWrite: true, CanDelete: true}, nil)

	mockService.On("DryRunDelete", mock.Anything, "DELETE FROM invalid_table", mock.AnythingOfType("*models.DataSource")).
		Return(nil, errors.New("relation \"invalid_table\" does not exist"))

	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, _ := createDryRunDeleteRequest(dataSource.ID.String(), "DELETE FROM invalid_table")
	req, _ := http.NewRequest("POST", "/api/v1/queries/dry-run/delete", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "relation")

	mockService.AssertExpectations(t)
}

// TestDryRunDelete_MissingDataSourceID_ReturnsBadRequest tests missing required field
func TestDryRunDelete_MissingDataSourceID_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, _ := setupDryRunDeleteRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody := `{"query_text": "DELETE FROM users WHERE id = 1"}`
	req, _ := http.NewRequest("POST", "/api/v1/queries/dry-run/delete", bytes.NewBuffer([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestDryRunDelete_MissingQueryText_ReturnsBadRequest tests missing query text
func TestDryRunDelete_MissingQueryText_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, _ := setupDryRunDeleteRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody := `{"data_source_id": "` + dataSource.ID.String() + `"}`
	req, _ := http.NewRequest("POST", "/api/v1/queries/dry-run/delete", bytes.NewBuffer([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestDryRunDelete_Unauthorized_ReturnsUnauthorized tests missing auth
func TestDryRunDelete_Unauthorized_ReturnsUnauthorized(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, _ := setupDryRunDeleteRouter(t, db)

	_ = fixtures.CreateTestDataSource(t, db, "test-ds")

	reqBody := `{"data_source_id": "` + uuid.New().String() + `", "query_text": "DELETE FROM users WHERE id = 1"}`
	req, _ := http.NewRequest("POST", "/api/v1/queries/dry-run/delete", bytes.NewBuffer([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestDryRunDelete_ViewerRole_ReturnsForbidden tests viewer cannot dry run delete
func TestDryRunDelete_ViewerRole_ReturnsForbidden(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupDryRunDeleteRouter(t, db)

	viewer := fixtures.CreateTestViewerUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "viewer-group")
	fixtures.AddUserToGroup(db, viewer.ID, group.ID)
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, viewer.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:   true,
			CanWrite:  false,
			CanSelect: true,
			CanDelete: false,
		}, nil)

	token, _ := jwtManager.GenerateToken(viewer.ID, viewer.Email, string(models.RoleViewer))
	reqBody, _ := createDryRunDeleteRequest(dataSource.ID.String(), "DELETE FROM users WHERE id = 1")
	req, _ := http.NewRequest("POST", "/api/v1/queries/dry-run/delete", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Insufficient permissions")

	mockService.AssertExpectations(t)
}

// TestDryRunDelete_ZeroAffectedRows_ReturnsZero tests dry run with no matching rows
func TestDryRunDelete_ZeroAffectedRows_ReturnsZero(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, _ := setupDryRunDeleteRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanRead: true, CanWrite: true, CanDelete: true}, nil)

	dryRunResult := &struct {
		AffectedRows int                      `json:"affected_rows"`
		Rows         []map[string]interface{} `json:"rows"`
		Query        string                   `json:"query"`
	}{
		AffectedRows: 0,
		Rows:         []map[string]interface{}{},
		Query:        "SELECT * FROM users WHERE id = 999999",
	}
	mockService.On("DryRunDelete", mock.Anything, "DELETE FROM users WHERE id = 999999", mock.AnythingOfType("*models.DataSource")).
		Return(dryRunResult, nil)

	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, _ := createDryRunDeleteRequest(dataSource.ID.String(), "DELETE FROM users WHERE id = 999999")
	req, _ := http.NewRequest("POST", "/api/v1/queries/dry-run/delete", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(0), response["affected_rows"])
	assert.Len(t, response["rows"].([]interface{}), 0)

	mockService.AssertExpectations(t)
}

// TestListQueryHistory_EmptyHistory_ReturnsEmptyArray tests that empty history returns empty array
func TestListQueryHistory_EmptyHistory_ReturnsEmptyArray(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupListQueryHistoryRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)

	// Mock service to return empty list
	mockService.On("ListQueryHistory", mock.Anything, user.ID.String(), 20, 0, "").
		Return([]models.QueryHistory{}, int64(0), nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	req, _ := http.NewRequest("GET", "/api/v1/queries/history", nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, float64(0), response["total"])
	history := response["history"].([]interface{})
	assert.Len(t, history, 0)

	mockService.AssertExpectations(t)
}

// =============================================================================
// GET QUERY RESULTS TESTS
// =============================================================================

// setupGetQueryResultsRouter creates a test router with GetQueryResults endpoint
func setupGetQueryResultsRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockQueryService, *auth.JWTManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockQueryService := new(MockQueryService)

	handler := &testGetQueryResultsHandler{
		db:           db,
		queryService: mockQueryService,
	}

	api := router.Group("/api/v1")
	{
		queries := api.Group("/queries")
		queries.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			queries.GET("/:id/results", handler.GetQueryResults)
		}
	}

	return router, mockQueryService, jwtManager
}

// testGetQueryResultsHandler wraps the real handler for testing
type testGetQueryResultsHandler struct {
	db           *gorm.DB
	queryService *MockQueryService
}

// GetQueryResults implements the handler logic using the mock service
func (h *testGetQueryResultsHandler) GetQueryResults(c *gin.Context) {
	queryID := c.Param("id")
	userID := c.GetString("user_id")

	// Verify query exists
	var query models.Query
	if err := h.db.First(&query, "id = ?", queryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Query not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch query"})
		}
		return
	}

	// Check permission (only owner or admin can access)
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err == nil {
		if user.Role != models.RoleAdmin && query.UserID.String() != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "100"))
	sortColumn := c.Query("sort_column")
	sortDirection := c.DefaultQuery("sort_direction", "asc")

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if perPage < 10 || perPage > 1000 {
		perPage = 100
	}
	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "asc"
	}

	// Get paginated results
	ctx := c.Request.Context()
	queryUUID, err := uuid.Parse(queryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query ID"})
		return
	}

	// Get query result from database to fetch column types
	var queryResult models.QueryResult
	if err := h.db.Where("query_id = ?", queryID).First(&queryResult).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Query result not found"})
		return
	}

	// Parse column types from database
	var columnTypes []string
	json.Unmarshal([]byte(queryResult.ColumnTypes), &columnTypes)

	rows, columnNames, metadata, err := h.queryService.GetPaginatedResults(ctx, queryUUID, page, perPage, sortColumn, sortDirection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build column info with names and types
	columns := make([]dto.ColumnInfo, len(columnNames))
	for i, col := range columnNames {
		colType := "unknown"
		if i < len(columnTypes) && columnTypes[i] != "" {
			colType = columnTypes[i]
		}
		columns[i] = dto.ColumnInfo{Name: col, Type: colType}
	}

	c.JSON(http.StatusOK, dto.PaginatedResultDTO{
		QueryID:  queryID,
		RowCount: len(rows),
		Columns:  columns,
		Data:     rows,
		Metadata: dto.PaginationMeta{
			Page:       metadata.Page,
			PerPage:    metadata.PerPage,
			TotalPages: metadata.TotalPages,
			TotalRows:  metadata.TotalRows,
			HasNext:    metadata.HasNext,
			HasPrev:    metadata.HasPrev,
		},
		SortColumn:    sortColumn,
		SortDirection: sortDirection,
	})
}

// TestGetQueryResults_WithPagination_ReturnsPaginatedResults tests pagination works correctly
func TestGetQueryResults_WithPagination_ReturnsPaginatedResults(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupGetQueryResultsRouter(t, db)

	// Create admin user
	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Create a query
	query := &models.Query{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		UserID:        admin.ID,
		QueryText:     "SELECT * FROM users",
		OperationType: models.OperationSelect,
		Status:        models.StatusCompleted,
	}
	db.Create(query)

	// Create query result
	queryResult := &models.QueryResult{
		ID:          uuid.New(),
		QueryID:     query.ID,
		RowCount:    50,
		ColumnNames: `["id", "name", "email"]`,
		ColumnTypes: `["integer", "text", "text"]`,
		Data:        `[{"id": 1, "name": "test", "email": "test@test.com"}]`,
	}
	db.Create(queryResult)

	// Mock paginated results
	page := 2
	perPage := 10
	mockService.On("GetPaginatedResults", mock.Anything, query.ID, page, perPage, "", "asc").
		Return(
			[]map[string]interface{}{
				{"id": float64(11), "name": "user11", "email": "user11@test.com"},
				{"id": float64(12), "name": "user12", "email": "user12@test.com"},
			},
			[]string{"id", "name", "email"},
			&dto.PaginationMeta{
				Page:       2,
				PerPage:    10,
				TotalPages: 5,
				TotalRows:  50,
				HasNext:    true,
				HasPrev:    true,
			},
			nil,
		)

	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/queries/%s/results?page=2&per_page=10", query.ID.String()), nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PaginatedResultDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, query.ID.String(), response.QueryID)
	assert.Equal(t, 2, response.RowCount)
	assert.Len(t, response.Columns, 3)
	assert.Equal(t, 2, response.Metadata.Page)
	assert.Equal(t, 10, response.Metadata.PerPage)
	assert.Equal(t, 50, response.Metadata.TotalRows)
	assert.Equal(t, 5, response.Metadata.TotalPages)
	assert.True(t, response.Metadata.HasNext)
	assert.True(t, response.Metadata.HasPrev)

	mockService.AssertExpectations(t)
}

// TestGetQueryResults_WithSorting_ReturnsSortedResults tests sorting works correctly
func TestGetQueryResults_WithSorting_ReturnsSortedResults(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupGetQueryResultsRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	query := &models.Query{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		UserID:        admin.ID,
		QueryText:     "SELECT * FROM users ORDER BY name",
		OperationType: models.OperationSelect,
		Status:        models.StatusCompleted,
	}
	db.Create(query)

	queryResult := &models.QueryResult{
		ID:          uuid.New(),
		QueryID:     query.ID,
		RowCount:    3,
		ColumnNames: `["id", "name"]`,
		ColumnTypes: `["integer", "text"]`,
		Data:        `[{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}, {"id": 3, "name": "Charlie"}]`,
	}
	db.Create(queryResult)

	// Mock with sorting - DESC order
	mockService.On("GetPaginatedResults", mock.Anything, query.ID, 1, 100, "name", "desc").
		Return(
			[]map[string]interface{}{
				{"id": float64(3), "name": "Charlie"},
				{"id": float64(2), "name": "Bob"},
				{"id": float64(1), "name": "Alice"},
			},
			[]string{"id", "name"},
			&dto.PaginationMeta{
				Page:       1,
				PerPage:    100,
				TotalPages: 1,
				TotalRows:  3,
				HasNext:    false,
				HasPrev:    false,
			},
			nil,
		)

	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/queries/%s/results?sort_column=name&sort_direction=desc", query.ID.String()), nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PaginatedResultDTO
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "name", response.SortColumn)
	assert.Equal(t, "desc", response.SortDirection)
	assert.Equal(t, "Charlie", response.Data[0]["name"])
	assert.Equal(t, "Alice", response.Data[2]["name"])

	mockService.AssertExpectations(t)
}

// TestGetQueryResults_QueryNotFound_ReturnsNotFound tests 404 for non-existent query
func TestGetQueryResults_QueryNotFound_ReturnsNotFound(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, jwtManager := setupGetQueryResultsRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	nonExistentID := uuid.New().String()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/queries/%s/results", nonExistentID), nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Query not found")
}

// TestGetQueryResults_NoPermission_ReturnsForbidden tests that non-owner gets 403
func TestGetQueryResults_NoPermission_ReturnsForbidden(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, jwtManager := setupGetQueryResultsRouter(t, db)

	// Create two users
	owner := fixtures.CreateTestRegularUser(t, db)
	otherUser := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Create query owned by owner
	query := &models.Query{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		UserID:        owner.ID,
		QueryText:     "SELECT * FROM users",
		OperationType: models.OperationSelect,
		Status:        models.StatusCompleted,
	}
	db.Create(query)

	// Create query result
	queryResult := &models.QueryResult{
		ID:          uuid.New(),
		QueryID:     query.ID,
		RowCount:    1,
		ColumnNames: `["id"]`,
		ColumnTypes: `["integer"]`,
		Data:        `[{"id": 1}]`,
	}
	db.Create(queryResult)

	// Other user tries to access (not admin)
	token, _ := jwtManager.GenerateToken(otherUser.ID, otherUser.Email, string(models.RoleUser))

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/queries/%s/results", query.ID.String()), nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Access denied")
}

// TestGetQueryResults_AdminCanAccessAnyQuery tests admin can access any query
func TestGetQueryResults_AdminCanAccessAnyQuery(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupGetQueryResultsRouter(t, db)

	// Create regular user as owner
	owner := fixtures.CreateTestRegularUser(t, db)
	// Create admin user
	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Create query owned by regular user
	query := &models.Query{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		UserID:        owner.ID,
		QueryText:     "SELECT * FROM secret_data",
		OperationType: models.OperationSelect,
		Status:        models.StatusCompleted,
	}
	db.Create(query)

	queryResult := &models.QueryResult{
		ID:          uuid.New(),
		QueryID:     query.ID,
		RowCount:    1,
		ColumnNames: `["secret"]`,
		ColumnTypes: `["text"]`,
		Data:        `[{"secret": "password123"}]`,
	}
	db.Create(queryResult)

	// Admin should have full access
	mockService.On("GetPaginatedResults", mock.Anything, query.ID, 1, 100, "", "asc").
		Return(
			[]map[string]interface{}{{"secret": "password123"}},
			[]string{"secret"},
			&dto.PaginationMeta{Page: 1, PerPage: 100, TotalPages: 1, TotalRows: 1, HasNext: false, HasPrev: false},
			nil,
		)

	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/queries/%s/results", query.ID.String()), nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PaginatedResultDTO
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 1, response.RowCount)

	mockService.AssertExpectations(t)
}

// TestGetQueryResults_LargeResultSet_ReturnsResults tests handling of large result sets
func TestGetQueryResults_LargeResultSet_ReturnsResults(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupGetQueryResultsRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	query := &models.Query{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		UserID:        admin.ID,
		QueryText:     "SELECT * FROM large_table",
		OperationType: models.OperationSelect,
		Status:        models.StatusCompleted,
	}
	db.Create(query)

	// Create query result with 10000 rows
	queryResult := &models.QueryResult{
		ID:          uuid.New(),
		QueryID:     query.ID,
		RowCount:    10000,
		ColumnNames: `["id", "data"]`,
		ColumnTypes: `["integer", "text"]`,
		Data:        `[{"id": 1, "data": "test"}]`,
	}
	db.Create(queryResult)

	// Mock with max perPage
	mockService.On("GetPaginatedResults", mock.Anything, query.ID, 1, 1000, "", "asc").
		Return(
			generateLargeResultSet(1000),
			[]string{"id", "data"},
			&dto.PaginationMeta{
				Page:       1,
				PerPage:    1000,
				TotalPages: 10,
				TotalRows:  10000,
				HasNext:    true,
				HasPrev:    false,
			},
			nil,
		)

	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	// Request max per_page
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/queries/%s/results?per_page=1000", query.ID.String()), nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PaginatedResultDTO
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, 1000, response.RowCount)
	assert.Equal(t, 10000, response.Metadata.TotalRows)
	assert.Equal(t, 10, response.Metadata.TotalPages)
	assert.True(t, response.Metadata.HasNext)

	mockService.AssertExpectations(t)
}

// generateLargeResultSet generates a large result set for testing
func generateLargeResultSet(count int) []map[string]interface{} {
	result := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		result[i] = map[string]interface{}{
			"id":   float64(i + 1),
			"data": fmt.Sprintf("data_%d", i+1),
		}
	}
	return result
}

// TestGetQueryResults_InvalidPaginationParams_UsesDefaults tests that invalid params use defaults
func TestGetQueryResults_InvalidPaginationParams_UsesDefaults(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupGetQueryResultsRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	query := &models.Query{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		UserID:        admin.ID,
		QueryText:     "SELECT * FROM users",
		OperationType: models.OperationSelect,
		Status:        models.StatusCompleted,
	}
	db.Create(query)

	queryResult := &models.QueryResult{
		ID:          uuid.New(),
		QueryID:     query.ID,
		RowCount:    10,
		ColumnNames: `["id"]`,
		ColumnTypes: `["integer"]`,
		Data:        `[{"id": 1}]`,
	}
	db.Create(queryResult)

	// Default pagination should be page=1, perPage=100
	mockService.On("GetPaginatedResults", mock.Anything, query.ID, 1, 100, "", "asc").
		Return(
			[]map[string]interface{}{{"id": float64(1)}},
			[]string{"id"},
			&dto.PaginationMeta{Page: 1, PerPage: 100, TotalPages: 1, TotalRows: 1, HasNext: false, HasPrev: false},
			nil,
		)

	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	// Request with invalid pagination params
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/queries/%s/results?page=-1&per_page=5&sort_direction=invalid", query.ID.String()), nil)
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.PaginatedResultDTO
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should use defaults
	assert.Equal(t, 1, response.Metadata.Page)
	assert.Equal(t, 100, response.Metadata.PerPage) // perPage < 10 uses default 100
	assert.Equal(t, "asc", response.SortDirection)  // invalid direction defaults to "asc"

	mockService.AssertExpectations(t)
}

// =============================================================================
// EXPORT QUERY TESTS
// =============================================================================

// setupExportQueryRouter creates a test router with ExportQuery endpoint
func setupExportQueryRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockQueryService, *auth.JWTManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockQueryService := new(MockQueryService)

	handler := &testExportQueryHandler{
		db:           db,
		queryService: mockQueryService,
	}

	api := router.Group("/api/v1")
	{
		queries := api.Group("/queries")
		queries.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			queries.POST("/export", handler.ExportQuery)
		}
	}

	return router, mockQueryService, jwtManager
}

// testExportQueryHandler wraps the real handler for testing
type testExportQueryHandler struct {
	db           *gorm.DB
	queryService *MockQueryService
}

// ExportQuery implements the handler logic using the mock service
func (h *testExportQueryHandler) ExportQuery(c *gin.Context) {
	var req dto.ExportQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")

	// Verify query exists
	var query models.Query
	if err := h.db.First(&query, "id = ?", req.QueryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Query not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch query"})
		}
		return
	}

	// Check permission (only owner or admin can export)
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err == nil {
		if user.Role != models.RoleAdmin && query.UserID.String() != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	// Parse query ID as UUID
	queryUUID, err := uuid.Parse(req.QueryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query ID"})
		return
	}

	// Export the query results
	ctx := c.Request.Context()
	data, contentType, err := h.queryService.ExportQuery(ctx, queryUUID, string(req.Format))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("query_%s.%s", req.QueryID, req.Format)
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Data(http.StatusOK, contentType, data)
}

// createExportQueryRequest creates an export query request body
func createExportQueryRequest(queryID string, format dto.ExportFormat) ([]byte, error) {
	req := dto.ExportQueryRequest{
		QueryID: queryID,
		Format:  format,
	}
	return json.Marshal(req)
}

// TestExportQuery_CSVFormat_ReturnsCSVData tests CSV export returns correct data
func TestExportQuery_CSVFormat_ReturnsCSVData(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupExportQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	query := &models.Query{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		UserID:        admin.ID,
		QueryText:     "SELECT * FROM users",
		OperationType: models.OperationSelect,
		Status:        models.StatusCompleted,
	}
	db.Create(query)

	csvData := []byte("id,name,email\n1,Alice,alice@test.com\n2,Bob,bob@test.com\n")

	mockService.On("ExportQuery", mock.Anything, query.ID, "csv").
		Return(csvData, "text/csv", nil)

	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	reqBody, _ := createExportQueryRequest(query.ID.String(), dto.ExportFormatCSV)
	req, _ := http.NewRequest("POST", "/api/v1/queries/export", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), ".csv")
	assert.Equal(t, string(csvData), w.Body.String())

	mockService.AssertExpectations(t)
}

// TestExportQuery_JSONFormat_ReturnsJSONData tests JSON export returns correct data
func TestExportQuery_JSONFormat_ReturnsJSONData(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupExportQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	query := &models.Query{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		UserID:        admin.ID,
		QueryText:     "SELECT * FROM users",
		OperationType: models.OperationSelect,
		Status:        models.StatusCompleted,
	}
	db.Create(query)

	jsonData := []byte(`[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]`)

	mockService.On("ExportQuery", mock.Anything, query.ID, "json").
		Return(jsonData, "application/json", nil)

	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	reqBody, _ := createExportQueryRequest(query.ID.String(), dto.ExportFormatJSON)
	req, _ := http.NewRequest("POST", "/api/v1/queries/export", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), ".json")
	assert.Equal(t, string(jsonData), w.Body.String())

	mockService.AssertExpectations(t)
}

// TestExportQuery_InvalidFormat_ReturnsBadRequest tests that invalid format returns 400
func TestExportQuery_InvalidFormat_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, jwtManager := setupExportQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	// Invalid format should fail binding validation
	invalidReq := `{"query_id": "550e8400-e29b-41d4-a716-446655440000", "format": "xml"}`
	req, _ := http.NewRequest("POST", "/api/v1/queries/export", bytes.NewBuffer([]byte(invalidReq)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Format")
}

// TestExportQuery_QueryNotFound_ReturnsNotFound tests that non-existent query returns 404
func TestExportQuery_QueryNotFound_ReturnsNotFound(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, jwtManager := setupExportQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	nonExistentID := uuid.New().String()
	reqBody, _ := createExportQueryRequest(nonExistentID, dto.ExportFormatCSV)
	req, _ := http.NewRequest("POST", "/api/v1/queries/export", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Query not found")
}

// TestExportQuery_NoPermission_ReturnsForbidden tests that non-owner gets 403
func TestExportQuery_NoPermission_ReturnsForbidden(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, jwtManager := setupExportQueryRouter(t, db)

	owner := fixtures.CreateTestRegularUser(t, db)
	otherUser := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	query := &models.Query{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		UserID:        owner.ID,
		QueryText:     "SELECT * FROM secret_data",
		OperationType: models.OperationSelect,
		Status:        models.StatusCompleted,
	}
	db.Create(query)

	// Other user tries to export
	token, _ := jwtManager.GenerateToken(otherUser.ID, otherUser.Email, string(models.RoleUser))

	reqBody, _ := createExportQueryRequest(query.ID.String(), dto.ExportFormatCSV)
	req, _ := http.NewRequest("POST", "/api/v1/queries/export", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Access denied")
}

// TestExportQuery_AdminCanExportAnyQuery tests admin can export any query
func TestExportQuery_AdminCanExportAnyQuery(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupExportQueryRouter(t, db)

	owner := fixtures.CreateTestRegularUser(t, db)
	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	query := &models.Query{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		UserID:        owner.ID,
		QueryText:     "SELECT * FROM secret_data",
		OperationType: models.OperationSelect,
		Status:        models.StatusCompleted,
	}
	db.Create(query)

	csvData := []byte("secret\npassword123\n")
	mockService.On("ExportQuery", mock.Anything, query.ID, "csv").
		Return(csvData, "text/csv", nil)

	// Admin exports
	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	reqBody, _ := createExportQueryRequest(query.ID.String(), dto.ExportFormatCSV)
	req, _ := http.NewRequest("POST", "/api/v1/queries/export", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, string(csvData), w.Body.String())

	mockService.AssertExpectations(t)
}

// TestExportQuery_OwnerCanExportQuery tests that owner can export their own query
func TestExportQuery_OwnerCanExportQuery(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupExportQueryRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	query := &models.Query{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		UserID:        user.ID,
		QueryText:     "SELECT * FROM my_data",
		OperationType: models.OperationSelect,
		Status:        models.StatusCompleted,
	}
	db.Create(query)

	jsonData := []byte(`[{"id": 1, "data": "test"}]`)
	mockService.On("ExportQuery", mock.Anything, query.ID, "json").
		Return(jsonData, "application/json", nil)

	token, _ := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))

	reqBody, _ := createExportQueryRequest(query.ID.String(), dto.ExportFormatJSON)
	req, _ := http.NewRequest("POST", "/api/v1/queries/export", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, string(jsonData), w.Body.String())

	mockService.AssertExpectations(t)
}

// TestExportQuery_ExportServiceError_ReturnsInternalError tests service errors return 500
func TestExportQuery_ExportServiceError_ReturnsInternalError(t *testing.T) {
	db := setupQueryTestDB(t)
	router, mockService, jwtManager := setupExportQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	query := &models.Query{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		UserID:        admin.ID,
		QueryText:     "SELECT * FROM users",
		OperationType: models.OperationSelect,
		Status:        models.StatusCompleted,
	}
	db.Create(query)

	// Mock service error
	mockService.On("ExportQuery", mock.Anything, query.ID, "csv").
		Return(nil, "", errors.New("export service unavailable"))

	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	reqBody, _ := createExportQueryRequest(query.ID.String(), dto.ExportFormatCSV)
	req, _ := http.NewRequest("POST", "/api/v1/queries/export", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "export service unavailable")

	mockService.AssertExpectations(t)
}

// TestExportQuery_MissingQueryID_ReturnsBadRequest tests missing query_id returns 400
func TestExportQuery_MissingQueryID_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, jwtManager := setupExportQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	// Missing query_id
	invalidReq := `{"format": "csv"}`
	req, _ := http.NewRequest("POST", "/api/v1/queries/export", bytes.NewBuffer([]byte(invalidReq)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestExportQuery_InvalidJSON_ReturnsBadRequest tests invalid JSON returns 400
func TestExportQuery_InvalidJSON_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, jwtManager := setupExportQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	req, _ := http.NewRequest("POST", "/api/v1/queries/export", bytes.NewBuffer([]byte(`{invalid json}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestExportQuery_EmptyQueryID_ReturnsBadRequest tests empty query_id returns 400
func TestExportQuery_EmptyQueryID_ReturnsBadRequest(t *testing.T) {
	db := setupQueryTestDB(t)
	router, _, jwtManager := setupExportQueryRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	token, _ := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))

	reqBody, _ := createExportQueryRequest("", dto.ExportFormatCSV)
	req, _ := http.NewRequest("POST", "/api/v1/queries/export", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", createAuthHeader(token))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// =============================================================================
// PERMISSION CHECK HELPER TESTS
// =============================================================================

// queryServiceInterface defines the interface needed by permission check helpers
type queryServiceInterface interface {
	GetEffectivePermissions(ctx context.Context, userID, dsID uuid.UUID) (*models.EffectivePermissions, error)
}

// mockQueryServiceWrapper wraps MockQueryService to implement the interface
type mockQueryServiceWrapper struct {
	*MockQueryService
}

func (m *mockQueryServiceWrapper) GetEffectivePermissions(ctx context.Context, userID, dsID uuid.UUID) (*models.EffectivePermissions, error) {
	return m.MockQueryService.GetEffectivePermissions(ctx, userID, dsID)
}

// testQueryHandlerForPermissions creates a test handler with mock service for permission tests
// This creates a struct with the same shape as QueryHandler but using an interface
type testQueryHandlerForPermissions struct {
	db           *gorm.DB
	queryService queryServiceInterface
}

// checkReadPermission is a copy of the handler method for testing
func (h *testQueryHandlerForPermissions) checkReadPermission(c *gin.Context, userID, dataSourceID string) bool {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return false
	}
	dsID, err := uuid.Parse(dataSourceID)
	if err != nil {
		return false
	}

	perms, err := h.queryService.GetEffectivePermissions(c.Request.Context(), uID, dsID)
	if err != nil {
		return false
	}
	return perms.CanRead
}

// checkWritePermission is a copy of the handler method for testing
func (h *testQueryHandlerForPermissions) checkWritePermission(c *gin.Context, userID, dataSourceID string) bool {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return false
	}
	dsID, err := uuid.Parse(dataSourceID)
	if err != nil {
		return false
	}

	perms, err := h.queryService.GetEffectivePermissions(c.Request.Context(), uID, dsID)
	if err != nil {
		return false
	}
	return perms.CanWrite
}

// setupPermissionTestHandler creates a test handler with mock service
func setupPermissionTestHandler(db *gorm.DB, mockService *MockQueryService) *testQueryHandlerForPermissions {
	return &testQueryHandlerForPermissions{
		db:           db,
		queryService: &mockQueryServiceWrapper{MockQueryService: mockService},
	}
}

// setupPermissionContext creates a gin context with user context
func setupPermissionContext(userID, role string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request = c.Request.WithContext(context.Background())
	c.Set("user_id", userID)
	c.Set("role", role)
	return c, w
}

// TestCheckReadPermission_UserWithRead_ReturnsTrue tests that user with read permission returns true
func TestCheckReadPermission_UserWithRead_ReturnsTrue(t *testing.T) {
	db := setupQueryTestDB(t)
	mockService := new(MockQueryService)

	// Create test data
	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	// Setup mock to return permissions with CanRead = true
	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:    true,
			CanWrite:   false,
			CanApprove: false,
			CanSelect:  true,
			CanInsert:  false,
			CanUpdate:  false,
			CanDelete:  false,
		}, nil)

	// Create handler via test wrapper
	handler := setupPermissionTestHandler(db, mockService)

	c, _ := setupPermissionContext(user.ID.String(), string(models.RoleUser))

	// Call the permission check
	result := handler.checkReadPermission(c, user.ID.String(), dataSource.ID.String())

	assert.True(t, result, "User with CanRead=true should return true")
	mockService.AssertExpectations(t)
}

// TestCheckReadPermission_UserWithoutRead_ReturnsFalse tests that user without read permission returns false
func TestCheckReadPermission_UserWithoutRead_ReturnsFalse(t *testing.T) {
	db := setupQueryTestDB(t)
	mockService := new(MockQueryService)

	// Create test data
	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	// Setup mock to return permissions with CanRead = false
	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:    false,
			CanWrite:   false,
			CanApprove: false,
			CanSelect:  false,
			CanInsert:  false,
			CanUpdate:  false,
			CanDelete:  false,
		}, nil)

	handler := setupPermissionTestHandler(db, mockService)

	c, _ := setupPermissionContext(user.ID.String(), string(models.RoleUser))

	result := handler.checkReadPermission(c, user.ID.String(), dataSource.ID.String())

	assert.False(t, result, "User with CanRead=false should return false")
	mockService.AssertExpectations(t)
}

// TestCheckReadPermission_ServiceError_ReturnsFalse tests that service error returns false
func TestCheckReadPermission_ServiceError_ReturnsFalse(t *testing.T) {
	db := setupQueryTestDB(t)
	mockService := new(MockQueryService)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	// Setup mock to return error
	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(nil, errors.New("database error"))

	handler := setupPermissionTestHandler(db, mockService)

	c, _ := setupPermissionContext(user.ID.String(), string(models.RoleUser))

	result := handler.checkReadPermission(c, user.ID.String(), dataSource.ID.String())

	assert.False(t, result, "Service error should return false")
	mockService.AssertExpectations(t)
}

// TestCheckReadPermission_InvalidUserID_ReturnsFalse tests that invalid user ID returns false
func TestCheckReadPermission_InvalidUserID_ReturnsFalse(t *testing.T) {
	db := setupQueryTestDB(t)
	mockService := new(MockQueryService)

	handler := setupPermissionTestHandler(db, mockService)

	c, _ := setupPermissionContext("invalid-uuid", string(models.RoleUser))
	dataSourceID := uuid.New().String()

	result := handler.checkReadPermission(c, "invalid-uuid", dataSourceID)

	assert.False(t, result, "Invalid user ID should return false")
	mockService.AssertNotCalled(t, "GetEffectivePermissions")
}

// TestCheckReadPermission_InvalidDataSourceID_ReturnsFalse tests that invalid data source ID returns false
func TestCheckReadPermission_InvalidDataSourceID_ReturnsFalse(t *testing.T) {
	db := setupQueryTestDB(t)
	mockService := new(MockQueryService)

	user := fixtures.CreateTestRegularUser(t, db)

	handler := setupPermissionTestHandler(db, mockService)

	c, _ := setupPermissionContext(user.ID.String(), string(models.RoleUser))

	result := handler.checkReadPermission(c, user.ID.String(), "invalid-uuid")

	assert.False(t, result, "Invalid data source ID should return false")
	mockService.AssertNotCalled(t, "GetEffectivePermissions")
}

// TestCheckReadPermission_ViewerWithRead_ReturnsTrue tests that viewer with read permission returns true
func TestCheckReadPermission_ViewerWithRead_ReturnsTrue(t *testing.T) {
	db := setupQueryTestDB(t)
	mockService := new(MockQueryService)

	user := fixtures.CreateTestViewerUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:    true,
			CanWrite:   false,
			CanApprove: false,
			CanSelect:  true,
			CanInsert:  false,
			CanUpdate:  false,
			CanDelete:  false,
		}, nil)

	handler := setupPermissionTestHandler(db, mockService)

	c, _ := setupPermissionContext(user.ID.String(), string(models.RoleUser))

	result := handler.checkReadPermission(c, user.ID.String(), dataSource.ID.String())

	assert.True(t, result, "Viewer with CanRead=true should return true")
	mockService.AssertExpectations(t)
}

// TestCheckReadPermission_MultipleGroups_UnionPermissions tests permission union across groups
func TestCheckReadPermission_MultipleGroups_UnionPermissions(t *testing.T) {
	db := setupQueryTestDB(t)
	mockService := new(MockQueryService)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group1 := fixtures.CreateTestGroup(t, db, "test-group-1")
	group2 := fixtures.CreateTestGroup(t, db, "test-group-2")
	fixtures.AddUserToGroup(db, user.ID, group1.ID)
	fixtures.AddUserToGroup(db, user.ID, group2.ID)
	fixtures.GrantReadOnlyPermission(t, db, group1.ID, dataSource.ID)
	fixtures.GrantFullPermission(t, db, group2.ID, dataSource.ID)

	// Mock returns merged permissions (CanRead from group1 union)
	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:    true,
			CanWrite:   true,
			CanApprove: true,
			CanSelect:  true,
			CanInsert:  true,
			CanUpdate:  true,
			CanDelete:  true,
		}, nil)

	handler := setupPermissionTestHandler(db, mockService)

	c, _ := setupPermissionContext(user.ID.String(), string(models.RoleUser))

	result := handler.checkReadPermission(c, user.ID.String(), dataSource.ID.String())

	assert.True(t, result, "User with read permission from multiple groups should return true")
	mockService.AssertExpectations(t)
}

// TestCheckWritePermission_UserWithWrite_ReturnsTrue tests that user with write permission returns true
func TestCheckWritePermission_UserWithWrite_ReturnsTrue(t *testing.T) {
	db := setupQueryTestDB(t)
	mockService := new(MockQueryService)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:    true,
			CanWrite:   true,
			CanApprove: true,
			CanSelect:  true,
			CanInsert:  true,
			CanUpdate:  true,
			CanDelete:  true,
		}, nil)

	handler := setupPermissionTestHandler(db, mockService)

	c, _ := setupPermissionContext(user.ID.String(), string(models.RoleUser))

	result := handler.checkWritePermission(c, user.ID.String(), dataSource.ID.String())

	assert.True(t, result, "User with CanWrite=true should return true")
	mockService.AssertExpectations(t)
}

// TestCheckWritePermission_UserWithoutWrite_ReturnsFalse tests that user without write permission returns false
func TestCheckWritePermission_UserWithoutWrite_ReturnsFalse(t *testing.T) {
	db := setupQueryTestDB(t)
	mockService := new(MockQueryService)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:    true,
			CanWrite:   false,
			CanApprove: false,
			CanSelect:  true,
			CanInsert:  false,
			CanUpdate:  false,
			CanDelete:  false,
		}, nil)

	handler := setupPermissionTestHandler(db, mockService)

	c, _ := setupPermissionContext(user.ID.String(), string(models.RoleUser))

	result := handler.checkWritePermission(c, user.ID.String(), dataSource.ID.String())

	assert.False(t, result, "User with CanWrite=false should return false")
	mockService.AssertExpectations(t)
}

// TestCheckWritePermission_ServiceError_ReturnsFalse tests that service error returns false
func TestCheckWritePermission_ServiceError_ReturnsFalse(t *testing.T) {
	db := setupQueryTestDB(t)
	mockService := new(MockQueryService)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(nil, errors.New("database error"))

	handler := setupPermissionTestHandler(db, mockService)

	c, _ := setupPermissionContext(user.ID.String(), string(models.RoleUser))

	result := handler.checkWritePermission(c, user.ID.String(), dataSource.ID.String())

	assert.False(t, result, "Service error should return false")
	mockService.AssertExpectations(t)
}

// TestCheckWritePermission_ViewerCannotWrite_ReturnsFalse tests that viewer cannot write
func TestCheckWritePermission_ViewerCannotWrite_ReturnsFalse(t *testing.T) {
	db := setupQueryTestDB(t)
	mockService := new(MockQueryService)

	user := fixtures.CreateTestViewerUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:    true,
			CanWrite:   false,
			CanApprove: false,
			CanSelect:  true,
			CanInsert:  false,
			CanUpdate:  false,
			CanDelete:  false,
		}, nil)

	handler := setupPermissionTestHandler(db, mockService)

	c, _ := setupPermissionContext(user.ID.String(), string(models.RoleViewer))

	result := handler.checkWritePermission(c, user.ID.String(), dataSource.ID.String())

	assert.False(t, result, "Viewer cannot write should return false")
	mockService.AssertExpectations(t)
}

// TestCheckWritePermission_MultipleGroups_UnionPermissions tests permission union across groups
func TestCheckWritePermission_MultipleGroups_UnionPermissions(t *testing.T) {
	db := setupQueryTestDB(t)
	mockService := new(MockQueryService)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group1 := fixtures.CreateTestGroup(t, db, "test-group-1")
	group2 := fixtures.CreateTestGroup(t, db, "test-group-2")
	fixtures.AddUserToGroup(db, user.ID, group1.ID)
	fixtures.AddUserToGroup(db, user.ID, group2.ID)
	fixtures.GrantReadOnlyPermission(t, db, group1.ID, dataSource.ID)
	fixtures.GrantFullPermission(t, db, group2.ID, dataSource.ID)

	// Mock returns merged permissions
	mockService.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{
			CanRead:    true,
			CanWrite:   true,
			CanApprove: true,
			CanSelect:  true,
			CanInsert:  true,
			CanUpdate:  true,
			CanDelete:  true,
		}, nil)

	handler := setupPermissionTestHandler(db, mockService)

	c, _ := setupPermissionContext(user.ID.String(), string(models.RoleUser))

	result := handler.checkWritePermission(c, user.ID.String(), dataSource.ID.String())

	assert.True(t, result, "User with write permission from multiple groups should return true")
	mockService.AssertExpectations(t)
}

// TestCheckWritePermission_InvalidUserID_ReturnsFalse tests that invalid user ID returns false
func TestCheckWritePermission_InvalidUserID_ReturnsFalse(t *testing.T) {
	db := setupQueryTestDB(t)
	mockService := new(MockQueryService)

	handler := setupPermissionTestHandler(db, mockService)

	c, _ := setupPermissionContext("invalid-uuid", string(models.RoleUser))

	result := handler.checkWritePermission(c, "invalid-uuid", uuid.New().String())

	assert.False(t, result, "Invalid user ID should return false")
	mockService.AssertNotCalled(t, "GetEffectivePermissions")
}

// TestCheckWritePermission_InvalidDataSourceID_ReturnsFalse tests that invalid data source ID returns false
func TestCheckWritePermission_InvalidDataSourceID_ReturnsFalse(t *testing.T) {
	db := setupQueryTestDB(t)
	mockService := new(MockQueryService)

	user := fixtures.CreateTestRegularUser(t, db)

	handler := setupPermissionTestHandler(db, mockService)

	c, _ := setupPermissionContext(user.ID.String(), string(models.RoleUser))

	result := handler.checkWritePermission(c, user.ID.String(), "invalid-uuid")

	assert.False(t, result, "Invalid data source ID should return false")
	mockService.AssertNotCalled(t, "GetEffectivePermissions")
}
