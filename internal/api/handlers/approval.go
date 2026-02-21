package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/querybase/internal/api/dto"
	"github.com/yourorg/querybase/internal/models"
	"github.com/yourorg/querybase/internal/service"
	"gorm.io/gorm"
)

// ApprovalHandler handles approval endpoints
type ApprovalHandler struct {
	db              *gorm.DB
	approvalService *service.ApprovalService
}

// NewApprovalHandler creates a new approval handler
func NewApprovalHandler(db *gorm.DB, approvalService *service.ApprovalService) *ApprovalHandler {
	return &ApprovalHandler{
		db:              db,
		approvalService: approvalService,
	}
}

// ListApprovals returns a list of approval requests
func (h *ApprovalHandler) ListApprovals(c *gin.Context) {
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

	approvals, total, err := h.approvalService.ListApprovals(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch approvals"})
		return
	}

	response := make([]gin.H, len(approvals))
	for i, approval := range approvals {
		response[i] = h.formatApprovalResponse(approval)
	}

	c.JSON(http.StatusOK, gin.H{
		"approvals": response,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

// GetApprovalCounts returns counts of approvals grouped by status
func (h *ApprovalHandler) GetApprovalCounts(c *gin.Context) {
	userID := c.GetString("user_id")

	var requestedBy string
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err == nil {
		if user.Role != models.RoleAdmin {
			isApprover := h.checkIsApprover(userID)
			if !isApprover {
				requestedBy = userID
			}
		}
	}

	counts, err := h.approvalService.GetApprovalCounts(c, requestedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch approval counts"})
		return
	}

	c.JSON(http.StatusOK, counts)
}

// GetApproval retrieves a single approval request
func (h *ApprovalHandler) GetApproval(c *gin.Context) {
	approvalID := c.Param("id")
	userID := c.GetString("user_id")

	approval, err := h.approvalService.GetApproval(c, approvalID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Approval not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch approval"})
		}
		return
	}

	// Check permission
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err == nil {
		if user.Role != models.RoleAdmin && approval.RequestedBy.String() != userID {
			// Check if user is an approver for this data source
			if !h.checkCanApprove(userID, approval.DataSourceID.String()) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, h.formatApprovalResponse(*approval))
}

// ReviewApproval adds a review to an approval request
func (h *ApprovalHandler) ReviewApproval(c *gin.Context) {
	approvalID := c.Param("id")
	userID := c.GetString("user_id")

	var req dto.ReviewApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify approval exists
	approval, err := h.approvalService.GetApproval(c, approvalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Approval not found"})
		return
	}

	// Check if user can approve this request
	if !h.checkCanApprove(userID, approval.DataSourceID.String()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to approve this request"})
		return
	}

	// Create review
	reviewInput := &service.ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: userID,
		Decision:   models.ApprovalDecision(req.Decision),
		Comments:   req.Comments,
	}

	review, err := h.approvalService.ReviewApproval(c, reviewInput)
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

// GetEligibleApprovers returns users who can approve for a data source
func (h *ApprovalHandler) GetEligibleApprovers(c *gin.Context) {
	dataSourceID := c.Param("id")

	approvers, err := h.approvalService.GetEligibleApprovers(c, dataSourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch approvers"})
		return
	}

	response := make([]gin.H, len(approvers))
	for i, approver := range approvers {
		response[i] = gin.H{
			"id":        approver.ID.String(),
			"email":     approver.Email,
			"username":  approver.Username,
			"full_name": approver.FullName,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"approvers": response,
		"count":     len(approvers),
	})
}

// checkIsApprover checks if user is an approver for any data source
func (h *ApprovalHandler) checkIsApprover(userID string) bool {
	var count int64
	h.db.Table("data_source_permissions").
		Joins("JOIN user_groups ON user_groups.group_id = data_source_permissions.group_id").
		Where("user_groups.user_id = ?", userID).
		Where("data_source_permissions.can_approve = ?", true).
		Count(&count)

	return count > 0
}

// checkCanApprove checks if user can approve requests for a data source
func (h *ApprovalHandler) checkCanApprove(userID, dataSourceID string) bool {
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

// formatApprovalResponse formats an approval request for API response
func (h *ApprovalHandler) formatApprovalResponse(approval models.ApprovalRequest) gin.H {
	reviews := make([]gin.H, len(approval.ApprovalReviews))
	for i, review := range approval.ApprovalReviews {
		reviews[i] = gin.H{
			"id":          review.ID.String(),
			"reviewer_id": review.ReviewerID,
			"decision":    string(review.Decision),
			"comments":    review.Comments,
			"reviewed_at": review.ReviewedAt,
		}
	}

	return gin.H{
		"id":               approval.ID.String(),
		"operation_type":   string(approval.OperationType),
		"data_source_id":   approval.DataSourceID.String(),
		"data_source_name": approval.DataSource.Name,
		"query_text":       approval.QueryText,
		"requested_by":     approval.RequestedBy.String(),
		"requester_name":   approval.RequestedByUser.FullName,
		"status":           string(approval.Status),
		"created_at":       approval.CreatedAt,
		"updated_at":       approval.UpdatedAt,
		"reviews":          reviews,
	}
}

// ValidateQuery validates a SQL query before submission
func (h *ApprovalHandler) ValidateQuery(c *gin.Context) {
	var req dto.ValidateQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Step 1: Validate SQL syntax
	err := service.ValidateSQL(req.QueryText)
	if err != nil {
		c.JSON(http.StatusOK, dto.ValidateQueryResponse{
			Valid:         false,
			Error:         err.Error(),
			OperationType: string(service.DetectOperationType(req.QueryText)),
		})
		return
	}

	// Step 2: Validate schema if data_source_id is provided
	if req.DataSourceID != "" {
		// Get data source
		var dataSource models.DataSource
		if err := h.db.First(&dataSource, "id = ?", req.DataSourceID).Error; err != nil {
			c.JSON(http.StatusOK, dto.ValidateQueryResponse{
				Valid:         false,
				Error:         "Data source not found",
				OperationType: string(service.DetectOperationType(req.QueryText)),
			})
			return
		}

		// Validate tables exist in data source
		// We need to access query service, but approval handler doesn't have it
		// For now, skip schema validation in this endpoint
		// The schema validation will be done when creating the approval request
	}

	// Detect operation type
	operationType := service.DetectOperationType(req.QueryText)

	c.JSON(http.StatusOK, dto.ValidateQueryResponse{
		Valid:         true,
		OperationType: string(operationType),
	})
}

// StartTransaction starts a transaction for an approval request
func (h *ApprovalHandler) StartTransaction(c *gin.Context) {
	userID := c.GetString("user_id")
	approvalID := c.Param("id")

	// Verify approval exists
	approval, err := h.approvalService.GetApproval(c, approvalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Approval not found"})
		return
	}

	// Check if user can approve this request
	if !h.checkCanApprove(userID, approval.DataSourceID.String()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to approve this request"})
		return
	}

	// Check if approval is still pending
	if approval.Status != models.ApprovalStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Approval is not pending"})
		return
	}

	// Start transaction
	transaction, err := h.approvalService.StartTransaction(c, approvalID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Parse preview data
	var previewData []map[string]interface{}
	var columns []string

	if transaction.PreviewData != "" {
		json.Unmarshal([]byte(transaction.PreviewData), &previewData)
	}

	// Get column names from the result
	if len(previewData) > 0 {
		for key := range previewData[0] {
			columns = append(columns, key)
		}
	}

	// Format response
	c.JSON(http.StatusOK, dto.TransactionResponse{
		TransactionID: transaction.ID.String(),
		ApprovalID:    transaction.ApprovalID.String(),
		Status:        string(transaction.Status),
		QueryText:     transaction.QueryText,
		DataSourceID:  transaction.DataSourceID.String(),
		StartedBy:     transaction.StartedBy.String(),
		StartedAt:     transaction.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		Preview: dto.TransactionPreview{
			RowCount: transaction.AffectedRows,
			Columns:  convertToColumnInfo(columns),
			Data:     previewData,
		},
	})
}

// CommitTransaction commits an active transaction
func (h *ApprovalHandler) CommitTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	userID := c.GetString("user_id")

	// Get transaction
	var transaction models.QueryTransaction
	if err := h.db.Preload("Approval").First(&transaction, "id = ?", transactionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Check if user can approve this request
	if !h.checkCanApprove(userID, transaction.DataSourceID.String()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to commit this transaction"})
		return
	}

	// Commit transaction
	err := h.approvalService.CommitTransaction(c, transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.CommitTransactionResponse{
		TransactionID: transactionID,
		Status:        "committed",
		Message:       "Transaction committed successfully",
		ApprovalID:    transaction.ApprovalID.String(),
	})
}

// RollbackTransaction rolls back an active transaction
func (h *ApprovalHandler) RollbackTransaction(c *gin.Context) {
	transactionID := c.Param("id")
	userID := c.GetString("user_id")

	// Get transaction
	var transaction models.QueryTransaction
	if err := h.db.Preload("Approval").First(&transaction, "id = ?", transactionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Check if user can approve this request
	if !h.checkCanApprove(userID, transaction.DataSourceID.String()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to rollback this transaction"})
		return
	}

	// Rollback transaction
	err := h.approvalService.RollbackTransaction(c, transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.RollbackTransactionResponse{
		TransactionID: transactionID,
		Status:        "rolled_back",
		Message:       "Transaction rolled back successfully",
		ApprovalID:    transaction.ApprovalID.String(),
	})
}

// GetTransactionStatus retrieves the status of a transaction
func (h *ApprovalHandler) GetTransactionStatus(c *gin.Context) {
	transactionID := c.Param("id")
	userID := c.GetString("user_id")

	// Get transaction
	var transaction models.QueryTransaction
	if err := h.db.Preload("Approval").Preload("DataSource").Preload("StartedByUser").
		First(&transaction, "id = ?", transactionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Check permission
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err == nil {
		if user.Role != models.RoleAdmin && transaction.StartedBy.String() != userID {
			if !h.checkCanApprove(userID, transaction.DataSourceID.String()) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
				return
			}
		}
	}

	// Format response
	response := dto.TransactionStatusResponse{
		TransactionID: transaction.ID.String(),
		ApprovalID:    transaction.ApprovalID.String(),
		Status:        string(transaction.Status),
		QueryText:     transaction.QueryText,
		DataSourceID:  transaction.DataSourceID.String(),
		StartedBy:     transaction.StartedBy.String(),
		StartedAt:     transaction.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		ErrorMessage:  transaction.ErrorMessage,
		AffectedRows:  transaction.AffectedRows,
	}

	if transaction.CompletedAt != nil {
		completedAt := transaction.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		response.CompletedAt = &completedAt
	}

	c.JSON(http.StatusOK, response)
}

// AddComment adds a comment to an approval request
func (h *ApprovalHandler) AddComment(c *gin.Context) {
	approvalID := c.Param("id")
	userID := c.GetString("user_id")

	var req dto.ApprovalCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify approval exists
	var approval models.ApprovalRequest
	if err := h.db.First(&approval, "id = ?", approvalID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Approval not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch approval"})
		}
		return
	}

	// Add comment
	comment, err := h.approvalService.AddComment(c, approvalID, userID, req.Comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ApprovalCommentResponse{
		ID:                comment.ID.String(),
		ApprovalRequestID: comment.ApprovalRequestID.String(),
		UserID:            comment.UserID.String(),
		Username:          comment.User.Username,
		FullName:          comment.User.FullName,
		Comment:           comment.Comment,
		CreatedAt:         comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         comment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetComments retrieves comments for an approval request
func (h *ApprovalHandler) GetComments(c *gin.Context) {
	approvalID := c.Param("id")

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if perPage > 100 {
		perPage = 100
	}

	// Get comments
	comments, total, err := h.approvalService.GetComments(c, approvalID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format response
	response := make([]dto.ApprovalCommentResponse, len(comments))
	for i, comment := range comments {
		response[i] = dto.ApprovalCommentResponse{
			ID:                comment.ID.String(),
			ApprovalRequestID: comment.ApprovalRequestID.String(),
			UserID:            comment.UserID.String(),
			Username:          comment.User.Username,
			FullName:          comment.User.FullName,
			Comment:           comment.Comment,
			CreatedAt:         comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:         comment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, dto.ApprovalCommentsListResponse{
		Comments: response,
		Total:    total,
		Page:     page,
		PerPage:  perPage,
	})
}

// DeleteComment deletes a comment
func (h *ApprovalHandler) DeleteComment(c *gin.Context) {
	commentID := c.Param("comment_id")
	userID := c.GetString("user_id")

	// Check if user is admin
	var user models.User
	isAdmin := false
	if err := h.db.First(&user, "id = ?", userID).Error; err == nil {
		isAdmin = user.Role == models.RoleAdmin
	}

	// Delete comment
	if err := h.approvalService.DeleteComment(c, commentID, userID, isAdmin); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		} else if err.Error() == "insufficient permissions to delete this comment" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}

// convertToColumnInfo converts string column names to ColumnInfo slice
func convertToColumnInfo(columns []string) []dto.ColumnInfo {
	result := make([]dto.ColumnInfo, len(columns))
	for i, col := range columns {
		result[i] = dto.ColumnInfo{
			Name: col,
			Type: "unknown",
		}
	}
	return result
}
