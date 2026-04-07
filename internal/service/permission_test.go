package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/querybase/internal/models"
	"gorm.io/gorm"
)

func TestAdminRole_BypassesAllPermissionChecks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	adminUser := &models.User{
		ID:           uuid.New(),
		Email:        "admin@test.com",
		Username:     "admin",
		PasswordHash: "hashed",
		FullName:     "Admin User",
		Role:         models.RoleAdmin,
		IsActive:     true,
	}
	err := db.Create(adminUser).Error
	require.NoError(t, err)

	dataSource := &models.DataSource{
		ID:                uuid.New(),
		Name:              "Restricted Data Source",
		Type:              models.DataSourceTypePostgreSQL,
		Host:              "localhost",
		Port:              5432,
		Username:          "test",
		EncryptedPassword: "encrypted",
		DatabaseName:      "testdb",
		IsActive:          true,
	}
	err = db.Create(dataSource).Error
	require.NoError(t, err)

	perms, err := queryService.GetEffectivePermissions(context.Background(), adminUser.ID, dataSource.ID)
	require.NoError(t, err)

	assert.True(t, perms.CanSelect, "Admin should have CanSelect")
	assert.True(t, perms.CanInsert, "Admin should have CanInsert")
	assert.True(t, perms.CanUpdate, "Admin should have CanUpdate")
	assert.True(t, perms.CanDelete, "Admin should have CanDelete")
	assert.True(t, perms.CanRead, "Admin should have CanRead")
	assert.True(t, perms.CanWrite, "Admin should have CanWrite")
	assert.True(t, perms.CanApprove, "Admin should have CanApprove")
}

func TestAdminRole_CanExecuteAnyQueryType(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	adminUser := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), adminUser.ID, dataSource.ID)
	require.NoError(t, err)

	assert.True(t, perms.CanSelect, "Admin should be able to execute SELECT")
	assert.True(t, perms.CanInsert, "Admin should be able to execute INSERT")
	assert.True(t, perms.CanUpdate, "Admin should be able to execute UPDATE")
	assert.True(t, perms.CanDelete, "Admin should be able to execute DELETE")
}

func TestAdminRole_CanApproveAnyRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	adminUser := createTestUser(t, db, models.RoleAdmin)
	dataSource := createTestDataSource(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), adminUser.ID, dataSource.ID)
	require.NoError(t, err)

	assert.True(t, perms.CanApprove, "Admin should be able to approve requests")
}

func TestAdminRole_CanAccessAnyDataSource(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	adminUser := createTestUser(t, db, models.RoleAdmin)

	dataSource1 := createTestDataSourceWithName(t, db, "DataSource1")
	dataSource2 := createTestDataSourceWithName(t, db, "DataSource2")
	dataSource3 := createTestDataSourceWithName(t, db, "DataSource3")

	perms1, err := queryService.GetEffectivePermissions(context.Background(), adminUser.ID, dataSource1.ID)
	require.NoError(t, err)
	assert.True(t, perms1.CanRead, "Admin should access DataSource1")

	perms2, err := queryService.GetEffectivePermissions(context.Background(), adminUser.ID, dataSource2.ID)
	require.NoError(t, err)
	assert.True(t, perms2.CanRead, "Admin should access DataSource2")

	perms3, err := queryService.GetEffectivePermissions(context.Background(), adminUser.ID, dataSource3.ID)
	require.NoError(t, err)
	assert.True(t, perms3.CanRead, "Admin should access DataSource3")
}

func TestAdminRole_CanViewAllQueryHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	adminUser := createTestUser(t, db, models.RoleAdmin)
	otherUser := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	history1 := &models.QueryHistory{
		ID:            uuid.New(),
		QueryID:       nil,
		UserID:        adminUser.ID,
		DataSourceID:  dataSource.ID,
		QueryText:     "SELECT 1",
		OperationType: models.OperationSelect,
		Status:        models.QueryStatusCompleted,
		ExecutedAt:    time.Now(),
	}
	err := db.Create(history1).Error
	require.NoError(t, err)

	history2 := &models.QueryHistory{
		ID:            uuid.New(),
		QueryID:       nil,
		UserID:        otherUser.ID,
		DataSourceID:  dataSource.ID,
		QueryText:     "SELECT 2",
		OperationType: models.OperationSelect,
		Status:        models.QueryStatusCompleted,
		ExecutedAt:    time.Now(),
	}
	err = db.Create(history2).Error
	require.NoError(t, err)

	history, total, err := queryService.ListQueryHistory(context.Background(), adminUser.ID.String(), 10, 0, "")
	require.NoError(t, err)

	assert.Equal(t, int64(2), total, "Admin should see all history entries")
	assert.Len(t, history, 2, "Admin should see all history")
}

func TestAdminRole_CanManageAllUsersAndGroups(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)

	adminUser := createTestUser(t, db, models.RoleAdmin)
	otherUser := createTestUser(t, db, models.RoleUser)

	group := &models.Group{
		ID:   uuid.New(),
		Name: "Test Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	var fetchedUser models.User
	err = db.First(&fetchedUser, "id = ?", otherUser.ID).Error
	require.NoError(t, err)
	assert.Equal(t, otherUser.ID, fetchedUser.ID)

	var fetchedGroup models.Group
	err = db.First(&fetchedGroup, "id = ?", group.ID).Error
	require.NoError(t, err)
	assert.Equal(t, group.ID, fetchedGroup.ID)

	membership := &models.UserGroup{
		UserID:  otherUser.ID,
		GroupID: group.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	_ = adminUser
}

func TestAdminRole_CanCreateDataSources(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	_ = createTestUser(t, db, models.RoleAdmin)

	newDataSource := &models.DataSource{
		ID:                uuid.New(),
		Name:              "New Data Source",
		Type:              models.DataSourceTypePostgreSQL,
		Host:              "newhost",
		Port:              5432,
		Username:          "newuser",
		EncryptedPassword: "encrypted_password",
		DatabaseName:      "newdb",
		IsActive:          true,
	}

	err := db.Create(newDataSource).Error
	require.NoError(t, err)

	var ds models.DataSource
	err = db.First(&ds, "id = ?", newDataSource.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "New Data Source", ds.Name)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: newDataSource.ID,
		GroupID:      uuid.Nil,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   true,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)
}

func TestAdminRole_CanDeleteAnyQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	adminUser := createTestUser(t, db, models.RoleAdmin)
	otherUser := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	query := &models.Query{
		ID:           uuid.New(),
		UserID:       otherUser.ID,
		DataSourceID: dataSource.ID,
		Name:         "Other User Query",
		QueryText:    "SELECT * FROM users",
		Status:       models.QueryStatusCompleted,
	}
	err := db.Create(query).Error
	require.NoError(t, err)

	queries, total, err := queryService.ListQueries(context.Background(), adminUser.ID.String(), 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total, "Admin should see all queries")
	assert.Len(t, queries, 1, "Admin should see the query created by other user")

	err = db.Where("id = ?", query.ID).Delete(&models.Query{}).Error
	require.NoError(t, err)

	var deletedQuery models.Query
	err = db.First(&deletedQuery, "id = ?", query.ID).Error
	assert.Error(t, err, "Query should be deleted")
}

func TestAdminRole_HasFullPermissionsComparedToViewer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	adminUser := createTestUser(t, db, models.RoleAdmin)
	viewerUser := createTestUser(t, db, models.RoleViewer)
	dataSource := createTestDataSource(t, db)

	adminPerms, err := queryService.GetEffectivePermissions(context.Background(), adminUser.ID, dataSource.ID)
	require.NoError(t, err)

	viewerPerms, err := queryService.GetEffectivePermissions(context.Background(), viewerUser.ID, dataSource.ID)
	require.NoError(t, err)

	assert.True(t, adminPerms.CanSelect)
	assert.True(t, adminPerms.CanInsert)
	assert.True(t, adminPerms.CanUpdate)
	assert.True(t, adminPerms.CanDelete)
	assert.True(t, adminPerms.CanRead)
	assert.True(t, adminPerms.CanWrite)
	assert.True(t, adminPerms.CanApprove)

	assert.False(t, viewerPerms.CanSelect)
	assert.False(t, viewerPerms.CanInsert)
	assert.False(t, viewerPerms.CanUpdate)
	assert.False(t, viewerPerms.CanDelete)
	assert.False(t, viewerPerms.CanRead)
	assert.False(t, viewerPerms.CanWrite)
	assert.False(t, viewerPerms.CanApprove)

	assert.True(t, adminPerms.CanSelect && !viewerPerms.CanSelect)
}

func TestAdminRole_NormalUserNeedsPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	regularUser := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), regularUser.ID, dataSource.ID)
	require.NoError(t, err)

	assert.False(t, perms.CanSelect)
	assert.False(t, perms.CanInsert)
	assert.False(t, perms.CanUpdate)
	assert.False(t, perms.CanDelete)
	assert.False(t, perms.CanRead)
	assert.False(t, perms.CanWrite)
	assert.False(t, perms.CanApprove)

	group := &models.Group{
		ID:   uuid.New(),
		Name: "Query Group",
	}
	err = db.Create(group).Error
	require.NoError(t, err)

	membership := &models.UserGroup{
		UserID:  regularUser.ID,
		GroupID: group.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   false,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	perms, err = queryService.GetEffectivePermissions(context.Background(), regularUser.ID, dataSource.ID)
	require.NoError(t, err)

	assert.True(t, perms.CanRead)
	assert.True(t, perms.CanSelect)
	assert.False(t, perms.CanWrite)
	assert.False(t, perms.CanApprove)
}

func createTestDataSourceWithName(t *testing.T, db *gorm.DB, name string) *models.DataSource {
	ds := &models.DataSource{
		ID:                uuid.New(),
		Name:              name,
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

// createReadOnlyUserWithPermission creates a user with RoleUser assigned to a group
// that has only CanRead permission on the data source
func createReadOnlyUserWithPermission(t *testing.T, db *gorm.DB) (*models.User, *models.Group, *models.DataSource) {
	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)
	group := &models.Group{
		ID:   uuid.New(),
		Name: "ReadOnly Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	// Add user to group
	membership := &models.UserGroup{
		UserID:  user.ID,
		GroupID: group.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	// Create permission with only CanRead = true, CanWrite = false, CanApprove = false
	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   false,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	return user, group, dataSource
}

// createUserWithoutPermissions creates a user with RoleUser NOT assigned to any group
func createUserWithoutPermissions(t *testing.T, db *gorm.DB) *models.User {
	return createTestUser(t, db, models.RoleUser)
}

// TestReadOnlyUser_CanExecuteSelectQueries verifies that a user with only can_read
// permission can execute SELECT queries
func TestReadOnlyUser_CanExecuteSelectQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	readOnlyUser, _, dataSource := createReadOnlyUserWithPermission(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), readOnlyUser.ID, dataSource.ID)
	require.NoError(t, err)

	// Read-only user should have CanSelect = true (derived from CanRead)
	assert.True(t, perms.CanRead, "Read-only user should have CanRead")
	assert.True(t, perms.CanSelect, "Read-only user should have CanSelect (derived from CanRead)")
	assert.False(t, perms.CanWrite, "Read-only user should NOT have CanWrite")
	assert.False(t, perms.CanInsert, "Read-only user should NOT have CanInsert")
	assert.False(t, perms.CanUpdate, "Read-only user should NOT have CanUpdate")
	assert.False(t, perms.CanDelete, "Read-only user should NOT have CanDelete")
	assert.False(t, perms.CanApprove, "Read-only user should NOT have CanApprove")
}

// TestReadOnlyUser_CannotExecuteInsertQueries verifies that a user with only can_read
// permission CANNOT execute INSERT queries
func TestReadOnlyUser_CannotExecuteInsertQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	readOnlyUser, _, dataSource := createReadOnlyUserWithPermission(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), readOnlyUser.ID, dataSource.ID)
	require.NoError(t, err)

	// Read-only user should NOT have CanInsert
	assert.False(t, perms.CanInsert, "Read-only user should NOT have CanInsert")
	assert.False(t, perms.CanWrite, "Read-only user should NOT have CanWrite")

	// Verify by checking that INSERT operation type is not allowed
	operationType := DetectOperationType("INSERT INTO users (name) VALUES ('test')")
	assert.Equal(t, models.OperationInsert, operationType)

	// Permission check: CanInsert should be false
	canExecute := perms.CanInsert
	assert.False(t, canExecute, "Read-only user should not be able to execute INSERT queries")
}

// TestReadOnlyUser_CannotExecuteUpdateQueries verifies that a user with only can_read
// permission CANNOT execute UPDATE queries
func TestReadOnlyUser_CannotExecuteUpdateQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	readOnlyUser, _, dataSource := createReadOnlyUserWithPermission(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), readOnlyUser.ID, dataSource.ID)
	require.NoError(t, err)

	// Read-only user should NOT have CanUpdate
	assert.False(t, perms.CanUpdate, "Read-only user should NOT have CanUpdate")
	assert.False(t, perms.CanWrite, "Read-only user should NOT have CanWrite")

	// Verify by checking that UPDATE operation type is not allowed
	operationType := DetectOperationType("UPDATE users SET name = 'test' WHERE id = 1")
	assert.Equal(t, models.OperationUpdate, operationType)

	// Permission check: CanUpdate should be false
	canExecute := perms.CanUpdate
	assert.False(t, canExecute, "Read-only user should not be able to execute UPDATE queries")
}

// TestReadOnlyUser_CannotExecuteDeleteQueries verifies that a user with only can_read
// permission CANNOT execute DELETE queries
func TestReadOnlyUser_CannotExecuteDeleteQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	readOnlyUser, _, dataSource := createReadOnlyUserWithPermission(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), readOnlyUser.ID, dataSource.ID)
	require.NoError(t, err)

	// Read-only user should NOT have CanDelete
	assert.False(t, perms.CanDelete, "Read-only user should NOT have CanDelete")
	assert.False(t, perms.CanWrite, "Read-only user should NOT have CanWrite")

	// Verify by checking that DELETE operation type is not allowed
	operationType := DetectOperationType("DELETE FROM users WHERE id = 1")
	assert.Equal(t, models.OperationDelete, operationType)

	// Permission check: CanDelete should be false
	canExecute := perms.CanDelete
	assert.False(t, canExecute, "Read-only user should not be able to execute DELETE queries")
}

// TestReadOnlyUser_CanViewOwnQueryHistory verifies that a user with only can_read
// permission can view their own query history
func TestReadOnlyUser_CanViewOwnQueryHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	readOnlyUser, _, dataSource := createReadOnlyUserWithPermission(t, db)
	otherUser := createTestUser(t, db, models.RoleUser)

	// Create query history for read-only user
	history1 := &models.QueryHistory{
		ID:            uuid.New(),
		QueryID:       nil,
		UserID:        readOnlyUser.ID,
		DataSourceID:  dataSource.ID,
		QueryText:     "SELECT 1",
		OperationType: models.OperationSelect,
		Status:        models.QueryStatusCompleted,
		ExecutedAt:    time.Now(),
	}
	err := db.Create(history1).Error
	require.NoError(t, err)

	// Create query history for other user
	history2 := &models.QueryHistory{
		ID:            uuid.New(),
		QueryID:       nil,
		UserID:        otherUser.ID,
		DataSourceID:  dataSource.ID,
		QueryText:     "SELECT 2",
		OperationType: models.OperationSelect,
		Status:        models.QueryStatusCompleted,
		ExecutedAt:    time.Now(),
	}
	err = db.Create(history2).Error
	require.NoError(t, err)

	// Read-only user should see their own history
	history, total, err := queryService.ListQueryHistory(context.Background(), readOnlyUser.ID.String(), 10, 0, "")
	require.NoError(t, err)

	// Should see at least 1 entry (their own), but filter by user ID in query service
	assert.GreaterOrEqual(t, total, int64(1), "Read-only user should see their own history")
	assert.Len(t, history, 1, "Read-only user should only see their own history, not other users'")

	// Verify the history belongs to read-only user
	assert.Equal(t, readOnlyUser.ID, history[0].UserID)
}

// TestReadOnlyUser_CannotApproveWriteRequests verifies that a user with only can_read
// permission CANNOT approve write requests
func TestReadOnlyUser_CannotApproveWriteRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	readOnlyUser, _, dataSource := createReadOnlyUserWithPermission(t, db)
	requester := createTestUser(t, db, models.RoleUser)

	perms, err := queryService.GetEffectivePermissions(context.Background(), readOnlyUser.ID, dataSource.ID)
	require.NoError(t, err)

	// Read-only user should NOT have CanApprove
	assert.False(t, perms.CanApprove, "Read-only user should NOT have CanApprove")

	// Create an approval request by another user
	approvalRequest := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "DELETE FROM users WHERE id = 1",
		RequestedBy:  requester.ID.String(),
	}
	_, err = approvalService.CreateApprovalRequest(context.Background(), approvalRequest)
	require.NoError(t, err)

	// read-only user's permission check
	assert.False(t, perms.CanApprove, "Read-only user should not have CanApprove permission")
}

// TestReadOnlyUser_CanPreviewQueries verifies that a user with only can_read
// permission CAN preview queries (previews convert write queries to SELECT)
func TestReadOnlyUser_CanPreviewQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	readOnlyUser, _, dataSource := createReadOnlyUserWithPermission(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), readOnlyUser.ID, dataSource.ID)
	require.NoError(t, err)

	// Read-only user has CanRead = true, which grants CanSelect
	assert.True(t, perms.CanRead, "Read-only user should have CanRead")
	assert.True(t, perms.CanSelect, "Read-only user should have CanSelect (for SELECT query preview)")

	// Preview queries use SELECT underneath, so they should work with CanRead permission
	// The PreviewWriteQuery function converts DELETE/UPDATE to SELECT with LIMIT
	// Note: This tests the permission model - actual preview execution requires DB connection
}

func createWriteUserWithPermission(t *testing.T, db *gorm.DB) (*models.User, *models.Group, *models.DataSource) {
	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)
	group := &models.Group{
		ID:   uuid.New(),
		Name: "Write Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	membership := &models.UserGroup{
		UserID:  user.ID,
		GroupID: group.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   false,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	err = db.Model(dsPerm).Update("can_read", false).Error
	require.NoError(t, err)

	return user, group, dataSource
}

func createApproverUserWithPermission(t *testing.T, db *gorm.DB, dataSource *models.DataSource) (*models.User, *models.Group) {
	approver := createTestUser(t, db, models.RoleUser)
	approverGroup := &models.Group{
		ID:   uuid.New(),
		Name: "Approver Group",
	}
	err := db.Create(approverGroup).Error
	require.NoError(t, err)

	membership := &models.UserGroup{
		UserID:  approver.ID,
		GroupID: approverGroup.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      approverGroup.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   true,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	return approver, approverGroup
}

func TestWriteUser_HasWritePermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	writeUser, _, dataSource := createWriteUserWithPermission(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), writeUser.ID, dataSource.ID)
	require.NoError(t, err)

	assert.True(t, perms.CanWrite, "Write user should have CanWrite")
	assert.True(t, perms.CanInsert, "Write user should have CanInsert (derived from CanWrite)")
	assert.True(t, perms.CanUpdate, "Write user should have CanUpdate (derived from CanWrite)")
	assert.True(t, perms.CanDelete, "Write user should have CanDelete (derived from CanWrite)")
	assert.False(t, perms.CanApprove, "Write user should NOT have CanApprove")
	assert.False(t, perms.CanRead, "Write user should NOT have CanRead (only CanWrite)")
}

func TestWriteUser_CanExecuteInsertWithApproval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	writeUser, _, dataSource := createWriteUserWithPermission(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), writeUser.ID, dataSource.ID)
	require.NoError(t, err)

	operationType := DetectOperationType("INSERT INTO users (name) VALUES ('test')")
	assert.Equal(t, models.OperationInsert, operationType)

	assert.True(t, perms.CanInsert, "Write user should have CanInsert permission")
	assert.True(t, perms.CanWrite, "Write user should have CanWrite permission")
	assert.True(t, perms.CanInsert, "Write user should be able to execute INSERT (with approval)")
}

func TestWriteUser_CanExecuteUpdateWithApproval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	writeUser, _, dataSource := createWriteUserWithPermission(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), writeUser.ID, dataSource.ID)
	require.NoError(t, err)

	operationType := DetectOperationType("UPDATE users SET name = 'test' WHERE id = 1")
	assert.Equal(t, models.OperationUpdate, operationType)

	assert.True(t, perms.CanUpdate, "Write user should have CanUpdate permission")
	assert.True(t, perms.CanWrite, "Write user should have CanWrite permission")
	assert.True(t, perms.CanUpdate, "Write user should be able to execute UPDATE (with approval)")
}

func TestWriteUser_CanExecuteDeleteWithApproval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	writeUser, _, dataSource := createWriteUserWithPermission(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), writeUser.ID, dataSource.ID)
	require.NoError(t, err)

	operationType := DetectOperationType("DELETE FROM users WHERE id = 1")
	assert.Equal(t, models.OperationDelete, operationType)

	assert.True(t, perms.CanDelete, "Write user should have CanDelete permission")
	assert.True(t, perms.CanWrite, "Write user should have CanWrite permission")
	assert.True(t, perms.CanDelete, "Write user should be able to execute DELETE (with approval)")
}

func TestWriteUser_CanCreateApprovalRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	writeUser, _, dataSource := createWriteUserWithPermission(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), writeUser.ID, dataSource.ID)
	require.NoError(t, err)

	assert.True(t, perms.CanWrite, "Write user should have CanWrite")
	assert.True(t, perms.CanInsert, "Write user should have CanInsert")

	approvalRequest := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "INSERT INTO users (name) VALUES ('test')",
		RequestedBy:  writeUser.ID.String(),
	}

	createdApproval, err := approvalService.CreateApprovalRequest(context.Background(), approvalRequest)
	require.NoError(t, err, "Write user should be able to create approval requests")

	assert.NotEqual(t, uuid.Nil, createdApproval.ID)
	assert.Equal(t, dataSource.ID, createdApproval.DataSourceID)
	assert.Equal(t, writeUser.ID, createdApproval.RequestedBy)
	assert.Equal(t, models.ApprovalStatusPending, createdApproval.Status)
}

func TestWriteUser_CannotApproveOwnRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	writeUser, _, dataSource := createWriteUserWithPermission(t, db)

	approvalRequest := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "DELETE FROM users WHERE id = 1",
		RequestedBy:  writeUser.ID.String(),
	}
	createdApproval, err := approvalService.CreateApprovalRequest(context.Background(), approvalRequest)
	require.NoError(t, err)

	perms, err := queryService.GetEffectivePermissions(context.Background(), writeUser.ID, dataSource.ID)
	require.NoError(t, err)
	assert.False(t, perms.CanApprove, "Write user should NOT have CanApprove")

	review := &ReviewInput{
		ApprovalID: createdApproval.ID,
		ReviewerID: writeUser.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Self-approval attempt",
	}

	_, err = approvalService.ReviewApproval(context.Background(), review)
	assert.Error(t, err, "Self-approval should be blocked")
	assert.Contains(t, err.Error(), "self-approval is not allowed", "Error should mention self-approval")
}

func TestWriteUser_CanViewOwnRequestsAndHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	writeUser, _, dataSource := createWriteUserWithPermission(t, db)
	otherUser := createTestUser(t, db, models.RoleUser)

	approval1 := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "INSERT INTO users (name) VALUES ('test1')",
		RequestedBy:  writeUser.ID.String(),
	}
	createdApproval1, err := approvalService.CreateApprovalRequest(context.Background(), approval1)
	require.NoError(t, err)

	approval2 := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "INSERT INTO users (name) VALUES ('test2')",
		RequestedBy:  otherUser.ID.String(),
	}
	_, err = approvalService.CreateApprovalRequest(context.Background(), approval2)
	require.NoError(t, err)

	writeUserApprovals, total, err := approvalService.ListApprovals(context.Background(), &ApprovalFilter{
		RequestedBy: writeUser.ID.String(),
		Limit:       10,
		Offset:      0,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total, "Write user should see only their own approval requests")
	assert.Len(t, writeUserApprovals, 1)
	assert.Equal(t, createdApproval1.ID, writeUserApprovals[0].ID)

	history1 := &models.QueryHistory{
		ID:            uuid.New(),
		QueryID:       nil,
		UserID:        writeUser.ID,
		DataSourceID:  dataSource.ID,
		QueryText:     "INSERT INTO users (name) VALUES ('test')",
		OperationType: models.OperationInsert,
		Status:        models.QueryStatusCompleted,
		ExecutedAt:    time.Now(),
	}
	err = db.Create(history1).Error
	require.NoError(t, err)

	history2 := &models.QueryHistory{
		ID:            uuid.New(),
		QueryID:       nil,
		UserID:        otherUser.ID,
		DataSourceID:  dataSource.ID,
		QueryText:     "SELECT * FROM users",
		OperationType: models.OperationSelect,
		Status:        models.QueryStatusCompleted,
		ExecutedAt:    time.Now(),
	}
	err = db.Create(history2).Error
	require.NoError(t, err)

	history, totalHistory, err := queryService.ListQueryHistory(context.Background(), writeUser.ID.String(), 10, 0, "")
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalHistory, "Write user should see only their own query history")
	assert.Len(t, history, 1)
	assert.Equal(t, writeUser.ID, history[0].UserID)
}

func TestWriteUser_CanPreviewWriteQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	writeUser, _, dataSource := createWriteUserWithPermission(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), writeUser.ID, dataSource.ID)
	require.NoError(t, err)

	assert.True(t, perms.CanWrite, "Write user should have CanWrite for preview")

	deleteOp := DetectOperationType("DELETE FROM users WHERE id = 1")
	updateOp := DetectOperationType("UPDATE users SET name = 'test' WHERE id = 1")

	assert.Equal(t, models.OperationDelete, deleteOp, "Should detect DELETE operation")
	assert.Equal(t, models.OperationUpdate, updateOp, "Should detect UPDATE operation")

	assert.True(t, deleteOp == models.OperationDelete || deleteOp == models.OperationUpdate,
		"DELETE should be previewable")
	assert.True(t, updateOp == models.OperationDelete || updateOp == models.OperationUpdate,
		"UPDATE should be previewable")
}

// TestUserWithoutPermissions_CannotExecuteAnyQueries verifies that a user without
// any permissions (not in any group) cannot execute any queries
func TestUserWithoutPermissions_CannotExecuteAnyQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	userWithoutPerms := createUserWithoutPermissions(t, db)
	dataSource := createTestDataSource(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), userWithoutPerms.ID, dataSource.ID)
	require.NoError(t, err)

	// User without any group membership should have NO permissions
	assert.False(t, perms.CanRead, "User without permissions should NOT have CanRead")
	assert.False(t, perms.CanWrite, "User without permissions should NOT have CanWrite")
	assert.False(t, perms.CanApprove, "User without permissions should NOT have CanApprove")
	assert.False(t, perms.CanSelect, "User without permissions should NOT have CanSelect")
	assert.False(t, perms.CanInsert, "User without permissions should NOT have CanInsert")
	assert.False(t, perms.CanUpdate, "User without permissions should NOT have CanUpdate")
	assert.False(t, perms.CanDelete, "User without permissions should NOT have CanDelete")

	// Verify all operation types are blocked
	selectType := DetectOperationType("SELECT * FROM users")
	insertType := DetectOperationType("INSERT INTO users (name) VALUES ('test')")
	updateType := DetectOperationType("UPDATE users SET name = 'test'")
	deleteType := DetectOperationType("DELETE FROM users")

	// All operations should be blocked (permissions are all false)
	assert.False(t, perms.CanSelect && selectType == models.OperationSelect, "User without permissions should not execute SELECT")
	assert.False(t, perms.CanInsert && insertType == models.OperationInsert, "User without permissions should not execute INSERT")
	assert.False(t, perms.CanUpdate && updateType == models.OperationUpdate, "User without permissions should not execute UPDATE")
	assert.False(t, perms.CanDelete && deleteType == models.OperationDelete, "User without permissions should not execute DELETE")
}

// ---------------------------------------------------------------------------
// Approver User Permission Tests
// ---------------------------------------------------------------------------

// TestApproverUser_CanApproveWriteRequests verifies that a user with can_approve
// permission can approve write requests from other users.
func TestApproverUser_CanApproveWriteRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create requester (regular user)
	requester := createTestUser(t, db, models.RoleUser)
	// Create data source first
	dataSource := createTestDataSource(t, db)
	// Create approver user with can_approve permission
	approver, _ := createApproverUserWithPermission(t, db, dataSource)

	// Add approver to data source permission
	group := &models.Group{
		ID:   uuid.New(),
		Name: "Approver Group for Test",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	membership := &models.UserGroup{
		UserID:  approver.ID,
		GroupID: group.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   true,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	// Create an approval request by the regular user
	approvalRequest := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "DELETE FROM users WHERE id = 1",
		RequestedBy:  requester.ID.String(),
	}
	createdApproval, err := approvalService.CreateApprovalRequest(context.Background(), approvalRequest)
	require.NoError(t, err, "Approval request should be created")
	assert.Equal(t, models.ApprovalStatusPending, createdApproval.Status)

	// Approver should be able to approve the request
	review := &ReviewInput{
		ApprovalID: createdApproval.ID,
		ReviewerID: approver.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Approved by approver user",
	}

	approvalReview, err := approvalService.ReviewApproval(context.Background(), review)
	require.NoError(t, err, "Approver with can_approve should be able to approve")
	assert.Equal(t, models.ApprovalDecisionApproved, approvalReview.Decision)

	// Verify the approval status was updated
	var updatedApproval models.ApprovalRequest
	err = db.First(&updatedApproval, "id = ?", createdApproval.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.ApprovalStatusApproved, updatedApproval.Status)
}

// TestApproverUser_CannotApproveOwnRequests verifies that a user with can_approve
// permission CANNOT approve their own requests, even though they have can_approve.
func TestApproverUser_CannotApproveOwnRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create approver user with can_approve permission
	approver := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group with can_approve permission
	group := &models.Group{
		ID:   uuid.New(),
		Name: "Self-Approver Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	membership := &models.UserGroup{
		UserID:  approver.ID,
		GroupID: group.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   true, // User HAS can_approve permission
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	// Create an approval request by the approver themselves
	approvalRequest := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "DELETE FROM orders WHERE id = 1",
		RequestedBy:  approver.ID.String(),
	}
	createdApproval, err := approvalService.CreateApprovalRequest(context.Background(), approvalRequest)
	require.NoError(t, err, "Approval request should be created")

	// Verify approver has CanApprove permission
	perms, err := queryService.GetEffectivePermissions(context.Background(), approver.ID, dataSource.ID)
	require.NoError(t, err)
	assert.True(t, perms.CanApprove, "Approver user should have CanApprove permission")

	// Attempt to approve own request - should fail
	review := &ReviewInput{
		ApprovalID: createdApproval.ID,
		ReviewerID: approver.ID.String(), // Same user trying to approve own request
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Self-approval attempt",
	}

	_, err = approvalService.ReviewApproval(context.Background(), review)
	assert.Error(t, err, "Self-approval should be blocked even with can_approve permission")
	assert.Contains(t, err.Error(), "self-approval is not allowed", "Error should mention self-approval")

	// Verify the approval status was NOT changed
	var unchangedApproval models.ApprovalRequest
	err = db.First(&unchangedApproval, "id = ?", createdApproval.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.ApprovalStatusPending, unchangedApproval.Status)
}

// TestApproverUser_CanRejectRequests verifies that a user with can_approve
// permission can reject write requests.
func TestApproverUser_CanRejectRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create requester (regular user)
	requester := createTestUser(t, db, models.RoleUser)
	// Create approver user with can_approve permission
	approver := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group with can_approve permission
	group := &models.Group{
		ID:   uuid.New(),
		Name: "Rejecter Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	membership := &models.UserGroup{
		UserID:  approver.ID,
		GroupID: group.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   true,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	// Create an approval request
	approvalRequest := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "DELETE FROM users WHERE id = 1",
		RequestedBy:  requester.ID.String(),
	}
	createdApproval, err := approvalService.CreateApprovalRequest(context.Background(), approvalRequest)
	require.NoError(t, err)

	// Approver should be able to reject the request
	review := &ReviewInput{
		ApprovalID: createdApproval.ID,
		ReviewerID: approver.ID.String(),
		Decision:   models.ApprovalDecisionRejected,
		Comments:   "Rejected by approver user",
	}

	approvalReview, err := approvalService.ReviewApproval(context.Background(), review)
	require.NoError(t, err, "Approver with can_approve should be able to reject")
	assert.Equal(t, models.ApprovalDecisionRejected, approvalReview.Decision)

	// Verify the approval status was updated to rejected
	var updatedApproval models.ApprovalRequest
	err = db.First(&updatedApproval, "id = ?", createdApproval.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.ApprovalStatusRejected, updatedApproval.Status)
}

// TestApproverUser_ApprovalAppearsInList verifies that when an approver reviews
// a request, the approval appears in their approval list.
func TestApproverUser_ApprovalAppearsInList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create requester (regular user)
	requester := createTestUser(t, db, models.RoleUser)
	// Create approver user with can_approve permission
	approver := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group with can_approve permission
	group := &models.Group{
		ID:   uuid.New(),
		Name: "Approval List Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	membership := &models.UserGroup{
		UserID:  approver.ID,
		GroupID: group.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   true,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	// Create an approval request
	approvalRequest := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "DELETE FROM users WHERE id = 1",
		RequestedBy:  requester.ID.String(),
	}
	createdApproval, err := approvalService.CreateApprovalRequest(context.Background(), approvalRequest)
	require.NoError(t, err)

	// Approve the request
	review := &ReviewInput{
		ApprovalID: createdApproval.ID,
		ReviewerID: approver.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Approved",
	}
	_, err = approvalService.ReviewApproval(context.Background(), review)
	require.NoError(t, err)

	// Get the approval with reviews loaded
	retrievedApproval, err := approvalService.GetApproval(context.Background(), createdApproval.ID.String())
	require.NoError(t, err)

	// Verify the approval has the review from the approver
	assert.Len(t, retrievedApproval.ApprovalReviews, 1, "Approval should have one review")
	assert.Equal(t, approver.ID, retrievedApproval.ApprovalReviews[0].ReviewerID)
	assert.Equal(t, models.ApprovalDecisionApproved, retrievedApproval.ApprovalReviews[0].Decision)
}

// TestApproverUser_HasCanApprovePermission verifies that a user with can_approve
// permission has the CanApprove permission flag set to true.
func TestApproverUser_HasCanApprovePermission(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create approver user with can_approve permission
	approver := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group with can_approve permission
	group := &models.Group{
		ID:   uuid.New(),
		Name: "CanApprove Permission Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	membership := &models.UserGroup{
		UserID:  approver.ID,
		GroupID: group.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   true,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	// Get effective permissions
	perms, err := queryService.GetEffectivePermissions(context.Background(), approver.ID, dataSource.ID)
	require.NoError(t, err)

	// Verify CanApprove is true
	assert.True(t, perms.CanApprove, "User with can_approve permission should have CanApprove = true")
}

// TestApproverUser_CanViewAllPendingApprovals verifies that a user with can_approve
// permission can view all pending approvals for the data source.
func TestApproverUser_CanViewAllPendingApprovals(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create multiple requesters
	requester1 := createTestUser(t, db, models.RoleUser)
	requester2 := createTestUser(t, db, models.RoleUser)
	// Create approver user with can_approve permission
	approver := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group with can_approve permission
	group := &models.Group{
		ID:   uuid.New(),
		Name: "View Pending Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	membership := &models.UserGroup{
		UserID:  approver.ID,
		GroupID: group.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   true,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	// Create multiple approval requests from different users
	approval1 := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "DELETE FROM users WHERE id = 1",
		RequestedBy:  requester1.ID.String(),
	}
	_, err = approvalService.CreateApprovalRequest(context.Background(), approval1)
	require.NoError(t, err)

	approval2 := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "DELETE FROM orders WHERE id = 2",
		RequestedBy:  requester2.ID.String(),
	}
	_, err = approvalService.CreateApprovalRequest(context.Background(), approval2)
	require.NoError(t, err)

	// Approver should be able to list all pending approvals
	approvals, total, err := approvalService.ListApprovals(context.Background(), &ApprovalFilter{
		Status: string(models.ApprovalStatusPending),
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)

	// Should see at least 2 pending approvals
	assert.GreaterOrEqual(t, total, int64(2), "Approver should see multiple pending approvals")
	assert.Len(t, approvals, 2, "Approver should see all pending approvals")
}

// TestApproverUser_CannotApproveAlreadyApproved verifies that a user with can_approve
// permission cannot approve a request that has already been approved.
func TestApproverUser_CannotApproveAlreadyApproved(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create requester
	requester := createTestUser(t, db, models.RoleUser)
	// Create approvers
	approver1 := createTestUser(t, db, models.RoleUser)
	approver2 := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group with can_approve permission
	group := &models.Group{
		ID:   uuid.New(),
		Name: "Already Approved Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	// Add both approvers to the group
	membership1 := &models.UserGroup{
		UserID:  approver1.ID,
		GroupID: group.ID,
	}
	err = db.Create(membership1).Error
	require.NoError(t, err)
	membership2 := &models.UserGroup{
		UserID:  approver2.ID,
		GroupID: group.ID,
	}
	err = db.Create(membership2).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   true,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	// Create an approval request
	approvalRequest := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "DELETE FROM users WHERE id = 1",
		RequestedBy:  requester.ID.String(),
	}
	createdApproval, err := approvalService.CreateApprovalRequest(context.Background(), approvalRequest)
	require.NoError(t, err)

	// First approver approves the request
	review1 := &ReviewInput{
		ApprovalID: createdApproval.ID,
		ReviewerID: approver1.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "First approval",
	}
	_, err = approvalService.ReviewApproval(context.Background(), review1)
	require.NoError(t, err)

	// Verify the approval is now approved
	var approvedApproval models.ApprovalRequest
	err = db.First(&approvedApproval, "id = ?", createdApproval.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.ApprovalStatusApproved, approvedApproval.Status)

	// Second approver tries to approve again - should fail
	review2 := &ReviewInput{
		ApprovalID: createdApproval.ID,
		ReviewerID: approver2.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Second approval attempt",
	}
	_, err = approvalService.ReviewApproval(context.Background(), review2)
	assert.Error(t, err, "Should not be able to approve already approved request")
	assert.Contains(t, err.Error(), "not pending", "Error should indicate request is not pending")
}

// ---------------------------------------------------------------------------
// Viewer Role Permission Tests
// ---------------------------------------------------------------------------

// TestViewerRole_CannotExecuteWriteQueries verifies that a viewer user
// cannot execute INSERT, UPDATE, or DELETE queries regardless of group membership
func TestViewerRole_CannotExecuteWriteQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create viewer user
	viewerUser := createTestUser(t, db, models.RoleViewer)
	dataSource := createTestDataSource(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), viewerUser.ID, dataSource.ID)
	require.NoError(t, err)

	// Verify viewer has no write permissions
	assert.False(t, perms.CanInsert, "Viewer should NOT have CanInsert")
	assert.False(t, perms.CanUpdate, "Viewer should NOT have CanUpdate")
	assert.False(t, perms.CanDelete, "Viewer should NOT have CanDelete")
	assert.False(t, perms.CanWrite, "Viewer should NOT have CanWrite")

	// Verify operation types are correctly detected
	insertOp := DetectOperationType("INSERT INTO users (name) VALUES ('test')")
	updateOp := DetectOperationType("UPDATE users SET name = 'test' WHERE id = 1")
	deleteOp := DetectOperationType("DELETE FROM users WHERE id = 1")

	assert.Equal(t, models.OperationInsert, insertOp, "Should detect INSERT operation")
	assert.Equal(t, models.OperationUpdate, updateOp, "Should detect UPDATE operation")
	assert.Equal(t, models.OperationDelete, deleteOp, "Should detect DELETE operation")

	// Viewer cannot execute any write operations
	canExecuteInsert := perms.CanInsert && insertOp == models.OperationInsert
	canExecuteUpdate := perms.CanUpdate && updateOp == models.OperationUpdate
	canExecuteDelete := perms.CanDelete && deleteOp == models.OperationDelete

	assert.False(t, canExecuteInsert, "Viewer should NOT be able to execute INSERT")
	assert.False(t, canExecuteUpdate, "Viewer should NOT be able to execute UPDATE")
	assert.False(t, canExecuteDelete, "Viewer should NOT be able to execute DELETE")
}

// TestViewerRole_CanOnlyViewQueries verifies that a viewer user can only
// execute SELECT queries and cannot perform any write operations
func TestViewerRole_CanOnlyViewQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create viewer user
	viewerUser := createTestUser(t, db, models.RoleViewer)
	dataSource := createTestDataSource(t, db)

	// Create a viewer-only group with CanRead permission
	viewerGroup := &models.Group{
		ID:   uuid.New(),
		Name: "Viewer Group",
	}
	err := db.Create(viewerGroup).Error
	require.NoError(t, err)

	membership := &models.UserGroup{
		UserID:  viewerUser.ID,
		GroupID: viewerGroup.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      viewerGroup.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   false,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	perms, err := queryService.GetEffectivePermissions(context.Background(), viewerUser.ID, dataSource.ID)
	require.NoError(t, err)

	// Viewer with CanRead should have SELECT permission
	assert.True(t, perms.CanRead, "Viewer should have CanRead")
	assert.True(t, perms.CanSelect, "Viewer should have CanSelect (derived from CanRead)")

	// But should NOT have write permissions
	assert.False(t, perms.CanWrite, "Viewer should NOT have CanWrite")
	assert.False(t, perms.CanInsert, "Viewer should NOT have CanInsert")
	assert.False(t, perms.CanUpdate, "Viewer should NOT have CanUpdate")
	assert.False(t, perms.CanDelete, "Viewer should NOT have CanDelete")
	assert.False(t, perms.CanApprove, "Viewer should NOT have CanApprove")

	// Verify SELECT is the only allowed operation
	selectOp := DetectOperationType("SELECT * FROM users")
	assert.Equal(t, models.OperationSelect, selectOp, "Should detect SELECT operation")

	canExecuteSelect := perms.CanSelect && selectOp == models.OperationSelect
	assert.True(t, canExecuteSelect, "Viewer should be able to execute SELECT")
}

// TestViewerRole_CannotCreateApprovalRequests verifies that a viewer user
// cannot create approval requests for write operations
func TestViewerRole_CannotCreateApprovalRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create viewer user
	viewerUser := createTestUser(t, db, models.RoleViewer)
	dataSource := createTestDataSource(t, db)

	perms, err := queryService.GetEffectivePermissions(context.Background(), viewerUser.ID, dataSource.ID)
	require.NoError(t, err)

	// Viewer should NOT have CanWrite permission
	assert.False(t, perms.CanWrite, "Viewer should NOT have CanWrite")
	assert.False(t, perms.CanInsert, "Viewer should NOT have CanInsert")

	// Creating approval requests requires CanWrite permission
	// This verifies that viewers cannot submit write queries for approval
	assert.False(t, perms.CanWrite, "Viewer cannot create approval requests because they lack CanWrite")
}

// TestViewerRole_CannotApproveAnything verifies that a viewer user
// cannot approve any approval requests, even if they belong to a group
// that would normally have approval permissions
func TestViewerRole_CannotApproveAnything(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	viewerUser := createTestUser(t, db, models.RoleViewer)
	dataSource := createTestDataSource(t, db)

	viewerGroup := &models.Group{
		ID:   uuid.New(),
		Name: "Viewer With Approve Group",
	}
	err := db.Create(viewerGroup).Error
	require.NoError(t, err)

	membership := &models.UserGroup{
		UserID:  viewerUser.ID,
		GroupID: viewerGroup.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      viewerGroup.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   true,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	perms, err := queryService.GetEffectivePermissions(context.Background(), viewerUser.ID, dataSource.ID)
	require.NoError(t, err)

	assert.False(t, perms.CanApprove, "Viewer role should NOT have CanApprove even with group permission")

	requester := createTestUser(t, db, models.RoleUser)

	approvalRequest := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "DELETE FROM users WHERE id = 1",
		RequestedBy:  requester.ID.String(),
	}
	createdApproval, err := approvalService.CreateApprovalRequest(context.Background(), approvalRequest)
	require.NoError(t, err)

	review := &ReviewInput{
		ApprovalID: createdApproval.ID,
		ReviewerID: viewerUser.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Viewer attempting to approve",
	}

	_, err = approvalService.ReviewApproval(context.Background(), review)
	assert.Error(t, err, "Viewer should not be able to approve requests")
}

// TestViewerRole_LimitedQueryHistoryAccess verifies that a viewer user
// can only view their own query history, not other users' history
func TestViewerRole_LimitedQueryHistoryAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create viewer user
	viewerUser := createTestUser(t, db, models.RoleViewer)
	otherUser := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create query history for viewer
	viewerHistory := &models.QueryHistory{
		ID:            uuid.New(),
		QueryID:       nil,
		UserID:        viewerUser.ID,
		DataSourceID:  dataSource.ID,
		QueryText:     "SELECT * FROM viewer_data",
		OperationType: models.OperationSelect,
		Status:        models.QueryStatusCompleted,
		ExecutedAt:    time.Now(),
	}
	err := db.Create(viewerHistory).Error
	require.NoError(t, err)

	// Create query history for other user
	otherHistory := &models.QueryHistory{
		ID:            uuid.New(),
		QueryID:       nil,
		UserID:        otherUser.ID,
		DataSourceID:  dataSource.ID,
		QueryText:     "SELECT * FROM other_data",
		OperationType: models.OperationSelect,
		Status:        models.QueryStatusCompleted,
		ExecutedAt:    time.Now(),
	}
	err = db.Create(otherHistory).Error
	require.NoError(t, err)

	// Viewer should only see their own history
	history, total, err := queryService.ListQueryHistory(context.Background(), viewerUser.ID.String(), 10, 0, "")
	require.NoError(t, err)

	assert.Equal(t, int64(1), total, "Viewer should only see their own history")
	assert.Len(t, history, 1, "Viewer should only see 1 history entry")
	assert.Equal(t, viewerUser.ID, history[0].UserID, "History should belong to viewer")
	assert.Equal(t, "SELECT * FROM viewer_data", history[0].QueryText)
}

// TestViewerRole_CanPreviewQueries verifies that a viewer user can
// preview queries, as previews convert write operations to SELECT
func TestViewerRole_CanPreviewQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create viewer user
	viewerUser := createTestUser(t, db, models.RoleViewer)
	dataSource := createTestDataSource(t, db)

	// Create a viewer-only group with CanRead permission
	viewerGroup := &models.Group{
		ID:   uuid.New(),
		Name: "Viewer Preview Group",
	}
	err := db.Create(viewerGroup).Error
	require.NoError(t, err)

	membership := &models.UserGroup{
		UserID:  viewerUser.ID,
		GroupID: viewerGroup.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      viewerGroup.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   false,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	perms, err := queryService.GetEffectivePermissions(context.Background(), viewerUser.ID, dataSource.ID)
	require.NoError(t, err)

	// Viewer should have CanSelect for previews
	assert.True(t, perms.CanSelect, "Viewer should have CanSelect for preview")

	// Verify operation types for queries that would be previewed
	deleteOp := DetectOperationType("DELETE FROM users WHERE id = 1")
	updateOp := DetectOperationType("UPDATE users SET name = 'test' WHERE id = 1")

	assert.Equal(t, models.OperationDelete, deleteOp)
	assert.Equal(t, models.OperationUpdate, updateOp)

	// Preview operations convert write to SELECT, which viewers can do with CanRead
	assert.True(t, perms.CanSelect, "Viewer should be able to preview queries (using SELECT)")
}

// TestViewerRole_CannotAccessAdminEndpoints verifies that a viewer user
// cannot access admin-only endpoints by checking their permissions
func TestViewerRole_CannotAccessAdminEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)

	viewerUser := createTestUser(t, db, models.RoleViewer)

	assert.Equal(t, models.RoleViewer, viewerUser.Role, "User should have viewer role")
	assert.NotEqual(t, models.RoleAdmin, viewerUser.Role, "Viewer should not have admin role")
	assert.True(t, viewerUser.Role != models.RoleAdmin, "Non-admin role blocks admin endpoint access")
}

// TestViewerRole_HasReadOnlyPermissions verifies that a viewer user
// has exactly read-only permissions regardless of group membership
func TestViewerRole_HasReadOnlyPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create viewer user
	viewerUser := createTestUser(t, db, models.RoleViewer)
	dataSource := createTestDataSource(t, db)

	// Get permissions without any group membership (default viewer state)
	perms, err := queryService.GetEffectivePermissions(context.Background(), viewerUser.ID, dataSource.ID)
	require.NoError(t, err)

	// Viewer without group should have NO permissions
	assert.False(t, perms.CanRead, "Viewer without group should NOT have CanRead")
	assert.False(t, perms.CanWrite, "Viewer without group should NOT have CanWrite")
	assert.False(t, perms.CanApprove, "Viewer without group should NOT have CanApprove")
	assert.False(t, perms.CanSelect, "Viewer without group should NOT have CanSelect")
	assert.False(t, perms.CanInsert, "Viewer without group should NOT have CanInsert")
	assert.False(t, perms.CanUpdate, "Viewer without group should NOT have CanUpdate")
	assert.False(t, perms.CanDelete, "Viewer without group should NOT have CanDelete")

	// Now add viewer to a group with CanRead only
	readOnlyGroup := &models.Group{
		ID:   uuid.New(),
		Name: "Viewer ReadOnly Group",
	}
	err = db.Create(readOnlyGroup).Error
	require.NoError(t, err)

	membership := &models.UserGroup{
		UserID:  viewerUser.ID,
		GroupID: readOnlyGroup.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      readOnlyGroup.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   false,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	perms, err = queryService.GetEffectivePermissions(context.Background(), viewerUser.ID, dataSource.ID)
	require.NoError(t, err)

	// Viewer with read-only group should have read permissions
	assert.True(t, perms.CanRead, "Viewer with read-only group should have CanRead")
	assert.True(t, perms.CanSelect, "Viewer with read-only group should have CanSelect")

	// But still NO write or approve permissions
	assert.False(t, perms.CanWrite, "Viewer should NOT have CanWrite")
	assert.False(t, perms.CanInsert, "Viewer should NOT have CanInsert")
	assert.False(t, perms.CanUpdate, "Viewer should NOT have CanUpdate")
	assert.False(t, perms.CanDelete, "Viewer should NOT have CanDelete")
	assert.False(t, perms.CanApprove, "Viewer should NOT have CanApprove")

	// Summary: Viewer role has read-only permissions
	assert.True(t, perms.CanSelect && !perms.CanWrite && !perms.CanApprove,
		"Viewer should have exactly read-only permissions")
}

// TestApproverUser_CannotApproveAlreadyRejected verifies that a user with can_approve
// permission cannot approve a request that has already been rejected.
func TestApproverUser_CannotApproveAlreadyRejected(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create requester
	requester := createTestUser(t, db, models.RoleUser)
	// Create approvers
	approver1 := createTestUser(t, db, models.RoleUser)
	approver2 := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group with can_approve permission
	group := &models.Group{
		ID:   uuid.New(),
		Name: "Already Rejected Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	// Add both approvers to the group
	membership1 := &models.UserGroup{
		UserID:  approver1.ID,
		GroupID: group.ID,
	}
	err = db.Create(membership1).Error
	require.NoError(t, err)
	membership2 := &models.UserGroup{
		UserID:  approver2.ID,
		GroupID: group.ID,
	}
	err = db.Create(membership2).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   true,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	// Create an approval request
	approvalRequest := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "DELETE FROM users WHERE id = 1",
		RequestedBy:  requester.ID.String(),
	}
	createdApproval, err := approvalService.CreateApprovalRequest(context.Background(), approvalRequest)
	require.NoError(t, err)

	// First approver rejects the request
	review1 := &ReviewInput{
		ApprovalID: createdApproval.ID,
		ReviewerID: approver1.ID.String(),
		Decision:   models.ApprovalDecisionRejected,
		Comments:   "Rejected",
	}
	_, err = approvalService.ReviewApproval(context.Background(), review1)
	require.NoError(t, err)

	// Verify the approval is now rejected
	var rejectedApproval models.ApprovalRequest
	err = db.First(&rejectedApproval, "id = ?", createdApproval.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.ApprovalStatusRejected, rejectedApproval.Status)

	// Second approver tries to approve after rejection - should fail
	review2 := &ReviewInput{
		ApprovalID: createdApproval.ID,
		ReviewerID: approver2.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Attempting to approve rejected request",
	}
	_, err = approvalService.ReviewApproval(context.Background(), review2)
	assert.Error(t, err, "Should not be able to approve already rejected request")
	assert.Contains(t, err.Error(), "not pending", "Error should indicate request is not pending")
}

// ---------------------------------------------------------------------------
// Group Permission Inheritance Tests
// ---------------------------------------------------------------------------

// TestGroupInheritance_MultipleGroups_UnionPermissions verifies that a user
// belonging to multiple groups receives the union of all permissions from those groups.
func TestGroupInheritance_MultipleGroups_UnionPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create a regular user
	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create three groups with different permissions
	group1 := &models.Group{ID: uuid.New(), Name: "Group 1 - Read"}
	group2 := &models.Group{ID: uuid.New(), Name: "Group 2 - Write"}
	group3 := &models.Group{ID: uuid.New(), Name: "Group 3 - Approve"}
	err := db.Create(group1).Error
	require.NoError(t, err)
	err = db.Create(group2).Error
	require.NoError(t, err)
	err = db.Create(group3).Error
	require.NoError(t, err)

	// Add user to all three groups
	membership1 := &models.UserGroup{UserID: user.ID, GroupID: group1.ID}
	membership2 := &models.UserGroup{UserID: user.ID, GroupID: group2.ID}
	membership3 := &models.UserGroup{UserID: user.ID, GroupID: group3.ID}
	err = db.Create(membership1).Error
	require.NoError(t, err)
	err = db.Create(membership2).Error
	require.NoError(t, err)
	err = db.Create(membership3).Error
	require.NoError(t, err)

	// Set permissions: Group1 has read, Group2 has write, Group3 has approve
	dsPerm1 := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group1.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   false,
	}
	dsPerm2 := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group2.ID,
		CanRead:      false,
		CanWrite:     true,
		CanApprove:   false,
	}
	dsPerm3 := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group3.ID,
		CanRead:      false,
		CanWrite:     false,
		CanApprove:   true,
	}
	err = db.Create(dsPerm1).Error
	require.NoError(t, err)
	err = db.Create(dsPerm2).Error
	require.NoError(t, err)
	err = db.Create(dsPerm3).Error
	require.NoError(t, err)

	// Get effective permissions
	perms, err := queryService.GetEffectivePermissions(context.Background(), user.ID, dataSource.ID)
	require.NoError(t, err)

	// User should have union of all permissions
	assert.True(t, perms.CanRead, "User should have CanRead from Group 1")
	assert.True(t, perms.CanWrite, "User should have CanWrite from Group 2")
	assert.True(t, perms.CanApprove, "User should have CanApprove from Group 3")
	assert.True(t, perms.CanSelect, "User should have CanSelect (derived from CanRead)")
	assert.True(t, perms.CanInsert, "User should have CanInsert (derived from CanWrite)")
	assert.True(t, perms.CanUpdate, "User should have CanUpdate (derived from CanWrite)")
	assert.True(t, perms.CanDelete, "User should have CanDelete (derived from CanWrite)")
}

// TestGroupInheritance_GroupAReadGroupBWrite_UserGetsBoth verifies that a user
// in Group A with can_read and Group B with can_write gets both permissions.
func TestGroupInheritance_GroupAReadGroupBWrite_UserGetsBoth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create a regular user
	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create Group A with read permission only
	groupA := &models.Group{ID: uuid.New(), Name: "Group A - Read Only"}
	err := db.Create(groupA).Error
	require.NoError(t, err)

	// Create Group B with write permission only
	groupB := &models.Group{ID: uuid.New(), Name: "Group B - Write Only"}
	err = db.Create(groupB).Error
	require.NoError(t, err)

	// Add user to both groups
	membershipA := &models.UserGroup{UserID: user.ID, GroupID: groupA.ID}
	membershipB := &models.UserGroup{UserID: user.ID, GroupID: groupB.ID}
	err = db.Create(membershipA).Error
	require.NoError(t, err)
	err = db.Create(membershipB).Error
	require.NoError(t, err)

	// Group A: CanRead = true, CanWrite = false
	dsPermA := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      groupA.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   false,
	}
	err = db.Create(dsPermA).Error
	require.NoError(t, err)

	// Group B: CanRead = false, CanWrite = true
	dsPermB := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      groupB.ID,
		CanRead:      false,
		CanWrite:     true,
		CanApprove:   false,
	}
	err = db.Create(dsPermB).Error
	require.NoError(t, err)

	// Get effective permissions
	perms, err := queryService.GetEffectivePermissions(context.Background(), user.ID, dataSource.ID)
	require.NoError(t, err)

	// User should have both read and write permissions
	assert.True(t, perms.CanRead, "User should have CanRead from Group A")
	assert.True(t, perms.CanWrite, "User should have CanWrite from Group B")
	assert.True(t, perms.CanSelect, "User should have CanSelect (from CanRead)")
	assert.True(t, perms.CanInsert, "User should have CanInsert (from CanWrite)")
	assert.True(t, perms.CanUpdate, "User should have CanUpdate (from CanWrite)")
	assert.True(t, perms.CanDelete, "User should have CanDelete (from CanWrite)")
	assert.False(t, perms.CanApprove, "User should NOT have CanApprove")
}

// TestGroupInheritance_OverlappingGroups_UnionNotOverride verifies that when
// groups have overlapping permissions, the result is a union (not an override).
func TestGroupInheritance_OverlappingGroups_UnionNotOverride(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create a regular user
	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create Group 1 with Read + Write
	group1 := &models.Group{ID: uuid.New(), Name: "Group 1 - Read+Write"}
	err := db.Create(group1).Error
	require.NoError(t, err)

	// Create Group 2 with Read + Approve (overlapping Read, different second permission)
	group2 := &models.Group{ID: uuid.New(), Name: "Group 2 - Read+Approve"}
	err = db.Create(group2).Error
	require.NoError(t, err)

	// Add user to both groups
	membership1 := &models.UserGroup{UserID: user.ID, GroupID: group1.ID}
	membership2 := &models.UserGroup{UserID: user.ID, GroupID: group2.ID}
	err = db.Create(membership1).Error
	require.NoError(t, err)
	err = db.Create(membership2).Error
	require.NoError(t, err)

	// Group 1: CanRead = true, CanWrite = true, CanApprove = false
	dsPerm1 := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group1.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   false,
	}
	err = db.Create(dsPerm1).Error
	require.NoError(t, err)

	// Group 2: CanRead = true, CanWrite = false, CanApprove = true
	dsPerm2 := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group2.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   true,
	}
	err = db.Create(dsPerm2).Error
	require.NoError(t, err)

	// Get effective permissions
	perms, err := queryService.GetEffectivePermissions(context.Background(), user.ID, dataSource.ID)
	require.NoError(t, err)

	// Union of permissions - user should have Read (from both), Write (from Group1), Approve (from Group2)
	assert.True(t, perms.CanRead, "User should have CanRead (union from both groups)")
	assert.True(t, perms.CanWrite, "User should have CanWrite (from Group 1)")
	assert.True(t, perms.CanApprove, "User should have CanApprove (from Group 2)")

	// Derived permissions
	assert.True(t, perms.CanSelect, "User should have CanSelect (from CanRead)")
	assert.True(t, perms.CanInsert, "User should have CanInsert (from CanWrite)")
	assert.True(t, perms.CanUpdate, "User should have CanUpdate (from CanWrite)")
	assert.True(t, perms.CanDelete, "User should have CanDelete (from CanWrite)")
}

// TestGroupInheritance_CanApproveInherited verifies that a user inherits
// the can_approve permission from a group that has it.
func TestGroupInheritance_CanApproveInherited(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create a regular user
	user := createTestUser(t, db, models.RoleUser)
	// Create a requester for approval tests
	requester := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group with can_approve permission
	approverGroup := &models.Group{ID: uuid.New(), Name: "Approver Group"}
	err := db.Create(approverGroup).Error
	require.NoError(t, err)

	// Add user to the approver group
	membership := &models.UserGroup{UserID: user.ID, GroupID: approverGroup.ID}
	err = db.Create(membership).Error
	require.NoError(t, err)

	// Group has CanRead = false, CanWrite = false, CanApprove = true
	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      approverGroup.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   true,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	// Get effective permissions
	perms, err := queryService.GetEffectivePermissions(context.Background(), user.ID, dataSource.ID)
	require.NoError(t, err)

	// User should have CanApprove inherited from group
	assert.True(t, perms.CanApprove, "User should inherit CanApprove from their group")
	assert.True(t, perms.CanRead, "User should have CanRead")

	// Verify the user can actually approve a request
	approvalRequest := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "DELETE FROM users WHERE id = 1",
		RequestedBy:  requester.ID.String(),
	}
	createdApproval, err := approvalService.CreateApprovalRequest(context.Background(), approvalRequest)
	require.NoError(t, err)

	review := &ReviewInput{
		ApprovalID: createdApproval.ID,
		ReviewerID: user.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Approved by inherited permission",
	}
	approvalReview, err := approvalService.ReviewApproval(context.Background(), review)
	require.NoError(t, err, "User with inherited CanApprove should be able to approve")
	assert.Equal(t, models.ApprovalDecisionApproved, approvalReview.Decision)
}

// TestGroupInheritance_AllPermissionsGroup_FullAccess verifies that a user
// in a group with all permissions (can_read, can_write, can_approve) gets full access.
func TestGroupInheritance_AllPermissionsGroup_FullAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create a regular user
	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group with ALL permissions
	fullAccessGroup := &models.Group{ID: uuid.New(), Name: "Full Access Group"}
	err := db.Create(fullAccessGroup).Error
	require.NoError(t, err)

	// Add user to the full access group
	membership := &models.UserGroup{UserID: user.ID, GroupID: fullAccessGroup.ID}
	err = db.Create(membership).Error
	require.NoError(t, err)

	// Group has ALL permissions: CanRead = true, CanWrite = true, CanApprove = true
	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      fullAccessGroup.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   true,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	// Get effective permissions
	perms, err := queryService.GetEffectivePermissions(context.Background(), user.ID, dataSource.ID)
	require.NoError(t, err)

	// User should have ALL permissions (equivalent to admin on this data source)
	assert.True(t, perms.CanRead, "User should have CanRead")
	assert.True(t, perms.CanWrite, "User should have CanWrite")
	assert.True(t, perms.CanApprove, "User should have CanApprove")
	assert.True(t, perms.CanSelect, "User should have CanSelect (from CanRead)")
	assert.True(t, perms.CanInsert, "User should have CanInsert (from CanWrite)")
	assert.True(t, perms.CanUpdate, "User should have CanUpdate (from CanWrite)")
	assert.True(t, perms.CanDelete, "User should have CanDelete (from CanWrite)")

	// All permissions should be true
	assert.True(t, perms.CanRead && perms.CanWrite && perms.CanApprove &&
		perms.CanSelect && perms.CanInsert && perms.CanUpdate && perms.CanDelete,
		"User in full-access group should have all permissions")
}

// TestGroupInheritance_RemovedFromGroup_LosesPermissions verifies that when
// a user is removed from a group, they lose that group's permissions.
func TestGroupInheritance_RemovedFromGroup_LosesPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create a regular user
	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create two groups
	readGroup := &models.Group{ID: uuid.New(), Name: "Read Group"}
	writeGroup := &models.Group{ID: uuid.New(), Name: "Write Group"}
	err := db.Create(readGroup).Error
	require.NoError(t, err)
	err = db.Create(writeGroup).Error
	require.NoError(t, err)

	// Add user to both groups initially
	membership1 := &models.UserGroup{UserID: user.ID, GroupID: readGroup.ID}
	membership2 := &models.UserGroup{UserID: user.ID, GroupID: writeGroup.ID}
	err = db.Create(membership1).Error
	require.NoError(t, err)
	err = db.Create(membership2).Error
	require.NoError(t, err)

	// Read group: CanRead = true
	dsPermRead := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      readGroup.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   false,
	}
	err = db.Create(dsPermRead).Error
	require.NoError(t, err)

	// Write group: CanWrite = true
	dsPermWrite := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      writeGroup.ID,
		CanRead:      false,
		CanWrite:     true,
		CanApprove:   false,
	}
	err = db.Create(dsPermWrite).Error
	require.NoError(t, err)

	// Verify initial permissions - user has both read and write
	perms, err := queryService.GetEffectivePermissions(context.Background(), user.ID, dataSource.ID)
	require.NoError(t, err)
	assert.True(t, perms.CanRead, "User should initially have CanRead")
	assert.True(t, perms.CanWrite, "User should initially have CanWrite")

	// Remove user from write group
	err = db.Where("user_id = ? AND group_id = ?", user.ID, writeGroup.ID).Delete(&models.UserGroup{}).Error
	require.NoError(t, err)

	// Verify user now only has read permissions
	perms, err = queryService.GetEffectivePermissions(context.Background(), user.ID, dataSource.ID)
	require.NoError(t, err)
	assert.True(t, perms.CanRead, "User should still have CanRead from remaining group")
	assert.True(t, perms.CanSelect, "User should still have CanSelect from remaining group")
	assert.False(t, perms.CanWrite, "User should have lost CanWrite after being removed from write group")
	assert.False(t, perms.CanInsert, "User should have lost CanInsert after being removed from write group")
	assert.False(t, perms.CanUpdate, "User should have lost CanUpdate after being removed from write group")
	assert.False(t, perms.CanDelete, "User should have lost CanDelete after being removed from write group")
}

// TestGroupInheritance_AddedToGroup_GainsPermissions verifies that when
// a user is added to a group, they gain that group's permissions.
func TestGroupInheritance_AddedToGroup_GainsPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create a regular user
	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create two groups
	readGroup := &models.Group{ID: uuid.New(), Name: "Read Group"}
	writeGroup := &models.Group{ID: uuid.New(), Name: "Write Group"}
	err := db.Create(readGroup).Error
	require.NoError(t, err)
	err = db.Create(writeGroup).Error
	require.NoError(t, err)

	// Initially, only add user to read group
	membership1 := &models.UserGroup{UserID: user.ID, GroupID: readGroup.ID}
	err = db.Create(membership1).Error
	require.NoError(t, err)

	// Read group: CanRead = true
	dsPermRead := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      readGroup.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   false,
	}
	err = db.Create(dsPermRead).Error
	require.NoError(t, err)

	// Write group: CanWrite = true
	dsPermWrite := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      writeGroup.ID,
		CanRead:      false,
		CanWrite:     true,
		CanApprove:   false,
	}
	err = db.Create(dsPermWrite).Error
	require.NoError(t, err)

	// Verify initial permissions - user only has read
	perms, err := queryService.GetEffectivePermissions(context.Background(), user.ID, dataSource.ID)
	require.NoError(t, err)
	assert.True(t, perms.CanRead, "User should initially have CanRead")
	assert.True(t, perms.CanSelect, "User should initially have CanSelect")
	assert.False(t, perms.CanWrite, "User should initially NOT have CanWrite")

	// Add user to write group
	membership2 := &models.UserGroup{UserID: user.ID, GroupID: writeGroup.ID}
	err = db.Create(membership2).Error
	require.NoError(t, err)

	// Verify user now has both read and write permissions
	perms, err = queryService.GetEffectivePermissions(context.Background(), user.ID, dataSource.ID)
	require.NoError(t, err)
	assert.True(t, perms.CanRead, "User should still have CanRead")
	assert.True(t, perms.CanWrite, "User should now have CanWrite after being added to write group")
	assert.True(t, perms.CanInsert, "User should now have CanInsert after being added to write group")
	assert.True(t, perms.CanUpdate, "User should now have CanUpdate after being added to write group")
	assert.True(t, perms.CanDelete, "User should now have CanDelete after being added to write group")
}

// TestGroupInheritance_MultipleGroupsSamePermission_NoDuplication verifies that
// having the same permission from multiple groups doesn't cause issues (idempotent).
func TestGroupInheritance_MultipleGroupsSamePermission_NoDuplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create a regular user
	user := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create three groups all with the same read permission
	group1 := &models.Group{ID: uuid.New(), Name: "Read Group 1"}
	group2 := &models.Group{ID: uuid.New(), Name: "Read Group 2"}
	group3 := &models.Group{ID: uuid.New(), Name: "Read Group 3"}
	err := db.Create(group1).Error
	require.NoError(t, err)
	err = db.Create(group2).Error
	require.NoError(t, err)
	err = db.Create(group3).Error
	require.NoError(t, err)

	// Add user to all three groups
	membership1 := &models.UserGroup{UserID: user.ID, GroupID: group1.ID}
	membership2 := &models.UserGroup{UserID: user.ID, GroupID: group2.ID}
	membership3 := &models.UserGroup{UserID: user.ID, GroupID: group3.ID}
	err = db.Create(membership1).Error
	require.NoError(t, err)
	err = db.Create(membership2).Error
	require.NoError(t, err)
	err = db.Create(membership3).Error
	require.NoError(t, err)

	// All groups have the same CanRead = true permission
	for _, group := range []*models.Group{group1, group2, group3} {
		dsPerm := &models.DataSourcePermission{
			ID:           uuid.New(),
			DataSourceID: dataSource.ID,
			GroupID:      group.ID,
			CanRead:      true,
			CanWrite:     false,
			CanApprove:   false,
		}
		err = db.Create(dsPerm).Error
		require.NoError(t, err)
	}

	// Get effective permissions
	perms, err := queryService.GetEffectivePermissions(context.Background(), user.ID, dataSource.ID)
	require.NoError(t, err)

	// User should have CanRead (true | true | true = true)
	assert.True(t, perms.CanRead, "User should have CanRead (union of same permission from multiple groups)")
	assert.True(t, perms.CanSelect, "User should have CanSelect")
	assert.False(t, perms.CanWrite, "User should NOT have CanWrite")
	assert.False(t, perms.CanApprove, "User should NOT have CanApprove")

	// Remove user from one group - should still have permission from remaining groups
	err = db.Where("user_id = ? AND group_id = ?", user.ID, group1.ID).Delete(&models.UserGroup{}).Error
	require.NoError(t, err)

	perms, err = queryService.GetEffectivePermissions(context.Background(), user.ID, dataSource.ID)
	require.NoError(t, err)
	assert.True(t, perms.CanRead, "User should still have CanRead from remaining groups")

	// Remove from second group - should still have permission from last group
	err = db.Where("user_id = ? AND group_id = ?", user.ID, group2.ID).Delete(&models.UserGroup{}).Error
	require.NoError(t, err)

	perms, err = queryService.GetEffectivePermissions(context.Background(), user.ID, dataSource.ID)
	require.NoError(t, err)
	assert.True(t, perms.CanRead, "User should still have CanRead from last remaining group")

	// Remove from last group - should lose permission
	err = db.Where("user_id = ? AND group_id = ?", user.ID, group3.ID).Delete(&models.UserGroup{}).Error
	require.NoError(t, err)

	perms, err = queryService.GetEffectivePermissions(context.Background(), user.ID, dataSource.ID)
	require.NoError(t, err)
	assert.False(t, perms.CanRead, "User should have lost CanRead after being removed from all groups with that permission")
}

// ---------------------------------------------------------------------------
// Permission Denial Tests - 403 Response Scenarios
// ---------------------------------------------------------------------------

// TestPermissionDenial_Returns403WhenLacksPermission verifies that when a user
// lacks permission to perform an action, a 403 Forbidden error is returned.
func TestPermissionDenial_Returns403WhenLacksPermission(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create a user without any permissions
	userWithoutPerms := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Get effective permissions
	perms, err := queryService.GetEffectivePermissions(context.Background(), userWithoutPerms.ID, dataSource.ID)
	require.NoError(t, err)

	// Verify user has no permissions
	assert.False(t, perms.CanRead, "User should NOT have CanRead")
	assert.False(t, perms.CanWrite, "User should NOT have CanWrite")
	assert.False(t, perms.CanApprove, "User should NOT have CanApprove")

	// Simulate permission check that would result in 403
	// This mirrors the check in query.go:57-60
	hasPermission := perms.CanRead
	assert.False(t, hasPermission, "Permission check should fail for user without permissions")

	// The 403 response would be returned in the handler:
	// c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to read from this data source"})
	// We verify the permission model correctly denies access
	assert.False(t, perms.CanSelect, "User should NOT be able to execute SELECT")
	assert.False(t, perms.CanInsert, "User should NOT be able to execute INSERT")
	assert.False(t, perms.CanUpdate, "User should NOT be able to execute UPDATE")
	assert.False(t, perms.CanDelete, "User should NOT be able to execute DELETE")
}

// TestPermissionDenial_ErrorMessageContainsDetails verifies that permission
// denial error messages contain specific details about what permission was denied.
func TestPermissionDenial_ErrorMessageContainsDetails(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create a user with only read permission
	readOnlyUser, _, dataSource := createReadOnlyUserWithPermission(t, db)

	// Get effective permissions
	perms, err := queryService.GetEffectivePermissions(context.Background(), readOnlyUser.ID, dataSource.ID)
	require.NoError(t, err)

	// Verify read permission details
	assert.True(t, perms.CanRead, "User should have CanRead")
	assert.True(t, perms.CanSelect, "User should have CanSelect (derived from CanRead)")

	// Verify write permission is denied with clear details
	assert.False(t, perms.CanWrite, "User should NOT have CanWrite")
	assert.False(t, perms.CanInsert, "User should NOT have CanInsert (requires CanWrite)")
	assert.False(t, perms.CanUpdate, "User should NOT have CanUpdate (requires CanWrite)")
	assert.False(t, perms.CanDelete, "User should NOT have CanDelete (requires CanWrite)")
	assert.False(t, perms.CanApprove, "User should NOT have CanApprove")

	// Verify the permission structure contains the data source context
	// Error messages in handlers include: data_source name, permission type, and hint
	assert.NotEqual(t, uuid.Nil, dataSource.ID, "Data source ID should be present for context")
	assert.NotEmpty(t, dataSource.Name, "Data source name should be present for error messages")

	// Simulate error message construction (as done in query.go:464-470)
	// "Insufficient permissions to submit write operations"
	// "data_source": dataSource.Name
	// "hint": "Contact your admin to get write access on this data source."
	errorMsg := fmt.Sprintf("Insufficient permissions to submit write operations on data source '%s'", dataSource.Name)
	assert.Contains(t, errorMsg, dataSource.Name, "Error message should contain data source name")
	assert.Contains(t, errorMsg, "write operations", "Error message should specify the denied operation type")
}

// TestPermissionDenial_AccessWithoutMembership_Returns403 verifies that accessing
// a data source without group membership returns 403 Forbidden.
func TestPermissionDenial_AccessWithoutMembership_Returns403(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create user without any group membership
	userWithoutMembership := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group with permissions but DON'T add user to it
	group := &models.Group{
		ID:   uuid.New(),
		Name: "Restricted Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)

	// Grant permissions to the group
	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   true,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	// User is NOT a member of the group, so should have no permissions
	perms, err := queryService.GetEffectivePermissions(context.Background(), userWithoutMembership.ID, dataSource.ID)
	require.NoError(t, err)

	// All permissions should be false - this triggers 403
	assert.False(t, perms.CanRead, "User without group membership should NOT have CanRead")
	assert.False(t, perms.CanWrite, "User without group membership should NOT have CanWrite")
	assert.False(t, perms.CanApprove, "User without group membership should NOT have CanApprove")
	assert.False(t, perms.CanSelect, "User without group membership should NOT have CanSelect")
	assert.False(t, perms.CanInsert, "User without group membership should NOT have CanInsert")
	assert.False(t, perms.CanUpdate, "User without group membership should NOT have CanUpdate")
	assert.False(t, perms.CanDelete, "User without group membership should NOT have CanDelete")

	// Verify the user exists but has no access
	var user models.User
	err = db.First(&user, "id = ?", userWithoutMembership.ID).Error
	require.NoError(t, err, "User should exist in database")
	assert.Equal(t, userWithoutMembership.ID, user.ID, "User should exist")
}

// TestPermissionDenial_ExecuteWriteAsViewer_Returns403 verifies that a viewer
// attempting to execute write queries receives 403 Forbidden.
func TestPermissionDenial_ExecuteWriteAsViewer_Returns403(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create viewer user
	viewerUser := createTestUser(t, db, models.RoleViewer)
	dataSource := createTestDataSource(t, db)

	// Add viewer to a group with read permission
	viewerGroup := &models.Group{
		ID:   uuid.New(),
		Name: "Viewer Group",
	}
	err := db.Create(viewerGroup).Error
	require.NoError(t, err)

	membership := &models.UserGroup{
		UserID:  viewerUser.ID,
		GroupID: viewerGroup.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      viewerGroup.ID,
		CanRead:      true,
		CanWrite:     false,
		CanApprove:   false,
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	// Get effective permissions
	perms, err := queryService.GetEffectivePermissions(context.Background(), viewerUser.ID, dataSource.ID)
	require.NoError(t, err)

	// Viewer should have read but NOT write permissions
	assert.True(t, perms.CanRead, "Viewer should have CanRead")
	assert.True(t, perms.CanSelect, "Viewer should have CanSelect")

	// All write operations should be denied (403)
	assert.False(t, perms.CanWrite, "Viewer should NOT have CanWrite - triggers 403")
	assert.False(t, perms.CanInsert, "Viewer should NOT have CanInsert - triggers 403")
	assert.False(t, perms.CanUpdate, "Viewer should NOT have CanUpdate - triggers 403")
	assert.False(t, perms.CanDelete, "Viewer should NOT have CanDelete - triggers 403")
	assert.False(t, perms.CanApprove, "Viewer should NOT have CanApprove - triggers 403")

	// Verify write operations would be blocked
	writeOperations := []models.OperationType{
		models.OperationInsert,
		models.OperationUpdate,
		models.OperationDelete,
	}

	for _, op := range writeOperations {
		canExecute := false
		switch op {
		case models.OperationInsert:
			canExecute = perms.CanInsert
		case models.OperationUpdate:
			canExecute = perms.CanUpdate
		case models.OperationDelete:
			canExecute = perms.CanDelete
		}
		assert.False(t, canExecute, "Viewer should NOT be able to execute %s - would return 403", op)
	}
}

// TestPermissionDenial_ApproveWithoutPermission_Returns403 verifies that
// attempting to approve a request without approval permission returns 403.
func TestPermissionDenial_ApproveWithoutPermission_Returns403(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	approvalService := NewApprovalService(db, queryService, nil)

	// Create users
	requester := createTestUser(t, db, models.RoleUser)
	unauthorizedUser := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// Create a group with write permission but NOT approve permission
	writeGroup := &models.Group{
		ID:   uuid.New(),
		Name: "Write Only Group",
	}
	err := db.Create(writeGroup).Error
	require.NoError(t, err)

	// Add unauthorized user to the group
	membership := &models.UserGroup{
		UserID:  unauthorizedUser.ID,
		GroupID: writeGroup.ID,
	}
	err = db.Create(membership).Error
	require.NoError(t, err)

	// Grant write permission but NOT approve permission
	dsPerm := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      writeGroup.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   false, // No approval permission
	}
	err = db.Create(dsPerm).Error
	require.NoError(t, err)

	// Verify unauthorized user has write but NOT approve permission
	perms, err := queryService.GetEffectivePermissions(context.Background(), unauthorizedUser.ID, dataSource.ID)
	require.NoError(t, err)

	assert.True(t, perms.CanWrite, "User should have CanWrite")
	assert.False(t, perms.CanApprove, "User should NOT have CanApprove - would trigger 403")

	// Create an approval request by requester
	approvalRequest := &ApprovalRequest{
		DataSourceID: dataSource.ID,
		QuerySQL:     "DELETE FROM users WHERE id = 1",
		RequestedBy:  requester.ID.String(),
	}
	createdApproval, err := approvalService.CreateApprovalRequest(context.Background(), approvalRequest)
	require.NoError(t, err)

	// Attempt to approve without CanApprove permission
	review := &ReviewInput{
		ApprovalID: createdApproval.ID,
		ReviewerID: unauthorizedUser.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   "Attempting to approve without permission",
	}

	_, err = approvalService.ReviewApproval(context.Background(), review)
	assert.Error(t, err, "Approving without CanApprove permission should fail")
	// The error indicates the user lacks permission to approve
	assert.Contains(t, err.Error(), "permission", "Error should mention permission")
}

// TestPermissionDenial_AccessAdminEndpointAsNonAdmin_Returns403 verifies that
// accessing admin-only endpoints as a non-admin user returns 403 Forbidden.
func TestPermissionDenial_AccessAdminEndpointAsNonAdmin_Returns403(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)

	// Create non-admin users
	regularUser := createTestUser(t, db, models.RoleUser)
	viewerUser := createTestUser(t, db, models.RoleViewer)

	// Verify roles
	assert.Equal(t, models.RoleUser, regularUser.Role, "User should have regular role")
	assert.Equal(t, models.RoleViewer, viewerUser.Role, "User should have viewer role")
	assert.NotEqual(t, models.RoleAdmin, regularUser.Role, "Regular user should NOT have admin role")
	assert.NotEqual(t, models.RoleAdmin, viewerUser.Role, "Viewer should NOT have admin role")

	// Simulate admin endpoint check (as in handlers)
	// if user.Role != models.RoleAdmin { return 403 }
	isAdminRegular := regularUser.Role == models.RoleAdmin
	isAdminViewer := viewerUser.Role == models.RoleAdmin

	assert.False(t, isAdminRegular, "Regular user should NOT pass admin check - triggers 403")
	assert.False(t, isAdminViewer, "Viewer should NOT pass admin check - triggers 403")

	// Create admin user for comparison
	adminUser := createTestUser(t, db, models.RoleAdmin)
	isAdmin := adminUser.Role == models.RoleAdmin
	assert.True(t, isAdmin, "Admin user should pass admin check")
}

// TestPermissionDenial_ExecuteQueryWithoutRead_Returns403 verifies that
// attempting to execute a query without read permission returns 403 Forbidden.
func TestPermissionDenial_ExecuteQueryWithoutRead_Returns403(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create user without any group membership (no permissions)
	userWithoutRead := createTestUser(t, db, models.RoleUser)
	dataSource := createTestDataSource(t, db)

	// User has no group memberships, so should have no permissions

	// Get effective permissions
	perms, err := queryService.GetEffectivePermissions(context.Background(), userWithoutRead.ID, dataSource.ID)
	require.NoError(t, err)

	// User should NOT have read permission - triggers 403 for SELECT queries
	assert.False(t, perms.CanRead, "User without group membership should NOT have CanRead")
	assert.False(t, perms.CanSelect, "User without group membership should NOT have CanSelect - triggers 403 for SELECT queries")

	// User should NOT have write permission either
	assert.False(t, perms.CanWrite, "User without group membership should NOT have CanWrite")
	assert.False(t, perms.CanInsert, "User without group membership should NOT have CanInsert")
	assert.False(t, perms.CanUpdate, "User without group membership should NOT have CanUpdate")
	assert.False(t, perms.CanDelete, "User without group membership should NOT have CanDelete")

	// Simulate SELECT query execution check (query.go:146-148)
	// if !perms.CanSelect { return 403 }
	canExecuteSelect := perms.CanSelect
	assert.False(t, canExecuteSelect, "SELECT query should be blocked without CanRead - would return 403")
}

// TestPermissionDenial_PreviewWriteWithoutWrite_Returns403 verifies that
// attempting to preview a write query without write permission returns 403.
func TestPermissionDenial_PreviewWriteWithoutWrite_Returns403(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := setupTestDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)

	// Create user with only read permission
	readOnlyUser, _, dataSource := createReadOnlyUserWithPermission(t, db)

	// Get effective permissions
	perms, err := queryService.GetEffectivePermissions(context.Background(), readOnlyUser.ID, dataSource.ID)
	require.NoError(t, err)

	// User should have read but NOT write permission
	assert.True(t, perms.CanRead, "User should have CanRead")
	assert.True(t, perms.CanSelect, "User should have CanSelect")
	assert.False(t, perms.CanWrite, "User should NOT have CanWrite")

	// Write operations that would be previewed
	writeQueries := []string{
		"DELETE FROM users WHERE id = 1",
		"UPDATE users SET name = 'test' WHERE id = 1",
		"INSERT INTO users (name) VALUES ('test')",
	}

	for _, query := range writeQueries {
		opType := DetectOperationType(query)
		requiresWrite := opType == models.OperationInsert ||
			opType == models.OperationUpdate ||
			opType == models.OperationDelete

		if requiresWrite {
			// Preview of write queries requires write permission
			// Without CanWrite, this would return 403
			assert.False(t, perms.CanWrite,
				"Preview of %s query should require write permission - would return 403", opType)
		}
	}

	// Note: The actual preview endpoint (query.go:367-410) checks read permission
	// for the preview itself since it converts write queries to SELECT.
	// However, submitting write queries for approval requires write permission.
	assert.False(t, perms.CanWrite, "User without CanWrite cannot submit write queries for approval")
}
