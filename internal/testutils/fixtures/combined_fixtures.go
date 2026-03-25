package fixtures

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/yourorg/querybase/internal/models"
)

// SetupAdminUser sets up a complete admin user scenario:
// - Creates an admin user
// - Creates a group
// - Adds user to the group
// - Creates a data source
// - Grants full permission to the group
// Returns all created entities.
func SetupAdminUser(db *gorm.DB) (*models.User, *models.Group, *models.DataSource, *models.DataSourcePermission) {
	// Create admin user
	passwordHash := "admin-hash-placeholder" // Simplified for fixture
	user := &models.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("admin-%s@querybase.local", uuid.New().String()[:8]),
		Username:     fmt.Sprintf("admin-%s", uuid.New().String()[:8]),
		PasswordHash: passwordHash,
		FullName:     "Admin User",
		Role:         models.RoleAdmin,
		IsActive:     true,
	}
	db.Create(user)

	// Create group
	group := &models.Group{
		ID:          uuid.New(),
		Name:        fmt.Sprintf("admin-group-%s", uuid.New().String()[:8]),
		Description: "Admin group for testing",
	}
	db.Create(group)

	// Add user to group
	userGroup := &models.UserGroup{
		UserID:  user.ID,
		GroupID: group.ID,
	}
	db.Create(userGroup)

	// Create data source
	dataSource := &models.DataSource{
		ID:                uuid.New(),
		Name:              fmt.Sprintf("admin-datasource-%s", uuid.New().String()[:8]),
		Type:              models.DataSourceTypePostgreSQL,
		Host:              "localhost",
		Port:              5432,
		DatabaseName:      "querybase",
		Username:          "admin",
		EncryptedPassword: "encrypted-password",
		IsActive:          true,
		IsHealthy:         true,
	}
	db.Create(dataSource)

	// Grant full permission
	permission := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      true,
		CanWrite:     true,
		CanApprove:   true,
	}
	db.Create(permission)

	return user, group, dataSource, permission
}

// SetupAdminUserWithCleanup sets up admin user with automatic cleanup.
func SetupAdminUserWithCleanup(t *testing.T, db *gorm.DB) (*models.User, *models.Group, *models.DataSource, *models.DataSourcePermission) {
	t.Helper()

	user, group, dataSource, permission := SetupAdminUser(db)

	t.Cleanup(func() {
		db.Unscoped().Where("id = ?", permission.ID).Delete(&models.DataSourcePermission{})
		db.Unscoped().Where("id = ?", dataSource.ID).Delete(&models.DataSource{})
		db.Unscoped().Where("user_id = ? AND group_id = ?", user.ID, group.ID).Delete(&models.UserGroup{})
		db.Unscoped().Where("id = ?", group.ID).Delete(&models.Group{})
		db.Unscoped().Where("id = ?", user.ID).Delete(&models.User{})
	})

	return user, group, dataSource, permission
}

// SetupRegularUser sets up a complete regular user scenario with specified permissions:
// - Creates a regular user
// - Creates a group
// - Adds user to the group
// - Creates a data source
// - Grants permissions based on canRead, canWrite, canApprove parameters
// Returns all created entities.
func SetupRegularUser(db *gorm.DB, canRead, canWrite, canApprove bool) (*models.User, *models.Group, *models.DataSource, *models.DataSourcePermission) {
	// Create regular user
	passwordHash := "user-hash-placeholder" // Simplified for fixture
	user := &models.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("user-%s@querybase.local", uuid.New().String()[:8]),
		Username:     fmt.Sprintf("user-%s", uuid.New().String()[:8]),
		PasswordHash: passwordHash,
		FullName:     "Regular User",
		Role:         models.RoleUser,
		IsActive:     true,
	}
	db.Create(user)

	// Create group
	group := &models.Group{
		ID:          uuid.New(),
		Name:        fmt.Sprintf("user-group-%s", uuid.New().String()[:8]),
		Description: "User group for testing",
	}
	db.Create(group)

	// Add user to group
	userGroup := &models.UserGroup{
		UserID:  user.ID,
		GroupID: group.ID,
	}
	db.Create(userGroup)

	// Create data source
	dataSource := &models.DataSource{
		ID:                uuid.New(),
		Name:              fmt.Sprintf("user-datasource-%s", uuid.New().String()[:8]),
		Type:              models.DataSourceTypePostgreSQL,
		Host:              "localhost",
		Port:              5432,
		DatabaseName:      "querybase",
		Username:          "user",
		EncryptedPassword: "encrypted-password",
		IsActive:          true,
		IsHealthy:         true,
	}
	db.Create(dataSource)

	// Grant permission
	permission := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dataSource.ID,
		GroupID:      group.ID,
		CanRead:      canRead,
		CanWrite:     canWrite,
		CanApprove:   canApprove,
	}
	db.Create(permission)

	return user, group, dataSource, permission
}

// SetupRegularUserWithCleanup sets up regular user with automatic cleanup.
func SetupRegularUserWithCleanup(t *testing.T, db *gorm.DB, canRead, canWrite, canApprove bool) (*models.User, *models.Group, *models.DataSource, *models.DataSourcePermission) {
	t.Helper()

	user, group, dataSource, permission := SetupRegularUser(db, canRead, canWrite, canApprove)

	t.Cleanup(func() {
		db.Unscoped().Where("id = ?", permission.ID).Delete(&models.DataSourcePermission{})
		db.Unscoped().Where("id = ?", dataSource.ID).Delete(&models.DataSource{})
		db.Unscoped().Where("user_id = ? AND group_id = ?", user.ID, group.ID).Delete(&models.UserGroup{})
		db.Unscoped().Where("id = ?", group.ID).Delete(&models.Group{})
		db.Unscoped().Where("id = ?", user.ID).Delete(&models.User{})
	})

	return user, group, dataSource, permission
}

// SetupViewerUser sets up a complete viewer user scenario:
// - Creates a viewer user
// - Creates a group
// - Adds user to the group
// - Creates a data source
// - Grants read-only permission to the group
// Returns all created entities.
func SetupViewerUser(db *gorm.DB) (*models.User, *models.Group, *models.DataSource, *models.DataSourcePermission) {
	return SetupRegularUser(db, true, false, false)
}

// SetupViewerUserWithCleanup sets up viewer user with automatic cleanup.
func SetupViewerUserWithCleanup(t *testing.T, db *gorm.DB) (*models.User, *models.Group, *models.DataSource, *models.DataSourcePermission) {
	t.Helper()
	return SetupRegularUserWithCleanup(t, db, true, false, false)
}

// SetupMultiGroupUser sets up a user that belongs to multiple groups with different permissions.
// Returns the user, groups (with permissions), and data sources.
func SetupMultiGroupUser(t *testing.T, db *gorm.DB, numGroups int) (*models.User, []*models.Group, []*models.DataSource) {
	t.Helper()

	if numGroups < 1 {
		numGroups = 1
	}

	// Create user
	passwordHash := "multigroup-hash-placeholder"
	user := &models.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("multigroup-%s@querybase.local", uuid.New().String()[:8]),
		Username:     fmt.Sprintf("multigroup-%s", uuid.New().String()[:8]),
		PasswordHash: passwordHash,
		FullName:     "Multi-Group User",
		Role:         models.RoleUser,
		IsActive:     true,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create multi-group user: %v", err)
	}

	groups := make([]*models.Group, numGroups)
	dataSources := make([]*models.DataSource, numGroups)

	for i := 0; i < numGroups; i++ {
		// Create group
		group := &models.Group{
			ID:          uuid.New(),
			Name:        fmt.Sprintf("multigroup-%s-group-%d", uuid.New().String()[:8], i),
			Description: fmt.Sprintf("Group %d for multi-group user", i),
		}
		if err := db.Create(group).Error; err != nil {
			t.Fatalf("Failed to create group: %v", err)
		}
		groups[i] = group

		// Add user to group
		userGroup := &models.UserGroup{
			UserID:  user.ID,
			GroupID: group.ID,
		}
		if err := db.Create(userGroup).Error; err != nil {
			t.Fatalf("Failed to add user to group: %v", err)
		}

		// Create data source
		dataSource := &models.DataSource{
			ID:                uuid.New(),
			Name:              fmt.Sprintf("multigroup-%s-ds-%d", uuid.New().String()[:8], i),
			Type:              models.DataSourceTypePostgreSQL,
			Host:              "localhost",
			Port:              5432,
			DatabaseName:      fmt.Sprintf("db_%d", i),
			Username:          "user",
			EncryptedPassword: "encrypted",
			IsActive:          true,
			IsHealthy:         true,
		}
		if err := db.Create(dataSource).Error; err != nil {
			t.Fatalf("Failed to create data source: %v", err)
		}
		dataSources[i] = dataSource

		// Grant permission (escalating permissions for each group)
		canRead := true
		canWrite := i > 0
		canApprove := i > 1

		permission := &models.DataSourcePermission{
			ID:           uuid.New(),
			DataSourceID: dataSource.ID,
			GroupID:      group.ID,
			CanRead:      canRead,
			CanWrite:     canWrite,
			CanApprove:   canApprove,
		}
		if err := db.Create(permission).Error; err != nil {
			t.Fatalf("Failed to grant permission: %v", err)
		}
	}

	t.Cleanup(func() {
		// Clean up permissions first
		for _, ds := range dataSources {
			db.Unscoped().Where("data_source_id = ?", ds.ID).Delete(&models.DataSourcePermission{})
		}
		// Clean up data sources
		for _, ds := range dataSources {
			db.Unscoped().Delete(ds)
		}
		// Clean up user groups
		db.Unscoped().Where("user_id = ?", user.ID).Delete(&models.UserGroup{})
		// Clean up groups
		for _, g := range groups {
			db.Unscoped().Delete(g)
		}
		// Clean up user
		db.Unscoped().Delete(user)
	})

	return user, groups, dataSources
}

// CleanupTestData manually cleans up test data in the correct order.
// Use this when you need manual control over cleanup timing.
func CleanupTestData(db *gorm.DB, entities ...interface{}) {
	for _, entity := range entities {
		db.Unscoped().Delete(entity)
	}
}
