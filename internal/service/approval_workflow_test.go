package service

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/querybase/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupWorkflowDB extends setupTestDB with the QueryTransactionStatement table
// which is needed for approval workflow tests but not migrated by the base helper.
func setupWorkflowDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
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
		&models.NotificationConfig{},
		&models.Notification{},
		&models.ApprovalComment{},
	)
	require.NoError(t, err)
	return db
}

func newApprovalServices(t *testing.T, db *gorm.DB) *ApprovalService {
	t.Helper()
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	return NewApprovalService(db, queryService, nil)
}

// ---------------------------------------------------------------------------
// Self-approval
// ---------------------------------------------------------------------------

// TestSelfApproval_Blocked verifies that a non-admin requester cannot approve their own request.
func TestSelfApproval_Blocked(t *testing.T) {
	db := setupWorkflowDB(t)
	svc := newApprovalServices(t, db)
	ctx := context.Background()

	requester := createTestUser(t, db, models.RoleUser) // non-admin user
	ds := createTestDataSource(t, db)

	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  ds.ID,
		QueryText:     "UPDATE users SET name = 'x' WHERE id = 1",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationUpdate,
	}
	require.NoError(t, db.Create(approval).Error)

	_, err := svc.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: requester.ID.String(), // same as requester → must fail
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "self-approve",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "self-approval is not allowed")

	// Ensure status was NOT changed
	var unchanged models.ApprovalRequest
	require.NoError(t, db.First(&unchanged, "id = ?", approval.ID).Error)
	assert.Equal(t, models.ApprovalStatusPending, unchanged.Status)
}

// TestSelfApproval_DifferentUser verifies that a different user can approve normally.
func TestSelfApproval_DifferentUser(t *testing.T) {
	db := setupWorkflowDB(t)
	svc := newApprovalServices(t, db)
	ctx := context.Background()

	requester := createTestUser(t, db, models.RoleUser)
	reviewer := createTestUser(t, db, models.RoleAdmin)
	ds := createTestDataSource(t, db)

	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  ds.ID,
		QueryText:     "DELETE FROM users WHERE id = 1",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationDelete,
	}
	require.NoError(t, db.Create(approval).Error)

	review, err := svc.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: reviewer.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "LGTM",
	})

	require.NoError(t, err)
	assert.Equal(t, models.ApprovalDecisionApproved, review.Decision)

	var updated models.ApprovalRequest
	require.NoError(t, db.First(&updated, "id = ?", approval.ID).Error)
	assert.Equal(t, models.ApprovalStatusApproved, updated.Status)
}

// ---------------------------------------------------------------------------
// StartTransaction status gate
// ---------------------------------------------------------------------------

// TestStartTransaction_RequiresPending verifies that StartTransaction fails
// when the approval status is still PENDING (not yet approved).
func TestStartTransaction_RequiresPending(t *testing.T) {
	db := setupWorkflowDB(t)
	svc := newApprovalServices(t, db)
	ctx := context.Background()

	user := createTestUser(t, db, models.RoleUser)
	ds := createTestDataSource(t, db)

	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  ds.ID,
		QueryText:     "INSERT INTO logs (msg) VALUES ('test')",
		RequestedBy:   user.ID,
		Status:        models.ApprovalStatusPending, // ← not approved yet
		OperationType: models.OperationInsert,
	}
	require.NoError(t, db.Create(approval).Error)

	_, err := svc.StartTransaction(ctx, approval.ID.String(), user.ID.String(), models.AuditModeCountOnly)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be approved before starting a transaction")
}

// TestStartTransaction_RequiresApproved_Rejected verifies that a REJECTED approval
// also cannot start a transaction.
func TestStartTransaction_RequiresApproved_Rejected(t *testing.T) {
	db := setupWorkflowDB(t)
	svc := newApprovalServices(t, db)
	ctx := context.Background()

	user := createTestUser(t, db, models.RoleUser)
	ds := createTestDataSource(t, db)

	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  ds.ID,
		QueryText:     "DELETE FROM events WHERE id = 99",
		RequestedBy:   user.ID,
		Status:        models.ApprovalStatusRejected,
		OperationType: models.OperationDelete,
	}
	require.NoError(t, db.Create(approval).Error)

	_, err := svc.StartTransaction(ctx, approval.ID.String(), user.ID.String(), models.AuditModeCountOnly)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be approved before starting a transaction")
}

// TestStartTransaction_PassesStatusCheck verifies that once an approval is APPROVED
// the status gate is passed (the error happens later due to the test data source
// not having a real connection, not because of the status check).
func TestStartTransaction_PassesStatusCheck(t *testing.T) {
	db := setupWorkflowDB(t)
	svc := newApprovalServices(t, db)
	ctx := context.Background()

	user := createTestUser(t, db, models.RoleUser)
	ds := createTestDataSource(t, db)

	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  ds.ID,
		QueryText:     "UPDATE users SET name = 'test' WHERE id = 1",
		RequestedBy:   user.ID,
		Status:        models.ApprovalStatusApproved, // ← approved
		OperationType: models.OperationUpdate,
	}
	require.NoError(t, db.Create(approval).Error)

	_, err := svc.StartTransaction(ctx, approval.ID.String(), user.ID.String(), models.AuditModeCountOnly)

	// The status check passes. The error (if any) comes from the DB connection
	// failing on a test data source — not from the approval status gate.
	if err != nil {
		assert.NotContains(t, err.Error(), "must be approved before starting a transaction",
			"status check should have passed for an APPROVED approval")
		assert.NotContains(t, err.Error(), "approval request is not pending",
			"old PENDING error should not appear for APPROVED approvals")
	}
}

// ---------------------------------------------------------------------------
// Full approval workflow
// ---------------------------------------------------------------------------

// TestFullApprovalWorkflow_CreateReviewApprove tests the complete flow from
// approval creation through review to approved status.
func TestFullApprovalWorkflow_CreateReviewApprove(t *testing.T) {
	db := setupWorkflowDB(t)
	svc := newApprovalServices(t, db)
	ctx := context.Background()

	requester := createTestUser(t, db, models.RoleUser)
	reviewer := createTestUser(t, db, models.RoleAdmin)
	ds := createTestDataSource(t, db)

	// Step 1: Create approval
	created, err := svc.CreateApprovalRequest(ctx, &ApprovalRequest{
		DataSourceID: ds.ID,
		QuerySQL:     "UPDATE products SET price = 99.99 WHERE id = 5",
		RequestedBy:  requester.ID.String(),
	})
	require.NoError(t, err)
	assert.Equal(t, models.ApprovalStatusPending, created.Status)

	// Step 2: Fetch and confirm PENDING
	fetched, err := svc.GetApproval(ctx, created.ID.String())
	require.NoError(t, err)
	assert.Equal(t, models.ApprovalStatusPending, fetched.Status)

	// Step 3: Reviewer approves
	review, err := svc.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: created.ID,
		ReviewerID: reviewer.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Approved after review",
	})
	require.NoError(t, err)
	assert.Equal(t, models.ApprovalDecisionApproved, review.Decision)

	// Step 4: Approval must now be APPROVED
	final, err := svc.GetApproval(ctx, created.ID.String())
	require.NoError(t, err)
	assert.Equal(t, models.ApprovalStatusApproved, final.Status)
	assert.NotNil(t, final.CompletedAt)
}

// TestFullApprovalWorkflow_Reject tests that a rejection closes the approval.
func TestFullApprovalWorkflow_Reject(t *testing.T) {
	db := setupWorkflowDB(t)
	svc := newApprovalServices(t, db)
	ctx := context.Background()

	requester := createTestUser(t, db, models.RoleUser)
	reviewer := createTestUser(t, db, models.RoleAdmin)
	ds := createTestDataSource(t, db)

	created, err := svc.CreateApprovalRequest(ctx, &ApprovalRequest{
		DataSourceID: ds.ID,
		QuerySQL:     "TRUNCATE TABLE audit_logs",
		RequestedBy:  requester.ID.String(),
	})
	require.NoError(t, err)

	_, err = svc.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: created.ID,
		ReviewerID: reviewer.ID.String(),
		Decision:   models.ApprovalDecisionRejected,
		Comments:   "Too destructive",
	})
	require.NoError(t, err)

	final, err := svc.GetApproval(ctx, created.ID.String())
	require.NoError(t, err)
	assert.Equal(t, models.ApprovalStatusRejected, final.Status)
}

// TestApprovalWorkflow_NoDuplicateReview verifies a reviewer cannot review twice.
func TestApprovalWorkflow_NoDuplicateReview(t *testing.T) {
	db := setupWorkflowDB(t)
	svc := newApprovalServices(t, db)
	ctx := context.Background()

	requester := createTestUser(t, db, models.RoleUser)
	reviewer := createTestUser(t, db, models.RoleAdmin)
	ds := createTestDataSource(t, db)

	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  ds.ID,
		QueryText:     "INSERT INTO items VALUES (1)",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationInsert,
	}
	require.NoError(t, db.Create(approval).Error)

	// First review
	_, err := svc.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: reviewer.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
	})
	require.NoError(t, err)

	// Second review from same reviewer must fail
	_, err = svc.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: reviewer.ID.String(),
		Decision:   models.ApprovalDecisionRejected,
	})
	require.Error(t, err)
	// After the first approval the status is APPROVED, so the "not pending" guard
	// fires before the duplicate-reviewer check. Both errors are acceptable.
	isExpectedErr := strings.Contains(err.Error(), "already reviewed") ||
		strings.Contains(err.Error(), "not pending")
	assert.True(t, isExpectedErr, "unexpected error: %v", err)
}

// TestApprovalWorkflow_CannotReviewNonPending verifies that already-closed approvals
// cannot receive new reviews.
func TestApprovalWorkflow_CannotReviewNonPending(t *testing.T) {
	db := setupWorkflowDB(t)
	svc := newApprovalServices(t, db)
	ctx := context.Background()

	tests := []struct {
		name   string
		status models.ApprovalStatus
	}{
		{"already approved", models.ApprovalStatusApproved},
		{"already rejected", models.ApprovalStatusRejected},
	}

	requester := createTestUser(t, db, models.RoleUser)
	reviewer := createTestUser(t, db, models.RoleAdmin)
	ds := createTestDataSource(t, db)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			approval := &models.ApprovalRequest{
				ID:            uuid.New(),
				DataSourceID:  ds.ID,
				QueryText:     "DELETE FROM tmp WHERE 1=1",
				RequestedBy:   requester.ID,
				Status:        tt.status,
				OperationType: models.OperationDelete,
			}
			require.NoError(t, db.Create(approval).Error)

			_, err := svc.ReviewApproval(ctx, &ReviewInput{
				ApprovalID: approval.ID,
				ReviewerID: reviewer.ID.String(),
				Decision:   models.ApprovalDecisionApproved,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not pending")
		})
	}
}

// ---------------------------------------------------------------------------
// StartTransaction — invalid inputs
// ---------------------------------------------------------------------------

func TestStartTransaction_InvalidInputs(t *testing.T) {
	db := setupWorkflowDB(t)
	svc := newApprovalServices(t, db)
	ctx := context.Background()

	user := createTestUser(t, db, models.RoleUser)
	ds := createTestDataSource(t, db)

	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  ds.ID,
		QueryText:     "UPDATE t SET x=1",
		RequestedBy:   user.ID,
		Status:        models.ApprovalStatusApproved,
		OperationType: models.OperationUpdate,
	}
	require.NoError(t, db.Create(approval).Error)

	tests := []struct {
		name        string
		approvalID  string
		startedBy   string
		errContains string
	}{
		{
			name:        "non-existent approval",
			approvalID:  uuid.New().String(),
			startedBy:   user.ID.String(),
			errContains: "not found",
		},
		{
			name:        "invalid started_by UUID",
			approvalID:  approval.ID.String(),
			startedBy:   "not-a-uuid",
			errContains: "invalid started_by UUID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.StartTransaction(ctx, tt.approvalID, tt.startedBy, models.AuditModeCountOnly)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

// ---------------------------------------------------------------------------
// GetApprovalCounts
// ---------------------------------------------------------------------------

func TestApprovalService_GetApprovalCounts(t *testing.T) {
	db := setupWorkflowDB(t)
	svc := newApprovalServices(t, db)
	ctx := context.Background()

	user := createTestUser(t, db, models.RoleUser)
	ds := createTestDataSource(t, db)

	statuses := []models.ApprovalStatus{
		models.ApprovalStatusPending,
		models.ApprovalStatusPending,
		models.ApprovalStatusApproved,
		models.ApprovalStatusRejected,
	}
	for _, s := range statuses {
		require.NoError(t, db.Create(&models.ApprovalRequest{
			ID:            uuid.New(),
			DataSourceID:  ds.ID,
			QueryText:     "SELECT 1",
			RequestedBy:   user.ID,
			Status:        s,
			OperationType: models.OperationUpdate,
		}).Error)
	}

	counts, err := svc.GetApprovalCounts(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, int64(4), counts["all"])
	assert.Equal(t, int64(2), counts["pending"])
	assert.Equal(t, int64(1), counts["approved"])
	assert.Equal(t, int64(1), counts["rejected"])
}

// ---------------------------------------------------------------------------
// RequiresApproval helper — operation type coverage
// ---------------------------------------------------------------------------

func TestRequiresApproval_AllOperationTypes(t *testing.T) {
	tests := []struct {
		op       models.OperationType
		expected bool
	}{
		{models.OperationSelect, false},
		{models.OperationSet, false},
		{models.OperationInsert, true},
		{models.OperationUpdate, true},
		{models.OperationDelete, true},
		{models.OperationCreateTable, true},
		{models.OperationDropTable, true},
		{models.OperationAlterTable, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.op), func(t *testing.T) {
			assert.Equal(t, tt.expected, RequiresApproval(tt.op))
		})
	}
}
