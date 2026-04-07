package service

import (
	"testing"

	"github.com/yourorg/querybase/internal/models"
)

// TestDetectOperationType tests the operation type detection
func TestDetectOperationType(t *testing.T) {
	tests := []struct {
		name         string
		sql          string
		expectedType models.OperationType
	}{
		{
			name:         "SELECT query",
			sql:          "SELECT * FROM users",
			expectedType: models.OperationSelect,
		},
		{
			name:         "INSERT query",
			sql:          "INSERT INTO users (name) VALUES ('John')",
			expectedType: models.OperationInsert,
		},
		{
			name:         "UPDATE query",
			sql:          "UPDATE users SET name = 'Jane'",
			expectedType: models.OperationUpdate,
		},
		{
			name:         "DELETE query",
			sql:          "DELETE FROM users WHERE id = 1",
			expectedType: models.OperationDelete,
		},
		{
			name:         "CREATE TABLE query",
			sql:          "CREATE TABLE test (id INT)",
			expectedType: models.OperationCreateTable,
		},
		{
			name:         "DROP TABLE query",
			sql:          "DROP TABLE old_table",
			expectedType: models.OperationDropTable,
		},
		{
			name:         "ALTER TABLE query",
			sql:          "ALTER TABLE users ADD COLUMN age INT",
			expectedType: models.OperationAlterTable,
		},
		{
			name:         "SELECT with lowercase",
			sql:          "select * from users",
			expectedType: models.OperationSelect,
		},
		{
			name:         "SELECT with extra whitespace",
			sql:          "  SELECT   *  FROM  users  ",
			expectedType: models.OperationSelect,
		},
		{
			name:         "GRANT query",
			sql:          "GRANT SELECT ON users TO role",
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

// TestRequiresApproval tests the approval requirement check
func TestRequiresApproval(t *testing.T) {
	tests := []struct {
		name          string
		operationType models.OperationType
		expected      bool
	}{
		{
			name:          "SELECT requires no approval",
			operationType: models.OperationSelect,
			expected:      false,
		},
		{
			name:          "INSERT requires approval",
			operationType: models.OperationInsert,
			expected:      true,
		},
		{
			name:          "UPDATE requires approval",
			operationType: models.OperationUpdate,
			expected:      true,
		},
		{
			name:          "DELETE requires approval",
			operationType: models.OperationDelete,
			expected:      true,
		},
		{
			name:          "CREATE TABLE requires approval",
			operationType: models.OperationCreateTable,
			expected:      true,
		},
		{
			name:          "DROP TABLE requires approval",
			operationType: models.OperationDropTable,
			expected:      true,
		},
		{
			name:          "ALTER TABLE requires approval",
			operationType: models.OperationAlterTable,
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RequiresApproval(tt.operationType)
			if result != tt.expected {
				t.Errorf("RequiresApproval() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestValidateSQL tests SQL syntax validation
func TestValidateSQL(t *testing.T) {
	tests := []struct {
		name        string
		sql         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid SELECT query",
			sql:         "SELECT * FROM users WHERE id = 1",
			expectError: false,
		},
		{
			name:        "Valid INSERT query",
			sql:         "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')",
			expectError: false,
		},
		{
			name:        "Valid UPDATE query",
			sql:         "UPDATE users SET name = 'Jane' WHERE id = 1",
			expectError: false,
		},
		{
			name:        "Valid DELETE query",
			sql:         "DELETE FROM users WHERE id = 1",
			expectError: false,
		},
		{
			name:        "Valid CREATE TABLE query",
			sql:         "CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(100))",
			expectError: false,
		},
		{
			name:        "Valid DROP TABLE query",
			sql:         "DROP TABLE IF EXISTS old_table",
			expectError: false,
		},
		{
			name:        "Empty query",
			sql:         "",
			expectError: true,
			errorMsg:    "SQL query cannot be empty",
		},
		{
			name:        "Whitespace only query",
			sql:         "   \t\n   ",
			expectError: true,
			errorMsg:    "SQL query cannot be empty",
		},
		{
			name:        "Unbalanced parentheses - missing closing",
			sql:         "SELECT * FROM users WHERE (id = 1",
			expectError: true,
			errorMsg:    "unbalanced parentheses",
		},
		{
			name:        "Unbalanced parentheses - missing opening",
			sql:         "SELECT * FROM users WHERE id = 1)",
			expectError: true,
			errorMsg:    "unbalanced parentheses",
		},
		{
			name:        "Unterminated string literal",
			sql:         "SELECT * FROM users WHERE name = 'John",
			expectError: true,
			errorMsg:    "unterminated string literal",
		},
		{
			name:        "INSERT without VALUES or SELECT",
			sql:         "INSERT INTO users",
			expectError: true,
			errorMsg:    "must include VALUES, SELECT, or SET clause",
		},
		{
			name:        "UPDATE without SET",
			sql:         "UPDATE users WHERE id = 1",
			expectError: true,
			errorMsg:    "must include SET clause",
		},
		{
			name:        "DELETE without FROM (valid in Postgres/SQL Server)",
			sql:         "DELETE users WHERE id = 1",
			expectError: false,
		},
		{
			name:        "CREATE TABLE without name",
			sql:         "CREATE TABLE",
			expectError: true,
			errorMsg:    "must specify table name",
		},
		{
			name:        "DROP TABLE without name",
			sql:         "DROP TABLE",
			expectError: true,
			errorMsg:    "must specify table name",
		},
		{
			name:        "Valid SELECT without FROM (simple expression)",
			sql:         "SELECT 1",
			expectError: false,
		},
		{
			name:        "Valid SELECT NOW()",
			sql:         "SELECT NOW()",
			expectError: false,
		},
		{
			name:        "Valid SELECT with escaped quote",
			sql:         "SELECT * FROM users WHERE name = 'O''Reilly'",
			expectError: false,
		},
		{
			name:        "Valid INSERT with SELECT",
			sql:         "INSERT INTO users_archive SELECT * FROM users",
			expectError: false,
		},
		{
			name:        "Valid INSERT with SET",
			sql:         "INSERT INTO users SET id = 1, name = 'John'",
			expectError: false,
		},
		// Multi-line queries — SET / VALUES / WHERE must be found even when
		// the keyword lands on a new line (regression for newline parsing bug).
		{
			name: "Valid multi-line UPDATE — SET on new line",
			sql: `UPDATE users
SET name = 'Jane', email = 'jane@example.com', updated_at = '2025-01-01 00:00:00'
WHERE id IN (1,2,3)`,
			expectError: false,
		},
		{
			name: "Valid multi-line INSERT — VALUES on new line",
			sql: `INSERT INTO customers (first_name, last_name, email)
VALUES
('John', 'Doe', 'john@example.com')`,
			expectError: false,
		},
		{
			name: "Valid multi-line SELECT — FROM on new line",
			sql: `SELECT id, first_name, last_name
FROM customers
WHERE is_active = true`,
			expectError: false,
		},
		{
			name: "Invalid multi-line UPDATE — genuinely missing SET",
			sql: `UPDATE customers
WHERE id = 1`,
			expectError: true,
			errorMsg:    "SET clause",
		},
		// UPDATE SET clause validation - AND instead of comma
		{
			name:        "Invalid UPDATE with AND instead of comma",
			sql:         "UPDATE users SET name = 'John' AND email = 'john@example.com' WHERE id = 1",
			expectError: true,
			errorMsg:    "use comma (,) not AND",
		},
		{
			name:        "Invalid UPDATE with AND instead of comma - multiple columns",
			sql:         "UPDATE products SET price = 10.99 AND stock = 100 AND updated_at = '2024-01-01' WHERE id = 5",
			expectError: true,
			errorMsg:    "use comma (,) not AND",
		},
		{
			name: "Valid UPDATE with comma between columns",
			sql:  "UPDATE users SET name = 'John', email = 'john@example.com' WHERE id = 1",
			expectError: false,
		},
		{
			name: "Valid UPDATE with multiple comma-separated columns",
			sql:  "UPDATE products SET price = 10.99, stock = 100, updated_at = '2024-01-01' WHERE id = 5",
			expectError: false,
		},
		{
			name: "Valid UPDATE with AND in WHERE clause only",
			sql:  "UPDATE users SET name = 'John' WHERE id = 1 AND status = 'active'",
			expectError: false,
		},
		{
			name: "Valid UPDATE with complex WHERE containing AND",
			sql:  "UPDATE orders SET status = 'shipped' WHERE id = 1 AND customer_id = 5 AND created_at > '2024-01-01'",
			expectError: false,
		},
		{
			name: "Valid multi-line UPDATE with comma-separated columns",
			sql: `UPDATE ecommerce_product_merchants 
SET created_at='2026-04-05 12:12:12', 
    updated_at='2026-04-05 12:12:12' 
WHERE product_merchant_id=19266006`,
			expectError: false,
		},
		{
			name: "Invalid multi-line UPDATE with AND instead of comma",
			sql: `UPDATE ecommerce_product_merchants 
SET created_at='2026-04-05 12:12:12' AND updated_at='2026-04-05 12:12:12' 
WHERE product_merchant_id=19266006`,
			expectError: true,
			errorMsg:    "use comma (,) not AND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSQL(tt.sql)
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateSQL() expected error but got none")
					return
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					// Check if error message contains expected text
					if !containsString(err.Error(), tt.errorMsg) {
						t.Errorf("ValidateSQL() error = %v, want error containing %v", err, tt.errorMsg)
					}
				}
			} else {
				if err != nil {
					t.Errorf("ValidateSQL() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestSanitizeSQL tests SQL sanitization (removing comments)
func TestSanitizeSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove single-line comment",
			input:    "SELECT * FROM users -- this is a comment\nWHERE id = 1",
			expected: "SELECT * FROM users WHERE id = 1",
		},
		{
			name:     "Remove multi-line comment",
			input:    "SELECT /* comment */ * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "Remove multiple comments",
			input:    "SELECT * -- comment1\nFROM /* comment2 */ users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "Collapse whitespace",
			input:    "SELECT   *    FROM   users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "No comments",
			input:    "SELECT * FROM users",
			expected: "SELECT * FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeSQL(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeSQL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
