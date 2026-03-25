package datasource

import (
	"fmt"
	"testing"

	"gorm.io/gorm"

	"github.com/yourorg/querybase/internal/models"
)

// SetupDataSourceWithTestTable creates a test table in the data source.
// The table is automatically dropped when the test completes.
func SetupDataSourceWithTestTable(t *testing.T, db *gorm.DB, ds *models.DataSource, tableName, columns string) {
	t.Helper()

	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, columns)

	if err := db.Exec(createSQL).Error; err != nil {
		t.Fatalf("Failed to create test table %s: %v", tableName, err)
	}

	t.Cleanup(func() {
		DropTestTable(t, db, ds, tableName)
	})
}

// DropTestTable drops a test table from the data source.
func DropTestTable(t *testing.T, db *gorm.DB, ds *models.DataSource, tableName string) {
	t.Helper()

	var dropSQL string
	switch ds.Type {
	case models.DataSourceTypePostgreSQL:
		dropSQL = fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	case models.DataSourceTypeMySQL:
		dropSQL = fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	default:
		t.Fatalf("Unsupported data source type: %s", ds.Type)
	}

	if err := db.Exec(dropSQL).Error; err != nil {
		t.Logf("Warning: Failed to drop test table %s: %v", tableName, err)
	}
}

// SetupUserTestTable creates the standard test_users table.
func SetupUserTestTable(t *testing.T, db *gorm.DB, ds *models.DataSource) {
	t.Helper()

	var createSQL string
	switch ds.Type {
	case models.DataSourceTypePostgreSQL:
		createSQL = `
			CREATE TABLE IF NOT EXISTS test_users (
				id SERIAL PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				email VARCHAR(255) UNIQUE,
				age INTEGER,
				status VARCHAR(50) DEFAULT 'active',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`
	case models.DataSourceTypeMySQL:
		createSQL = `
			CREATE TABLE IF NOT EXISTS test_users (
				id INT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				email VARCHAR(255) UNIQUE,
				age INT,
				status VARCHAR(50) DEFAULT 'active',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`
	default:
		t.Fatalf("Unsupported data source type: %s", ds.Type)
	}

	if err := db.Exec(createSQL).Error; err != nil {
		t.Fatalf("Failed to create test_users table: %v", err)
	}

	t.Cleanup(func() {
		DropTestTable(t, db, ds, "test_users")
	})
}

// SetupOrderTestTable creates the standard test_orders table.
func SetupOrderTestTable(t *testing.T, db *gorm.DB, ds *models.DataSource) {
	t.Helper()

	var createSQL string
	switch ds.Type {
	case models.DataSourceTypePostgreSQL:
		createSQL = `
			CREATE TABLE IF NOT EXISTS test_orders (
				id SERIAL PRIMARY KEY,
				user_id INTEGER,
				product VARCHAR(100),
				quantity INTEGER DEFAULT 1,
				price DECIMAL(10,2),
				status VARCHAR(50) DEFAULT 'pending',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`
	case models.DataSourceTypeMySQL:
		createSQL = `
			CREATE TABLE IF NOT EXISTS test_orders (
				id INT AUTO_INCREMENT PRIMARY KEY,
				user_id INT,
				product VARCHAR(100),
				quantity INT DEFAULT 1,
				price DECIMAL(10,2),
				status VARCHAR(50) DEFAULT 'pending',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`
	default:
		t.Fatalf("Unsupported data source type: %s", ds.Type)
	}

	if err := db.Exec(createSQL).Error; err != nil {
		t.Fatalf("Failed to create test_orders table: %v", err)
	}

	t.Cleanup(func() {
		DropTestTable(t, db, ds, "test_orders")
	})
}

// InsertTestData inserts test data into a table.
func InsertTestData(t *testing.T, db *gorm.DB, tableName string, columns string, values string) {
	t.Helper()

	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tableName, columns, values)

	if err := db.Exec(insertSQL).Error; err != nil {
		t.Fatalf("Failed to insert test data into %s: %v", tableName, err)
	}
}

// InsertTestUsers inserts test user data.
func InsertTestUsers(t *testing.T, db *gorm.DB, users []map[string]any) {
	t.Helper()

	for _, user := range users {
		if err := db.Table("test_users").Create(user).Error; err != nil {
			t.Fatalf("Failed to insert test user: %v", err)
		}
	}
}

// InsertTestOrders inserts test order data.
func InsertTestOrders(t *testing.T, db *gorm.DB, orders []map[string]any) {
	t.Helper()

	for _, order := range orders {
		if err := db.Table("test_orders").Create(order).Error; err != nil {
			t.Fatalf("Failed to insert test order: %v", err)
		}
	}
}

// GetRowCount returns the number of rows in a table.
func GetRowCount(t *testing.T, db *gorm.DB, tableName string) int64 {
	t.Helper()

	var count int64
	if err := db.Table(tableName).Count(&count).Error; err != nil {
		t.Fatalf("Failed to count rows in %s: %v", tableName, err)
	}
	return count
}

// TableExists checks if a table exists in the database.
func TableExists(t *testing.T, db *gorm.DB, ds *models.DataSource, tableName string) bool {
	t.Helper()

	var exists bool
	var err error

	switch ds.Type {
	case models.DataSourceTypePostgreSQL:
		err = db.Raw(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = ?
			)
		`, tableName).Scan(&exists).Error
	case models.DataSourceTypeMySQL:
		err = db.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = DATABASE()
				AND table_name = ?
			)
		`, tableName).Scan(&exists).Error
	default:
		t.Fatalf("Unsupported data source type: %s", ds.Type)
	}

	if err != nil {
		t.Fatalf("Failed to check if table exists: %v", err)
	}

	return exists
}

// CreateTableWithForeignKey creates a table with a foreign key constraint.
func CreateTableWithForeignKey(t *testing.T, db *gorm.DB, ds *models.DataSource, tableName, columns, foreignKey string) {
	t.Helper()

	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s, %s)", tableName, columns, foreignKey)

	if err := db.Exec(createSQL).Error; err != nil {
		t.Fatalf("Failed to create table %s with foreign key: %v", tableName, err)
	}

	t.Cleanup(func() {
		DropTestTable(t, db, ds, tableName)
	})
}

// TruncateTable truncates a table (removes all rows).
func TruncateTable(t *testing.T, db *gorm.DB, tableName string) {
	t.Helper()

	var truncateSQL string

	if db.Config.Dialector.Name() == "mysql" {
		truncateSQL = fmt.Sprintf("TRUNCATE TABLE %s", tableName)
	} else {
		truncateSQL = fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName)
	}

	if err := db.Exec(truncateSQL).Error; err != nil {
		t.Fatalf("Failed to truncate table %s: %v", tableName, err)
	}
}
