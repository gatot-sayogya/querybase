package datasource

import (
	"database/sql"
	"fmt"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/yourorg/querybase/internal/models"
)

// TestConnectionConfig holds connection configuration for test data sources.
type TestConnectionConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	DatabaseName string
	Type         models.DataSourceType
}

// GetTestPostgreSQLConfig returns a test PostgreSQL data source configuration.
func GetTestPostgreSQLConfig() *models.DataSource {
	return &models.DataSource{
		ID:           [16]byte{}, // Will be overwritten
		Name:         "test-postgresql",
		Type:         models.DataSourceTypePostgreSQL,
		Host:         "localhost",
		Port:         5432,
		DatabaseName: "test_db",
		Username:     "testuser",
	}
}

// GetTestMySQLConfig returns a test MySQL data source configuration.
func GetTestMySQLConfig() *models.DataSource {
	return &models.DataSource{
		ID:           [16]byte{}, // Will be overwritten
		Name:         "test-mysql",
		Type:         models.DataSourceTypeMySQL,
		Host:         "localhost",
		Port:         3306,
		DatabaseName: "test_db",
		Username:     "testuser",
	}
}

// GetTestPostgreSQLDSN returns the PostgreSQL DSN for a data source.
func GetTestPostgreSQLDSN(ds *models.DataSource, password string) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		ds.Host,
		ds.Port,
		ds.Username,
		password,
		ds.DatabaseName,
	)
}

// GetTestMySQLDSN returns the MySQL DSN for a data source.
func GetTestMySQLDSN(ds *models.DataSource, password string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		ds.Username,
		password,
		ds.Host,
		ds.Port,
		ds.DatabaseName,
	)
}

// ConnectToDataSource establishes a connection to the data source and returns the GORM DB.
func ConnectToDataSource(t *testing.T, ds *models.DataSource, password string) *gorm.DB {
	t.Helper()

	var db *gorm.DB
	var err error

	switch ds.Type {
	case models.DataSourceTypePostgreSQL:
		db, err = gorm.Open(postgres.Open(GetTestPostgreSQLDSN(ds, password)), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	case models.DataSourceTypeMySQL:
		db, err = gorm.Open(mysql.Open(GetTestMySQLDSN(ds, password)), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	default:
		t.Fatalf("Unsupported data source type: %s", ds.Type)
	}

	if err != nil {
		t.Fatalf("Failed to connect to data source: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	})

	return db
}

// ConnectToDataSourceWithConfig establishes a connection using test connection config.
func ConnectToDataSourceWithConfig(t *testing.T, config *TestConnectionConfig) *gorm.DB {
	t.Helper()

	var db *gorm.DB
	var err error

	switch config.Type {
	case models.DataSourceTypePostgreSQL:
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Host,
			config.Port,
			config.Username,
			config.Password,
			config.DatabaseName,
		)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	case models.DataSourceTypeMySQL:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.DatabaseName,
		)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	default:
		t.Fatalf("Unsupported data source type: %s", config.Type)
	}

	if err != nil {
		t.Fatalf("Failed to connect to data source: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	})

	return db
}

// TestConnection tests connectivity to a data source.
func TestConnection(t *testing.T, ds *models.DataSource, password string) error {
	t.Helper()

	var db *gorm.DB
	var err error

	switch ds.Type {
	case models.DataSourceTypePostgreSQL:
		db, err = gorm.Open(postgres.Open(GetTestPostgreSQLDSN(ds, password)), &gorm.Config{})
	case models.DataSourceTypeMySQL:
		db, err = gorm.Open(mysql.Open(GetTestMySQLDSN(ds, password)), &gorm.Config{})
	default:
		return fmt.Errorf("unsupported data source type: %s", ds.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying DB: %w", err)
	}
	defer sqlDB.Close()

	return sqlDB.Ping()
}

// GetRawDBConnection returns the underlying sql.DB connection.
func GetRawDBConnection(ds *models.DataSource, password string) (*sql.DB, error) {
	var db *gorm.DB
	var err error

	switch ds.Type {
	case models.DataSourceTypePostgreSQL:
		db, err = gorm.Open(postgres.Open(GetTestPostgreSQLDSN(ds, password)), &gorm.Config{})
	case models.DataSourceTypeMySQL:
		db, err = gorm.Open(mysql.Open(GetTestMySQLDSN(ds, password)), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported data source type: %s", ds.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return db.DB()
}

// CreateTestDatabase creates a test database for the given data source type.
func CreateTestDatabase(t *testing.T, baseDS *models.DataSource, basePassword, newDBName string) *models.DataSource {
	t.Helper()

	newDS := &models.DataSource{
		ID:           baseDS.ID,
		Name:         baseDS.Name,
		Type:         baseDS.Type,
		Host:         baseDS.Host,
		Port:         baseDS.Port,
		Username:     baseDS.Username,
		DatabaseName: newDBName,
	}

	var createSQL string
	switch newDS.Type {
	case models.DataSourceTypePostgreSQL:
		createSQL = fmt.Sprintf("CREATE DATABASE %s", newDBName)
	case models.DataSourceTypeMySQL:
		createSQL = fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", newDBName)
	}

	baseDS.DatabaseName = "postgres"
	if newDS.Type == models.DataSourceTypeMySQL {
		baseDS.DatabaseName = ""
	}

	db := ConnectToDataSource(t, baseDS, basePassword)
	if err := db.Exec(createSQL).Error; err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	t.Cleanup(func() {
		var dropSQL string
		switch newDS.Type {
		case models.DataSourceTypePostgreSQL:
			dropSQL = fmt.Sprintf("DROP DATABASE IF EXISTS %s", newDBName)
		case models.DataSourceTypeMySQL:
			dropSQL = fmt.Sprintf("DROP DATABASE IF EXISTS %s", newDBName)
		}
		db.Exec(dropSQL)
	})

	newDS.EncryptedPassword = baseDS.EncryptedPassword
	return newDS
}
