package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/querybase/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate all models
	// Note: SQLite doesn't support uuid_generate_v4() default, so we'll set IDs explicitly in tests
	err = db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.DataSource{},
		&models.DataSourcePermission{},
		&models.Query{},
		&models.QueryResult{},
		&models.QueryHistory{},
		&models.ApprovalRequest{},
		&models.ApprovalReview{},
		&models.QueryTransaction{},
		&models.NotificationConfig{},
		&models.Notification{},
	)
	require.NoError(t, err)

	return db
}

// createTestUser creates a test user in the database
func createTestUser(t *testing.T, db *gorm.DB, role models.UserRole) *models.User {
	user := &models.User{
		ID:           uuid.New(),
		Email:        uuid.New().String() + "@test.com",
		Username:     uuid.New().String(),
		PasswordHash: "hashed_password",
		FullName:     "Test User",
		Role:         role,
		IsActive:     true,
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

// createTestDataSource creates a test data source
func createTestDataSource(t *testing.T, db *gorm.DB) *models.DataSource {
	ds := &models.DataSource{
		ID:                uuid.New(),
		Name:              "Test Data Source",
		Type:              models.DataSourceTypePostgreSQL,
		Host:              "localhost",
		Port:              5432,
		Username:          "test",
		EncryptedPassword: "encrypted_password",
		DatabaseName:      "testdb",
		IsActive:          true,
	}
	err := db.Create(ds).Error
	require.NoError(t, err)
	return ds
}

// TestApprovalService_CreateApprovalRequest tests creating approval requests
func TestApprovalService_CreateApprovalRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := &QueryService{}
	approvalService := NewApprovalService(db, queryService, nil)

	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	tests := []struct {
		name        string
		request     *ApprovalRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid approval request",
			request: &ApprovalRequest{
				DataSourceID: dataSource.ID,
				QuerySQL:     "INSERT INTO users (name) VALUES ('Test')",
				RequestedBy:  user.ID.String(),
			},
			expectError: false,
		},
		{
			name: "Invalid requested_by UUID",
			request: &ApprovalRequest{
				DataSourceID: dataSource.ID,
				QuerySQL:     "INSERT INTO users (name) VALUES ('Test')",
				RequestedBy:  "invalid-uuid",
			},
			expectError: true,
			errorMsg:    "invalid requested_by UUID",
		},
		{
			name: "Empty query",
			request: &ApprovalRequest{
				DataSourceID: dataSource.ID,
				QuerySQL:     "",
				RequestedBy:  user.ID.String(),
			},
			expectError: false, // Service doesn't validate query, just creates record
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			approval, err := approvalService.CreateApprovalRequest(ctx, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, approval)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, approval)
				assert.NotEqual(t, uuid.UUID{}, approval.ID)
				assert.Equal(t, models.ApprovalStatusPending, approval.Status)
				assert.Equal(t, tt.request.DataSourceID, approval.DataSourceID)
				assert.Equal(t, tt.request.QuerySQL, approval.QueryText)
			}
		})
	}
}

// TestApprovalService_GetApproval tests retrieving approval requests
func TestApprovalService_GetApproval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}
	db := setupTestDB(t)
	queryService := &QueryService{}
	approvalService := NewApprovalService(db, queryService, nil)

	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create approval request
	approval := &models.ApprovalRequest{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		QueryText:    "UPDATE users SET name = 'Test'",
		RequestedBy:  user.ID,
		Status:       models.ApprovalStatusPending,
	}
	err := db.Create(approval).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		approvalID  string
		expectError bool
	}{
		{
			name:        "Existing approval",
			approvalID:  approval.ID.String(),
			expectError: false,
		},
		{
			name:        "Non-existing approval",
			approvalID:  uuid.New().String(),
			expectError: true,
		},
		{
			name:        "Invalid UUID",
			approvalID:  "invalid-uuid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := approvalService.GetApproval(ctx, tt.approvalID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, approval.ID, result.ID)
			}
		})
	}
}

// TestApprovalService_ListApprovals tests listing approval requests
func TestApprovalService_ListApprovals(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}
	db := setupTestDB(t)
	queryService := &QueryService{}
	approvalService := NewApprovalService(db, queryService, nil)

	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create multiple approval requests
	approvals := []*models.ApprovalRequest{
		{
			ID:           uuid.New(),
			DataSourceID: dataSource.ID,
			QueryText:    "INSERT INTO users VALUES (1)",
			RequestedBy:  user.ID,
			Status:       models.ApprovalStatusPending,
		},
		{
			ID:           uuid.New(),
			DataSourceID: dataSource.ID,
			QueryText:    "UPDATE users SET name = 'Test'",
			RequestedBy:  user.ID,
			Status:       models.ApprovalStatusApproved,
		},
		{
			ID:           uuid.New(),
			DataSourceID: dataSource.ID,
			QueryText:    "DELETE FROM users WHERE id = 1",
			RequestedBy:  user.ID,
			Status:       models.ApprovalStatusPending,
		},
	}

	for _, a := range approvals {
		err := db.Create(a).Error
		require.NoError(t, err)
	}

	tests := []struct {
		name        string
		filter      *ApprovalFilter
		minCount    int
		maxCount    int
		expectError bool
	}{
		{
			name: "List all approvals",
			filter: &ApprovalFilter{
				Limit:  10,
				Offset: 0,
			},
			minCount:    3,
			maxCount:    3,
			expectError: false,
		},
		{
			name: "Filter by status - pending",
			filter: &ApprovalFilter{
				Status: string(models.ApprovalStatusPending),
				Limit:  10,
				Offset: 0,
			},
			minCount:    2,
			maxCount:    2,
			expectError: false,
		},
		{
			name: "Filter by status - approved",
			filter: &ApprovalFilter{
				Status: string(models.ApprovalStatusApproved),
				Limit:  10,
				Offset: 0,
			},
			minCount:    1,
			maxCount:    1,
			expectError: false,
		},
		{
			name: "Filter with pagination",
			filter: &ApprovalFilter{
				Limit:  2,
				Offset: 0,
			},
			minCount:    0,
			maxCount:    2,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, total, err := approvalService.ListApprovals(ctx, tt.filter)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, len(result), tt.minCount)
				assert.LessOrEqual(t, len(result), tt.maxCount)
				assert.GreaterOrEqual(t, total, int64(tt.minCount))
			}
		})
	}
}

// TestApprovalService_ReviewApproval tests reviewing approval requests
func TestApprovalService_ReviewApproval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}
	db := setupTestDB(t)
	queryService := &QueryService{}
	approvalService := NewApprovalService(db, queryService, nil)

	user := createTestUser(t, db, models.RoleUser)
	reviewer := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	// Create approval request
	approval := &models.ApprovalRequest{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		QueryText:    "UPDATE users SET name = 'Test'",
		RequestedBy:  user.ID,
		Status:       models.ApprovalStatusPending,
	}
	err := db.Create(approval).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		approvalID  uuid.UUID
		review      *ReviewInput
		expectError bool
		errorMsg    string
	}{
		{
			name:       "Valid approval",
			approvalID: approval.ID,
			review: &ReviewInput{
				ApprovalID: approval.ID,
				ReviewerID: reviewer.ID.String(),
				Decision:   models.ApprovalDecisionApproved,
				Comments:   "Looks good",
			},
			expectError: false,
		},
		{
			name:       "Valid rejection",
			approvalID: approval.ID,
			review: &ReviewInput{
				ApprovalID: approval.ID,
				ReviewerID: reviewer.ID.String(),
				Decision:   models.ApprovalDecisionRejected,
				Comments:   "Not safe",
			},
			expectError: false,
		},
		{
			name:       "Non-existing approval",
			approvalID: uuid.New(),
			review: &ReviewInput{
				ApprovalID: uuid.New(),
				ReviewerID: reviewer.ID.String(),
				Decision:   models.ApprovalDecisionApproved,
				Comments:   "Test",
			},
			expectError: true,
			errorMsg:    "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// For multiple reviews on same approval, we need different approval instances
			if tt.name == "Valid rejection" {
				// Create new approval for this test
				newApproval := &models.ApprovalRequest{
					ID:           uuid.New(),
					DataSourceID: dataSource.ID,
					QueryText:    "DELETE FROM users",
					RequestedBy:  user.ID,
					Status:       models.ApprovalStatusPending,
				}
				err := db.Create(newApproval).Error
				require.NoError(t, err)
				tt.approvalID = newApproval.ID
				tt.review.ApprovalID = newApproval.ID
			}

			result, err := approvalService.ReviewApproval(ctx, tt.review)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.review.Decision, result.Decision)
				assert.Equal(t, tt.review.Comments, result.Comments)

				// Check that approval status was updated
				var updatedApproval models.ApprovalRequest
				err := db.First(&updatedApproval, "id = ?", tt.approvalID).Error
				require.NoError(t, err)
				assert.NotEqual(t, models.ApprovalStatusPending, updatedApproval.Status)
			}
		})
	}
}

// TestApprovalService_GetEligibleApprovers tests getting eligible approvers
func TestApprovalService_GetEligibleApprovers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}
	db := setupTestDB(t)
	queryService := &QueryService{}
	approvalService := NewApprovalService(db, queryService, nil)

	// Create users and data source
	user2 := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	// Create group with approve permission
	group := &models.Group{
		ID:   uuid.New(),
		Name: "Approvers",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	// Add user2 to group (many-to-many relationship)
	err = db.Model(user2).Association("Groups").Append(group)
	require.NoError(t, err)

	// Grant approve permission to group
	permission := &models.DataSourcePermission{
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   true,
	}
	err = db.Create(permission).Error
	require.NoError(t, err)

	// Test getting eligible approvers
	ctx := context.Background()
	approvers, err := approvalService.GetEligibleApprovers(ctx, dataSource.ID.String())

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(approvers), 1)

	// Check that user2 (admin) is in the list
	found := false
	for _, approver := range approvers {
		if approver.ID == user2.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Admin user should be in eligible approvers list")
}

// TestApprovalService_StartTransaction tests starting transactions
func TestApprovalService_StartTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}
	db := setupTestDB(t)
	queryService := &QueryService{}
	approvalService := NewApprovalService(db, queryService, nil)

	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create approval request
	approval := &models.ApprovalRequest{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		QueryText:    "INSERT INTO users (name) VALUES ('Test')",
		RequestedBy:  user.ID,
		Status:       models.ApprovalStatusPending,
	}
	err := db.Create(approval).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		approvalID  string
		startedBy   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Non-existing approval",
			approvalID:  uuid.New().String(),
			startedBy:   user.ID.String(),
			expectError: true,
			errorMsg:    "not found",
		},
		{
			name:        "Invalid started_by UUID",
			approvalID:  approval.ID.String(),
			startedBy:   "invalid-uuid",
			expectError: true,
			errorMsg:    "invalid started_by UUID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			transaction, err := approvalService.StartTransaction(ctx, tt.approvalID, tt.startedBy)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, transaction)
			}
		})
	}
}

// TestApprovalService_UpdateApprovalStatus tests status updates after reviews
func TestApprovalService_UpdateApprovalStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}
	db := setupTestDB(t)
	queryService := &QueryService{}
	approvalService := NewApprovalService(db, queryService, nil)

	user := createTestUser(t, db, models.RoleUser)
	reviewer := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	t.Run("Approval on approve decision", func(t *testing.T) {
		approval := &models.ApprovalRequest{
			ID:           uuid.New(),
			DataSourceID: dataSource.ID,
			QueryText:    "INSERT INTO users VALUES (1)",
			RequestedBy:  user.ID,
			Status:       models.ApprovalStatusPending,
		}
		err := db.Create(approval).Error
		require.NoError(t, err)

		ctx := context.Background()
		review := &ReviewInput{
			ApprovalID: approval.ID,
			ReviewerID: reviewer.ID.String(),
			Decision:   models.ApprovalDecisionApproved,
			Comments:   "Approved",
		}

		_, err = approvalService.ReviewApproval(ctx, review)
		assert.NoError(t, err)

		// Check approval status
		var updatedApproval models.ApprovalRequest
		err = db.First(&updatedApproval, "id = ?", approval.ID).Error
		require.NoError(t, err)
		assert.Equal(t, models.ApprovalStatusApproved, updatedApproval.Status)
	})

	t.Run("Rejection on reject decision", func(t *testing.T) {
		approval := &models.ApprovalRequest{
			ID:           uuid.New(),
			DataSourceID: dataSource.ID,
			QueryText:    "DELETE FROM users",
			RequestedBy:  user.ID,
			Status:       models.ApprovalStatusPending,
		}
		err := db.Create(approval).Error
		require.NoError(t, err)

		ctx := context.Background()
		review := &ReviewInput{
			ApprovalID: approval.ID,
			ReviewerID: reviewer.ID.String(),
			Decision:   models.ApprovalDecisionRejected,
			Comments:   "Rejected",
		}

		_, err = approvalService.ReviewApproval(ctx, review)
		assert.NoError(t, err)

		// Check approval status
		var updatedApproval models.ApprovalRequest
		err = db.First(&updatedApproval, "id = ?", approval.ID).Error
		require.NoError(t, err)
		assert.Equal(t, models.ApprovalStatusRejected, updatedApproval.Status)
	})
}

// TestApprovalService_DuplicateReview tests that duplicate reviews are rejected
func TestApprovalService_DuplicateReview(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}
	db := setupTestDB(t)
	queryService := &QueryService{}
	approvalService := NewApprovalService(db, queryService, nil)

	user := createTestUser(t, db, models.RoleUser)
	reviewer := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	approval := &models.ApprovalRequest{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		QueryText:    "UPDATE users SET name = 'Test'",
		RequestedBy:  user.ID,
		Status:       models.ApprovalStatusPending,
	}
	err := db.Create(approval).Error
	require.NoError(t, err)

	ctx := context.Background()

	// First review - manually insert to avoid status update
	firstReview := &models.ApprovalReview{
		ID:         uuid.New(),
		ApprovalID: approval.ID,
		ReviewerID: reviewer.ID,
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "First review",
		ReviewedAt: time.Now(),
	}
	err = db.Create(firstReview).Error
	require.NoError(t, err)

	// Second review from same reviewer - should fail
	review2 := &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: reviewer.ID.String(),
		Decision:   models.ApprovalDecisionRejected,
		Comments:   "Second review",
	}
	_, err = approvalService.ReviewApproval(ctx, review2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already reviewed")
}

// TestApprovalService_ReviewNonPendingApproval tests reviewing non-pending approvals
func TestApprovalService_ReviewNonPendingApproval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}
	db := setupTestDB(t)
	queryService := &QueryService{}
	approvalService := NewApprovalService(db, queryService, nil)

	user := createTestUser(t, db, models.RoleUser)
	reviewer := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	// Create already approved approval
	approval := &models.ApprovalRequest{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		QueryText:    "INSERT INTO users VALUES (1)",
		RequestedBy:  user.ID,
		Status:       models.ApprovalStatusApproved,
	}
	err := db.Create(approval).Error
	require.NoError(t, err)

	ctx := context.Background()
	review := &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: reviewer.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Trying to review already approved request",
	}

	_, err = approvalService.ReviewApproval(ctx, review)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not pending")
}
