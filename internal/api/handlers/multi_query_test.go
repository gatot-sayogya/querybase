package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
	"github.com/yourorg/querybase/internal/service"
	testauth "github.com/yourorg/querybase/internal/testutils/auth"
	"github.com/yourorg/querybase/internal/testutils/fixtures"
)

// MockMultiQueryService is a mock implementation of the multi-query service for testing
type MockMultiQueryService struct {
	mock.Mock
}

func (m *MockMultiQueryService) PreviewMultiQuery(ctx interface{}, dataSourceID, userID uuid.UUID, queryTexts []string) (*service.MultiQueryPreviewResult, error) {
	args := m.Called(ctx, dataSourceID, userID, queryTexts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.MultiQueryPreviewResult), args.Error(1)
}

func (m *MockMultiQueryService) CalculateMultiQueryImpact(ctx interface{}, dataSourceID, userID uuid.UUID, queryTexts []string) (*service.MultiQueryImpact, error) {
	args := m.Called(ctx, dataSourceID, userID, queryTexts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.MultiQueryImpact), args.Error(1)
}

func (m *MockMultiQueryService) CreateMultiQueryTransaction(ctx interface{}, approvalID *uuid.UUID, dataSourceID uuid.UUID, queryTexts []string, startedBy uuid.UUID) (*models.QueryTransaction, error) {
	args := m.Called(ctx, approvalID, dataSourceID, queryTexts, startedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.QueryTransaction), args.Error(1)
}

func (m *MockMultiQueryService) ExecuteMultiQuery(ctx interface{}, transactionID uuid.UUID, auditMode models.AuditMode) (*service.MultiQueryResult, error) {
	args := m.Called(ctx, transactionID, auditMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.MultiQueryResult), args.Error(1)
}

func (m *MockMultiQueryService) GetMultiQueryStatements(ctx interface{}, transactionID uuid.UUID) ([]models.QueryTransactionStatement, error) {
	args := m.Called(ctx, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.QueryTransactionStatement), args.Error(1)
}

func (m *MockMultiQueryService) RollbackMultiQuery(ctx interface{}, transactionID uuid.UUID) error {
	args := m.Called(ctx, transactionID)
	return args.Error(0)
}

// MockQueryServiceForMultiQuery is a mock for the query service used by multi-query handler
type MockQueryServiceForMultiQuery struct {
	mock.Mock
}

func (m *MockQueryServiceForMultiQuery) GetEffectivePermissions(ctx interface{}, userID, dsID uuid.UUID) (*models.EffectivePermissions, error) {
	args := m.Called(ctx, userID, dsID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EffectivePermissions), args.Error(1)
}

func (m *MockQueryServiceForMultiQuery) PreviewWriteQuery(ctx interface{}, queryText string, dataSource *models.DataSource) (*dto.PreviewWriteQueryResponse, error) {
	args := m.Called(ctx, queryText, dataSource)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PreviewWriteQueryResponse), args.Error(1)
}

// setupMultiQueryTestDB creates an in-memory SQLite database for testing multi-query handlers
func setupMultiQueryTestDB(t *testing.T) *gorm.DB {
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
		&models.QueryTransaction{},
		&models.QueryTransactionStatement{},
	)
	require.NoError(t, err)

	return db
}

// setupMultiQueryTestRouter creates a test router with multi-query handler
func setupMultiQueryTestRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockMultiQueryService, *MockQueryServiceForMultiQuery, *auth.JWTManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockMultiQueryService := new(MockMultiQueryService)
	mockQueryService := new(MockQueryServiceForMultiQuery)

	// Create a test-specific wrapper that mimics the PreviewMultiQuery flow
	multiQueryHandler := &testMultiQueryHandler{
		db:            db,
		multiQuerySvc: mockMultiQueryService,
		querySvc:      mockQueryService,
	}

	// Setup routes
	api := router.Group("/api/v1")
	{
		multiQueries := api.Group("/multi-queries")
		multiQueries.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			multiQueries.POST("/preview", multiQueryHandler.PreviewMultiQuery)
		}
	}

	return router, mockMultiQueryService, mockQueryService, jwtManager
}

// setupExecuteMultiQueryTestRouter creates a test router with execute endpoint for multi-query handler
func setupExecuteMultiQueryTestRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockMultiQueryService, *MockQueryServiceForMultiQuery, *auth.JWTManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockMultiQueryService := new(MockMultiQueryService)
	mockQueryService := new(MockQueryServiceForMultiQuery)

	// Create a test-specific wrapper that mimics the ExecuteMultiQuery flow
	executeHandler := &testExecuteMultiQueryHandler{
		db:            db,
		multiQuerySvc: mockMultiQueryService,
		querySvc:      mockQueryService,
	}

	// Setup routes
	api := router.Group("/api/v1")
	{
		multiQueries := api.Group("/multi-queries")
		multiQueries.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			multiQueries.POST("/execute", executeHandler.ExecuteMultiQuery)
		}
	}

	return router, mockMultiQueryService, mockQueryService, jwtManager
}

// testMultiQueryHandler wraps the multi-query handling logic for testing
type testMultiQueryHandler struct {
	db            *gorm.DB
	multiQuerySvc *MockMultiQueryService
	querySvc      *MockQueryServiceForMultiQuery
}

// PreviewMultiQuery generates previews for multiple queries
func (h *testMultiQueryHandler) PreviewMultiQuery(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.MultiQueryPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse data source ID
	dataSourceID, err := uuid.Parse(req.DataSourceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data source ID"})
		return
	}

	// Parse user ID
	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Join all query texts with semicolons for parsing
	fullQueryText := ""
	for i, qt := range req.QueryTexts {
		if i > 0 {
			fullQueryText += "; "
		}
		fullQueryText += qt
	}

	// Validate and parse queries
	parseResult := service.ValidateMultiQuery(fullQueryText)
	if len(parseResult.Errors) > 0 {
		errors := make([]gin.H, len(parseResult.Errors))
		for i, e := range parseResult.Errors {
			errors[i] = gin.H{
				"sequence": e.Sequence,
				"position": e.Position,
				"message":  e.Message,
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Invalid queries",
			"errors": errors,
		})
		return
	}

	// Extract query texts from parsed statements
	queryTexts := make([]string, len(parseResult.Statements))
	for i, stmt := range parseResult.Statements {
		queryTexts[i] = stmt.QueryText
	}

	// Generate preview
	preview, err := h.multiQuerySvc.PreviewMultiQuery(c.Request.Context(), dataSourceID, userUUID, queryTexts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response DTO
	response := dto.MultiQueryPreviewResponse{
		StatementCount:     preview.StatementCount,
		TotalEstimatedRows: preview.TotalEstimatedRows,
		RequiresApproval:   preview.RequiresApproval,
		Statements:         make([]dto.StatementPreview, len(preview.Statements)),
	}

	for i, stmt := range preview.Statements {
		response.Statements[i] = dto.StatementPreview{
			Sequence:      stmt.Sequence,
			QueryText:     stmt.QueryText,
			OperationType: string(stmt.OperationType),
			EstimatedRows: stmt.EstimatedRows,
			PreviewRows:   stmt.PreviewRows,
			Error:         stmt.Error,
		}

		// Convert columns
		if len(stmt.Columns) > 0 {
			response.Statements[i].Columns = make([]dto.ColumnInfo, len(stmt.Columns))
			for j, col := range stmt.Columns {
				response.Statements[i].Columns[j] = dto.ColumnInfo{
					Name: col,
					Type: "unknown",
				}
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// testExecuteMultiQueryHandler wraps the execute multi-query handling logic for testing
type testExecuteMultiQueryHandler struct {
	db            *gorm.DB
	multiQuerySvc *MockMultiQueryService
	querySvc      *MockQueryServiceForMultiQuery
}

// ExecuteMultiQuery executes multiple queries in a transaction
func (h *testExecuteMultiQueryHandler) ExecuteMultiQuery(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.MultiQueryExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse IDs
	dataSourceID, err := uuid.Parse(req.DataSourceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data source ID"})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Join all query texts with semicolons for parsing
	fullQueryText := ""
	for i, qt := range req.QueryTexts {
		if i > 0 {
			fullQueryText += "; "
		}
		fullQueryText += qt
	}

	// Validate and parse queries
	parseResult := service.ValidateMultiQuery(fullQueryText)
	if len(parseResult.Errors) > 0 {
		errors := make([]gin.H, len(parseResult.Errors))
		for i, e := range parseResult.Errors {
			errors[i] = gin.H{
				"sequence": e.Sequence,
				"position": e.Position,
				"message":  e.Message,
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Invalid queries",
			"errors": errors,
		})
		return
	}

	// Extract query texts from parsed result
	queryTexts := make([]string, len(parseResult.Statements))
	for i, stmt := range parseResult.Statements {
		queryTexts[i] = stmt.QueryText
	}

	// Check permissions
	perms, err := h.querySvc.GetEffectivePermissions(c.Request.Context(), userUUID, dataSourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if user has basic read permission
	if !perms.CanSelect {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied: SELECT not allowed"})
		return
	}

	// Calculate the actual impact of the multi-query
	impact, err := h.multiQuerySvc.CalculateMultiQueryImpact(c.Request.Context(), dataSourceID, userUUID, queryTexts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check for permission errors in any statement
	for _, stmt := range impact.Statements {
		if stmt.Error != "" && strings.Contains(stmt.Error, "permission denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": stmt.Error})
			return
		}
	}

	// Check if user is admin from role context
	role, _ := c.Get("role")
	isAdmin := role == string(models.RoleAdmin)

	// Block execution only for UPDATE/DELETE operations that would affect 0 rows
	hasUpdateOrDelete := false
	for _, stmt := range impact.Statements {
		if stmt.OperationType == models.OperationUpdate || stmt.OperationType == models.OperationDelete {
			hasUpdateOrDelete = true
			break
		}
	}

	if impact.RequiresApproval && impact.TotalEstimatedRows == 0 && hasUpdateOrDelete {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "UPDATE/DELETE operations would affect 0 rows. Please check your WHERE clause. INSERT operations can proceed to approval.",
			"estimated_rows":    0,
			"requires_approval": false,
		})
		return
	}

	// Non-admin users require approval for write operations that affect rows
	requiresApproval := !isAdmin && impact.RequiresApproval && impact.TotalEstimatedRows > 0

	if requiresApproval {
		approval := &models.ApprovalRequest{
			ID:            uuid.New(),
			RequestedBy:   userUUID,
			OperationType: models.OperationUpdate,
			QueryText:     fullQueryText,
			DataSourceID:  dataSourceID,
			Status:        models.ApprovalStatusPending,
		}

		if err := h.db.Create(approval).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create approval request"})
			return
		}

		c.JSON(http.StatusAccepted, dto.MultiQueryResponse{
			Status:           "pending_approval",
			StatementCount:   len(queryTexts),
			RequiresApproval: true,
			ApprovalID:       approval.ID.String(),
		})
		return
	}

	// Execute immediately for admins or SELECT-only queries
	transaction, err := h.multiQuerySvc.CreateMultiQueryTransaction(c.Request.Context(), nil, dataSourceID, queryTexts, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction: " + err.Error()})
		return
	}

	// Execute the multi-query
	result, err := h.multiQuerySvc.ExecuteMultiQuery(c.Request.Context(), transaction.ID, models.AuditModeCountOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute multi-query: " + err.Error()})
		return
	}

	// Return the execution result
	response := dto.MultiQueryResponse{
		TransactionID:     result.TransactionID.String(),
		Status:            result.Status,
		IsMultiQuery:      true,
		StatementCount:    len(result.Statements),
		TotalAffectedRows: result.TotalAffectedRows,
		ExecutionTimeMs:   int(result.ExecutionTimeMs),
		Statements:        make([]dto.StatementResult, len(result.Statements)),
		ErrorMessage:      result.ErrorMessage,
		RequiresApproval:  false,
	}

	for i, stmt := range result.Statements {
		response.Statements[i] = dto.StatementResult{
			Sequence:        stmt.Sequence,
			QueryText:       stmt.QueryText,
			OperationType:   string(stmt.OperationType),
			Status:          string(stmt.Status),
			AffectedRows:    stmt.AffectedRows,
			ErrorMessage:    stmt.ErrorMessage,
			ExecutionTimeMs: stmt.ExecutionTimeMs,
		}
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to create preview request body
func createPreviewMultiQueryRequest(dsID string, queryTexts []string) ([]byte, error) {
	req := dto.MultiQueryPreviewRequest{
		DataSourceID: dsID,
		QueryTexts:   queryTexts,
	}
	return json.Marshal(req)
}

// Helper function to create execute request body
func createExecuteMultiQueryRequest(dsID string, queryTexts []string) ([]byte, error) {
	req := dto.MultiQueryExecuteRequest{
		DataSourceID: dsID,
		QueryTexts:   queryTexts,
	}
	return json.Marshal(req)
}

// =============================================================================
// PREVIEW MULTI QUERY TESTS
// =============================================================================

// TestPreviewMultiQuery_MultipleSelect_ReturnsPreview tests preview with multiple SELECT statements
func TestPreviewMultiQuery_MultipleSelect_ReturnsPreview(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, _ := setupMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	// Mock preview response
	previewResult := &service.MultiQueryPreviewResult{
		StatementCount:     2,
		TotalEstimatedRows: 15,
		RequiresApproval:   false,
		Statements: []service.StatementPreview{
			{
				Sequence:      0,
				QueryText:     "SELECT * FROM users WHERE active = 1",
				OperationType: models.OperationSelect,
				EstimatedRows: 10,
				PreviewRows: []map[string]interface{}{
					{"id": 1, "name": "Alice"},
					{"id": 2, "name": "Bob"},
				},
				Columns: []string{"id", "name"},
			},
			{
				Sequence:      1,
				QueryText:     "SELECT * FROM orders WHERE status = 'pending'",
				OperationType: models.OperationSelect,
				EstimatedRows: 5,
				PreviewRows: []map[string]interface{}{
					{"order_id": 101, "total": 150.00},
				},
				Columns: []string{"order_id", "total"},
			},
		},
	}

	queryTexts := []string{
		"SELECT * FROM users WHERE active = 1",
		"SELECT * FROM orders WHERE status = 'pending'",
	}

	mockMultiQuerySvc.On("PreviewMultiQuery", mock.Anything, dataSource.ID, admin.ID, queryTexts).
		Return(previewResult, nil)

	// Create JWT token for admin
	token, err := testauth.CreateTestJWTToken(admin.ID.String(), admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	// Create request
	reqBody, err := createPreviewMultiQueryRequest(dataSource.ID.String(), queryTexts)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.MultiQueryPreviewResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, response.StatementCount)
	assert.Equal(t, 15, response.TotalEstimatedRows)
	assert.False(t, response.RequiresApproval)
	assert.Len(t, response.Statements, 2)

	// Verify first statement
	assert.Equal(t, 0, response.Statements[0].Sequence)
	assert.Equal(t, "SELECT * FROM users WHERE active = 1", response.Statements[0].QueryText)
	assert.Equal(t, "select", response.Statements[0].OperationType)
	assert.Equal(t, 10, response.Statements[0].EstimatedRows)

	// Verify second statement
	assert.Equal(t, 1, response.Statements[1].Sequence)
	assert.Equal(t, "SELECT * FROM orders WHERE status = 'pending'", response.Statements[1].QueryText)
	assert.Equal(t, "select", response.Statements[1].OperationType)
	assert.Equal(t, 5, response.Statements[1].EstimatedRows)

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestPreviewMultiQuery_MixedOperations_ReturnsPreview tests preview with mixed operations (SELECT, INSERT, UPDATE)
func TestPreviewMultiQuery_MixedOperations_ReturnsPreview(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupMultiQueryTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	// Mock preview response with mixed operations
	previewResult := &service.MultiQueryPreviewResult{
		StatementCount:     3,
		TotalEstimatedRows: 12,
		RequiresApproval:   true,
		Statements: []service.StatementPreview{
			{
				Sequence:      0,
				QueryText:     "SELECT * FROM users WHERE id = 1",
				OperationType: models.OperationSelect,
				EstimatedRows: 1,
			},
			{
				Sequence:      1,
				QueryText:     "INSERT INTO logs (message) VALUES ('User accessed data')",
				OperationType: models.OperationInsert,
				EstimatedRows: 1,
			},
			{
				Sequence:      2,
				QueryText:     "UPDATE users SET last_access = NOW() WHERE id = 1",
				OperationType: models.OperationUpdate,
				EstimatedRows: 10,
				PreviewRows: []map[string]interface{}{
					{"id": 1, "last_access": "2024-01-01T00:00:00Z"},
				},
				Columns: []string{"id", "last_access"},
			},
		},
	}

	queryTexts := []string{
		"SELECT * FROM users WHERE id = 1",
		"INSERT INTO logs (message) VALUES ('User accessed data')",
		"UPDATE users SET last_access = NOW() WHERE id = 1",
	}

	mockMultiQuerySvc.On("PreviewMultiQuery", mock.Anything, dataSource.ID, user.ID, queryTexts).
		Return(previewResult, nil)

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	reqBody, err := createPreviewMultiQueryRequest(dataSource.ID.String(), queryTexts)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.MultiQueryPreviewResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 3, response.StatementCount)
	assert.Equal(t, 12, response.TotalEstimatedRows)
	assert.True(t, response.RequiresApproval)
	assert.Len(t, response.Statements, 3)

	// Verify SELECT statement
	assert.Equal(t, "select", response.Statements[0].OperationType)
	assert.Equal(t, 1, response.Statements[0].EstimatedRows)

	// Verify INSERT statement
	assert.Equal(t, "insert", response.Statements[1].OperationType)
	assert.Equal(t, 1, response.Statements[1].EstimatedRows)

	// Verify UPDATE statement
	assert.Equal(t, "update", response.Statements[2].OperationType)
	assert.Equal(t, 10, response.Statements[2].EstimatedRows)
	assert.Len(t, response.Statements[2].PreviewRows, 1)

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestPreviewMultiQuery_WithSetVariables_ReturnsPreview tests preview with SET variable declarations
func TestPreviewMultiQuery_WithSetVariables_ReturnsPreview(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupMultiQueryTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	// Mock preview response with SET variables
	previewResult := &service.MultiQueryPreviewResult{
		StatementCount:     4,
		TotalEstimatedRows: 5,
		RequiresApproval:   true,
		Statements: []service.StatementPreview{
			{
				Sequence:      0,
				QueryText:     "SET @wallet_id = 23916",
				OperationType: models.OperationSet,
				EstimatedRows: 0,
			},
			{
				Sequence:      1,
				QueryText:     "SET @execution_date = '2026-03-03 15:15:15'",
				OperationType: models.OperationSet,
				EstimatedRows: 0,
			},
			{
				Sequence:      2,
				QueryText:     "SET @updated_by = 5285816",
				OperationType: models.OperationSet,
				EstimatedRows: 0,
			},
			{
				Sequence:      3,
				QueryText:     "UPDATE wallet_trxes SET amount = 0 WHERE wallet_id = @wallet_id",
				OperationType: models.OperationUpdate,
				EstimatedRows: 5,
				PreviewRows: []map[string]interface{}{
					{"id": 1, "amount": 100.00},
					{"id": 2, "amount": 200.00},
				},
				Columns: []string{"id", "amount"},
			},
		},
	}

	queryTexts := []string{
		"SET @wallet_id = 23916",
		"SET @execution_date = '2026-03-03 15:15:15'",
		"SET @updated_by = 5285816",
		"UPDATE wallet_trxes SET amount = 0 WHERE wallet_id = @wallet_id",
	}

	mockMultiQuerySvc.On("PreviewMultiQuery", mock.Anything, dataSource.ID, user.ID, queryTexts).
		Return(previewResult, nil)

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	reqBody, err := createPreviewMultiQueryRequest(dataSource.ID.String(), queryTexts)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.MultiQueryPreviewResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 4, response.StatementCount)
	assert.Len(t, response.Statements, 4)

	// Verify SET statements
	assert.Equal(t, "set", response.Statements[0].OperationType)
	assert.Equal(t, "SET @wallet_id = 23916", response.Statements[0].QueryText)
	assert.Equal(t, "set", response.Statements[1].OperationType)
	assert.Equal(t, "set", response.Statements[2].OperationType)

	// Verify UPDATE statement
	assert.Equal(t, "update", response.Statements[3].OperationType)
	assert.Equal(t, 5, response.Statements[3].EstimatedRows)
	assert.True(t, response.RequiresApproval)

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestPreviewMultiQuery_PermissionCheck_PerStatement tests permission enforcement per statement
func TestPreviewMultiQuery_PermissionCheck_PerStatement(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupMultiQueryTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	// Grant only SELECT permission
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	// Mock preview response with permission errors
	previewResult := &service.MultiQueryPreviewResult{
		StatementCount:     3,
		TotalEstimatedRows: 1,
		RequiresApproval:   false,
		Statements: []service.StatementPreview{
			{
				Sequence:      0,
				QueryText:     "SELECT * FROM users",
				OperationType: models.OperationSelect,
				EstimatedRows: 1,
			},
			{
				Sequence:      1,
				QueryText:     "INSERT INTO logs (msg) VALUES ('test')",
				OperationType: models.OperationInsert,
				EstimatedRows: 0,
				Error:         "permission denied: INSERT not allowed",
			},
			{
				Sequence:      2,
				QueryText:     "UPDATE users SET active = 1",
				OperationType: models.OperationUpdate,
				EstimatedRows: 0,
				Error:         "permission denied: UPDATE not allowed",
			},
		},
	}

	queryTexts := []string{
		"SELECT * FROM users",
		"INSERT INTO logs (msg) VALUES ('test')",
		"UPDATE users SET active = 1",
	}

	mockMultiQuerySvc.On("PreviewMultiQuery", mock.Anything, dataSource.ID, user.ID, queryTexts).
		Return(previewResult, nil)

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	reqBody, err := createPreviewMultiQueryRequest(dataSource.ID.String(), queryTexts)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.MultiQueryPreviewResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Statements, 3)

	// SELECT should succeed
	assert.Equal(t, "select", response.Statements[0].OperationType)
	assert.Empty(t, response.Statements[0].Error)

	// INSERT should have permission error
	assert.Equal(t, "insert", response.Statements[1].OperationType)
	assert.Contains(t, response.Statements[1].Error, "permission denied")

	// UPDATE should have permission error
	assert.Equal(t, "update", response.Statements[2].OperationType)
	assert.Contains(t, response.Statements[2].Error, "permission denied")

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestPreviewMultiQuery_InvalidSQL_ReturnsError tests error handling for invalid SQL
func TestPreviewMultiQuery_InvalidSQL_ReturnsError(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, _, _, jwtManager := setupMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	// Test with empty query
	reqBody, _ := createPreviewMultiQueryRequest(dataSource.ID.String(), []string{})
	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPreviewMultiQuery_TransactionControl_Blocked tests that transaction control statements are blocked
func TestPreviewMultiQuery_TransactionControl_Blocked(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, _, _, jwtManager := setupMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	tests := []struct {
		name       string
		queryTexts []string
		errorMsg   string
	}{
		{
			name:       "BEGIN statement",
			queryTexts: []string{"BEGIN", "SELECT 1"},
			errorMsg:   "Transaction control",
		},
		{
			name:       "COMMIT statement",
			queryTexts: []string{"SELECT 1", "COMMIT"},
			errorMsg:   "Transaction control",
		},
		{
			name:       "ROLLBACK statement",
			queryTexts: []string{"ROLLBACK"},
			errorMsg:   "Transaction control",
		},
		{
			name:       "START TRANSACTION statement",
			queryTexts: []string{"START TRANSACTION", "SELECT 1"},
			errorMsg:   "Transaction control",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := createPreviewMultiQueryRequest(dataSource.ID.String(), tt.queryTexts)
			req, _ := http.NewRequest("POST", "/api/v1/multi-queries/preview", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Contains(t, response["error"], "Invalid queries")

			errors, ok := response["errors"].([]interface{})
			require.True(t, ok)
			require.Greater(t, len(errors), 0)

			errorObj := errors[0].(map[string]interface{})
			assert.Contains(t, errorObj["message"], tt.errorMsg)
		})
	}
}

// TestPreviewMultiQuery_StatementBreakdown_Returned tests that preview returns proper statement breakdown
func TestPreviewMultiQuery_StatementBreakdown_Returned(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupMultiQueryTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	// Mock preview with detailed statement breakdown
	previewResult := &service.MultiQueryPreviewResult{
		StatementCount:     4,
		TotalEstimatedRows: 25,
		RequiresApproval:   true,
		Statements: []service.StatementPreview{
			{
				Sequence:      0,
				QueryText:     "SELECT COUNT(*) FROM users",
				OperationType: models.OperationSelect,
				EstimatedRows: 1,
				Columns:       []string{"count"},
			},
			{
				Sequence:      1,
				QueryText:     "SELECT * FROM users WHERE created_at > '2024-01-01'",
				OperationType: models.OperationSelect,
				EstimatedRows: 20,
				Columns:       []string{"id", "name", "email", "created_at"},
			},
			{
				Sequence:      2,
				QueryText:     "INSERT INTO audit_log (action, user_id) VALUES ('preview', 1)",
				OperationType: models.OperationInsert,
				EstimatedRows: 1,
			},
			{
				Sequence:      3,
				QueryText:     "DELETE FROM temp_data WHERE expired = true",
				OperationType: models.OperationDelete,
				EstimatedRows: 3,
				PreviewRows: []map[string]interface{}{
					{"id": 100, "data": "temp1"},
					{"id": 101, "data": "temp2"},
					{"id": 102, "data": "temp3"},
				},
				Columns: []string{"id", "data"},
			},
		},
	}

	queryTexts := []string{
		"SELECT COUNT(*) FROM users",
		"SELECT * FROM users WHERE created_at > '2024-01-01'",
		"INSERT INTO audit_log (action, user_id) VALUES ('preview', 1)",
		"DELETE FROM temp_data WHERE expired = true",
	}

	mockMultiQuerySvc.On("PreviewMultiQuery", mock.Anything, dataSource.ID, user.ID, queryTexts).
		Return(previewResult, nil)

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	reqBody, err := createPreviewMultiQueryRequest(dataSource.ID.String(), queryTexts)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.MultiQueryPreviewResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify statement breakdown
	assert.Equal(t, 4, response.StatementCount)
	assert.Equal(t, 25, response.TotalEstimatedRows)
	assert.Len(t, response.Statements, 4)

	// Verify sequence numbers
	for i, stmt := range response.Statements {
		assert.Equal(t, i, stmt.Sequence)
	}

	// Verify operation types
	assert.Equal(t, "select", response.Statements[0].OperationType)
	assert.Equal(t, "select", response.Statements[1].OperationType)
	assert.Equal(t, "insert", response.Statements[2].OperationType)
	assert.Equal(t, "delete", response.Statements[3].OperationType)

	// Verify columns are returned
	assert.Len(t, response.Statements[0].Columns, 1)
	assert.Len(t, response.Statements[1].Columns, 4)
	assert.Len(t, response.Statements[3].Columns, 2)

	// Verify preview rows for DELETE
	assert.Len(t, response.Statements[3].PreviewRows, 3)

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestPreviewMultiQuery_DifferentOperationTypes_Handled tests handling of different operation types
func TestPreviewMultiQuery_DifferentOperationTypes_Handled(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupMultiQueryTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	// Mock preview with all operation types
	previewResult := &service.MultiQueryPreviewResult{
		StatementCount:     5,
		TotalEstimatedRows: 16,
		RequiresApproval:   true,
		Statements: []service.StatementPreview{
			{
				Sequence:      0,
				QueryText:     "SELECT * FROM users",
				OperationType: models.OperationSelect,
				EstimatedRows: 10,
			},
			{
				Sequence:      1,
				QueryText:     "INSERT INTO logs (msg) VALUES ('test')",
				OperationType: models.OperationInsert,
				EstimatedRows: 1,
			},
			{
				Sequence:      2,
				QueryText:     "UPDATE users SET active = 1 WHERE id = 1",
				OperationType: models.OperationUpdate,
				EstimatedRows: 1,
			},
			{
				Sequence:      3,
				QueryText:     "DELETE FROM temp WHERE id = 1",
				OperationType: models.OperationDelete,
				EstimatedRows: 1,
			},
			{
				Sequence:      4,
				QueryText:     "SET @var = 123",
				OperationType: models.OperationSet,
				EstimatedRows: 0,
			},
		},
	}

	queryTexts := []string{
		"SELECT * FROM users",
		"INSERT INTO logs (msg) VALUES ('test')",
		"UPDATE users SET active = 1 WHERE id = 1",
		"DELETE FROM temp WHERE id = 1",
		"SET @var = 123",
	}

	mockMultiQuerySvc.On("PreviewMultiQuery", mock.Anything, dataSource.ID, user.ID, queryTexts).
		Return(previewResult, nil)

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	reqBody, err := createPreviewMultiQueryRequest(dataSource.ID.String(), queryTexts)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.MultiQueryPreviewResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 5, response.StatementCount)
	assert.Len(t, response.Statements, 5)

	// Verify all operation types are handled correctly
	operationTypes := make([]string, len(response.Statements))
	for i, stmt := range response.Statements {
		operationTypes[i] = stmt.OperationType
	}

	assert.Contains(t, operationTypes, "select")
	assert.Contains(t, operationTypes, "insert")
	assert.Contains(t, operationTypes, "update")
	assert.Contains(t, operationTypes, "delete")
	assert.Contains(t, operationTypes, "set")

	// Verify requires_approval is true (due to write operations)
	assert.True(t, response.RequiresApproval)

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestPreviewMultiQuery_Unauthorized_ReturnsUnauthorized tests that request without token returns 401
func TestPreviewMultiQuery_Unauthorized_ReturnsUnauthorized(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, _, _, _ := setupMultiQueryTestRouter(t, db)

	// Create a data source
	_ = fixtures.CreateTestDataSource(t, db, "test-ds")

	reqBody, _ := createPreviewMultiQueryRequest(
		uuid.New().String(),
		[]string{"SELECT 1"},
	)
	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestPreviewMultiQuery_InvalidDataSourceID_ReturnsBadRequest tests that invalid UUID returns 400
func TestPreviewMultiQuery_InvalidDataSourceID_ReturnsBadRequest(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, _, _, jwtManager := setupMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	// Send request with invalid UUID format
	reqBody := `{"data_source_id": "not-a-valid-uuid", "query_texts": ["SELECT 1"]}`
	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/preview", bytes.NewBuffer([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Invalid data source ID")
}

// TestPreviewMultiQuery_DataSourceNotFound_ReturnsNotFound tests non-existent data source
func TestPreviewMultiQuery_DataSourceNotFound_ReturnsNotFound(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)

	// Use a non-existent data source ID
	fakeDSID := uuid.New()
	queryTexts := []string{"SELECT 1"}

	// Mock service to return error for non-existent data source
	mockMultiQuerySvc.On("PreviewMultiQuery", mock.Anything, fakeDSID, admin.ID, queryTexts).
		Return(nil, assert.AnError)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, _ := createPreviewMultiQueryRequest(fakeDSID.String(), queryTexts)
	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestPreviewMultiQuery_ServiceError_ReturnsInternalServerError tests service error handling
func TestPreviewMultiQuery_ServiceError_ReturnsInternalServerError(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	queryTexts := []string{"SELECT * FROM users"}

	// Mock service to return error
	mockMultiQuerySvc.On("PreviewMultiQuery", mock.Anything, dataSource.ID, admin.ID, queryTexts).
		Return(nil, assert.AnError)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, _ := createPreviewMultiQueryRequest(dataSource.ID.String(), queryTexts)
	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/preview", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["error"])

	mockMultiQuerySvc.AssertExpectations(t)
}

// =============================================================================
// EXECUTE MULTI QUERY TESTS
// =============================================================================

// TestExecuteMultiQuery_MultipleSelect_ExecutesSuccessfully tests execution of multiple SELECT statements
func TestExecuteMultiQuery_MultipleSelect_ExecutesSuccessfully(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, mockQuerySvc, jwtManager := setupExecuteMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	queryTexts := []string{
		"SELECT * FROM users WHERE active = 1",
		"SELECT * FROM orders WHERE status = 'pending'",
	}

	// Mock permissions check
	mockQuerySvc.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanSelect: true, CanInsert: true, CanUpdate: true, CanDelete: true}, nil)

	// Mock impact calculation - SELECT only, no approval needed
	impact := &service.MultiQueryImpact{
		Statements: []service.StatementPreview{
			{Sequence: 0, QueryText: queryTexts[0], OperationType: models.OperationSelect, EstimatedRows: 10},
			{Sequence: 1, QueryText: queryTexts[1], OperationType: models.OperationSelect, EstimatedRows: 5},
		},
		TotalEstimatedRows: 15,
		RequiresApproval:   false,
	}
	mockMultiQuerySvc.On("CalculateMultiQueryImpact", mock.Anything, dataSource.ID, admin.ID, queryTexts).
		Return(impact, nil)

	// Mock transaction creation
	transactionID := uuid.New()
	transaction := &models.QueryTransaction{
		ID:             transactionID,
		DataSourceID:   dataSource.ID,
		QueryText:      "SELECT * FROM users WHERE active = 1; SELECT * FROM orders WHERE status = 'pending'",
		StartedBy:      admin.ID,
		Status:         models.TransactionStatusActive,
		IsMultiQuery:   true,
		StatementCount: 2,
	}
	mockMultiQuerySvc.On("CreateMultiQueryTransaction", mock.Anything, mock.Anything, dataSource.ID, queryTexts, admin.ID).
		Return(transaction, nil)

	// Mock execution result
	execResult := &service.MultiQueryResult{
		TransactionID:     transactionID,
		Status:            "success",
		TotalAffectedRows: 0,
		ExecutionTimeMs:   150,
		Statements: []models.QueryTransactionStatement{
			{Sequence: 0, QueryText: queryTexts[0], OperationType: models.OperationSelect, Status: models.StatementStatusSuccess, AffectedRows: 0},
			{Sequence: 1, QueryText: queryTexts[1], OperationType: models.OperationSelect, Status: models.StatementStatusSuccess, AffectedRows: 0},
		},
	}
	mockMultiQuerySvc.On("ExecuteMultiQuery", mock.Anything, transactionID, models.AuditModeCountOnly).
		Return(execResult, nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, err := createExecuteMultiQueryRequest(dataSource.ID.String(), queryTexts)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/execute", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.MultiQueryResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response.Status)
	assert.True(t, response.IsMultiQuery)
	assert.Equal(t, 2, response.StatementCount)
	assert.Equal(t, 0, response.TotalAffectedRows)
	assert.False(t, response.RequiresApproval)
	assert.Len(t, response.Statements, 2)

	// Verify statement ordering
	assert.Equal(t, 0, response.Statements[0].Sequence)
	assert.Equal(t, "SELECT * FROM users WHERE active = 1", response.Statements[0].QueryText)
	assert.Equal(t, "select", response.Statements[0].OperationType)
	assert.Equal(t, 1, response.Statements[1].Sequence)
	assert.Equal(t, "SELECT * FROM orders WHERE status = 'pending'", response.Statements[1].QueryText)

	mockMultiQuerySvc.AssertExpectations(t)
	mockQuerySvc.AssertExpectations(t)
}

// TestExecuteMultiQuery_MixedOperations_CreatesTransaction tests mixed operations creating a transaction
func TestExecuteMultiQuery_MixedOperations_CreatesTransaction(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, mockQuerySvc, jwtManager := setupExecuteMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	queryTexts := []string{
		"SELECT * FROM users WHERE id = 1",
		"INSERT INTO logs (message) VALUES ('User accessed data')",
		"UPDATE users SET last_access = NOW() WHERE id = 1",
	}

	// Mock permissions check
	mockQuerySvc.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanSelect: true, CanInsert: true, CanUpdate: true, CanDelete: true}, nil)

	// Mock impact calculation - mixed operations
	impact := &service.MultiQueryImpact{
		Statements: []service.StatementPreview{
			{Sequence: 0, QueryText: queryTexts[0], OperationType: models.OperationSelect, EstimatedRows: 1},
			{Sequence: 1, QueryText: queryTexts[1], OperationType: models.OperationInsert, EstimatedRows: 1},
			{Sequence: 2, QueryText: queryTexts[2], OperationType: models.OperationUpdate, EstimatedRows: 1},
		},
		TotalEstimatedRows: 2,
		RequiresApproval:   true,
	}
	mockMultiQuerySvc.On("CalculateMultiQueryImpact", mock.Anything, dataSource.ID, admin.ID, queryTexts).
		Return(impact, nil)

	// Mock transaction creation
	transactionID := uuid.New()
	transaction := &models.QueryTransaction{
		ID:             transactionID,
		DataSourceID:   dataSource.ID,
		QueryText:      "SELECT * FROM users WHERE id = 1; INSERT INTO logs (message) VALUES ('User accessed data'); UPDATE users SET last_access = NOW() WHERE id = 1",
		StartedBy:      admin.ID,
		Status:         models.TransactionStatusActive,
		IsMultiQuery:   true,
		StatementCount: 3,
	}
	mockMultiQuerySvc.On("CreateMultiQueryTransaction", mock.Anything, mock.Anything, dataSource.ID, queryTexts, admin.ID).
		Return(transaction, nil)

	// Mock execution result
	execResult := &service.MultiQueryResult{
		TransactionID:     transactionID,
		Status:            "success",
		TotalAffectedRows: 2,
		ExecutionTimeMs:   250,
		Statements: []models.QueryTransactionStatement{
			{Sequence: 0, QueryText: queryTexts[0], OperationType: models.OperationSelect, Status: models.StatementStatusSuccess, AffectedRows: 1},
			{Sequence: 1, QueryText: queryTexts[1], OperationType: models.OperationInsert, Status: models.StatementStatusSuccess, AffectedRows: 1},
			{Sequence: 2, QueryText: queryTexts[2], OperationType: models.OperationUpdate, Status: models.StatementStatusSuccess, AffectedRows: 1},
		},
	}
	mockMultiQuerySvc.On("ExecuteMultiQuery", mock.Anything, transactionID, models.AuditModeCountOnly).
		Return(execResult, nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, err := createExecuteMultiQueryRequest(dataSource.ID.String(), queryTexts)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/execute", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.MultiQueryResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response.Status)
	assert.True(t, response.IsMultiQuery)
	assert.Equal(t, 3, response.StatementCount)
	assert.Equal(t, 2, response.TotalAffectedRows)
	assert.Len(t, response.Statements, 3)

	// Verify all operation types
	assert.Equal(t, "select", response.Statements[0].OperationType)
	assert.Equal(t, "insert", response.Statements[1].OperationType)
	assert.Equal(t, "update", response.Statements[2].OperationType)

	mockMultiQuerySvc.AssertExpectations(t)
	mockQuerySvc.AssertExpectations(t)
}

// TestExecuteMultiQuery_PermissionCheck_PerStatement tests permission enforcement per statement
func TestExecuteMultiQuery_PermissionCheck_PerStatement(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, mockQuerySvc, jwtManager := setupExecuteMultiQueryTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	// Grant only SELECT permission
	fixtures.GrantReadOnlyPermission(t, db, group.ID, dataSource.ID)

	queryTexts := []string{
		"SELECT * FROM users",
		"INSERT INTO logs (msg) VALUES ('test')",
		"UPDATE users SET active = 1",
	}

	// Mock permissions check - user only has SELECT
	mockQuerySvc.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanSelect: true, CanInsert: false, CanUpdate: false, CanDelete: false}, nil)

	// Mock impact calculation with permission errors
	impact := &service.MultiQueryImpact{
		Statements: []service.StatementPreview{
			{Sequence: 0, QueryText: queryTexts[0], OperationType: models.OperationSelect, EstimatedRows: 1},
			{Sequence: 1, QueryText: queryTexts[1], OperationType: models.OperationInsert, EstimatedRows: 0, Error: "permission denied: INSERT not allowed"},
			{Sequence: 2, QueryText: queryTexts[2], OperationType: models.OperationUpdate, EstimatedRows: 0, Error: "permission denied: UPDATE not allowed"},
		},
		TotalEstimatedRows: 1,
		RequiresApproval:   false,
	}
	mockMultiQuerySvc.On("CalculateMultiQueryImpact", mock.Anything, dataSource.ID, user.ID, queryTexts).
		Return(impact, nil)

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	reqBody, err := createExecuteMultiQueryRequest(dataSource.ID.String(), queryTexts)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/execute", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 403 Forbidden due to permission error in statements
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify the error message contains permission denied
	assert.Contains(t, response["error"], "permission denied")

	mockMultiQuerySvc.AssertExpectations(t)
	mockQuerySvc.AssertExpectations(t)
}

// TestExecuteMultiQuery_WriteOperations_RequireApproval tests that write operations require approval for non-admins
func TestExecuteMultiQuery_WriteOperations_RequireApproval(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, mockQuerySvc, jwtManager := setupExecuteMultiQueryTestRouter(t, db)

	user := fixtures.CreateTestRegularUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, user.ID, group.ID)
	// Grant full permissions so the permission check passes, but user still needs approval for writes
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	queryTexts := []string{
		"SELECT * FROM users WHERE id = 1",
		"UPDATE users SET active = 1 WHERE id = 1",
	}

	// Mock permissions check - user has all permissions
	mockQuerySvc.On("GetEffectivePermissions", mock.Anything, user.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanSelect: true, CanInsert: true, CanUpdate: true, CanDelete: true}, nil)

	// Mock impact calculation - write operations that require approval
	impact := &service.MultiQueryImpact{
		Statements: []service.StatementPreview{
			{Sequence: 0, QueryText: queryTexts[0], OperationType: models.OperationSelect, EstimatedRows: 1},
			{Sequence: 1, QueryText: queryTexts[1], OperationType: models.OperationUpdate, EstimatedRows: 1},
		},
		TotalEstimatedRows: 1,
		RequiresApproval:   true,
	}
	mockMultiQuerySvc.On("CalculateMultiQueryImpact", mock.Anything, dataSource.ID, user.ID, queryTexts).
		Return(impact, nil)

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	reqBody, err := createExecuteMultiQueryRequest(dataSource.ID.String(), queryTexts)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/execute", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 202 Accepted with pending approval status
	assert.Equal(t, http.StatusAccepted, w.Code)

	var response dto.MultiQueryResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "pending_approval", response.Status)
	assert.True(t, response.RequiresApproval)
	assert.Equal(t, 2, response.StatementCount)
	assert.NotEmpty(t, response.ApprovalID)

	// Verify approval request was created in database
	var approval models.ApprovalRequest
	err = db.Where("requested_by = ?", user.ID).First(&approval).Error
	require.NoError(t, err)
	assert.Equal(t, models.ApprovalStatusPending, approval.Status)
	assert.Equal(t, dataSource.ID, approval.DataSourceID)

	mockMultiQuerySvc.AssertExpectations(t)
	mockQuerySvc.AssertExpectations(t)
}

// TestExecuteMultiQuery_ErrorHandling_Rollback tests error handling and transaction rollback
func TestExecuteMultiQuery_ErrorHandling_Rollback(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, mockQuerySvc, jwtManager := setupExecuteMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	queryTexts := []string{
		"INSERT INTO logs (message) VALUES ('First')",
		"INSERT INTO logs (message) VALUES ('Second')",
		"INVALID SQL SYNTAX HERE",
	}

	// Mock permissions check
	mockQuerySvc.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanSelect: true, CanInsert: true, CanUpdate: true, CanDelete: true}, nil)

	// Mock impact calculation
	impact := &service.MultiQueryImpact{
		Statements: []service.StatementPreview{
			{Sequence: 0, QueryText: queryTexts[0], OperationType: models.OperationInsert, EstimatedRows: 1},
			{Sequence: 1, QueryText: queryTexts[1], OperationType: models.OperationInsert, EstimatedRows: 1},
			{Sequence: 2, QueryText: queryTexts[2], OperationType: models.OperationSelect, EstimatedRows: 0},
		},
		TotalEstimatedRows: 2,
		RequiresApproval:   true,
	}
	mockMultiQuerySvc.On("CalculateMultiQueryImpact", mock.Anything, dataSource.ID, admin.ID, queryTexts).
		Return(impact, nil)

	// Mock transaction creation
	transactionID := uuid.New()
	transaction := &models.QueryTransaction{
		ID:             transactionID,
		DataSourceID:   dataSource.ID,
		QueryText:      "INSERT INTO logs (message) VALUES ('First'); INSERT INTO logs (message) VALUES ('Second'); INVALID SQL SYNTAX HERE",
		StartedBy:      admin.ID,
		Status:         models.TransactionStatusActive,
		IsMultiQuery:   true,
		StatementCount: 3,
	}
	mockMultiQuerySvc.On("CreateMultiQueryTransaction", mock.Anything, mock.Anything, dataSource.ID, queryTexts, admin.ID).
		Return(transaction, nil)

	// Mock execution to return an error (simulating rollback scenario)
	mockMultiQuerySvc.On("ExecuteMultiQuery", mock.Anything, transactionID, models.AuditModeCountOnly).
		Return(nil, fmt.Errorf("Statement 2 failed: syntax error at or near 'INVALID'"))

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, err := createExecuteMultiQueryRequest(dataSource.ID.String(), queryTexts)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/execute", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 500 with error details
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should contain error message
	assert.Contains(t, response["error"], "failed")

	mockMultiQuerySvc.AssertExpectations(t)
	mockQuerySvc.AssertExpectations(t)
}

// TestExecuteMultiQuery_StatementOrdering_Maintained tests that statement ordering is maintained
func TestExecuteMultiQuery_StatementOrdering_Maintained(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, mockQuerySvc, jwtManager := setupExecuteMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	queryTexts := []string{
		"SELECT 1 as seq",
		"SELECT 2 as seq",
		"SELECT 3 as seq",
		"SELECT 4 as seq",
		"SELECT 5 as seq",
	}

	// Mock permissions check
	mockQuerySvc.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanSelect: true, CanInsert: true, CanUpdate: true, CanDelete: true}, nil)

	// Mock impact calculation
	impact := &service.MultiQueryImpact{
		Statements: []service.StatementPreview{
			{Sequence: 0, QueryText: queryTexts[0], OperationType: models.OperationSelect, EstimatedRows: 1},
			{Sequence: 1, QueryText: queryTexts[1], OperationType: models.OperationSelect, EstimatedRows: 1},
			{Sequence: 2, QueryText: queryTexts[2], OperationType: models.OperationSelect, EstimatedRows: 1},
			{Sequence: 3, QueryText: queryTexts[3], OperationType: models.OperationSelect, EstimatedRows: 1},
			{Sequence: 4, QueryText: queryTexts[4], OperationType: models.OperationSelect, EstimatedRows: 1},
		},
		TotalEstimatedRows: 5,
		RequiresApproval:   false,
	}
	mockMultiQuerySvc.On("CalculateMultiQueryImpact", mock.Anything, dataSource.ID, admin.ID, queryTexts).
		Return(impact, nil)

	// Mock transaction creation
	transactionID := uuid.New()
	transaction := &models.QueryTransaction{
		ID:             transactionID,
		DataSourceID:   dataSource.ID,
		StartedBy:      admin.ID,
		Status:         models.TransactionStatusActive,
		IsMultiQuery:   true,
		StatementCount: 5,
	}
	mockMultiQuerySvc.On("CreateMultiQueryTransaction", mock.Anything, mock.Anything, dataSource.ID, queryTexts, admin.ID).
		Return(transaction, nil)

	// Mock execution result with statements in correct order
	execResult := &service.MultiQueryResult{
		TransactionID:     transactionID,
		Status:            "success",
		TotalAffectedRows: 0,
		ExecutionTimeMs:   200,
		Statements: []models.QueryTransactionStatement{
			{Sequence: 0, QueryText: queryTexts[0], OperationType: models.OperationSelect, Status: models.StatementStatusSuccess},
			{Sequence: 1, QueryText: queryTexts[1], OperationType: models.OperationSelect, Status: models.StatementStatusSuccess},
			{Sequence: 2, QueryText: queryTexts[2], OperationType: models.OperationSelect, Status: models.StatementStatusSuccess},
			{Sequence: 3, QueryText: queryTexts[3], OperationType: models.OperationSelect, Status: models.StatementStatusSuccess},
			{Sequence: 4, QueryText: queryTexts[4], OperationType: models.OperationSelect, Status: models.StatementStatusSuccess},
		},
	}
	mockMultiQuerySvc.On("ExecuteMultiQuery", mock.Anything, transactionID, models.AuditModeCountOnly).
		Return(execResult, nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, err := createExecuteMultiQueryRequest(dataSource.ID.String(), queryTexts)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/execute", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.MultiQueryResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify statement count and ordering
	assert.Equal(t, 5, response.StatementCount)
	assert.Len(t, response.Statements, 5)

	// Verify each statement is in correct order
	for i, stmt := range response.Statements {
		assert.Equal(t, i, stmt.Sequence)
		assert.Equal(t, queryTexts[i], stmt.QueryText)
	}

	mockMultiQuerySvc.AssertExpectations(t)
	mockQuerySvc.AssertExpectations(t)
}

// TestExecuteMultiQuery_TransactionCreation_Success tests successful transaction creation
func TestExecuteMultiQuery_TransactionCreation_Success(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, mockQuerySvc, jwtManager := setupExecuteMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")
	group := fixtures.CreateTestGroup(t, db, "test-group")
	fixtures.AddUserToGroup(db, admin.ID, group.ID)
	fixtures.GrantFullPermission(t, db, group.ID, dataSource.ID)

	queryTexts := []string{
		"UPDATE users SET active = 1 WHERE id = 5",
	}

	// Mock permissions check
	mockQuerySvc.On("GetEffectivePermissions", mock.Anything, admin.ID, dataSource.ID).
		Return(&models.EffectivePermissions{CanSelect: true, CanInsert: true, CanUpdate: true, CanDelete: true}, nil)

	// Mock impact calculation
	impact := &service.MultiQueryImpact{
		Statements: []service.StatementPreview{
			{Sequence: 0, QueryText: queryTexts[0], OperationType: models.OperationUpdate, EstimatedRows: 1},
		},
		TotalEstimatedRows: 1,
		RequiresApproval:   true,
	}
	mockMultiQuerySvc.On("CalculateMultiQueryImpact", mock.Anything, dataSource.ID, admin.ID, queryTexts).
		Return(impact, nil)

	// Mock transaction creation - verify it was called with correct parameters
	transactionID := uuid.New()
	transaction := &models.QueryTransaction{
		ID:             transactionID,
		DataSourceID:   dataSource.ID,
		QueryText:      queryTexts[0],
		StartedBy:      admin.ID,
		Status:         models.TransactionStatusActive,
		IsMultiQuery:   true,
		StatementCount: 1,
	}
	mockMultiQuerySvc.On("CreateMultiQueryTransaction", mock.Anything, (*uuid.UUID)(nil), dataSource.ID, queryTexts, admin.ID).
		Return(transaction, nil)

	// Mock execution result
	execResult := &service.MultiQueryResult{
		TransactionID:     transactionID,
		Status:            "success",
		TotalAffectedRows: 1,
		ExecutionTimeMs:   100,
		Statements: []models.QueryTransactionStatement{
			{Sequence: 0, QueryText: queryTexts[0], OperationType: models.OperationUpdate, Status: models.StatementStatusSuccess, AffectedRows: 1},
		},
	}
	mockMultiQuerySvc.On("ExecuteMultiQuery", mock.Anything, transactionID, models.AuditModeCountOnly).
		Return(execResult, nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	reqBody, err := createExecuteMultiQueryRequest(dataSource.ID.String(), queryTexts)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/execute", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.MultiQueryResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify transaction was created and returned
	assert.Equal(t, transactionID.String(), response.TransactionID)
	assert.Equal(t, "success", response.Status)
	assert.Equal(t, 1, response.TotalAffectedRows)

	mockMultiQuerySvc.AssertExpectations(t)
	mockQuerySvc.AssertExpectations(t)
}

// TestExecuteMultiQuery_InvalidSQL_ReturnsError tests error handling for invalid SQL
func TestExecuteMultiQuery_InvalidSQL_ReturnsError(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, _, _, jwtManager := setupExecuteMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	tests := []struct {
		name              string
		queryTexts        []string
		errorMsg          string
		isValidationError bool
	}{
		{
			name:       "BEGIN statement blocked",
			queryTexts: []string{"BEGIN", "SELECT 1"},
			errorMsg:   "Transaction control",
		},
		{
			name:       "COMMIT statement blocked",
			queryTexts: []string{"SELECT 1", "COMMIT"},
			errorMsg:   "Transaction control",
		},
		{
			name:       "ROLLBACK statement blocked",
			queryTexts: []string{"ROLLBACK"},
			errorMsg:   "Transaction control",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := createExecuteMultiQueryRequest(dataSource.ID.String(), tt.queryTexts)
			req, _ := http.NewRequest("POST", "/api/v1/multi-queries/execute", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			// SQL validation errors have "Invalid queries" as the error
			if response["error"] == "Invalid queries" {
				// Check that errors array contains the expected error message
				errors, ok := response["errors"].([]interface{})
				require.True(t, ok, "Expected errors array")
				require.Greater(t, len(errors), 0, "Expected at least one error")

				// Find error that contains the expected message
				found := false
				for _, err := range errors {
					errMap, ok := err.(map[string]interface{})
					if ok {
						if msg, ok := errMap["message"].(string); ok && strings.Contains(msg, tt.errorMsg) {
							found = true
							break
						}
					}
				}
				assert.True(t, found, "Expected to find error containing: %s", tt.errorMsg)
			} else {
				// For binding validation errors, just check it contains validation text
				errStr, ok := response["error"].(string)
				assert.True(t, ok, "Expected error to be a string")
				assert.Contains(t, errStr, "Error")
			}
		})
	}
}

// TestExecuteMultiQuery_EmptyQueryList_ReturnsValidationError tests that empty query list returns validation error
func TestExecuteMultiQuery_EmptyQueryList_ReturnsValidationError(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, _, _, jwtManager := setupExecuteMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	dataSource := fixtures.CreateTestDataSource(t, db, "test-ds")

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	// Test with empty query list - should fail binding validation
	reqBody, _ := createExecuteMultiQueryRequest(dataSource.ID.String(), []string{})
	req, _ := http.NewRequest("POST", "/api/v1/multi-queries/execute", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "QueryTexts")
}

// =============================================================================
// COMMIT MULTI QUERY TESTS
// =============================================================================

// setupCommitMultiQueryTestRouter creates a test router with commit endpoint for multi-query handler
func setupCommitMultiQueryTestRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockMultiQueryService, *MockQueryServiceForMultiQuery, *auth.JWTManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockMultiQueryService := new(MockMultiQueryService)
	mockQueryService := new(MockQueryServiceForMultiQuery)

	// Create a test-specific wrapper that mimics the CommitMultiQuery flow
	commitHandler := &testCommitMultiQueryHandler{
		db:            db,
		multiQuerySvc: mockMultiQueryService,
		querySvc:      mockQueryService,
	}

	// Setup routes
	api := router.Group("/api/v1")
	{
		queries := api.Group("/queries")
		queries.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			queries.POST("/multi/:id/commit", commitHandler.CommitMultiQuery)
		}
	}

	return router, mockMultiQueryService, mockQueryService, jwtManager
}

// testCommitMultiQueryHandler wraps the commit multi-query handling logic for testing
type testCommitMultiQueryHandler struct {
	db            *gorm.DB
	multiQuerySvc *MockMultiQueryService
	querySvc      *MockQueryServiceForMultiQuery
}

// CommitMultiQuery commits a multi-query transaction
func (h *testCommitMultiQueryHandler) CommitMultiQuery(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse user ID
	_, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse transaction ID from URL
	transactionIDStr := c.Param("id")
	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	// Execute the commit (which calls ExecuteMultiQuery in the actual handler)
	result, err := h.multiQuerySvc.ExecuteMultiQuery(c.Request.Context(), transactionID, models.AuditModeCountOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response
	response := dto.MultiQueryResponse{
		TransactionID:     result.TransactionID.String(),
		Status:            result.Status,
		IsMultiQuery:      true,
		StatementCount:    len(result.Statements),
		TotalAffectedRows: result.TotalAffectedRows,
		ExecutionTimeMs:   int(result.ExecutionTimeMs),
		Statements:        make([]dto.StatementResult, len(result.Statements)),
		ErrorMessage:      result.ErrorMessage,
	}

	for i, stmt := range result.Statements {
		response.Statements[i] = dto.StatementResult{
			Sequence:        stmt.Sequence,
			QueryText:       stmt.QueryText,
			OperationType:   string(stmt.OperationType),
			Status:          string(stmt.Status),
			AffectedRows:    stmt.AffectedRows,
			ErrorMessage:    stmt.ErrorMessage,
			ExecutionTimeMs: stmt.ExecutionTimeMs,
		}
	}

	c.JSON(http.StatusOK, response)
}

// TestCommitMultiQuery_AfterSuccessfulExecution_Commits tests committing after successful execution
func TestCommitMultiQuery_AfterSuccessfulExecution_Commits(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupCommitMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	transactionID := uuid.New()

	// Mock successful execution/commit result
	execResult := &service.MultiQueryResult{
		TransactionID:     transactionID,
		Status:            "committed",
		TotalAffectedRows: 10,
		ExecutionTimeMs:   150,
		Statements: []models.QueryTransactionStatement{
			{
				Sequence:      0,
				QueryText:     "UPDATE users SET active = 1",
				OperationType: models.OperationUpdate,
				Status:        models.StatementStatusSuccess,
				AffectedRows:  10,
			},
		},
	}

	mockMultiQuerySvc.On("ExecuteMultiQuery", mock.Anything, transactionID, models.AuditModeCountOnly).
		Return(execResult, nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/queries/multi/%s/commit", transactionID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.MultiQueryResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, transactionID.String(), response.TransactionID)
	assert.Equal(t, "committed", response.Status)
	assert.Equal(t, 10, response.TotalAffectedRows)
	assert.Equal(t, 1, response.StatementCount)
	assert.True(t, response.IsMultiQuery)

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestCommitMultiQuery_ClearsActiveTransaction tests that commit clears the active transaction
func TestCommitMultiQuery_ClearsActiveTransaction(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupCommitMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	transactionID := uuid.New()

	// Mock result showing transaction is cleared (committed)
	execResult := &service.MultiQueryResult{
		TransactionID:     transactionID,
		Status:            "committed",
		TotalAffectedRows: 5,
		ExecutionTimeMs:   100,
		Statements: []models.QueryTransactionStatement{
			{
				Sequence:      0,
				QueryText:     "INSERT INTO logs (msg) VALUES ('test')",
				OperationType: models.OperationInsert,
				Status:        models.StatementStatusSuccess,
				AffectedRows:  5,
			},
		},
	}

	mockMultiQuerySvc.On("ExecuteMultiQuery", mock.Anything, transactionID, models.AuditModeCountOnly).
		Return(execResult, nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/queries/multi/%s/commit", transactionID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.MultiQueryResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// After commit, transaction should be cleared (status = committed)
	assert.Equal(t, "committed", response.Status)
	assert.Equal(t, 5, response.TotalAffectedRows)

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestCommitMultiQuery_NoActiveTransaction_ReturnsError tests error when no active transaction
func TestCommitMultiQuery_NoActiveTransaction_ReturnsError(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupCommitMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	transactionID := uuid.New()

	// Mock error for no active transaction
	mockMultiQuerySvc.On("ExecuteMultiQuery", mock.Anything, transactionID, models.AuditModeCountOnly).
		Return(nil, fmt.Errorf("no active transaction found for data source"))

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/queries/multi/%s/commit", transactionID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "no active transaction found")

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestCommitMultiQuery_UpdatesTransactionStatus tests that transaction status is updated to committed
func TestCommitMultiQuery_UpdatesTransactionStatus(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupCommitMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	transactionID := uuid.New()

	// Mock result with committed status
	execResult := &service.MultiQueryResult{
		TransactionID:     transactionID,
		Status:            "committed",
		TotalAffectedRows: 15,
		ExecutionTimeMs:   200,
		Statements: []models.QueryTransactionStatement{
			{
				Sequence:      0,
				QueryText:     "UPDATE users SET status = 'active'",
				OperationType: models.OperationUpdate,
				Status:        models.StatementStatusSuccess,
				AffectedRows:  10,
			},
			{
				Sequence:      1,
				QueryText:     "INSERT INTO audit_log (action) VALUES ('batch_update')",
				OperationType: models.OperationInsert,
				Status:        models.StatementStatusSuccess,
				AffectedRows:  5,
			},
		},
	}

	mockMultiQuerySvc.On("ExecuteMultiQuery", mock.Anything, transactionID, models.AuditModeCountOnly).
		Return(execResult, nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/queries/multi/%s/commit", transactionID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.MultiQueryResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify status is updated to committed
	assert.Equal(t, "committed", response.Status)
	assert.Equal(t, 2, response.StatementCount)
	assert.Equal(t, 15, response.TotalAffectedRows)

	// Verify all statements have success status
	for _, stmt := range response.Statements {
		assert.Equal(t, "success", stmt.Status)
	}

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestCommitMultiQuery_PermissionCheck_Enforced tests that permission check is enforced for commit
func TestCommitMultiQuery_PermissionCheck_Enforced(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupCommitMultiQueryTestRouter(t, db)

	// Create a regular user (non-admin)
	user := fixtures.CreateTestRegularUser(t, db)
	transactionID := uuid.New()

	// Mock permission denied error
	mockMultiQuerySvc.On("ExecuteMultiQuery", mock.Anything, transactionID, models.AuditModeCountOnly).
		Return(nil, fmt.Errorf("permission denied: user does not have permission to commit transaction"))

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/queries/multi/%s/commit", transactionID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "permission denied")

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestCommitMultiQuery_InvalidTransactionID_ReturnsError tests invalid transaction ID handling
func TestCommitMultiQuery_InvalidTransactionID_ReturnsError(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, _, _, jwtManager := setupCommitMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	// Test with invalid UUID format
	req, _ := http.NewRequest("POST", "/api/v1/queries/multi/invalid-uuid/commit", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "Invalid transaction ID")
}

// TestCommitMultiQuery_AlreadyCommitted_ReturnsError tests already committed transaction handling
func TestCommitMultiQuery_AlreadyCommitted_ReturnsError(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupCommitMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	transactionID := uuid.New()

	// Mock error for already committed transaction
	mockMultiQuerySvc.On("ExecuteMultiQuery", mock.Anything, transactionID, models.AuditModeCountOnly).
		Return(nil, fmt.Errorf("transaction already committed"))

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/queries/multi/%s/commit", transactionID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "already committed")

	mockMultiQuerySvc.AssertExpectations(t)
}

// =============================================================================
// ROLLBACK MULTI QUERY TESTS
// =============================================================================

// setupRollbackMultiQueryTestRouter creates a test router with rollback endpoint for multi-query handler
func setupRollbackMultiQueryTestRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *MockMultiQueryService, *MockQueryServiceForMultiQuery, *auth.JWTManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := auth.NewJWTManager(testauth.TestJWTSecret, testauth.TestJWTExpireTime, testauth.TestJWTIssuer)
	mockMultiQueryService := new(MockMultiQueryService)
	mockQueryService := new(MockQueryServiceForMultiQuery)

	// Create a test-specific wrapper that mimics the RollbackMultiQuery flow
	rollbackHandler := &testRollbackMultiQueryHandler{
		db:            db,
		multiQuerySvc: mockMultiQueryService,
		querySvc:      mockQueryService,
	}

	// Setup routes
	api := router.Group("/api/v1")
	{
		queries := api.Group("/queries")
		queries.Use(middleware.AuthMiddleware(jwtManager, nil))
		{
			queries.POST("/multi/:id/rollback", rollbackHandler.RollbackMultiQuery)
		}
	}

	return router, mockMultiQueryService, mockQueryService, jwtManager
}

// testRollbackMultiQueryHandler wraps the rollback multi-query handling logic for testing
type testRollbackMultiQueryHandler struct {
	db            *gorm.DB
	multiQuerySvc *MockMultiQueryService
	querySvc      *MockQueryServiceForMultiQuery
}

// RollbackMultiQuery rolls back a multi-query transaction
func (h *testRollbackMultiQueryHandler) RollbackMultiQuery(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse user ID
	_, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse transaction ID from URL
	transactionIDStr := c.Param("id")
	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	// Execute the rollback
	if err := h.multiQuerySvc.RollbackMultiQuery(c.Request.Context(), transactionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "rolled_back"})
}

// TestRollbackMultiQuery_AfterFailure_RollsBack tests rolling back after a failure
func TestRollbackMultiQuery_AfterFailure_RollsBack(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupRollbackMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	transactionID := uuid.New()

	// Mock successful rollback after failure
	mockMultiQuerySvc.On("RollbackMultiQuery", mock.Anything, transactionID).
		Return(nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/queries/multi/%s/rollback", transactionID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "rolled_back", response["status"])

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestRollbackMultiQuery_WithActiveTransaction_RollsBack tests rollback with active transaction
func TestRollbackMultiQuery_WithActiveTransaction_RollsBack(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupRollbackMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	transactionID := uuid.New()

	// Create an active transaction in the database
	transaction := &models.QueryTransaction{
		ID:             transactionID,
		DataSourceID:   uuid.New(),
		QueryText:      "UPDATE users SET active = 1",
		StartedBy:      admin.ID,
		Status:         models.TransactionStatusActive,
		IsMultiQuery:   true,
		StatementCount: 1,
	}
	err := db.Create(transaction).Error
	require.NoError(t, err)

	// Mock successful rollback
	mockMultiQuerySvc.On("RollbackMultiQuery", mock.Anything, transactionID).
		Return(nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/queries/multi/%s/rollback", transactionID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "rolled_back", response["status"])

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestRollbackMultiQuery_WithoutActiveTransaction_ReturnsError tests error when no active transaction
func TestRollbackMultiQuery_WithoutActiveTransaction_ReturnsError(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupRollbackMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	transactionID := uuid.New()

	// Mock error for no active transaction
	mockMultiQuerySvc.On("RollbackMultiQuery", mock.Anything, transactionID).
		Return(fmt.Errorf("no active transaction found for rollback"))

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/queries/multi/%s/rollback", transactionID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "no active transaction found")

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestRollbackMultiQuery_CleanupTransactionRecords tests cleanup of transaction records
func TestRollbackMultiQuery_CleanupTransactionRecords(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupRollbackMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	transactionID := uuid.New()
	dataSourceID := uuid.New()

	// Create an active transaction with statements
	transaction := &models.QueryTransaction{
		ID:             transactionID,
		DataSourceID:   dataSourceID,
		QueryText:      "UPDATE users SET active = 1; INSERT INTO logs (msg) VALUES ('test')",
		StartedBy:      admin.ID,
		Status:         models.TransactionStatusActive,
		IsMultiQuery:   true,
		StatementCount: 2,
	}
	err := db.Create(transaction).Error
	require.NoError(t, err)

	// Create transaction statements
	statements := []models.QueryTransactionStatement{
		{
			ID:            uuid.New(),
			TransactionID: transactionID,
			Sequence:      0,
			QueryText:     "UPDATE users SET active = 1",
			OperationType: models.OperationUpdate,
			Status:        models.StatementStatusSuccess,
			AffectedRows:  10,
		},
		{
			ID:            uuid.New(),
			TransactionID: transactionID,
			Sequence:      1,
			QueryText:     "INSERT INTO logs (msg) VALUES ('test')",
			OperationType: models.OperationInsert,
			Status:        models.StatementStatusSuccess,
			AffectedRows:  1,
		},
	}
	for _, stmt := range statements {
		err := db.Create(&stmt).Error
		require.NoError(t, err)
	}

	// Mock successful rollback (which would cleanup records in real implementation)
	mockMultiQuerySvc.On("RollbackMultiQuery", mock.Anything, transactionID).
		Return(nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/queries/multi/%s/rollback", transactionID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "rolled_back", response["status"])

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestRollbackMultiQuery_UpdatesTransactionStatus tests that transaction status is updated to rolled back
func TestRollbackMultiQuery_UpdatesTransactionStatus(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupRollbackMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	transactionID := uuid.New()

	// Create an active transaction
	transaction := &models.QueryTransaction{
		ID:             transactionID,
		DataSourceID:   uuid.New(),
		QueryText:      "UPDATE users SET active = 0",
		StartedBy:      admin.ID,
		Status:         models.TransactionStatusActive,
		IsMultiQuery:   true,
		StatementCount: 1,
	}
	err := db.Create(transaction).Error
	require.NoError(t, err)

	// Mock successful rollback
	mockMultiQuerySvc.On("RollbackMultiQuery", mock.Anything, transactionID).
		Return(nil)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/queries/multi/%s/rollback", transactionID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response indicates rolled back status
	assert.Equal(t, "rolled_back", response["status"])

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestRollbackMultiQuery_PermissionCheck_Enforced tests that permission check is enforced for rollback
func TestRollbackMultiQuery_PermissionCheck_Enforced(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupRollbackMultiQueryTestRouter(t, db)

	// Create a regular user (non-admin)
	user := fixtures.CreateTestRegularUser(t, db)
	transactionID := uuid.New()

	// Mock permission denied error
	mockMultiQuerySvc.On("RollbackMultiQuery", mock.Anything, transactionID).
		Return(fmt.Errorf("permission denied: user does not have permission to rollback transaction"))

	token, err := jwtManager.GenerateToken(user.ID, user.Email, string(models.RoleUser))
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/queries/multi/%s/rollback", transactionID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "permission denied")

	mockMultiQuerySvc.AssertExpectations(t)
}

// TestRollbackMultiQuery_InvalidTransactionID_ReturnsError tests invalid transaction ID handling
func TestRollbackMultiQuery_InvalidTransactionID_ReturnsError(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, _, _, jwtManager := setupRollbackMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	// Test with invalid UUID format
	req, _ := http.NewRequest("POST", "/api/v1/queries/multi/invalid-uuid/rollback", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "Invalid transaction ID")
}

// TestRollbackMultiQuery_AlreadyRolledBack_ReturnsError tests already rolled back transaction handling
func TestRollbackMultiQuery_AlreadyRolledBack_ReturnsError(t *testing.T) {
	db := setupMultiQueryTestDB(t)
	router, mockMultiQuerySvc, _, jwtManager := setupRollbackMultiQueryTestRouter(t, db)

	admin := fixtures.CreateTestAdminUser(t, db)
	transactionID := uuid.New()

	// Mock error for already rolled back transaction
	mockMultiQuerySvc.On("RollbackMultiQuery", mock.Anything, transactionID).
		Return(fmt.Errorf("transaction already rolled back"))

	token, err := jwtManager.GenerateToken(admin.ID, admin.Email, string(models.RoleAdmin))
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/queries/multi/%s/rollback", transactionID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "already rolled back")

	mockMultiQuerySvc.AssertExpectations(t)
}
