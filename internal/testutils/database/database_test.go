package database

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/yourorg/querybase/internal/models"
)

func TestSetupTestDB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := SetupTestDB(t)
	require.NotNil(t, db, "SetupTestDB should return a non-nil database")

	var count int64
	err := db.Raw("SELECT COUNT(*) FROM sqlite_master").Scan(&count).Error
	require.NoError(t, err, "should be able to query the database")

	assert.True(t, count > 0, "database should have tables")
}

func TestMigrateTestSchema(t *testing.T) {
	db := SetupTestDB(t)

	tables := []string{
		"users",
		"groups",
		"user_groups",
		"data_sources",
		"data_source_permissions",
		"queries",
		"query_results",
		"query_history",
		"approval_requests",
		"approval_reviews",
		"approval_comments",
		"query_transactions",
		"query_transaction_statements",
		"notification_configs",
		"notifications",
	}

	for _, table := range tables {
		var count int64
		err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count).Error
		require.NoError(t, err, "should query table existence")
		assert.Equal(t, int64(1), count, "table %s should exist", table)
	}
}

func TestCleanupTestDB(t *testing.T) {
	db := SetupTestDB(t)
	CleanupTestDB(db)
}

func TestRunTestWithTransaction(t *testing.T) {
	db := SetupTestDB(t)

	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Username: "testuser",
	}

	RunTestWithTransaction(t, db, func(tx *gorm.DB) {
		err := tx.Create(user).Error
		require.NoError(t, err, "should create user in transaction")

		var count int64
		err = tx.Model(&models.User{}).Count(&count).Error
		require.NoError(t, err, "should count users in transaction")
		assert.Equal(t, int64(1), count, "should have 1 user in transaction")
	})

	var totalCount int64
	err := db.Model(&models.User{}).Count(&totalCount).Error
	require.NoError(t, err, "should count users in main db")
	assert.Equal(t, int64(0), totalCount, "transaction should be rolled back")
}

func TestRunTestInTransaction(t *testing.T) {
	db := SetupTestDB(t)

	user := &models.User{
		ID:       uuid.New(),
		Email:    "test2@example.com",
		Username: "testuser2",
	}

	tx := RunTestInTransaction(t, db, func(txDB *gorm.DB) {})

	err := tx.Create(user).Error
	require.NoError(t, err, "should create user in transaction")

	var count int64
	err = tx.Model(&models.User{}).Count(&count).Error
	require.NoError(t, err, "should count users in transaction")
	assert.Equal(t, int64(1), count, "should have 1 user in transaction")

	CleanupTransaction(tx)

	var totalCount int64
	err = db.Model(&models.User{}).Count(&totalCount).Error
	require.NoError(t, err, "should count users in main db")
	assert.Equal(t, int64(0), totalCount, "transaction should be rolled back")
}

func TestGetTestConfig(t *testing.T) {
	config := GetTestConfig()

	require.NotNil(t, config, "GetTestConfig should return a non-nil config")
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, "debug", config.Server.Mode)
	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port)
	assert.Equal(t, "querybase_test", config.Database.Name)
	assert.Equal(t, "postgresql", config.Database.Dialect)
	assert.Equal(t, "test-secret-key-for-unit-tests-only", config.JWT.Secret)
}

func TestSetupTestDBWithPostgres(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := GetTestConfig()
	db := SetupTestDBWithPostgres(t, &config.Database)

	if db != nil {
		var count int64
		err := db.Raw("SELECT 1").Scan(&count).Error
		require.NoError(t, err, "should be able to query the database")
	}
}

func TestSetupTestDBWithContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := SetupTestDBWithContainer(t)
	if db == nil {
		t.Skip("Testcontainers not available")
	}
	defer cleanup()

	require.NotNil(t, db, "SetupTestDBWithContainer should return a non-nil database")

	var result int
	err := db.Raw("SELECT 1").Scan(&result).Error
	require.NoError(t, err, "should be able to query the database")
	assert.Equal(t, 1, result, "should get expected result")
}

func TestGetContainerConfig(t *testing.T) {
	config := GetContainerConfig("localhost", "5432")

	require.NotNil(t, config, "GetContainerConfig should return a non-nil config")
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 5432, config.Port)
	assert.Equal(t, "test", config.User)
	assert.Equal(t, "test", config.Password)
	assert.Equal(t, "test", config.Name)
	assert.Equal(t, "postgresql", config.Dialect)
}

func TestUserModelCreation(t *testing.T) {
	db := SetupTestDB(t)

	user := &models.User{
		ID:       uuid.New(),
		Email:    "model-test@example.com",
		Username: "modeluser",
		FullName: "Model Test User",
		Role:     models.RoleUser,
	}

	RunTestWithTransaction(t, db, func(tx *gorm.DB) {
		err := tx.Create(user).Error
		require.NoError(t, err, "should create user")

		var found models.User
		err = tx.First(&found, "id = ?", user.ID).Error
		require.NoError(t, err, "should find user")
		assert.Equal(t, user.Email, found.Email)
		assert.Equal(t, user.Username, found.Username)
	})
}
