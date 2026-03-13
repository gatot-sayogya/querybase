package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/querybase/internal/models"
	"gorm.io/gorm"
)

// MultiQueryService handles multi-query transaction operations
type MultiQueryService struct {
	db              *gorm.DB
	queryService    *QueryService
	auditService    *AuditService
	approvalService *ApprovalService
}

// NewMultiQueryService creates a new multi-query service
func NewMultiQueryService(db *gorm.DB, queryService *QueryService, auditService *AuditService, approvalService *ApprovalService) *MultiQueryService {
	return &MultiQueryService{
		db:              db,
		queryService:    queryService,
		auditService:    auditService,
		approvalService: approvalService,
	}
}

// StatementPreview represents a preview for a single statement
type StatementPreview struct {
	Sequence      int                      `json:"sequence"`
	QueryText     string                   `json:"query_text"`
	OperationType models.OperationType     `json:"operation_type"`
	EstimatedRows int                      `json:"estimated_rows"`
	PreviewRows   []map[string]interface{} `json:"preview_rows,omitempty"`
	Columns       []string                 `json:"columns,omitempty"`
	Error         string                   `json:"error,omitempty"`
}

// MultiQueryPreviewResult represents the preview for all statements
type MultiQueryPreviewResult struct {
	StatementCount     int                `json:"statement_count"`
	TotalEstimatedRows int                `json:"total_estimated_rows"`
	Statements         []StatementPreview `json:"statements"`
	RequiresApproval   bool               `json:"requires_approval"`
}

// MultiQueryResult represents the execution result
type MultiQueryResult struct {
	TransactionID     uuid.UUID                          `json:"transaction_id"`
	Status            string                             `json:"status"`
	TotalAffectedRows int                                `json:"total_affected_rows"`
	ExecutionTimeMs   int64                              `json:"execution_time_ms"`
	Statements        []models.QueryTransactionStatement `json:"statements"`
	ErrorMessage      string                             `json:"error_message,omitempty"`
}

// PreviewMultiQuery generates previews for all statements in a multi-query
func (s *MultiQueryService) PreviewMultiQuery(ctx context.Context, dataSourceID uuid.UUID, userID uuid.UUID, queryTexts []string) (*MultiQueryPreviewResult, error) {
	// Get data source
	var dataSource models.DataSource
	if err := s.db.First(&dataSource, "id = ?", dataSourceID).Error; err != nil {
		return nil, fmt.Errorf("data source not found: %w", err)
	}

	// Check permissions
	perms, err := s.queryService.GetEffectivePermissions(ctx, userID, dataSourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}

	result := &MultiQueryPreviewResult{
		Statements: make([]StatementPreview, 0, len(queryTexts)),
	}

	for i, queryText := range queryTexts {
		preview := StatementPreview{
			Sequence:      i,
			QueryText:     queryText,
			OperationType: DetectOperationType(queryText),
		}

		// Check permissions for this operation
		switch preview.OperationType {
		case models.OperationSelect:
			if !perms.CanSelect {
				preview.Error = "permission denied: SELECT not allowed"
			}
		case models.OperationInsert:
			if !perms.CanInsert {
				preview.Error = "permission denied: INSERT not allowed"
				result.RequiresApproval = true
			}
		case models.OperationUpdate:
			if !perms.CanUpdate {
				preview.Error = "permission denied: UPDATE not allowed"
				result.RequiresApproval = true
			}
		case models.OperationDelete:
			if !perms.CanDelete {
				preview.Error = "permission denied: DELETE not allowed"
				result.RequiresApproval = true
			}
		}

		// Generate preview for write operations
		if preview.OperationType == models.OperationUpdate || preview.OperationType == models.OperationDelete {
			result.RequiresApproval = true

			if preview.Error == "" { // Only generate preview if permissions OK
				writePreview, err := s.queryService.PreviewWriteQuery(ctx, queryText, &dataSource)
				if err != nil {
					preview.Error = fmt.Sprintf("preview generation failed: %v", err)
				} else {
					preview.EstimatedRows = writePreview.TotalAffected
					preview.PreviewRows = writePreview.PreviewRows
					preview.Columns = writePreview.Columns
				}
			}
		}

		result.Statements = append(result.Statements, preview)
		result.TotalEstimatedRows += preview.EstimatedRows
	}

	result.StatementCount = len(queryTexts)
	return result, nil
}

// CreateMultiQueryTransaction creates a new multi-query transaction
func (s *MultiQueryService) CreateMultiQueryTransaction(ctx context.Context, approvalID uuid.UUID, queryTexts []string, startedBy uuid.UUID) (*models.QueryTransaction, error) {
	// Get approval
	var approval models.ApprovalRequest
	if err := s.db.First(&approval, "id = ?", approvalID).Error; err != nil {
		return nil, fmt.Errorf("approval not found: %w", err)
	}

	// Create transaction record
	transaction := &models.QueryTransaction{
		ID:             uuid.New(),
		ApprovalID:     approvalID,
		DataSourceID:   approval.DataSourceID,
		QueryText:      strings.Join(queryTexts, "; "),
		StartedBy:      startedBy,
		Status:         models.TransactionStatusActive,
		IsMultiQuery:   true,
		StatementCount: len(queryTexts),
	}

	if err := s.db.Create(transaction).Error; err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Create statement records
	for i, queryText := range queryTexts {
		stmt := &models.QueryTransactionStatement{
			ID:            uuid.New(),
			TransactionID: transaction.ID,
			Sequence:      i,
			QueryText:     queryText,
			OperationType: DetectOperationType(queryText),
			Status:        models.StatementStatusPending,
		}

		if err := s.db.Create(stmt).Error; err != nil {
			return nil, fmt.Errorf("failed to create statement record: %w", err)
		}
	}

	return transaction, nil
}

// ExecuteMultiQuery executes all statements in a multi-query transaction
func (s *MultiQueryService) ExecuteMultiQuery(ctx context.Context, transactionID uuid.UUID, auditMode models.AuditMode) (*MultiQueryResult, error) {
	startTime := time.Now()

	// Get transaction
	var transaction models.QueryTransaction
	if err := s.db.Preload("Statements").First(&transaction, "id = ?", transactionID).Error; err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	if transaction.Status != models.TransactionStatusActive {
		return nil, fmt.Errorf("transaction is not active")
	}

	// Get data source
	var dataSource models.DataSource
	if err := s.db.First(&dataSource, "id = ?", transaction.DataSourceID).Error; err != nil {
		return nil, fmt.Errorf("data source not found: %w", err)
	}

	// Connect to data source
	dataSourceDB, err := s.queryService.connectToDataSource(&dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to data source: %w", err)
	}

	// Begin database transaction
	tx := dataSourceDB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	result := &MultiQueryResult{
		TransactionID: transactionID,
		Status:        "success",
		Statements:    make([]models.QueryTransactionStatement, 0, len(transaction.Statements)),
	}

	// Execute statements sequentially
	for _, stmt := range transaction.Statements {
		stmtStartTime := time.Now()

		// Execute based on operation type
		if stmt.OperationType == models.OperationSelect {
			// For SELECT, just record that it was executed (no actual execution in write transaction)
			stmt.Status = models.StatementStatusSuccess
			stmt.AffectedRows = 0
		} else {
			// Execute write operation
			execResult := tx.Exec(stmt.QueryText)
			if execResult.Error != nil {
				// Rollback on error
				tx.Rollback()

				stmt.Status = models.StatementStatusFailed
				stmt.ErrorMessage = execResult.Error.Error()
				s.db.Save(&stmt)

				// Update transaction status
				transaction.Status = models.TransactionStatusFailed
				transaction.ErrorMessage = fmt.Sprintf("Statement %d failed: %v", stmt.Sequence, execResult.Error)
				s.db.Save(&transaction)

				result.Status = "failed"
				result.ErrorMessage = transaction.ErrorMessage
				result.ExecutionTimeMs = time.Since(startTime).Milliseconds()
				return result, nil
			}

			stmt.Status = models.StatementStatusSuccess
			stmt.AffectedRows = int(execResult.RowsAffected)
			result.TotalAffectedRows += stmt.AffectedRows
		}

		stmt.ExecutionTimeMs = int(time.Since(stmtStartTime).Milliseconds())
		result.Statements = append(result.Statements, stmt)

		// Save statement status
		s.db.Save(&stmt)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		transaction.Status = models.TransactionStatusFailed
		transaction.ErrorMessage = fmt.Sprintf("Commit failed: %v", err)
		s.db.Save(&transaction)

		result.Status = "failed"
		result.ErrorMessage = transaction.ErrorMessage
		result.ExecutionTimeMs = time.Since(startTime).Milliseconds()
		return result, nil
	}

	// Update transaction status
	now := time.Now()
	transaction.Status = models.TransactionStatusCommitted
	transaction.CompletedAt = &now
	transaction.AffectedRows = result.TotalAffectedRows
	if err := s.db.Save(&transaction).Error; err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	result.ExecutionTimeMs = time.Since(startTime).Milliseconds()
	return result, nil
}

// GetMultiQueryStatements retrieves all statements for a transaction
func (s *MultiQueryService) GetMultiQueryStatements(ctx context.Context, transactionID uuid.UUID) ([]models.QueryTransactionStatement, error) {
	var statements []models.QueryTransactionStatement
	if err := s.db.Where("transaction_id = ?", transactionID).Order("sequence").Find(&statements).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch statements: %w", err)
	}
	return statements, nil
}

// RollbackMultiQuery rolls back a multi-query transaction
func (s *MultiQueryService) RollbackMultiQuery(ctx context.Context, transactionID uuid.UUID) error {
	var transaction models.QueryTransaction
	if err := s.db.First(&transaction, "id = ?", transactionID).Error; err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	if transaction.Status != models.TransactionStatusActive {
		return fmt.Errorf("transaction is not active")
	}

	transaction.Status = models.TransactionStatusRolledBack
	now := time.Now()
	transaction.CompletedAt = &now

	if err := s.db.Save(&transaction).Error; err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}
