package database

import (
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/yourorg/querybase/internal/config"
	"github.com/yourorg/querybase/internal/models"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to create in-memory SQLite database: %v", err)
	}

	if err := MigrateTestSchema(db); err != nil {
		t.Fatalf("failed to migrate test schema: %v", err)
	}

	t.Cleanup(func() {
		CleanupTestDB(db)
	})

	return db
}

func SetupTestDBWithPostgres(t *testing.T, dbConfig *config.DatabaseConfig) *gorm.DB {
	t.Helper()

	dsn := dbConfig.GetDatabaseDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("cannot connect to test PostgreSQL: %v", err)
		return nil
	}

	if err := MigrateTestSchema(db); err != nil {
		t.Fatalf("failed to migrate test schema: %v", err)
	}

	t.Cleanup(func() {
		CleanupTestDB(db)
	})

	return db
}

func CleanupTestDB(db *gorm.DB) {
	if db == nil {
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		return
	}

	sqlDB.Close()
}

func MigrateTestSchema(db *gorm.DB) error {
	return db.AutoMigrate(
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
		&models.ApprovalComment{},
		&models.QueryTransaction{},
		&models.QueryTransactionStatement{},
		&models.NotificationConfig{},
		&models.Notification{},
	)
}

func GetTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
			Mode: "debug",
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "querybase",
			Password: "querybase",
			Name:     "querybase_test",
			SSLMode:  "disable",
			Dialect:  "postgresql",
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		JWT: config.JWTConfig{
			Secret:      "test-secret-key-for-unit-tests-only",
			ExpireHours: 24,
			Issuer:      "querybase-test",
		},
		CORS: config.CORSConfig{
			AllowedOrigins:   "http://localhost:3000",
			AllowCredentials: true,
			MaxAge:           86400,
		},
	}
}
