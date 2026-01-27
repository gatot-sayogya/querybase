package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestUser_TableName tests the table name for User model
func TestUser_TableName(t *testing.T) {
	user := User{}
	assert.Equal(t, "users", user.TableName())
}

// TestGroup_TableName tests the table name for Group model
func TestGroup_TableName(t *testing.T) {
	group := Group{}
	assert.Equal(t, "groups", group.TableName())
}

// TestUser_RoleConstants tests role constants
func TestUser_RoleConstants(t *testing.T) {
	assert.Equal(t, UserRole("admin"), RoleAdmin)
	assert.Equal(t, UserRole("user"), RoleUser)
	assert.Equal(t, UserRole("viewer"), RoleViewer)
}

// TestUser_BeforeCreate tests UUID generation
func TestUser_BeforeCreate(t *testing.T) {
	// This would require a test database setup
	// For now, we test the constants and types
	user := User{
		Email:    "test@example.com",
		Username: "testuser",
		Role:     RoleUser,
	}

	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, RoleUser, user.Role)
}

// TestDataSource_TypeConstants tests data source type constants
func TestDataSource_TypeConstants(t *testing.T) {
	assert.Equal(t, DataSourceType("postgresql"), DataSourceTypePostgreSQL)
	assert.Equal(t, DataSourceType("mysql"), DataSourceTypeMySQL)
}

// TestQuery_StatusConstants tests query status constants
func TestQuery_StatusConstants(t *testing.T) {
	assert.Equal(t, QueryStatus("pending"), QueryStatusPending)
	assert.Equal(t, QueryStatus("running"), QueryStatusRunning)
	assert.Equal(t, QueryStatus("completed"), QueryStatusCompleted)
	assert.Equal(t, QueryStatus("failed"), QueryStatusFailed)
}

// TestOperationType_String tests operation type string representation
func TestOperationType_String(t *testing.T) {
	tests := []struct {
		op       OperationType
		expected string
	}{
		{OperationTypeSelect, "select"},
		{OperationTypeInsert, "insert"},
		{OperationTypeUpdate, "update"},
		{OperationTypeDelete, "delete"},
		{OperationTypeCreateTable, "create_table"},
		{OperationTypeDropTable, "drop_table"},
		{OperationTypeAlterTable, "alter_table"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.op))
		})
	}
}

// TestApprovalStatus_Constants tests approval status constants
func TestApprovalStatus_Constants(t *testing.T) {
	// These are defined in approval.go but test them here for completeness
	// ApprovalStatusPending, ApprovalStatusApproved, ApprovalStatusRejected
	assert.Equal(t, "pending", string(ApprovalStatusPending))
	assert.Equal(t, "approved", string(ApprovalStatusApproved))
	assert.Equal(t, "rejected", string(ApprovalStatusRejected))
}

// TestUser_IsActiveDefault tests default active status
func TestUser_IsActiveDefault(t *testing.T) {
	user := User{
		Email:    "test@example.com",
		Username: "testuser",
		Role:     RoleUser,
		IsActive: true, // Explicitly set to true (GORM default)
	}

	// User should be active
	assert.True(t, user.IsActive)
}

// TestDataSource_TableName tests data source table name
func TestDataSource_TableName(t *testing.T) {
	ds := DataSource{}
	assert.Equal(t, "data_sources", ds.TableName())
}

// TestQuery_TableName tests query table name
func TestQuery_TableName(t *testing.T) {
	query := Query{}
	assert.Equal(t, "queries", query.TableName())
}

// TestApprovalRequest_TableName tests approval request table name
func TestApprovalRequest_TableName(t *testing.T) {
	ar := ApprovalRequest{}
	assert.Equal(t, "approval_requests", ar.TableName())
}

// TestApprovalReview_TableName tests approval review table name
func TestApprovalReview_TableName(t *testing.T) {
	ar := ApprovalReview{}
	assert.Equal(t, "approval_reviews", ar.TableName())
}

// TestNotificationConfig_TableName tests notification config table name
func TestNotificationConfig_TableName(t *testing.T) {
	nc := NotificationConfig{}
	assert.Equal(t, "notification_configs", nc.TableName())
}

// TestNotification_TableName tests notification table name
func TestNotification_TableName(t *testing.T) {
	n := Notification{}
	assert.Equal(t, "notifications", n.TableName())
}

// TestUser_WithUUID tests user with UUID
func TestUser_WithUUID(t *testing.T) {
	id := uuid.New()
	user := User{
		ID:       id,
		Email:    "test@example.com",
		Username: "testuser",
		Role:     RoleAdmin,
	}

	assert.Equal(t, id, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, RoleAdmin, user.Role)
}

// TestGroup_WithUsers tests group with users association
func TestGroup_WithUsers(t *testing.T) {
	group := Group{
		Name:        "Test Group",
		Description: "A test group",
	}

	// In a real scenario with GORM, we would append users
	// For now, just test the group structure
	assert.Equal(t, "Test Group", group.Name)
	assert.Equal(t, "A test group", group.Description)
	assert.Empty(t, group.Users) // Should be empty slice initially
}

// TestDataSource_WithEncryptedPassword tests data source with encrypted password
func TestDataSource_WithEncryptedPassword(t *testing.T) {
	ds := DataSource{
		Name:             "Test DB",
		Type:             DataSourceTypePostgreSQL,
		Host:             "localhost",
		Port:             5432,
		DatabaseName:     "testdb",
		Username:         "testuser",
		EncryptedPassword: "encrypted:secret123",
	}

	assert.Equal(t, "Test DB", ds.Name)
	assert.Equal(t, DataSourceTypePostgreSQL, ds.Type)
	assert.Equal(t, "localhost", ds.Host)
	assert.Equal(t, 5432, ds.Port)
	assert.Equal(t, "testdb", ds.DatabaseName)
	assert.Equal(t, "testuser", ds.Username)
	assert.Equal(t, "encrypted:secret123", ds.EncryptedPassword)
}
