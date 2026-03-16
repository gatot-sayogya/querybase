package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourorg/querybase/internal/api/dto"
	"github.com/yourorg/querybase/internal/models"
	"github.com/yourorg/querybase/internal/service"
	"gorm.io/gorm"
)

// MultiQueryHandler handles multi-query transaction operations
type MultiQueryHandler struct {
	db            *gorm.DB
	multiQuerySvc *service.MultiQueryService
	querySvc      *service.QueryService
	approvalSvc   *service.ApprovalService
}

// NewMultiQueryHandler creates a new multi-query handler
func NewMultiQueryHandler(db *gorm.DB, multiQuerySvc *service.MultiQueryService, querySvc *service.QueryService, approvalSvc *service.ApprovalService) *MultiQueryHandler {
	return &MultiQueryHandler{
		db:            db,
		multiQuerySvc: multiQuerySvc,
		querySvc:      querySvc,
		approvalSvc:   approvalSvc,
	}
}

// PreviewMultiQuery generates previews for multiple queries
func (h *MultiQueryHandler) PreviewMultiQuery(c *gin.Context) {
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
	fullQueryText := strings.Join(req.QueryTexts, "; ")

	// Validate and parse queries
	parseResult := service.ValidateMultiQuery(fullQueryText)
	if len(parseResult.Errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Invalid queries",
			"errors": parseResult.Errors,
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

// ExecuteMultiQuery executes multiple queries in a transaction
func (h *MultiQueryHandler) ExecuteMultiQuery(c *gin.Context) {
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
	fullQueryText := strings.Join(req.QueryTexts, "; ")

	// Validate and parse queries
	parseResult := service.ValidateMultiQuery(fullQueryText)
	if len(parseResult.Errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Invalid queries",
			"errors": parseResult.Errors,
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

	// Check if user is admin
	user, _ := c.Get("user")
	isAdmin := false
	if u, ok := user.(*models.User); ok {
		isAdmin = u.Role == models.RoleAdmin
	}

	// Block execution when write operations would affect 0 rows
	// This prevents both direct execution and pointless approval requests
	if impact.RequiresApproval && impact.TotalEstimatedRows == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "No rows would be affected by these queries. Please check your WHERE clause.",
			"estimated_rows":    0,
			"requires_approval": false,
		})
		return
	}

	// Non-admin users require approval for write operations that affect rows
	// Only admins can execute write queries directly
	requiresApproval := !isAdmin && impact.RequiresApproval && impact.TotalEstimatedRows > 0

	if requiresApproval {
		approval := &models.ApprovalRequest{
			ID:            uuid.New(),
			RequestedBy:   userUUID,
			OperationType: models.OperationUpdate, // Use most restrictive type
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
	// Create a temporary transaction and execute it (no approval needed, so pass nil)
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

// GetMultiQueryStatements retrieves all statements for a transaction
func (h *MultiQueryHandler) GetMultiQueryStatements(c *gin.Context) {
	var req dto.GetMultiQueryStatementsRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transactionID, err := uuid.Parse(req.TransactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	statements, err := h.multiQuerySvc.GetMultiQueryStatements(c.Request.Context(), transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to DTO
	result := make([]dto.StatementResult, len(statements))
	for i, stmt := range statements {
		result[i] = dto.StatementResult{
			Sequence:        stmt.Sequence,
			QueryText:       stmt.QueryText,
			OperationType:   string(stmt.OperationType),
			Status:          string(stmt.Status),
			AffectedRows:    stmt.AffectedRows,
			ErrorMessage:    stmt.ErrorMessage,
			ExecutionTimeMs: stmt.ExecutionTimeMs,
		}
	}

	c.JSON(http.StatusOK, result)
}

// CommitMultiQuery commits a multi-query transaction
func (h *MultiQueryHandler) CommitMultiQuery(c *gin.Context) {
	var req dto.CommitMultiQueryRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transactionID, err := uuid.Parse(req.TransactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

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

// RollbackMultiQuery rolls back a multi-query transaction
func (h *MultiQueryHandler) RollbackMultiQuery(c *gin.Context) {
	var req dto.RollbackMultiQueryRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transactionID, err := uuid.Parse(req.TransactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	if err := h.multiQuerySvc.RollbackMultiQuery(c.Request.Context(), transactionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "rolled_back"})
}
