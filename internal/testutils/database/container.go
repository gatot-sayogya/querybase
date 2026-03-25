package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/yourorg/querybase/internal/config"
	"github.com/yourorg/querybase/internal/models"
)

const (
	postgresImage    = "postgres:15-alpine"
	postgresUser     = "test"
	postgresPass     = "test"
	postgresDB       = "test"
	containerRetries = 3
)

func SetupTestDBWithContainer(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Skipf("Docker not available: %v", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var container *postgres.PostgresContainer
	var err error

	for i := 0; i < containerRetries; i++ {
		container, err = postgres.Run(ctx,
			postgresImage,
			postgres.WithDatabase(postgresDB),
			postgres.WithUsername(postgresUser),
			postgres.WithPassword(postgresPass),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(10*time.Second),
			),
		)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		t.Skipf("failed to start PostgreSQL container: %v", err)
		return nil, func() {}
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get container port: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		port.Port(),
		postgresUser,
		postgresPass,
		postgresDB,
	)

	db, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to connect to PostgreSQL container: %v", err)
	}

	if err := migrateSchema(db); err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to migrate schema: %v", err)
	}

	cleanup := func() {
		if db != nil {
			sqlDB, _ := db.DB()
			sqlDB.Close()
		}
		container.Terminate(ctx)
	}

	t.Cleanup(cleanup)

	return db, cleanup
}

func migrateSchema(db *gorm.DB) error {
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

// GetContainerConfig returns a database config for connecting to a test container.
func GetContainerConfig(host string, port string) *config.DatabaseConfig {
	return &config.DatabaseConfig{
		Host:     host,
		Port:     5432,
		User:     postgresUser,
		Password: postgresPass,
		Name:     postgresDB,
		SSLMode:  "disable",
		Dialect:  "postgresql",
	}
}
