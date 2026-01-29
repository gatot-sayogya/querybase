package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/yourorg/querybase/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestExtractTableNames tests table name extraction from SQL queries
func TestExtractTableNames(t *testing.T) {
	queryService := &QueryService{} // We only need the struct, no initialization needed

	tests := []struct {
		name           string
		sql            string
		expectedTables []string
	}{
		{
			name:           "Simple SELECT",
			sql:            "SELECT * FROM users",
			expectedTables: []string{"users"},
		},
		{
			name:           "SELECT with JOIN",
			sql:            "SELECT * FROM users INNER JOIN orders ON users.id = orders.user_id",
			expectedTables: []string{"users", "orders"},
		},
		{
			name:           "INSERT query",
			sql:            "INSERT INTO users (name) VALUES ('John')",
			expectedTables: []string{"users"},
		},
		{
			name:           "UPDATE query",
			sql:            "UPDATE users SET name = 'Jane'",
			expectedTables: []string{"users"},
		},
		{
			name:           "DELETE query",
			sql:            "DELETE FROM users WHERE id = 1",
			expectedTables: []string{"users"},
		},
		{
			name:           "CREATE TABLE",
			sql:            "CREATE TABLE new_users (id INT)",
			expectedTables: []string{"new_users"},
		},
		{
			name:           "DROP TABLE",
			sql:            "DROP TABLE old_users",
			expectedTables: []string{"old_users"},
		},
		{
			name:           "ALTER TABLE",
			sql:            "ALTER TABLE users ADD COLUMN age INT",
			expectedTables: []string{"users"},
		},
		{
			name:           "Multiple JOINs",
			sql:            "SELECT * FROM orders JOIN users ON orders.user_id = users.id JOIN products ON orders.product_id = products.id",
			expectedTables: []string{"orders", "users", "products"},
		},
		{
			name:           "LEFT JOIN",
			sql:            "SELECT * FROM users LEFT JOIN orders ON users.id = orders.user_id",
			expectedTables: []string{"users", "orders"},
		},
		{
			name:           "RIGHT JOIN",
			sql:            "SELECT * FROM users RIGHT JOIN orders ON users.id = orders.user_id",
			expectedTables: []string{"users", "orders"},
		},
		{
			name:           "FULL OUTER JOIN",
			sql:            "SELECT * FROM users FULL OUTER JOIN orders ON users.id = orders.user_id",
			expectedTables: []string{"users", "orders"},
		},
		{
			name:           "Schema qualified table",
			sql:            "SELECT * FROM public.users",
			expectedTables: []string{"public.users"},
		},
		{
			name:           "Quoted table name",
			sql:            `SELECT * FROM "users"`,
			expectedTables: []string{"users"},
		},
		{
			name:           "SELECT with subquery",
			sql:            "SELECT * FROM users WHERE id IN (SELECT user_id FROM orders)",
			expectedTables: []string{"users", "orders"},
		},
		{
			name:           "Duplicate tables",
			sql:            "SELECT * FROM users JOIN orders ON users.id = orders.user_id JOIN products ON orders.product_id = products.id",
			expectedTables: []string{"users", "orders", "products"},
		},
		{
			name:           "INSERT with SELECT",
			sql:            "INSERT INTO users_archive SELECT * FROM users",
			expectedTables: []string{"users_archive", "users"},
		},
		{
			name:           "Lowercase query",
			sql:            "select * from users",
			expectedTables: []string{"users"},
		},
		{
			name:           "Mixed case query",
			sql:            "SeLeCt * FrOm users",
			expectedTables: []string{"users"},
		},
		{
			name:           "Query with comments",
			sql:            "SELECT * FROM users -- comment\nWHERE id = 1",
			expectedTables: []string{"users"},
		},
		{
			name:           "CREATE IF NOT EXISTS",
			sql:            "CREATE TABLE IF NOT EXISTS users (id INT)",
			expectedTables: []string{"users"},
		},
		{
			name:           "DROP IF EXISTS",
			sql:            "DROP TABLE IF EXISTS users",
			expectedTables: []string{"users"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tables, err := queryService.extractTableNames(tt.sql)
			if err != nil {
				t.Errorf("extractTableNames() unexpected error: %v", err)
				return
			}

			// Check if we got the expected number of tables
			if len(tables) != len(tt.expectedTables) {
				t.Errorf("extractTableNames() returned %d tables, want %d", len(tables), len(tt.expectedTables))
				t.Logf("Got: %v", tables)
				t.Logf("Want: %v", tt.expectedTables)
				return
			}

			// Check if all expected tables are present (using a map for easier comparison)
			tableMap := make(map[string]bool)
			for _, table := range tables {
				tableMap[table] = true
			}

			for _, expectedTable := range tt.expectedTables {
				if !tableMap[expectedTable] {
					t.Errorf("extractTableNames() missing table %s. Got: %v", expectedTable, tables)
				}
			}
		})
	}
}

// TestTableExists tests the table existence check (requires database)
func TestTableExists(t *testing.T) {
	// This test requires a real database connection
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	// Try to connect to test database
	dsn := "host=localhost port=5432 user=querybase password=querybase dbname=querybase sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
		return
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Create a test table
	db.Exec("CREATE TABLE IF NOT EXISTS test_table_exists (id INT)")
	defer db.Exec("DROP TABLE IF EXISTS test_table_exists")

	dataSource := &models.DataSource{
		Type: models.DataSourceTypePostgreSQL,
	}

	tests := []struct {
		name         string
		tableName    string
		expectExists bool
	}{
		{
			name:         "Existing table",
			tableName:    "test_table_exists",
			expectExists: true,
		},
		{
			name:         "Non-existing table",
			tableName:    "nonexistent_table_xyz",
			expectExists: false,
		},
		{
			name:         "Table with different case",
			tableName:    "TEST_TABLE_EXISTS",
			expectExists: true, // Should find it (case-insensitive)
		},
	}

	queryService := &QueryService{} // No initialization needed for tableExists

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists := queryService.tableExists(db, dataSource, tt.tableName)
			if exists != tt.expectExists {
				t.Errorf("tableExists(%s) = %v, want %v", tt.tableName, exists, tt.expectExists)
			}
		})
	}
}

// TestValidateQuerySchema tests schema validation (requires database)
func TestValidateQuerySchema(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	// Try to connect to test database
	dsn := "host=localhost port=5432 user=querybase password=querybase dbname=querybase sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
		return
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Create test tables
	db.Exec("CREATE TABLE IF NOT EXISTS test_users (id INT PRIMARY KEY, name VARCHAR(100))")
	db.Exec("CREATE TABLE IF NOT EXISTS test_orders (id INT PRIMARY KEY, user_id INT)")
	defer db.Exec("DROP TABLE IF EXISTS test_users")
	defer db.Exec("DROP TABLE IF EXISTS test_orders")

	queryService := &QueryService{}
	dataSource := &models.DataSource{
		Type:     models.DataSourceTypePostgreSQL,
		Host:     "localhost",
		Port:     5432,
		DatabaseName: "querybase",
		Username: "querybase",
	}

	tests := []struct {
		name        string
		sql         string
		expectError  bool
		errorMsg    string
	}{
		{
			name:       "Valid query with existing tables",
			sql:        "SELECT * FROM test_users",
			expectError: false,
		},
		{
			name:       "Valid query with JOIN",
			sql:        "SELECT * FROM test_users JOIN test_orders ON test_users.id = test_orders.user_id",
			expectError: false,
		},
		{
			name:       "Invalid query with non-existing table",
			sql:        "SELECT * FROM nonexistent_table",
			expectError: true,
			errorMsg:   "does not exist",
		},
		{
			name:       "Valid INSERT",
			sql:        "INSERT INTO test_users VALUES (1, 'John')",
			expectError: false,
		},
		{
			name:       "Valid UPDATE",
			sql:        "UPDATE test_users SET name = 'Jane'",
			expectError: false,
		},
		{
			name:       "Valid DELETE",
			sql:        "DELETE FROM test_users WHERE id = 1",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := queryService.ValidateQuerySchema(ctx, tt.sql, dataSource)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateQuerySchema() expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateQuerySchema() error = %v, want error containing %v", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateQuerySchema() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestDetectOperationTypeEdgeCases tests edge cases for operation detection
func TestDetectOperationTypeEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		sql          string
		expectedType models.OperationType
	}{
		{
			name:         "SELECT with newline",
			sql:         "SELECT *\nFROM\nusers",
			expectedType: models.OperationSelect,
		},
		{
			name:         "SELECT with tabs",
			sql:         "SELECT\t*\tFROM\tusers",
			expectedType: models.OperationSelect,
		},
		{
			name:         "SELECT with comments before",
			sql:         "-- comment\nSELECT * FROM users",
			expectedType: models.OperationSelect,
		},
		{
			name:         "INSERT with lowercase",
			sql:         "insert into users values (1)",
			expectedType: models.OperationInsert,
		},
		{
			name:         "UPDATE with newlines",
			sql:         "UPDATE users\nSET name = 'Jane'\nWHERE id = 1",
			expectedType: models.OperationUpdate,
		},
		{
			name:         "Complex SELECT with subqueries",
			sql:         "SELECT * FROM (SELECT * FROM users) AS u",
			expectedType: models.OperationSelect,
		},
		{
			name:         "Transaction BEGIN",
			sql:         "BEGIN TRANSACTION",
			expectedType: models.OperationUpdate,
		},
		{
			name:         "Transaction COMMIT",
			sql:         "COMMIT",
			expectedType: models.OperationUpdate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectOperationType(tt.sql)
			if result != tt.expectedType {
				t.Errorf("DetectOperationType() = %v, want %v", result, tt.expectedType)
			}
		})
	}
}

// TestConvertDeleteToSelect tests the DELETE to SELECT conversion
func TestConvertDeleteToSelect(t *testing.T) {
	tests := []struct {
		name     string
		deleteSQL string
		expectedSelectSQL string
	}{
		{
			name:     "Simple DELETE",
			deleteSQL: "DELETE FROM users WHERE id = 1",
			expectedSelectSQL: "SELECT * FROM users WHERE id = 1",
		},
		{
			name:     "DELETE with multiple conditions",
			deleteSQL: "DELETE FROM users WHERE id = 1 AND status = 'active'",
			expectedSelectSQL: "SELECT * FROM users WHERE id = 1 AND status = 'active'",
		},
		{
			name:     "DELETE with complex WHERE",
			deleteSQL: "DELETE FROM orders WHERE created_at < '2025-01-01' AND status IN ('cancelled', 'expired')",
			expectedSelectSQL: "SELECT * FROM orders WHERE created_at < '2025-01-01' AND status IN ('cancelled', 'expired')",
		},
		{
			name:     "DELETE with subquery",
			deleteSQL: "DELETE FROM users WHERE id IN (SELECT user_id FROM inactive_users)",
			expectedSelectSQL: "SELECT * FROM users WHERE id IN (SELECT user_id FROM inactive_users)",
		},
		{
			name:     "DELETE without WHERE",
			deleteSQL: "DELETE FROM users",
			expectedSelectSQL: "SELECT * FROM users",
		},
		{
			name:     "DELETE with JOIN",
			deleteSQL: "DELETE FROM users USING orders WHERE users.id = orders.user_id AND orders.status = 'cancelled'",
			expectedSelectSQL: "SELECT * FROM users USING orders WHERE users.id = orders.user_id AND orders.status = 'cancelled'",
		},
		{
			name:     "DELETE with lowercase",
			deleteSQL: "delete from users where id = 1",
			expectedSelectSQL: "SELECT * FROM users where id = 1",
		},
		{
			name:     "DELETE with extra whitespace",
			deleteSQL: "  DELETE   FROM  users  WHERE  id = 1  ",
			expectedSelectSQL: "SELECT * FROM users WHERE id = 1",
		},
		{
			name:     "DELETE with comments",
			deleteSQL: "-- Remove inactive users\nDELETE FROM users WHERE status = 'inactive'",
			expectedSelectSQL: "SELECT * FROM users WHERE status = 'inactive'",
		},
		{
			name:     "DELETE with schema-qualified table",
			deleteSQL: "DELETE FROM public.users WHERE id = 1",
			expectedSelectSQL: "SELECT * FROM public.users WHERE id = 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertDeleteToSelect(tt.deleteSQL)
			if result != tt.expectedSelectSQL {
				t.Errorf("convertDeleteToSelect() = %q, want %q", result, tt.expectedSelectSQL)
			}
		})
	}
}

// TestExplainQuery_Integration tests EXPLAIN query (requires database)
func TestExplainQuery_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	// Try to connect to test database
	dsn := "host=localhost port=5432 user=querybase password=querybase dbname=querybase sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
		return
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Create test table
	db.Exec("CREATE TABLE IF NOT EXISTS test_explain (id INT PRIMARY KEY, name VARCHAR(100))")
	db.Exec("INSERT INTO test_explain VALUES (1, 'Alice'), (2, 'Bob')")
	defer db.Exec("DROP TABLE IF EXISTS test_explain")

	dataSource := &models.DataSource{
		Type:     models.DataSourceTypePostgreSQL,
		Host:     "localhost",
		Port:     5432,
		DatabaseName: "querybase",
		Username: "querybase",
	}

	queryService := &QueryService{}

	tests := []struct {
		name        string
		query       string
		analyze     bool
		expectError bool
	}{
		{
			name:        "EXPLAIN SELECT",
			query:       "SELECT * FROM test_explain WHERE id = 1",
			analyze:     false,
			expectError: false,
		},
		{
			name:        "EXPLAIN ANALYZE SELECT",
			query:       "SELECT * FROM test_explain",
			analyze:     true,
			expectError: false,
		},
		{
			name:        "EXPLAIN with JOIN",
			query:       "SELECT * FROM test_explain t1 JOIN test_explain t2 ON t1.id = t2.id",
			analyze:     false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := queryService.ExplainQuery(ctx, tt.query, dataSource, tt.analyze)

			if tt.expectError {
				if err == nil {
					t.Errorf("ExplainQuery() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ExplainQuery() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("ExplainQuery() returned nil result")
				return
			}

			if len(result.Plan) == 0 {
				t.Errorf("ExplainQuery() returned empty plan")
			}

			if result.RawOutput == "" {
				t.Errorf("ExplainQuery() returned empty raw output")
			}
		})
	}
}

// TestDryRunDelete_Integration tests dry run DELETE (requires database)
func TestDryRunDelete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	// Try to connect to test database
	dsn := "host=localhost port=5432 user=querybase password=querybase dbname=querybase sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
		return
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Create test table with data
	db.Exec("CREATE TABLE IF NOT EXISTS test_dryrun (id INT PRIMARY KEY, name VARCHAR(100), status VARCHAR(50))")
	db.Exec("INSERT INTO test_dryrun VALUES (1, 'Alice', 'active'), (2, 'Bob', 'inactive'), (3, 'Charlie', 'active')")
	defer db.Exec("DROP TABLE IF EXISTS test_dryrun")

	dataSource := &models.DataSource{
		Type:     models.DataSourceTypePostgreSQL,
		Host:     "localhost",
		Port:     5432,
		DatabaseName: "querybase",
		Username: "querybase",
	}

	queryService := &QueryService{}

	tests := []struct {
		name               string
		deleteQuery        string
		expectedAffectedRows int
		expectError        bool
	}{
		{
			name:               "DELETE single row",
			deleteQuery:        "DELETE FROM test_dryrun WHERE id = 1",
			expectedAffectedRows: 1,
			expectError:        false,
		},
		{
			name:               "DELETE multiple rows",
			deleteQuery:        "DELETE FROM test_dryrun WHERE status = 'active'",
			expectedAffectedRows: 2,
			expectError:        false,
		},
		{
			name:               "DELETE no rows",
			deleteQuery:        "DELETE FROM test_dryrun WHERE id = 999",
			expectedAffectedRows: 0,
			expectError:        false,
		},
		{
			name:               "DELETE all rows",
			deleteQuery:        "DELETE FROM test_dryrun",
			expectedAffectedRows: 3,
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := queryService.DryRunDelete(ctx, tt.deleteQuery, dataSource)

			if tt.expectError {
				if err == nil {
					t.Errorf("DryRunDelete() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("DryRunDelete() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("DryRunDelete() returned nil result")
				return
			}

			if result.AffectedRows != tt.expectedAffectedRows {
				t.Errorf("DryRunDelete() affectedRows = %d, want %d", result.AffectedRows, tt.expectedAffectedRows)
			}

			if !strings.Contains(result.Query, "SELECT") {
				t.Errorf("DryRunDelete() query should be SELECT, got: %s", result.Query)
			}

			if len(result.Rows) != tt.expectedAffectedRows {
				t.Errorf("DryRunDelete() returned %d rows, want %d", len(result.Rows), tt.expectedAffectedRows)
			}
		})
	}
}

// TestDryRunDelete_NonDeleteQuery tests that non-DELETE queries are rejected
func TestDryRunDelete_NonDeleteQuery(t *testing.T) {
	queryService := &QueryService{}
	dataSource := &models.DataSource{
		Type: models.DataSourceTypePostgreSQL,
	}

	tests := []struct {
		name        string
		query       string
		expectError string
	}{
		{
			name:        "SELECT query",
			query:       "SELECT * FROM users",
			expectError: "dry run is only supported for DELETE queries",
		},
		{
			name:        "UPDATE query",
			query:       "UPDATE users SET name = 'test'",
			expectError: "dry run is only supported for DELETE queries",
		},
		{
			name:        "INSERT query",
			query:       "INSERT INTO users VALUES (1, 'test')",
			expectError: "dry run is only supported for DELETE queries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := queryService.DryRunDelete(ctx, tt.query, dataSource)

			if err == nil {
				t.Errorf("DryRunDelete() expected error but got none")
				return
			}

			if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("DryRunDelete() error = %v, want error containing %v", err, tt.expectError)
			}
		})
	}
}

// TestSortRows tests the sorting functionality
func TestSortRows(t *testing.T) {
	queryService := &QueryService{}

	tests := []struct {
		name           string
		rows           []map[string]interface{}
		sortColumn     string
		sortDirection  string
		expectedFirst  interface{}
		expectedLast   interface{}
	}{
		{
			name: "Sort by int column ascending",
			rows: []map[string]interface{}{
				{"id": 3, "name": "Charlie"},
				{"id": 1, "name": "Alice"},
				{"id": 2, "name": "Bob"},
			},
			sortColumn:     "id",
			sortDirection:  "asc",
			expectedFirst:  1,
			expectedLast:   3,
		},
		{
			name: "Sort by int column descending",
			rows: []map[string]interface{}{
				{"id": 1, "name": "Alice"},
				{"id": 3, "name": "Charlie"},
				{"id": 2, "name": "Bob"},
			},
			sortColumn:     "id",
			sortDirection:  "desc",
			expectedFirst:  3,
			expectedLast:   1,
		},
		{
			name: "Sort by string column ascending",
			rows: []map[string]interface{}{
				{"name": "Charlie", "id": 3},
				{"name": "Alice", "id": 1},
				{"name": "Bob", "id": 2},
			},
			sortColumn:     "name",
			sortDirection:  "asc",
			expectedFirst:  "Alice",
			expectedLast:   "Charlie",
		},
		{
			name: "Sort by string column descending",
			rows: []map[string]interface{}{
				{"name": "Alice", "id": 1},
				{"name": "Charlie", "id": 3},
				{"name": "Bob", "id": 2},
			},
			sortColumn:     "name",
			sortDirection:  "desc",
			expectedFirst:  "Charlie",
			expectedLast:   "Alice",
		},
		{
			name: "Sort with nil values",
			rows: []map[string]interface{}{
				{"id": nil, "name": "Unknown"},
				{"id": 2, "name": "Bob"},
				{"id": 1, "name": "Alice"},
			},
			sortColumn:     "id",
			sortDirection:  "asc",
			expectedFirst:  nil,
			expectedLast:   2,
		},
		{
			name: "Sort by float column",
			rows: []map[string]interface{}{
				{"price": 19.99, "name": "Item A"},
				{"price": 9.99, "name": "Item B"},
				{"price": 29.99, "name": "Item C"},
			},
			sortColumn:     "price",
			sortDirection:  "asc",
			expectedFirst:  9.99,
			expectedLast:   29.99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := queryService.sortRows(tt.rows, tt.sortColumn, tt.sortDirection)

			if len(result) != len(tt.rows) {
				t.Errorf("sortRows() returned %d rows, want %d", len(result), len(tt.rows))
				return
			}

			firstValue := result[0][tt.sortColumn]
			lastValue := result[len(result)-1][tt.sortColumn]

			// For nil values, we need special comparison
			if tt.expectedFirst == nil {
				if firstValue != nil {
					t.Errorf("sortRows() first value = %v, want nil", firstValue)
				}
			} else {
				if firstValue != tt.expectedFirst {
					t.Errorf("sortRows() first value = %v, want %v", firstValue, tt.expectedFirst)
				}
			}

			if tt.expectedLast == nil {
				if lastValue != nil {
					t.Errorf("sortRows() last value = %v, want nil", lastValue)
				}
			} else {
				if lastValue != tt.expectedLast {
					t.Errorf("sortRows() last value = %v, want %v", lastValue, tt.expectedLast)
				}
			}
		})
	}
}

// TestCompareValues tests the value comparison function
func TestCompareValues(t *testing.T) {
	queryService := &QueryService{}

	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected int // -1 if a < b, 0 if a == b, 1 if a > b
	}{
		{"int less", 1, 2, -1},
		{"int equal", 2, 2, 0},
		{"int greater", 3, 2, 1},
		{"float less", 1.5, 2.5, -1},
		{"float equal", 2.5, 2.5, 0},
		{"float greater", 3.5, 2.5, 1},
		{"string less", "Alice", "Bob", -1},
		{"string equal", "Bob", "Bob", 0},
		{"string greater", "Charlie", "Bob", 1},
		{"nil first", nil, 1, -1},
		{"nil second", 1, nil, 1},
		{"nil both", nil, nil, 0},
		{"mixed int and float", 2, 2.0, 0},
		{"string number less", "10", "2", 1}, // String comparison: "10" > "2" lexicographically
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := queryService.compareValues(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("compareValues(%v, %v) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestToFloat64 tests the float64 conversion function
func TestToFloat64(t *testing.T) {
	queryService := &QueryService{}

	tests := []struct {
		name     string
		value    interface{}
		expected float64
		ok       bool
	}{
		{"int", 42, 42.0, true},
		{"int32", int32(32), 32.0, true},
		{"int64", int64(64), 64.0, true},
		{"float32", float32(3.0), 3.0, true}, // Use exact value to avoid precision issues
		{"float64", 2.718, 2.718, true},
		{"numeric string", "123.45", 123.45, true},
		{"non-numeric string", "hello", 0, false},
		{"nil", nil, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := queryService.toFloat64(tt.value)
			if ok != tt.ok {
				t.Errorf("toFloat64(%v) ok = %v, want %v", tt.value, ok, tt.ok)
				return
			}
			if ok && result != tt.expected {
				t.Errorf("toFloat64(%v) = %f, want %f", tt.value, result, tt.expected)
			}
		})
	}
}

// TestExportToCSV tests the CSV export functionality
func TestExportToCSV(t *testing.T) {
	queryService := &QueryService{}

	tests := []struct {
		name     string
		rows     []map[string]interface{}
		columns  []string
		expected []string // Strings that should be present in the output
	}{
		{
			name: "Simple CSV export",
			rows: []map[string]interface{}{
				{"id": 1, "name": "Alice", "age": 30},
				{"id": 2, "name": "Bob", "age": 25},
			},
			columns:  []string{"id", "name", "age"},
			expected: []string{"\"id\",\"name\",\"age\"", "\"1\",\"Alice\",\"30\"", "\"2\",\"Bob\",\"25\""},
		},
		{
			name: "CSV with special characters",
			rows: []map[string]interface{}{
				{"name": "John \"The Boss\"", "quote": "He said \"Hello\""},
			},
			columns:  []string{"name", "quote"},
			expected: []string{"\"name\",\"quote\"", "\"John \"\"The Boss\"\"\"", "\"He said \"\"Hello\"\"\""},
		},
		{
			name: "CSV with nil values",
			rows: []map[string]interface{}{
				{"id": 1, "name": nil, "email": "test@example.com"},
			},
			columns:  []string{"id", "name", "email"},
			expected: []string{"\"id\",\"name\",\"email\"", "\"1\",\"\",\"test@example.com\""},
		},
		{
			name:    "Empty result set",
			rows:    []map[string]interface{}{},
			columns: []string{"id", "name"},
			expected: []string{"\"id\",\"name\""}, // Only header row
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := queryService.exportToCSV(tt.rows, tt.columns)
			if err != nil {
				t.Errorf("exportToCSV() unexpected error: %v", err)
				return
			}

			resultStr := string(result)

			for _, expected := range tt.expected {
				if !strings.Contains(resultStr, expected) {
					t.Errorf("exportToCSV() output does not contain expected string %q\nGot:\n%s", expected, resultStr)
				}
			}
		})
	}
}

// TestExportToJSON tests the JSON export functionality
func TestExportToJSON(t *testing.T) {
	queryService := &QueryService{}

	tests := []struct {
		name     string
		rows     []map[string]interface{}
		columns  []string
		expected []string // Strings that should be present in the output
	}{
		{
			name: "Simple JSON export",
			rows: []map[string]interface{}{
				{"id": 1, "name": "Alice"},
				{"id": 2, "name": "Bob"},
			},
			columns:  []string{"id", "name"},
			expected: []string{"\"columns\"", "\"row_count\"", "\"data\"", "\"id\"", "\"name\""},
		},
		{
			name: "JSON with various data types",
			rows: []map[string]interface{}{
				{"id": 1, "name": "Test", "score": 95.5, "active": true},
			},
			columns:  []string{"id", "name", "score", "active"},
			expected: []string{"\"score\": 95.5", "\"active\": true"},
		},
		{
			name:     "Empty result set",
			rows:     []map[string]interface{}{},
			columns:  []string{"id"},
			expected: []string{"\"row_count\": 0", "\"data\": []"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := queryService.exportToJSON(tt.rows, tt.columns)
			if err != nil {
				t.Errorf("exportToJSON() unexpected error: %v", err)
				return
			}

			resultStr := string(result)

			for _, expected := range tt.expected {
				if !strings.Contains(resultStr, expected) {
					t.Errorf("exportToJSON() output does not contain expected string %q\nGot:\n%s", expected, resultStr)
				}
			}
		})
	}
}

