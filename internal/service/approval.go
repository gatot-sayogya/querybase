package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/querybase/internal/models"
	"gorm.io/gorm"
)

// ApprovalService handles approval workflow logic
type ApprovalService struct {
	db           *gorm.DB
	queryService *QueryService
	statsService *StatsService
	auditService *AuditService
}

// NewApprovalService creates a new approval service
func NewApprovalService(db *gorm.DB, queryService *QueryService, statsService *StatsService) *ApprovalService {
	return &ApprovalService{
		db:           db,
		queryService: queryService,
		statsService: statsService,
		auditService: NewAuditService(db),
	}
}

// CreateApprovalRequest creates a new approval request for a write operation
func (s *ApprovalService) CreateApprovalRequest(ctx context.Context, req *ApprovalRequest) (*models.ApprovalRequest, error) {
	requestedByUUID, err := uuid.Parse(req.RequestedBy)
	if err != nil {
		return nil, fmt.Errorf("invalid requested_by UUID: %w", err)
	}

	approval := &models.ApprovalRequest{
		ID:           uuid.New(),
		DataSourceID: req.DataSourceID,
		QueryText:    req.QuerySQL,
		RequestedBy:  requestedByUUID,
		Status:       models.ApprovalStatusPending,
	}

	if err := s.db.Create(approval).Error; err != nil {
		return nil, fmt.Errorf("failed to create approval request: %w", err)
	}

	// TODO: Send notifications to eligible approvers

	// Trigger stats update
	if s.statsService != nil {
		s.statsService.TriggerStatsChanged(approval.RequestedBy.String())
	}

	return approval, nil
}

// GetApproval retrieves an approval by ID
func (s *ApprovalService) GetApproval(ctx context.Context, approvalID string) (*models.ApprovalRequest, error) {
	var approval models.ApprovalRequest
	err := s.db.Preload("DataSource").
		Preload("RequestedByUser").
		Preload("ApprovalReviews").
		Preload("ApprovalReviews.Reviewer").
		First(&approval, "id = ?", approvalID).Error

	if err != nil {
		return nil, err
	}

	return &approval, nil
}

// ListApprovals retrieves a list of approvals with filters
func (s *ApprovalService) ListApprovals(ctx context.Context, filter *ApprovalFilter) ([]models.ApprovalRequest, int64, error) {
	var approvals []models.ApprovalRequest
	var total int64

	query := s.db.Model(&models.ApprovalRequest{})

	// Apply filters
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.DataSourceID != "" {
		query = query.Where("data_source_id = ?", filter.DataSourceID)
	}
	if filter.RequestedBy != "" {
		query = query.Where("requested_by = ?", filter.RequestedBy)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results with relations
	err := query.Preload("DataSource").
		Preload("RequestedByUser").
		Preload("ApprovalReviews").
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&approvals).Error

	return approvals, total, err
}

// GetApprovalCounts retrieves counts of approvals grouped by status
func (s *ApprovalService) GetApprovalCounts(ctx context.Context, requestedBy string) (map[string]int64, error) {
	counts := make(map[string]int64)
	counts["all"] = 0
	counts["pending"] = 0
	counts["approved"] = 0
	counts["rejected"] = 0

	type result struct {
		Status string
		Count  int64
	}
	var results []result

	query := s.db.Model(&models.ApprovalRequest{}).Select("status, count(*) as count").Group("status")
	if requestedBy != "" {
		query = query.Where("requested_by = ?", requestedBy)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to count approvals: %w", err)
	}

	var total int64
	for _, r := range results {
		counts[r.Status] = r.Count
		total += r.Count
	}
	counts["all"] = total

	return counts, nil
}

// ReviewApproval adds a review to an approval request
func (s *ApprovalService) ReviewApproval(ctx context.Context, review *ReviewInput) (*models.ApprovalReview, error) {
	// Get the approval request
	var approval models.ApprovalRequest
	if err := s.db.First(&approval, "id = ?", review.ApprovalID).Error; err != nil {
		return nil, fmt.Errorf("approval request not found: %w", err)
	}

	// Check if approval is still pending
	if approval.Status != models.ApprovalStatusPending {
		return nil, fmt.Errorf("approval request is not pending")
	}

	// Check if reviewer has already reviewed
	var existingReview models.ApprovalReview
	reviewerUUID, _ := uuid.Parse(review.ReviewerID)
	checkErr := s.db.Where("approval_request_id = ? AND reviewed_by = ?", review.ApprovalID, reviewerUUID).First(&existingReview).Error
	if checkErr == nil {
		return nil, fmt.Errorf("already reviewed")
	}

	// Create the review
	approvalReview := &models.ApprovalReview{
		ID:         uuid.New(),
		ApprovalID: approval.ID,
		ReviewerID: reviewerUUID,
		Decision:   review.Decision,
		Comments:   review.Comments,
		ReviewedAt: time.Now(),
	}

	if err := s.db.Create(approvalReview).Error; err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	// Update approval status based on reviews
	s.updateApprovalStatus(ctx, approval.ID)

	// TODO: Send notification to requester

	return approvalReview, nil
}

// GetEligibleApprovers returns users who can approve an approval request
func (s *ApprovalService) GetEligibleApprovers(ctx context.Context, dataSourceID string) ([]models.User, error) {
	var users []models.User

	// Get users with approve permission on the data source
	err := s.db.
		Joins("JOIN user_groups ON users.id = user_groups.user_id").
		Joins("JOIN data_source_permissions ON data_source_permissions.group_id = user_groups.group_id").
		Where("data_source_permissions.data_source_id = ?", dataSourceID).
		Where("data_source_permissions.can_approve = ?", true).
		Where("users.is_active = ?", true).
		Find(&users).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get eligible approvers: %w", err)
	}

	return users, nil
}

// updateApprovalStatus updates the approval status based on reviews
func (s *ApprovalService) updateApprovalStatus(ctx context.Context, approvalID uuid.UUID) {
	var reviews []models.ApprovalReview
	s.db.Where("approval_request_id = ?", approvalID).Find(&reviews)

	if len(reviews) == 0 {
		return
	}

	// If any review is rejected, mark as rejected
	for _, review := range reviews {
		if review.Decision == models.ApprovalDecisionRejected {
			now := time.Now()
			s.db.Model(&models.ApprovalRequest{}).
				Where("id = ?", approvalID).
				Updates(map[string]interface{}{
					"status":       models.ApprovalStatusRejected,
					"completed_at": now,
				})
			return
		}
	}

	// If we have at least one approval, mark as approved
	// For multi-stage approval, you would check for required number of approvals
	for _, review := range reviews {
		if review.Decision == models.ApprovalDecisionApproved {
			now := time.Now()
			s.db.Model(&models.ApprovalRequest{}).
				Where("id = ?", approvalID).
				Updates(map[string]interface{}{
					"status":       models.ApprovalStatusApproved,
					"completed_at": now,
				})

			// Trigger stats update
			if s.statsService != nil {
				var tempApp models.ApprovalRequest
				s.db.Select("requested_by").First(&tempApp, "id = ?", approvalID)
				s.statsService.TriggerStatsChanged(tempApp.RequestedBy.String())
			}

			// Execute the approved query
			var approval models.ApprovalRequest
			if err := s.db.Preload("DataSource").First(&approval, "id = ?", approvalID).Error; err == nil {
				// We create a transient query executed by the reviewer (who has permissions) to bypass the direct execution block
				query := &models.Query{
					ID:            uuid.New(),
					DataSourceID:  approval.DataSourceID,
					UserID:        review.ReviewerID, // Execute as the approver
					QueryText:     approval.QueryText,
					OperationType: approval.OperationType,
					Name:          "Approved execution",
					Status:        models.StatusRunning,
				}

				// The transient query must exist for QueryResult to avoid f-key violations
				s.db.Create(query)

				// Execute in background to avoid blocking the API response
				go func(q *models.Query, ds *models.DataSource) {
					bgCtx := context.Background()
					_, _ = s.queryService.ExecuteQuery(bgCtx, q, ds)
				}(query, &approval.DataSource)
			}
			return
		}
	}
}

// StartTransaction starts a transaction for an approval request and executes the query in preview mode
func (s *ApprovalService) StartTransaction(ctx context.Context, approvalID, startedBy string, auditMode models.AuditMode) (*models.QueryTransaction, error) {
	// Get the approval request
	var approval models.ApprovalRequest
	if err := s.db.Preload("DataSource").First(&approval, "id = ?", approvalID).Error; err != nil {
		return nil, fmt.Errorf("approval request not found: %w", err)
	}

	// Check if approval is still pending
	if approval.Status != models.ApprovalStatusPending {
		return nil, fmt.Errorf("approval request is not pending")
	}

	// Check if an active transaction already exists — return it directly
	var existingTx models.QueryTransaction
	err := s.db.Where("approval_id = ? AND status = ?", approvalID, models.TransactionStatusActive).First(&existingTx).Error
	if err == nil {
		return &existingTx, nil // Return existing active transaction
	}

	// Hard-delete any stale transaction records for this approval so the user can retry cleanly.
	// We wipe all non-committed rows since the unique constraint on approval_id would block a new insert.
	s.db.Where("approval_id = ? AND status != ?", approvalID, models.TransactionStatusCommitted).
		Delete(&models.QueryTransaction{})

	// Parse startedBy as UUID
	startedByUUID, err := uuid.Parse(startedBy)
	if err != nil {
		return nil, fmt.Errorf("invalid started_by UUID: %w", err)
	}

	// Resolve audit mode based on capability
	effectiveMode := s.auditService.ResolveAuditMode(auditMode, approval.DataSource.AuditCapability)

	// Estimate affected rows
	dataSourceDB, err := s.queryService.connectToDataSource(&approval.DataSource)
	var estimatedRows int
	var caution bool
	var cautionMsg string
	if err == nil {
		estimatedRows, _ = s.auditService.EstimateAffectedRows(ctx, approval.QueryText, dataSourceDB, &approval.DataSource)
		caution, cautionMsg = s.auditService.CheckCaution(estimatedRows, &approval.DataSource)

		// Test audit capability lazily if unknown
		if approval.DataSource.AuditCapability == models.AuditCapabilityUnknown {
			cap, _ := s.auditService.TestAuditCapability(ctx, dataSourceDB, &approval.DataSource)
			effectiveMode = s.auditService.ResolveAuditMode(auditMode, cap)
		}
	} else {
		log.Printf("[StartTransaction] WARNING: failed to connect to data source for row estimation: %v", err)
	}

	// Build the transaction record to save
	emptyArray := "[]"
	transaction := &models.QueryTransaction{
		ID:            uuid.New(),
		ApprovalID:    &approval.ID,
		DataSourceID:  approval.DataSourceID,
		QueryText:     approval.QueryText,
		StartedBy:     startedByUUID,
		Status:        models.TransactionStatusActive,
		AuditMode:     effectiveMode,
		EstimatedRows: estimatedRows,
		BeforeData:    &emptyArray,
		AfterData:     &emptyArray,
		PreviewData:   &emptyArray,
	}

	log.Printf("[StartTransaction] Executing query for approval=%s, queryText=%q", approvalID, approval.QueryText)

	// Execute query in transaction mode — returns query results + audit data (before/after rows)
	result, auditResult, err := s.queryService.ExecuteQueryInTransaction(ctx, &approval, &approval.DataSource)
	if err != nil {
		log.Printf("[StartTransaction] ExecuteQueryInTransaction failed: %v", err)
		transaction.Status = models.TransactionStatusFailed
		transaction.ErrorMessage = err.Error()
		// Best-effort save of failed record; ignore error (the wipe above cleared any blocker)
		s.db.Create(transaction)
		return transaction, fmt.Errorf("query execution failed: %w", err)
	}

	// Store preview results. For write queries in audit mode, the actual rows
	// are in auditResult.BeforeData. The result.Data is just metadata (e.g. "operation, rows_affected...").
	if auditResult != nil && len(auditResult.BeforeData) > 0 {
		if beforeJSON, err := json.Marshal(auditResult.BeforeData); err == nil {
			beforeStr := string(beforeJSON)
			transaction.PreviewData = &beforeStr
		}
	} else if result.Data != "" {
		transaction.PreviewData = &result.Data
	}
	transaction.AffectedRows = result.RowCount

	// Store audit before/after row data if available
	if auditResult != nil {
		if len(auditResult.BeforeData) > 0 {
			if beforeJSON, err := json.Marshal(auditResult.BeforeData); err == nil {
				beforeStr := string(beforeJSON)
				transaction.BeforeData = &beforeStr
			}
		}
		if len(auditResult.AfterData) > 0 {
			if afterJSON, err := json.Marshal(auditResult.AfterData); err == nil {
				afterStr := string(afterJSON)
				transaction.AfterData = &afterStr
			}
		}
		if auditResult.AffectedRows > 0 {
			transaction.AffectedRows = auditResult.AffectedRows
		}
		transaction.AuditMode = auditResult.AuditMode
	}

	_ = caution
	_ = cautionMsg

	// Save transaction record
	if err := s.db.Create(transaction).Error; err != nil {
		log.Printf("[StartTransaction] Failed to save transaction record: %v", err)
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}
	beforeCount, afterCount := 0, 0
	if auditResult != nil {
		beforeCount = len(auditResult.BeforeData)
		afterCount = len(auditResult.AfterData)
	}
	log.Printf("[StartTransaction] Success: transaction=%s created for approval=%s, beforeRows=%d, afterRows=%d",
		transaction.ID, approvalID, beforeCount, afterCount)
	return transaction, nil
}

// CommitTransaction commits an active transaction
func (s *ApprovalService) CommitTransaction(ctx context.Context, transactionID string) error {
	// Get the transaction
	var transaction models.QueryTransaction
	if err := s.db.Preload("Approval").Preload("DataSource").First(&transaction, "id = ?", transactionID).Error; err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	// Check if transaction is active
	if transaction.Status != models.TransactionStatusActive {
		return fmt.Errorf("transaction is not active")
	}

	// Commit the transaction in the data source
	err := s.queryService.CommitTransaction(ctx, &transaction.DataSource)
	if err != nil {
		transaction.Status = models.TransactionStatusFailed
		transaction.ErrorMessage = err.Error()
		s.db.Save(&transaction)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Update transaction status
	now := time.Now()
	transaction.Status = models.TransactionStatusCommitted
	transaction.CompletedAt = &now
	s.db.Save(&transaction)

	// Update approval status to approved (only if there's an associated approval)
	if transaction.ApprovalID != nil {
		s.db.Model(&models.ApprovalRequest{}).
			Where("id = ?", *transaction.ApprovalID).
			Updates(map[string]interface{}{
				"status":       models.ApprovalStatusApproved,
				"completed_at": now,
			})
	}

	return nil
}

// RollbackTransaction rolls back an active transaction
func (s *ApprovalService) RollbackTransaction(ctx context.Context, transactionID string) error {
	// Get the transaction
	var transaction models.QueryTransaction
	if err := s.db.Preload("DataSource").First(&transaction, "id = ?", transactionID).Error; err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	// Check if transaction is active
	if transaction.Status != models.TransactionStatusActive {
		return fmt.Errorf("transaction is not active")
	}

	// Rollback the transaction in the data source
	err := s.queryService.RollbackTransaction(ctx, &transaction.DataSource)
	if err != nil {
		transaction.Status = models.TransactionStatusFailed
		transaction.ErrorMessage = err.Error()
		s.db.Save(&transaction)
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	// Update transaction status
	now := time.Now()
	transaction.Status = models.TransactionStatusRolledBack
	transaction.CompletedAt = &now
	s.db.Save(&transaction)

	// Update approval status to rejected (rolled back) - only if there's an associated approval
	if transaction.ApprovalID != nil {
		s.db.Model(&models.ApprovalRequest{}).
			Where("id = ?", *transaction.ApprovalID).
			Updates(map[string]interface{}{
				"status":           models.ApprovalStatusRejected,
				"rejection_reason": "Transaction rolled back by approver",
				"completed_at":     now,
			})
	}

	return nil
}

// GetActiveTransaction gets the active transaction for an approval
func (s *ApprovalService) GetActiveTransaction(ctx context.Context, approvalID string) (*models.QueryTransaction, error) {
	var transaction models.QueryTransaction
	err := s.db.Preload("DataSource").
		Preload("StartedByUser").
		Where("approval_id = ? AND status = ?", approvalID, models.TransactionStatusActive).
		First(&transaction).Error

	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

// AddComment adds a comment to an approval request
func (s *ApprovalService) AddComment(ctx context.Context, approvalID, userID, comment string) (*models.ApprovalComment, error) {
	// Verify approval request exists
	var approval models.ApprovalRequest
	if err := s.db.First(&approval, "id = ?", approvalID).Error; err != nil {
		return nil, fmt.Errorf("approval request not found: %w", err)
	}

	// Parse UUIDs
	approvalUUID, err := uuid.Parse(approvalID)
	if err != nil {
		return nil, fmt.Errorf("invalid approval ID: %w", err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Create comment
	approvalComment := &models.ApprovalComment{
		ID:                uuid.New(),
		ApprovalRequestID: approvalUUID,
		UserID:            userUUID,
		Comment:           comment,
	}

	if err := s.db.Create(approvalComment).Error; err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Preload user data for response
	s.db.Preload("User").First(approvalComment, approvalComment.ID)

	return approvalComment, nil
}

// GetComments retrieves comments for an approval request with pagination
func (s *ApprovalService) GetComments(ctx context.Context, approvalID string, page, perPage int) ([]models.ApprovalComment, int64, error) {
	var comments []models.ApprovalComment
	var total int64

	// Get total count
	if err := s.db.Model(&models.ApprovalComment{}).Where("approval_request_id = ?", approvalID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count comments: %w", err)
	}

	// Get paginated results with user data
	offset := (page - 1) * perPage
	err := s.db.Where("approval_request_id = ?", approvalID).
		Preload("User").
		Order("created_at ASC").
		Limit(perPage).
		Offset(offset).
		Find(&comments).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch comments: %w", err)
	}

	return comments, total, nil
}

// DeleteComment deletes a comment (only by the author or admin)
func (s *ApprovalService) DeleteComment(ctx context.Context, commentID, userID string, isAdmin bool) error {
	var comment models.ApprovalComment
	if err := s.db.First(&comment, "id = ?", commentID).Error; err != nil {
		return fmt.Errorf("comment not found: %w", err)
	}

	// Check permission: admin or comment author
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	if !isAdmin && comment.UserID != userUUID {
		return fmt.Errorf("insufficient permissions to delete this comment")
	}

	// Delete comment
	if err := s.db.Delete(&comment).Error; err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}

// ApprovalRequest represents the input to create an approval request
type ApprovalRequest struct {
	DataSourceID uuid.UUID
	QuerySQL     string
	RequestedBy  string
}

// ApprovalFilter represents filters for listing approvals
type ApprovalFilter struct {
	Status       string
	DataSourceID string
	RequestedBy  string
	Limit        int
	Offset       int
}

// ReviewInput represents the input to review an approval
type ReviewInput struct {
	ApprovalID uuid.UUID
	ReviewerID string
	Decision   models.ApprovalDecision
	Comments   string
}
