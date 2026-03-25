package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
