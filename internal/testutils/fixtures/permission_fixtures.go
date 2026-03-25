package fixtures

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/yourorg/querybase/internal/models"
)

// CreateTestDataSource creates a test data source with the specified name.
// The data source is automatically cleaned up when the test completes.
func CreateTestDataSource(t *testing.T, db *gorm.DB, name string) *models.DataSource {
	t.Helper()

	dataSource := &models.DataSource{
		ID:                uuid.New(),
		Name:              name,
		Type:              models.DataSourceTypePostgreSQL,
		Host:              "localhost",
		Port:              5432,
		DatabaseName:      fmt.Sprintf("test_db_%s", uuid.New().String()[:8]),
		Username:          "testuser",
		EncryptedPassword: "encrypted-password-placeholder",
		IsActive:          true,
		IsHealthy:         true,
		AuditRowThreshold: 1000,
	}

	if err := db.Create(dataSource).Error; err != nil {
		t.Fatalf("Failed to create test data source: %v", err)
	}

	t.Cleanup(func() {
		db.Unscoped().Delete(dataSource)
	})

	return dataSource
}

// CreateTestDataSourceWithType creates a test data source with a specific type.
func CreateTestDataSourceWithType(t *testing.T, db *gorm.DB, name string, dsType models.DataSourceType) *models.DataSource {
	t.Helper()

	port := 5432
	if dsType == models.DataSourceTypeMySQL {
		port = 3306
	}

	dataSource := &models.DataSource{
		ID:                uuid.New(),
		Name:              name,
		Type:              dsType,
		Host:              "localhost",
		Port:              port,
		DatabaseName:      fmt.Sprintf("test_db_%s", uuid.New().String()[:8]),
		Username:          "testuser",
		EncryptedPassword: "encrypted-password-placeholder",
		IsActive:          true,
		IsHealthy:         true,
		AuditRowThreshold: 1000,
	}

	if err := db.Create(dataSource).Error; err != nil {
		t.Fatalf("Failed to create test data source: %v", err)
	}

	t.Cleanup(func() {
		db.Unscoped().Delete(dataSource)
	})

	return dataSource
}

// CreateTestMySQLDataSource creates a test MySQL data source.
func CreateTestMySQLDataSource(t *testing.T, db *gorm.DB, name string) *models.DataSource {
	return CreateTestDataSourceWithType(t, db, name, models.DataSourceTypeMySQL)
}

// GrantPermission grants permissions to a group for a data source.
// Returns the created permission.
func GrantPermission(db *gorm.DB, groupID, dsID uuid.UUID, canRead, canWrite, canApprove bool) (*models.DataSourcePermission, error) {
	permission := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dsID,
		GroupID:      groupID,
		CanRead:      canRead,
		CanWrite:     canWrite,
		CanApprove:   canApprove,
	}

	if err := db.Create(permission).Error; err != nil {
		return nil, err
	}

	return permission, nil
}

// GrantPermissionWithCleanup grants permissions and registers cleanup for a test.
func GrantPermissionWithCleanup(t *testing.T, db *gorm.DB, groupID, dsID uuid.UUID, canRead, canWrite, canApprove bool) *models.DataSourcePermission {
	t.Helper()

	permission, err := GrantPermission(db, groupID, dsID, canRead, canWrite, canApprove)
	if err != nil {
		t.Fatalf("Failed to grant permission: %v", err)
	}

	t.Cleanup(func() {
		db.Unscoped().Delete(permission)
	})

	return permission
}

// GrantReadOnlyPermission grants read-only permission to a group.
func GrantReadOnlyPermission(t *testing.T, db *gorm.DB, groupID, dsID uuid.UUID) *models.DataSourcePermission {
	return GrantPermissionWithCleanup(t, db, groupID, dsID, true, false, false)
}

// GrantWritePermission grants read and write permission to a group.
func GrantWritePermission(t *testing.T, db *gorm.DB, groupID, dsID uuid.UUID) *models.DataSourcePermission {
	return GrantPermissionWithCleanup(t, db, groupID, dsID, true, true, false)
}

// GrantFullPermission grants full permission (read, write, approve) to a group.
func GrantFullPermission(t *testing.T, db *gorm.DB, groupID, dsID uuid.UUID) *models.DataSourcePermission {
	return GrantPermissionWithCleanup(t, db, groupID, dsID, true, true, true)
}

// GetDataSourcePermissions retrieves all permissions for a data source.
func GetDataSourcePermissions(db *gorm.DB, dsID uuid.UUID) ([]models.DataSourcePermission, error) {
	var permissions []models.DataSourcePermission
	err := db.Where("data_source_id = ?", dsID).Find(&permissions).Error
	return permissions, err
}

// GetGroupPermissions retrieves all permissions for a group.
func GetGroupPermissions(db *gorm.DB, groupID uuid.UUID) ([]models.DataSourcePermission, error) {
	var permissions []models.DataSourcePermission
	err := db.Where("group_id = ?", groupID).Find(&permissions).Error
	return permissions, err
}

// RevokePermission deletes a specific permission.
func RevokePermission(db *gorm.DB, permissionID uuid.UUID) error {
	return db.Unscoped().Delete(&models.DataSourcePermission{}, "id = ?", permissionID).Error
}
