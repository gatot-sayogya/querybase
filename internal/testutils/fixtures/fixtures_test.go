package fixtures

import (
	"testing"

	"github.com/yourorg/querybase/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func getTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "host=localhost port=5432 user=querybase password=querybase dbname=querybase sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
		return nil
	}

	sqlDB, _ := db.DB()
	t.Cleanup(func() {
		sqlDB.Close()
	})

	return db
}

func TestCreateTestUser(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}

	user := CreateTestUser(t, db, models.RoleUser)

	if user.ID.String() == "" {
		t.Error("Expected user ID to be set")
	}
	if user.Email == "" {
		t.Error("Expected user email to be set")
	}
	if user.Role != models.RoleUser {
		t.Errorf("Expected role %s, got %s", models.RoleUser, user.Role)
	}
	if !user.IsActive {
		t.Error("Expected user to be active")
	}
}

func TestCreateTestGroup(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}

	group := CreateTestGroupWithUniqueName(t, db)

	if group.ID.String() == "" {
		t.Error("Expected group ID to be set")
	}
	if group.Name == "" {
		t.Error("Expected group name to be set")
	}
}

func TestAddUserToGroup(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}

	user := CreateTestUser(t, db, models.RoleUser)
	group := CreateTestGroupWithUniqueName(t, db)

	err := AddUserToGroup(db, user.ID, group.ID)
	if err != nil {
		t.Errorf("Failed to add user to group: %v", err)
	}

	var count int64
	db.Model(&models.UserGroup{}).Where("user_id = ? AND group_id = ?", user.ID, group.ID).Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 user group entry, got %d", count)
	}
}

func TestCreateTestDataSource(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}

	ds := CreateTestDataSource(t, db, "test-datasource")

	if ds.ID.String() == "" {
		t.Error("Expected data source ID to be set")
	}
	if ds.Name != "test-datasource" {
		t.Errorf("Expected name 'test-datasource', got '%s'", ds.Name)
	}
	if ds.Type != models.DataSourceTypePostgreSQL {
		t.Errorf("Expected type PostgreSQL, got %s", ds.Type)
	}
}

func TestGrantPermission(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}

	group := CreateTestGroupWithUniqueName(t, db)
	ds := CreateTestDataSource(t, db, "test-datasource-perm")

	perm, err := GrantPermission(db, group.ID, ds.ID, true, false, false)
	if err != nil {
		t.Errorf("Failed to grant permission: %v", err)
	}

	if perm.CanRead != true {
		t.Error("Expected CanRead to be true")
	}
	if perm.CanWrite != false {
		t.Error("Expected CanWrite to be false")
	}
	if perm.CanApprove != false {
		t.Error("Expected CanApprove to be false")
	}
}

func TestSetupAdminUserWithCleanup(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}

	user, group, ds, perm := SetupAdminUserWithCleanup(t, db)

	if user.ID.String() == "" {
		t.Error("Expected user ID to be set")
	}
	if user.Role != models.RoleAdmin {
		t.Errorf("Expected admin role, got %s", user.Role)
	}
	if group.ID.String() == "" {
		t.Error("Expected group ID to be set")
	}
	if ds.ID.String() == "" {
		t.Error("Expected data source ID to be set")
	}
	if !perm.CanRead || !perm.CanWrite || !perm.CanApprove {
		t.Error("Expected full permissions")
	}
}

func TestSetupViewerUserWithCleanup(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}

	user, _, _, perm := SetupViewerUserWithCleanup(t, db)

	if user.Role != models.RoleUser {
		t.Errorf("Expected user role, got %s", user.Role)
	}
	if !perm.CanRead {
		t.Error("Expected CanRead to be true")
	}
	if perm.CanWrite {
		t.Error("Expected CanWrite to be false")
	}
	if perm.CanApprove {
		t.Error("Expected CanApprove to be false")
	}
}

func TestSetupRegularUserWithCleanup(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}

	_, _, _, perm := SetupRegularUserWithCleanup(t, db, true, true, false) //nolint:govet

	if !perm.CanRead {
		t.Error("Expected CanRead to be true")
	}
	if !perm.CanWrite {
		t.Error("Expected CanWrite to be true")
	}
	if perm.CanApprove {
		t.Error("Expected CanApprove to be false")
	}
}
