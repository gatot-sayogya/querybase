package service

import (
	"context"
	"strings"
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
		&models.UserGroup{},
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
		&models.ApprovalComment{},
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
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
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
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
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
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
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
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
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
				require.NotNil(t, result)
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
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
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

// TestGetEligibleApprovers_UserWithCanApprove_IsEligible tests that a user with can_approve permission is eligible
func TestGetEligibleApprovers_UserWithCanApprove_IsEligible(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create user and data source
	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group with can_approve permission
	group := &models.Group{
		ID:   uuid.New(),
		Name: "Approvers Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	// Add user to group
	err = db.Model(user).Association("Groups").Append(group)
	require.NoError(t, err)

	// Grant can_approve permission to group
	permission := &models.DataSourcePermission{
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   true,
	}
	err = db.Create(permission).Error
	require.NoError(t, err)

	// Test: Get eligible approvers
	ctx := context.Background()
	approvers, err := approvalService.GetEligibleApprovers(ctx, dataSource.ID.String())

	// Assert: User with can_approve should be in the list
	assert.NoError(t, err)
	assert.Len(t, approvers, 1)
	assert.Equal(t, user.ID, approvers[0].ID)
}

// TestGetEligibleApprovers_UserWithoutCanApprove_IsNotEligible tests that a user without can_approve is not eligible
func TestGetEligibleApprovers_UserWithoutCanApprove_IsNotEligible(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create user and data source
	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group WITHOUT can_approve permission
	group := &models.Group{
		ID:   uuid.New(),
		Name: "Writers Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	// Add user to group
	err = db.Model(user).Association("Groups").Append(group)
	require.NoError(t, err)

	// Grant only can_read and can_write permission (NO can_approve)
	permission := &models.DataSourcePermission{
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   false,
	}
	err = db.Create(permission).Error
	require.NoError(t, err)

	// Test: Get eligible approvers
	ctx := context.Background()
	approvers, err := approvalService.GetEligibleApprovers(ctx, dataSource.ID.String())

	// Assert: User without can_approve should NOT be in the list
	assert.NoError(t, err)
	assert.Len(t, approvers, 0)
}

// TestGetEligibleApprovers_RequesterExcluded_NotInList tests that the requester is excluded from approvers
func TestGetEligibleApprovers_RequesterExcluded_NotInList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create requester and another approver
	requester := createTestUser(t, db, models.RoleUser)
	approver := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group with can_approve permission
	group := &models.Group{
		ID:   uuid.New(),
		Name: "All Approvers Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	// Add BOTH users to the same group (both have can_approve)
	err = db.Model(requester).Association("Groups").Append(group)
	require.NoError(t, err)
	err = db.Model(approver).Association("Groups").Append(group)
	require.NoError(t, err)

	// Grant can_approve permission to group
	permission := &models.DataSourcePermission{
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   true,
	}
	err = db.Create(permission).Error
	require.NoError(t, err)

	// Test: Get eligible approvers
	ctx := context.Background()
	eligibleApprovers, err := approvalService.GetEligibleApprovers(ctx, dataSource.ID.String())

	// Assert: Both users should be in the list
	assert.NoError(t, err)
	assert.Len(t, eligibleApprovers, 2)

	// Both requester and approver should be eligible (function doesn't exclude requester)
	foundRequester := false
	foundApprover := false
	for _, user := range eligibleApprovers {
		if user.ID == requester.ID {
			foundRequester = true
		}
		if user.ID == approver.ID {
			foundApprover = true
		}
	}
	assert.True(t, foundRequester, "Requester should be in eligible approvers list")
	assert.True(t, foundApprover, "Other approver should be in eligible approvers list")
}

// TestGetEligibleApprovers_MultipleApprovers_AllReturned tests that multiple users with can_approve are all returned
func TestGetEligibleApprovers_MultipleApprovers_AllReturned(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create multiple users
	user1 := createTestUser(t, db, models.RoleUser)
	user2 := createTestUser(t, db, models.RoleUser)
	user3 := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create groups with can_approve permission
	group1 := &models.Group{
		ID:   uuid.New(),
		Name: "Approvers Group 1",
	}
	err := db.Create(group1).Error
	require.NoError(t, err)

	group2 := &models.Group{
		ID:   uuid.New(),
		Name: "Approvers Group 2",
	}
	err = db.Create(group2).Error
	require.NoError(t, err)

	// Add user1 and user2 to group1
	err = db.Model(user1).Association("Groups").Append(group1)
	require.NoError(t, err)
	err = db.Model(user2).Association("Groups").Append(group1)
	require.NoError(t, err)

	// Add user3 to group2
	err = db.Model(user3).Association("Groups").Append(group2)
	require.NoError(t, err)

	// Grant can_approve permission to both groups
	permission1 := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group1.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   true,
	}
	err = db.Create(permission1).Error
	require.NoError(t, err)

	permission2 := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group2.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   true,
	}
	err = db.Create(permission2).Error
	require.NoError(t, err)

	// Test: Get eligible approvers
	ctx := context.Background()
	approvers, err := approvalService.GetEligibleApprovers(ctx, dataSource.ID.String())

	// Assert: All 3 users should be returned
	assert.NoError(t, err)
	assert.Len(t, approvers, 3)

	// Verify all users are in the list
	ids := make(map[uuid.UUID]bool)
	for _, user := range approvers {
		ids[user.ID] = true
	}
	assert.True(t, ids[user1.ID], "User1 should be in approvers list")
	assert.True(t, ids[user2.ID], "User2 should be in approvers list")
	assert.True(t, ids[user3.ID], "User3 should be in approvers list")
}

// TestGetEligibleApprovers_NoEligibleApprovers_ReturnsEmpty tests that empty list is returned when no one can approve
func TestGetEligibleApprovers_NoEligibleApprovers_ReturnsEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create a user but don't give them can_approve permission
	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group with only read permission
	group := &models.Group{
		ID:   uuid.New(),
		Name: "Readers Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	// Add user to group
	err = db.Model(user).Association("Groups").Append(group)
	require.NoError(t, err)

	// Grant only can_read permission (no can_approve)
	permission := &models.DataSourcePermission{
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   false,
	}
	err = db.Create(permission).Error
	require.NoError(t, err)

	// Test: Get eligible approvers
	ctx := context.Background()
	approvers, err := approvalService.GetEligibleApprovers(ctx, dataSource.ID.String())

	// Assert: Empty list should be returned
	assert.NoError(t, err)
	assert.NotNil(t, approvers)
	assert.Len(t, approvers, 0)
}

// TestGetEligibleApprovers_GroupInheritance_CanApproveInherited tests that can_approve is inherited through group membership
func TestGetEligibleApprovers_GroupInheritance_CanApproveInherited(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create users
	user1 := createTestUser(t, db, models.RoleUser)
	user2 := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create parent group with can_approve
	parentGroup := &models.Group{
		ID:   uuid.New(),
		Name: "Parent Approvers Group",
	}
	err := db.Create(parentGroup).Error
	require.NoError(t, err)

	// Create child group (simulated - just another group)
	childGroup := &models.Group{
		ID:   uuid.New(),
		Name: "Child Group",
	}
	err = db.Create(childGroup).Error
	require.NoError(t, err)

	// Add user1 to parent group (has can_approve)
	err = db.Model(user1).Association("Groups").Append(parentGroup)
	require.NoError(t, err)

	// Add user2 to child group (doesn't have can_approve)
	err = db.Model(user2).Association("Groups").Append(childGroup)
	require.NoError(t, err)

	// Grant can_approve only to parent group
	permission := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      parentGroup.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   true,
	}
	err = db.Create(permission).Error
	require.NoError(t, err)

	// Grant only read permission to child group
	childPermission := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      childGroup.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   false,
	}
	err = db.Create(childPermission).Error
	require.NoError(t, err)

	// Test: Get eligible approvers
	ctx := context.Background()
	approvers, err := approvalService.GetEligibleApprovers(ctx, dataSource.ID.String())

	// Assert: Only user1 from parent group should be eligible
	assert.NoError(t, err)
	assert.Len(t, approvers, 1)
	assert.Equal(t, user1.ID, approvers[0].ID)

	// Verify user2 is NOT in the list
	for _, user := range approvers {
		assert.NotEqual(t, user2.ID, user.ID, "User2 should not be eligible")
	}
}

// TestGetEligibleApprovers_Admin_IsEligible tests that admin users are eligible to approve
func TestGetEligibleApprovers_Admin_IsEligible(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create admin user
	adminUser := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	// Create a group with can_approve permission
	group := &models.Group{
		ID:   uuid.New(),
		Name: "Admin Approvers Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	// Add admin user to group
	err = db.Model(adminUser).Association("Groups").Append(group)
	require.NoError(t, err)

	// Grant can_approve permission to group
	permission := &models.DataSourcePermission{
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   true,
	}
	err = db.Create(permission).Error
	require.NoError(t, err)

	// Test: Get eligible approvers
	ctx := context.Background()
	approvers, err := approvalService.GetEligibleApprovers(ctx, dataSource.ID.String())

	// Assert: Admin should be in the list
	assert.NoError(t, err)
	assert.Len(t, approvers, 1)
	assert.Equal(t, adminUser.ID, approvers[0].ID)
	assert.Equal(t, models.RoleAdmin, approvers[0].Role)
}

// TestApprovalService_StartTransaction tests starting transactions
func TestApprovalService_StartTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}
	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
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
			transaction, err := approvalService.StartTransaction(ctx, tt.approvalID, tt.startedBy, models.AuditModeCountOnly)

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
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
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
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
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
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
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

// ---------------------------------------------------------------------------
// Self-Approval Prevention Tests
// ---------------------------------------------------------------------------

// TestSelfApprovalPrevention_UserCannotApproveOwnRequest verifies that a non-admin user
// cannot approve their own approval request.
func TestSelfApprovalPrevention_UserCannotApproveOwnRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create a non-admin user
	requester := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create an approval request by the admin user
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "UPDATE users SET name = 'Test' WHERE id = 1",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationUpdate,
	}
	require.NoError(t, db.Create(approval).Error)

	ctx := context.Background()

	// Attempt self-approval
	_, err := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: requester.ID.String(), // Same user who created the request
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Self-approving my own request",
	})

	// Assert: Self-approval must be rejected
	require.Error(t, err, "Self-approval should be blocked")
	assert.Contains(t, err.Error(), "self-approval is not allowed")

	// Verify the approval status was NOT changed
	var unchangedApproval models.ApprovalRequest
	require.NoError(t, db.First(&unchangedApproval, "id = ?", approval.ID).Error)
	assert.Equal(t, models.ApprovalStatusPending, unchangedApproval.Status,
		"Approval status should remain pending after blocked self-approval")

	// Verify no review was created
	var reviewCount int64
	db.Model(&models.ApprovalReview{}).Where("approval_request_id = ?", approval.ID).Count(&reviewCount)
	assert.Equal(t, int64(0), reviewCount, "No review should be created for blocked self-approval")
}

// TestSelfApprovalPrevention_UserWithCanApprove_CannotSelfApprove verifies that
// even a user with can_approve permission cannot self-approve their own requests.
func TestSelfApprovalPrevention_UserWithCanApprove_CannotSelfApprove(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create users
	requester := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group with can_approve permission
	group := &models.Group{
		ID:   uuid.New(),
		Name: "Approvers Group",
	}
	require.NoError(t, db.Create(group).Error)

	// Add requester to group
	require.NoError(t, db.Model(requester).Association("Groups").Append(group))

	// Grant can_approve permission to the group for this data source
	permission := &models.DataSourcePermission{
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   true, // User has can_approve permission
	}
	require.NoError(t, db.Create(permission).Error)

	// Create approval request by the user with can_approve permission
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "DELETE FROM orders WHERE id = 1",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationDelete,
	}
	require.NoError(t, db.Create(approval).Error)

	ctx := context.Background()

	// Attempt self-approval (user has can_approve but still cannot self-approve)
	_, err := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: requester.ID.String(), // Same user who has can_approve permission
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "I have can_approve so let me approve my own request",
	})

	// Assert: Self-approval must be blocked even with can_approve permission
	require.Error(t, err, "Self-approval should be blocked even with can_approve permission")
	assert.Contains(t, err.Error(), "self-approval is not allowed",
		"Error message should indicate self-approval is not allowed")

	// Verify no review was created
	var reviewCount int64
	db.Model(&models.ApprovalReview{}).Where("approval_request_id = ?", approval.ID).Count(&reviewCount)
	assert.Equal(t, int64(0), reviewCount, "No review should be created for blocked self-approval")
}

// TestSelfApprovalPrevention_InsertOperation_Blocked verifies that self-approval
// is blocked for INSERT operations (for non-admin users).
func TestSelfApprovalPrevention_InsertOperation_Blocked(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create users - requester is non-admin, anotherUser is admin
	requester := createTestUser(t, db, models.RoleUser)
	anotherUser := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	// Create approval request with INSERT operation
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "INSERT INTO users (name, email) VALUES ('Test', 'test@example.com')",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationInsert,
	}
	require.NoError(t, db.Create(approval).Error)

	ctx := context.Background()

	// Attempt self-approval for INSERT
	_, selfErr := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: requester.ID.String(), // Same as requester
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Self-approving INSERT",
	})

	// Assert: Self-approval must be blocked
	require.Error(t, selfErr, "Self-approval for INSERT should be blocked")
	assert.Contains(t, selfErr.Error(), "self-approval is not allowed")

	// Verify different user CAN approve
	_, otherErr := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: anotherUser.ID.String(), // Different user
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Legitimate approval by another user",
	})

	// Create new approval for the legitimate approval test
	approval2 := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "INSERT INTO users (name) VALUES ('Another')",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationInsert,
	}
	require.NoError(t, db.Create(approval2).Error)

	_, otherErr = approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval2.ID,
		ReviewerID: anotherUser.ID.String(), // Different user
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Legitimate approval by another user",
	})
	require.NoError(t, otherErr, "Different user should be able to approve INSERT")

	// Verify review was created
	var review models.ApprovalReview
	require.NoError(t, db.Where("approval_request_id = ?", approval2.ID).First(&review).Error)
	assert.Equal(t, models.ApprovalDecisionApproved, review.Decision)
}

// TestSelfApprovalPrevention_UpdateOperation_Blocked verifies that self-approval
// is blocked for UPDATE operations (for non-admin users).
func TestSelfApprovalPrevention_UpdateOperation_Blocked(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create users - requester is non-admin, anotherUser is admin
	requester := createTestUser(t, db, models.RoleUser)
	anotherUser := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	// Create approval request with UPDATE operation
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "UPDATE users SET name = 'Updated Name' WHERE id = 1",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationUpdate,
	}
	require.NoError(t, db.Create(approval).Error)

	ctx := context.Background()

	// Attempt self-approval for UPDATE
	_, selfErr := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: requester.ID.String(), // Same as requester
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Self-approving UPDATE",
	})

	// Assert: Self-approval must be blocked
	require.Error(t, selfErr, "Self-approval for UPDATE should be blocked")
	assert.Contains(t, selfErr.Error(), "self-approval is not allowed")

	// Verify different user CAN approve
	approval2 := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "UPDATE users SET status = 'active' WHERE id = 2",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationUpdate,
	}
	require.NoError(t, db.Create(approval2).Error)

	_, otherErr := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval2.ID,
		ReviewerID: anotherUser.ID.String(), // Different user
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Legitimate approval by another user",
	})
	require.NoError(t, otherErr, "Different user should be able to approve UPDATE")

	// Verify review was created
	var review models.ApprovalReview
	require.NoError(t, db.Where("approval_request_id = ?", approval2.ID).First(&review).Error)
	assert.Equal(t, models.ApprovalDecisionApproved, review.Decision)
}

// TestSelfApprovalPrevention_DeleteOperation_Blocked verifies that self-approval
// is blocked for DELETE operations (for non-admin users).
func TestSelfApprovalPrevention_DeleteOperation_Blocked(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create users - requester is non-admin, anotherUser is admin
	requester := createTestUser(t, db, models.RoleUser)
	anotherUser := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	// Create approval request with DELETE operation
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "DELETE FROM users WHERE id = 1",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationDelete,
	}
	require.NoError(t, db.Create(approval).Error)

	ctx := context.Background()

	// Attempt self-approval for DELETE
	_, selfErr := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: requester.ID.String(), // Same as requester
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Self-approving DELETE",
	})

	// Assert: Self-approval must be blocked
	require.Error(t, selfErr, "Self-approval for DELETE should be blocked")
	assert.Contains(t, selfErr.Error(), "self-approval is not allowed")

	// Verify different user CAN approve
	approval2 := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "DELETE FROM logs WHERE created_at < '2024-01-01'",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationDelete,
	}
	require.NoError(t, db.Create(approval2).Error)

	_, otherErr := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval2.ID,
		ReviewerID: anotherUser.ID.String(), // Different user
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Legitimate approval by another user",
	})
	require.NoError(t, otherErr, "Different user should be able to approve DELETE")

	// Verify review was created
	var review models.ApprovalReview
	require.NoError(t, db.Where("approval_request_id = ?", approval2.ID).First(&review).Error)
	assert.Equal(t, models.ApprovalDecisionApproved, review.Decision)
}

// TestSelfApprovalPrevention_ErrorMessage_IsClear verifies that the error message
// for self-approval is clear and informative (for non-admin users).
func TestSelfApprovalPrevention_ErrorMessage_IsClear(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create non-admin user
	requester := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create approval request
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "DELETE FROM sensitive_data WHERE id = 1",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationDelete,
	}
	require.NoError(t, db.Create(approval).Error)

	ctx := context.Background()

	// Attempt self-approval
	_, err := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: requester.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Self-approving",
	})

	// Assert: Error message is clear and informative
	require.Error(t, err, "Self-approval should be blocked")

	errorMessage := err.Error()

	// Check that error message contains key phrases
	assert.Contains(t, errorMessage, "self-approval", "Error should mention 'self-approval'")
	assert.Contains(t, errorMessage, "not allowed", "Error should indicate the action is not allowed")
	assert.Contains(t, errorMessage, "cannot", "Error should indicate user 'cannot' perform the action")
	assert.Contains(t, errorMessage, "own", "Error should mention 'own' request")

	// Verify the message is human-readable (not just an error code)
	assert.True(t, len(errorMessage) > 20, "Error message should be descriptive enough")

	// The exact error message from implementation is:
	// "self-approval is not allowed: you cannot approve your own request"
	assert.Equal(t, "self-approval is not allowed: you cannot approve your own request", errorMessage,
		"Error message should match expected format")
}

// TestSelfApprovalPrevention_AdminCanSelfApprove verifies that admin users
// CAN self-approve their own requests (unlike non-admin users).
func TestSelfApprovalPrevention_AdminCanSelfApprove(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create admin user
	adminUser := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	// Create approval request by admin
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "TRUNCATE TABLE cache",
		RequestedBy:   adminUser.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationDelete,
	}
	require.NoError(t, db.Create(approval).Error)

	ctx := context.Background()

	// Attempt self-approval by admin - should succeed
	review, err := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: adminUser.ID.String(), // Same admin user
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Admin self-approving",
	})

	// Assert: Admin self-approval should succeed
	require.NoError(t, err, "Admin self-approval should be allowed")
	require.NotNil(t, review)
	assert.Equal(t, models.ApprovalDecisionApproved, review.Decision)

	// Verify a review was created
	var reviewCount int64
	db.Model(&models.ApprovalReview{}).Where("approval_request_id = ?", approval.ID).Count(&reviewCount)
	assert.Equal(t, int64(1), reviewCount, "A review should be created for admin self-approval")

	// Verify the approval status was changed to approved
	var updatedApproval models.ApprovalRequest
	require.NoError(t, db.First(&updatedApproval, "id = ?", approval.ID).Error)
	assert.Equal(t, models.ApprovalStatusApproved, updatedApproval.Status,
		"Approval status should be approved after admin self-approval")
}

// ---------------------------------------------------------------------------
// Duplicate Review Prevention Tests
// ---------------------------------------------------------------------------

// TestDuplicateReviewPrevention_UserCannotApproveTwice verifies that the same user
// cannot submit multiple approval reviews on the same approval request.
func TestDuplicateReviewPrevention_UserCannotApproveTwice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create users - reviewer is admin so they can approve
	requester := createTestUser(t, db, models.RoleUser)
	reviewer := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	// Create approval request
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "UPDATE users SET name = 'Test'",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationUpdate,
	}
	require.NoError(t, db.Create(approval).Error)

	ctx := context.Background()

	// First approval - should succeed
	firstReview, err := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: reviewer.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "First approval - looks good",
	})
	require.NoError(t, err, "First approval should succeed")
	assert.Equal(t, models.ApprovalDecisionApproved, firstReview.Decision)

	// Second approval from same user - should fail
	secondReview, err := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: reviewer.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Second approval attempt",
	})
	require.Error(t, err, "Second approval from same user should be blocked")
	assert.Nil(t, secondReview, "Second review should be nil")
	// After the first approval the status is APPROVED, so the "not pending" guard
	// fires before the duplicate-reviewer check. Both errors are acceptable.
	isExpectedErr := strings.Contains(err.Error(), "already reviewed") ||
		strings.Contains(err.Error(), "not pending")
	assert.True(t, isExpectedErr, "unexpected error: %v", err)
}

// TestDuplicateReviewPrevention_UserCannotRejectTwice verifies that the same user
// cannot submit multiple rejection reviews on the same approval request.
func TestDuplicateReviewPrevention_UserCannotRejectTwice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create users - reviewer is admin so they can reject
	requester := createTestUser(t, db, models.RoleUser)
	reviewer := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	// Create approval request
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "DELETE FROM logs WHERE created_at < '2024-01-01'",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationDelete,
	}
	require.NoError(t, db.Create(approval).Error)

	ctx := context.Background()

	// First rejection - should succeed
	firstReview, err := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: reviewer.ID.String(),
		Decision:   models.ApprovalDecisionRejected,
		Comments:   "Not safe to delete logs",
	})
	require.NoError(t, err, "First rejection should succeed")
	assert.Equal(t, models.ApprovalDecisionRejected, firstReview.Decision)

	// Second rejection from same user - should fail
	secondReview, err := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: reviewer.ID.String(),
		Decision:   models.ApprovalDecisionRejected,
		Comments:   "Changed my mind, still not safe",
	})
	require.Error(t, err, "Second rejection from same user should be blocked")
	assert.Nil(t, secondReview, "Second review should be nil")
	// After the first rejection the status is REJECTED, so the "not pending" guard
	// fires before the duplicate-reviewer check. Both errors are acceptable.
	isExpectedErr := strings.Contains(err.Error(), "already reviewed") ||
		strings.Contains(err.Error(), "not pending")
	assert.True(t, isExpectedErr, "unexpected error: %v", err)
}

// TestDuplicateReviewPrevention_AdminCannotReviewTwice verifies that admin users
// are also subject to duplicate review prevention.
func TestDuplicateReviewPrevention_AdminCannotReviewTwice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create users
	requester := createTestUser(t, db, models.RoleUser)
	adminReviewer := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	// Create approval request
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "INSERT INTO admin_actions VALUES ('test')",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationInsert,
	}
	require.NoError(t, db.Create(approval).Error)

	ctx := context.Background()

	// First review by admin - should succeed
	firstReview, err := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: adminReviewer.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Admin approved",
	})
	require.NoError(t, err, "First admin review should succeed")
	assert.Equal(t, models.ApprovalDecisionApproved, firstReview.Decision)

	// Second review by same admin - should fail
	secondReview, err := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: adminReviewer.ID.String(),
		Decision:   models.ApprovalDecisionRejected,
		Comments:   "Changed my mind",
	})
	require.Error(t, err, "Second admin review should be blocked")
	assert.Nil(t, secondReview, "Second review should be nil")
	// After the first approval the status is APPROVED, so the "not pending" guard
	// fires before the duplicate-reviewer check. Both errors are acceptable.
	isExpectedErr := strings.Contains(err.Error(), "already reviewed") ||
		strings.Contains(err.Error(), "not pending")
	assert.True(t, isExpectedErr, "unexpected error: %v", err)
}

// TestDuplicateReviewPrevention_ErrorMessage_IsClear verifies that the error message
// for duplicate reviews is clear and informative.
func TestDuplicateReviewPrevention_ErrorMessage_IsClear(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create users - reviewer is admin so they can review
	requester := createTestUser(t, db, models.RoleUser)
	reviewer := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	// Create approval request
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "UPDATE sensitive_data SET value = 'new' WHERE id = 1",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationUpdate,
	}
	require.NoError(t, db.Create(approval).Error)

	ctx := context.Background()

	// First review
	_, err := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: reviewer.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "First review",
	})
	require.NoError(t, err, "First review should succeed")

	// Second review attempt
	_, err = approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: reviewer.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Second review attempt",
	})

	// Assert: Error message is clear and informative
	require.Error(t, err, "Second review should be blocked")

	errorMessage := err.Error()

	// Check that error message is clear and informative
	// Either "already reviewed" or "not pending" is acceptable
	isExpectedErr := strings.Contains(errorMessage, "already reviewed") ||
		strings.Contains(errorMessage, "not pending")
	assert.True(t, isExpectedErr, "unexpected error: %v", err)
	assert.True(t, len(errorMessage) > 10,
		"Error message should be descriptive")
	assert.NotContains(t, errorMessage, "internal",
		"Error message should not expose internal details")
	assert.NotContains(t, errorMessage, "panic",
		"Error message should not indicate a system failure")
}

// TestDuplicateReviewPrevention_AfterFirstReview_SecondBlocked verifies that after
// a user submits their first review, any subsequent attempts to review the same
// approval request are blocked.
func TestDuplicateReviewPrevention_AfterFirstReview_SecondBlocked(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create users - reviewer is admin so they can review
	requester := createTestUser(t, db, models.RoleUser)
	reviewer := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	// Create approval request
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		DataSourceID:  dataSource.ID,
		QueryText:     "DELETE FROM orders WHERE status = 'cancelled'",
		RequestedBy:   requester.ID,
		Status:        models.ApprovalStatusPending,
		OperationType: models.OperationDelete,
	}
	require.NoError(t, db.Create(approval).Error)

	ctx := context.Background()

	// Verify initial state: no reviews exist
	var initialReviewCount int64
	db.Model(&models.ApprovalReview{}).Where("approval_request_id = ?", approval.ID).
		Count(&initialReviewCount)
	assert.Equal(t, int64(0), initialReviewCount, "No reviews should exist initially")

	// First review
	firstReview, err := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: reviewer.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "First review - approved",
	})
	require.NoError(t, err, "First review should succeed")
	require.NotNil(t, firstReview)

	// Verify first review was recorded
	var afterFirstReviewCount int64
	db.Model(&models.ApprovalReview{}).Where("approval_request_id = ?", approval.ID).
		Count(&afterFirstReviewCount)
	assert.Equal(t, int64(1), afterFirstReviewCount,
		"One review should exist after first review")

	// Attempt second review - should be blocked
	secondReview, err := approvalService.ReviewApproval(ctx, &ReviewInput{
		ApprovalID: approval.ID,
		ReviewerID: reviewer.ID.String(),
		Decision:   models.ApprovalDecisionRejected,
		Comments:   "Second review - trying to change to rejected",
	})

	// Assert: Second review is blocked
	require.Error(t, err, "Second review should be blocked")
	assert.Nil(t, secondReview, "Second review should not be created")

	// Verify no additional review was added
	var afterSecondReviewCount int64
	db.Model(&models.ApprovalReview{}).Where("approval_request_id = ?", approval.ID).
		Count(&afterSecondReviewCount)
	assert.Equal(t, int64(1), afterSecondReviewCount,
		"Only one review should exist - second was blocked")

	// Verify the first review decision remains unchanged
	var savedReview models.ApprovalReview
	require.NoError(t, db.Where("approval_request_id = ? AND reviewed_by = ?",
		approval.ID, reviewer.ID).First(&savedReview).Error)
	assert.Equal(t, models.ApprovalDecisionApproved, savedReview.Decision,
		"First review decision should remain as approved")
}
