package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourorg/querybase/internal/api/dto"
	"github.com/yourorg/querybase/internal/models"
	"github.com/yourorg/querybase/internal/service"
	"gorm.io/gorm"
)

// QueryHandler handles query endpoints
type QueryHandler struct {
	db            *gorm.DB
	queryService  *service.QueryService
}

// NewQueryHandler creates a new query handler
func NewQueryHandler(db *gorm.DB, queryService *service.QueryService) *QueryHandler {
	return &QueryHandler{
		db:           db,
		queryService: queryService,
	}
}

// ExecuteQuery executes a SQL query
func (h *QueryHandler) ExecuteQuery(c *gin.Context) {
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

	// Check user permissions
	if !h.checkReadPermission(userID, dataSource.ID.String()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to read from this data source"})
		return
	}

	// Detect operation type
	operationType := service.DetectOperationType(req.QueryText)

	// For write operations, create approval request
	if service.RequiresApproval(operationType) {
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
		Name:          req.Name,
		Description:   req.Description,
		OperationType: operationType,
		Status:        models.StatusRunning,
	}

	// Save query to database BEFORE executing (needed for foreign key constraint)
	if err := h.db.Create(query).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save query"})
		return
	}

	startTime := time.Now()

	// Execute the query
	result, err := h.queryService.ExecuteQuery(c, query, &dataSource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	executionTime := int(time.Since(startTime).Milliseconds())

	// Parse results
	var data []map[string]interface{}
	json.Unmarshal([]byte(result.Data), &data)

	// Parse column names from JSON string
	var columnNames []string
	json.Unmarshal([]byte(result.ColumnNames), &columnNames)

	columns := make([]dto.ColumnInfo, len(columnNames))
	for i, col := range columnNames {
		columns[i] = dto.ColumnInfo{Name: col, Type: "unknown"}
	}

	c.JSON(http.StatusOK, dto.ExecuteQueryResponse{
		QueryID:       query.ID.String(),
		Status:        "completed",
		RowCount:      &result.RowCount,
		ExecutionTime: &executionTime,
		Data:          data,
		Columns:       columns,
		RequiresApproval: false,
	})
}

// SaveQuery saves a query for later use
func (h *QueryHandler) SaveQuery(c *gin.Context) {
	var req dto.SaveQueryRequest
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

	// Detect operation type
	operationType := service.DetectOperationType(req.QueryText)

	// Parse userID as UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	query := &models.Query{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		UserID:        userUUID,
		QueryText:     req.QueryText,
		Name:          req.Name,
		Description:   req.Description,
		OperationType: operationType,
		Status:        models.StatusPending,
	}

	if err := h.queryService.SaveQuery(c, query); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save query"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"query_id": query.ID.String(),
		"message":  "Query saved successfully",
	})
}

// GetQuery retrieves a query by ID
func (h *QueryHandler) GetQuery(c *gin.Context) {
	queryID := c.Param("id")
	userID := c.GetString("user_id")

	query, err := h.queryService.GetQuery(c, queryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Query not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch query"})
		}
		return
	}

	// Check permission
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err == nil {
		if user.Role != models.RoleAdmin && query.UserID.String() != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	// Get latest result
	var result models.QueryResult
	h.db.Where("query_id = ?", queryID).Order("stored_at DESC").First(&result)

	var data []map[string]interface{}
	if result.Data != "" {
		json.Unmarshal([]byte(result.Data), &data)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             query.ID.String(),
		"name":           query.Name,
		"description":    query.Description,
		"query_text":     query.QueryText,
		"data_source_id": query.DataSourceID.String(),
		"operation_type": string(query.OperationType),
		"status":         string(query.Status),
		"user_id":        query.UserID.String(),
		"created_at":     query.CreatedAt,
		"result": gin.H{
			"row_count": result.RowCount,
			"columns":   result.ColumnNames,
			"data":      data,
		},
	})
}

// ListQueries retrieves a list of queries
func (h *QueryHandler) ListQueries(c *gin.Context) {
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
			"id":              query.ID.String(),
			"name":            query.Name,
			"description":     query.Description,
			"data_source_id":  query.DataSourceID.String(),
			"operation_type":  string(query.OperationType),
			"status":          string(query.Status),
			"created_at":      query.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"queries": response,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// DeleteQuery deletes a saved query
func (h *QueryHandler) DeleteQuery(c *gin.Context) {
	queryID := c.Param("id")
	userID := c.GetString("user_id")

	var query models.Query
	if err := h.db.First(&query, "id = ?", queryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Query not found"})
		return
	}

	// Check permission (only owner or admin can delete)
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err == nil {
		if user.Role != models.RoleAdmin && query.UserID.String() != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	// Delete query and related records
	h.db.Where("query_id = ?", queryID).Delete(&models.QueryResult{})
	h.db.Where("query_id = ?", queryID).Delete(&models.QueryHistory{})
	h.db.Delete(&query)

	c.JSON(http.StatusOK, gin.H{"message": "Query deleted successfully"})
}

// createApprovalForQuery creates an approval request for write operations
func (h *QueryHandler) createApprovalForQuery(c *gin.Context, req dto.ExecuteQueryRequest, dataSource models.DataSource, userID string, operationType models.OperationType) {
	// Check if user has write permission
	if !h.checkWritePermission(userID, dataSource.ID.String()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to submit write operations"})
		return
	}

	// Step 1: Validate SQL syntax
	if err := service.ValidateSQL(req.QueryText); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid SQL syntax",
			"details": err.Error(),
		})
		return
	}

	// Step 2: Validate schema (tables exist in data source)
	// This prevents approvers from reviewing queries with non-existent tables
	if err := h.queryService.ValidateQuerySchema(c, req.QueryText, &dataSource); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Schema validation failed",
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

	// Send notification to approvers
	// This will be implemented when we add the notification service

	c.JSON(http.StatusAccepted, dto.ExecuteQueryResponse{
		QueryID:         uuid.New().String(),
		Status:          "pending_approval",
		RequiresApproval: true,
		ApprovalID:      approval.ID.String(),
	})
}

// checkReadPermission checks if user has read permission on data source
func (h *QueryHandler) checkReadPermission(userID, dataSourceID string) bool {
	var user models.User
	if err := h.db.Preload("Groups").First(&user, "id = ?", userID).Error; err != nil {
		return false
	}

	// Admin has all permissions
	if user.Role == models.RoleAdmin {
		return true
	}

	// Check group permissions
	var count int64
	h.db.Table("data_source_permissions").
		Joins("JOIN user_groups ON user_groups.group_id = data_source_permissions.group_id").
		Where("user_groups.user_id = ? AND data_source_permissions.data_source_id = ?", userID, dataSourceID).
		Where("data_source_permissions.can_read = ?", true).
		Count(&count)

	return count > 0
}

// checkWritePermission checks if user has write permission on data source
func (h *QueryHandler) checkWritePermission(userID, dataSourceID string) bool {
	var user models.User
	if err := h.db.Preload("Groups").First(&user, "id = ?", userID).Error; err != nil {
		return false
	}

	// Admin has all permissions
	if user.Role == models.RoleAdmin {
		return true
	}

	// Check group permissions
	var count int64
	h.db.Table("data_source_permissions").
		Joins("JOIN user_groups ON user_groups.group_id = data_source_permissions.group_id").
		Where("user_groups.user_id = ? AND data_source_permissions.data_source_id = ?", userID, dataSourceID).
		Where("data_source_permissions.can_write = ?", true).
		Count(&count)

	return count > 0
}

// ListQueryHistory retrieves query execution history
func (h *QueryHandler) ListQueryHistory(c *gin.Context) {
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

	history, total, err := h.queryService.ListQueryHistory(c, userID, limit, offset)
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
			"id":               entry.ID.String(),
			"query_id":         entry.QueryID,
			"user_id":          entry.UserID.String(),
			"data_source_id":   entry.DataSourceID.String(),
			"query_text":       entry.QueryText,
			"operation_type":   string(entry.OperationType),
			"status":           string(entry.Status),
			"row_count":        entry.RowCount,
			"execution_time_ms": entry.ExecutionTimeMs,
			"error_message":    entry.ErrorMessage,
			"executed_at":      entry.ExecutedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"history": response,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// ExplainQuery explains a query execution plan
func (h *QueryHandler) ExplainQuery(c *gin.Context) {
	var req dto.ExplainQueryRequest
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

	// Check user permissions
	if !h.checkReadPermission(userID, dataSource.ID.String()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to read from this data source"})
		return
	}

	// Execute EXPLAIN query
	ctx := c.Request.Context()
	result, err := h.queryService.ExplainQuery(ctx, req.QueryText, &dataSource, req.Analyze)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DryRunDelete performs a dry run for DELETE queries
func (h *QueryHandler) DryRunDelete(c *gin.Context) {
	var req dto.DryRunRequest
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

	// Check user permissions (require write permission for dry run)
	if !h.checkWritePermission(userID, dataSource.ID.String()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to write to this data source"})
		return
	}

	// Execute dry run
	ctx := c.Request.Context()
	result, err := h.queryService.DryRunDelete(ctx, req.QueryText, &dataSource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetQueryResults retrieves paginated query results
func (h *QueryHandler) GetQueryResults(c *gin.Context) {
	queryID := c.Param("id")
	userID := c.GetString("user_id")

	// Verify query exists and user has access
	var query models.Query
	if err := h.db.First(&query, "id = ?", queryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Query not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch query"})
		}
		return
	}

	// Check permission
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

	rows, columnNames, metadata, err := h.queryService.GetPaginatedResults(ctx, queryUUID, page, perPage, sortColumn, sortDirection)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build column info
	columns := make([]dto.ColumnInfo, len(columnNames))
	for i, col := range columnNames {
		columns[i] = dto.ColumnInfo{Name: col, Type: "unknown"}
	}

	c.JSON(http.StatusOK, dto.PaginatedResultDTO{
		QueryID:       queryID,
		RowCount:      len(rows),
		Columns:       columns,
		Data:          rows,
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

// ExportQuery exports query results in CSV or JSON format
func (h *QueryHandler) ExportQuery(c *gin.Context) {
	var req dto.ExportQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")

	// Verify query exists and user has access
	var query models.Query
	if err := h.db.First(&query, "id = ?", req.QueryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Query not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch query"})
		}
		return
	}

	// Check permission
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
