package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourorg/querybase/internal/models"
)

func TestParseMultipleQueries_SETVariables(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedStmts []string
		expectedTypes []models.OperationType
	}{
		{
			name: "SET variable declarations",
			query: `SET @wallet_id = 23916;
SET @execution_date = '2026-03-03 15:15:15';
SET @updated_by = 5285816;`,
			expectedCount: 3,
			expectedStmts: []string{
				"SET @wallet_id = 23916",
				"SET @execution_date = '2026-03-03 15:15:15'",
				"SET @updated_by = 5285816",
			},
			expectedTypes: []models.OperationType{
				models.OperationSet,
				models.OperationSet,
				models.OperationSet,
			},
		},
		{
			name: "SET variables with UPDATE using variables",
			query: `SET @wallet_id = 23916;
SET @execution_date = '2026-03-03 15:15:15';
SET @updated_by = 5285816;
UPDATE wallet_trxes SET amount = 0 WHERE wallet_id = @wallet_id;
UPDATE wallets SET total_balance = 0 WHERE id = @wallet_id;`,
			expectedCount: 5,
			expectedStmts: []string{
				"SET @wallet_id = 23916",
				"SET @execution_date = '2026-03-03 15:15:15'",
				"SET @updated_by = 5285816",
				"UPDATE wallet_trxes SET amount = 0 WHERE wallet_id = @wallet_id",
				"UPDATE wallets SET total_balance = 0 WHERE id = @wallet_id",
			},
			expectedTypes: []models.OperationType{
				models.OperationSet,
				models.OperationSet,
				models.OperationSet,
				models.OperationUpdate,
				models.OperationUpdate,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMultipleQueries(tt.query)
			assert.Len(t, result.Statements, tt.expectedCount, "Expected %d statements, got %d", tt.expectedCount, len(result.Statements))
			assert.Empty(t, result.Errors, "Expected no errors")

			for i, expectedStmt := range tt.expectedStmts {
				if i < len(result.Statements) {
					assert.Equal(t, expectedStmt, result.Statements[i].QueryText, "Statement %d mismatch", i)
					assert.Equal(t, i, result.Statements[i].Sequence, "Statement %d sequence mismatch", i)
					assert.Equal(t, tt.expectedTypes[i], result.Statements[i].OperationType, "Statement %d operation type mismatch", i)
				}
			}
		})
	}
}

func TestParseMultipleQueries_WithComments(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedStmts []string
	}{
		{
			name: "Line comments before statements",
			query: `-- Set variable
SET @wallet_id = 23916;
-- Update semua nominal di tabel wallet_trxes menjadi 0
UPDATE wallet_trxes SET amount = 0 WHERE wallet_id = @wallet_id;`,
			expectedCount: 2,
			expectedStmts: []string{
				"-- Set variable\nSET @wallet_id = 23916",
				"-- Update semua nominal di tabel wallet_trxes menjadi 0\nUPDATE wallet_trxes SET amount = 0 WHERE wallet_id = @wallet_id",
			},
		},
		{
			name: "Line comments inline",
			query: `SELECT * FROM users; -- Get all users
UPDATE users SET active = 1; -- Activate users`,
			expectedCount: 3,
			expectedStmts: []string{
				"SELECT * FROM users",
				"-- Get all users\nUPDATE users SET active = 1",
				"-- Activate users",
			},
		},
		{
			name: "Block comments",
			query: `/* Start transaction block */
UPDATE wallet_trxes SET amount = 0 WHERE wallet_id = 1;
/* End transaction block */
SELECT * FROM wallets;`,
			expectedCount: 2,
			expectedStmts: []string{
				"/* Start transaction block */\nUPDATE wallet_trxes SET amount = 0 WHERE wallet_id = 1",
				"/* End transaction block */\nSELECT * FROM wallets",
			},
		},
		{
			name: "Mixed comments with SET variables",
			query: `-- Initialize variables
SET @id = 100;
/* Multi-line
   comment */
SET @name = 'test';
-- Final query
UPDATE table1 SET name = @name WHERE id = @id;`,
			expectedCount: 3,
			expectedStmts: []string{
				"-- Initialize variables\nSET @id = 100",
				"/* Multi-line\n   comment */\nSET @name = 'test'",
				"-- Final query\nUPDATE table1 SET name = @name WHERE id = @id",
			},
		},
		{
			name: "Indonesian language comments",
			query: `-- Update semua nominal di tabel wallet_trxes menjadi 0
UPDATE wallet_trxes SET amount = 0;
-- Update kolom saldo di tabel wallets menjadi 0
UPDATE wallets SET total_balance = 0;`,
			expectedCount: 2,
			expectedStmts: []string{
				"-- Update semua nominal di tabel wallet_trxes menjadi 0\nUPDATE wallet_trxes SET amount = 0",
				"-- Update kolom saldo di tabel wallets menjadi 0\nUPDATE wallets SET total_balance = 0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMultipleQueries(tt.query)
			assert.Len(t, result.Statements, tt.expectedCount, "Expected %d statements, got %d", tt.expectedCount, len(result.Statements))
			assert.Empty(t, result.Errors, "Expected no errors")

			for i, expectedStmt := range tt.expectedStmts {
				if i < len(result.Statements) {
					assert.Equal(t, expectedStmt, result.Statements[i].QueryText, "Statement %d mismatch", i)
				}
			}
		})
	}
}

func TestParseMultipleQueries_NewlinesAndWhitespace(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedStmts []string
	}{
		{
			name: "Single query with newlines",
			query: `SELECT 
    id,
    name,
    email
FROM 
    users
WHERE 
    active = 1;`,
			expectedCount: 1,
			expectedStmts: []string{
				"SELECT \n    id,\n    name,\n    email\nFROM \n    users\nWHERE \n    active = 1",
			},
		},
		{
			name: "Multiple queries with excessive whitespace",
			query: `

SELECT * FROM table1;


UPDATE table2 SET col = 1;

`,
			expectedCount: 2,
			expectedStmts: []string{
				"SELECT * FROM table1",
				"UPDATE table2 SET col = 1",
			},
		},
		{
			name: "Complex query with newlines and indentation",
			query: `SET @wallet_id = 23916;
UPDATE wallet_trxes
SET
  amount = 0,
  customer_amount = 0,
  currency_amount = 0,
  updated_by = @updated_by,
  updated_at = @execution_date
WHERE wallet_id = @wallet_id;
UPDATE wallets
SET
  total_balance = 0,
  customer_balance = 0,
  customer_amount = 0,
  updated_by = @updated_by,
  updated_at = @execution_date
WHERE id = @wallet_id;`,
			expectedCount: 3,
			expectedStmts: []string{
				"SET @wallet_id = 23916",
				"UPDATE wallet_trxes\nSET\n  amount = 0,\n  customer_amount = 0,\n  currency_amount = 0,\n  updated_by = @updated_by,\n  updated_at = @execution_date\nWHERE wallet_id = @wallet_id",
				"UPDATE wallets\nSET\n  total_balance = 0,\n  customer_balance = 0,\n  customer_amount = 0,\n  updated_by = @updated_by,\n  updated_at = @execution_date\nWHERE id = @wallet_id",
			},
		},
		{
			name:          "Tabs and spaces mixed",
			query:         "SELECT\t*\tFROM\tusers;\t\n\tUPDATE\tusers\tSET\tactive\t=\t1;",
			expectedCount: 2,
			expectedStmts: []string{
				"SELECT\t*\tFROM\tusers",
				"UPDATE\tusers\tSET\tactive\t=\t1",
			},
		},
		{
			name: "No trailing semicolon",
			query: `SELECT * FROM users
WHERE id = 1`,
			expectedCount: 1,
			expectedStmts: []string{
				"SELECT * FROM users\nWHERE id = 1",
			},
		},
		{
			name:          "Empty statements between semicolons",
			query:         "SELECT 1;; ; SELECT 2;",
			expectedCount: 2,
			expectedStmts: []string{
				"SELECT 1",
				"SELECT 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMultipleQueries(tt.query)
			assert.Len(t, result.Statements, tt.expectedCount, "Expected %d statements, got %d", tt.expectedCount, len(result.Statements))
			assert.Empty(t, result.Errors, "Expected no errors")

			for i, expectedStmt := range tt.expectedStmts {
				if i < len(result.Statements) {
					assert.Equal(t, expectedStmt, result.Statements[i].QueryText, "Statement %d mismatch", i)
				}
			}
		})
	}
}

func TestParseMultipleQuotes(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedStmts []string
	}{
		{
			name:          "Single quotes with semicolon",
			query:         "INSERT INTO test VALUES ('a;b'); SELECT 1;",
			expectedCount: 2,
			expectedStmts: []string{
				"INSERT INTO test VALUES ('a;b')",
				"SELECT 1",
			},
		},
		{
			name:          "Double quotes with semicolon",
			query:         `INSERT INTO test VALUES ("a;b"); SELECT 1;`,
			expectedCount: 2,
			expectedStmts: []string{
				`INSERT INTO test VALUES ("a;b")`,
				"SELECT 1",
			},
		},
		{
			name:          "Escaped quotes",
			query:         `INSERT INTO test VALUES ('it\'s; working'); SELECT 1;`,
			expectedCount: 2,
			expectedStmts: []string{
				`INSERT INTO test VALUES ('it\'s; working')`,
				"SELECT 1",
			},
		},
		{
			name:          "SET with string containing semicolon",
			query:         "SET @msg = 'Hello; World'; SELECT @msg;",
			expectedCount: 2,
			expectedStmts: []string{
				"SET @msg = 'Hello; World'",
				"SELECT @msg",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMultipleQueries(tt.query)
			assert.Len(t, result.Statements, tt.expectedCount, "Expected %d statements, got %d", tt.expectedCount, len(result.Statements))
			assert.Empty(t, result.Errors, "Expected no errors")

			for i, expectedStmt := range tt.expectedStmts {
				if i < len(result.Statements) {
					assert.Equal(t, expectedStmt, result.Statements[i].QueryText, "Statement %d mismatch", i)
				}
			}
		})
	}
}

func TestIsMultiQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "Single query",
			query:    "SELECT * FROM users",
			expected: false,
		},
		{
			name:     "Multiple queries",
			query:    "SELECT 1; SELECT 2",
			expected: true,
		},
		{
			name:     "SET and UPDATE",
			query:    "SET @x = 1; UPDATE t SET col = @x",
			expected: true,
		},
		{
			name:     "Empty string",
			query:    "",
			expected: false,
		},
		{
			name:     "Whitespace only",
			query:    "   \n\t  ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMultiQuery(tt.query)
			assert.Equal(t, tt.expected, result, "IsMultiQuery mismatch")
		})
	}
}

func TestValidateMultiQuery(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedErrors int
		errorContains  string
	}{
		{
			name:           "Valid multi-query",
			query:          "SELECT 1; SELECT 2",
			expectedErrors: 0,
		},
		{
			name:           "BEGIN not allowed",
			query:          "BEGIN; SELECT 1",
			expectedErrors: 1,
			errorContains:  "Transaction control",
		},
		{
			name:           "COMMIT not allowed",
			query:          "SELECT 1; COMMIT",
			expectedErrors: 1,
			errorContains:  "Transaction control",
		},
		{
			name:           "ROLLBACK not allowed",
			query:          "ROLLBACK",
			expectedErrors: 1,
			errorContains:  "Transaction control",
		},
		{
			name:           "START TRANSACTION not allowed",
			query:          "START TRANSACTION; SELECT 1",
			expectedErrors: 1,
			errorContains:  "Transaction control",
		},
		{
			name:           "Multiple transaction statements",
			query:          "BEGIN; COMMIT; ROLLBACK",
			expectedErrors: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateMultiQuery(tt.query)
			assert.Len(t, result.Errors, tt.expectedErrors, "Expected %d errors, got %d", tt.expectedErrors, len(result.Errors))

			if tt.errorContains != "" && len(result.Errors) > 0 {
				for _, err := range result.Errors {
					assert.Contains(t, err.Message, tt.errorContains, "Error message should contain %q", tt.errorContains)
				}
			}
		})
	}
}

func TestParseMultipleQueries_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedCount int
	}{
		{
			name:          "Only semicolons",
			query:         ";;;",
			expectedCount: 0,
		},
		{
			name:          "Comment only",
			query:         "-- Just a comment",
			expectedCount: 1,
		},
		{
			name:          "Block comment only",
			query:         "/* Just a block comment */",
			expectedCount: 1,
		},
		{
			name:          "MySQL-style backtick identifiers - backticks treated as regular chars so semicolon splits",
			query:         "SELECT `column;name` FROM table1; SELECT 1",
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMultipleQueries(tt.query)
			// For edge cases, we just check it doesn't panic
			assert.NotNil(t, result)
			if tt.expectedCount > 0 {
				assert.Len(t, result.Statements, tt.expectedCount, "Expected %d statements, got %d", tt.expectedCount, len(result.Statements))
			}
		})
	}
}

// TestParseMultipleQueries_VariousSQLPatterns_ParsesCorrectly tests parsing of various SQL patterns
func TestParseMultipleQueries_VariousSQLPatterns_ParsesCorrectly(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedStmts []string
		expectedTypes []models.OperationType
	}{
		{
			name:          "SELECT statements",
			query:         "SELECT * FROM users; SELECT id, name FROM products",
			expectedCount: 2,
			expectedStmts: []string{
				"SELECT * FROM users",
				"SELECT id, name FROM products",
			},
			expectedTypes: []models.OperationType{
				models.OperationSelect,
				models.OperationSelect,
			},
		},
		{
			name:          "INSERT statements",
			query:         "INSERT INTO users (name) VALUES ('John'); INSERT INTO logs (msg) VALUES ('created')",
			expectedCount: 2,
			expectedStmts: []string{
				"INSERT INTO users (name) VALUES ('John')",
				"INSERT INTO logs (msg) VALUES ('created')",
			},
			expectedTypes: []models.OperationType{
				models.OperationInsert,
				models.OperationInsert,
			},
		},
		{
			name:          "UPDATE statements",
			query:         "UPDATE users SET active = 1; UPDATE products SET stock = stock - 1",
			expectedCount: 2,
			expectedStmts: []string{
				"UPDATE users SET active = 1",
				"UPDATE products SET stock = stock - 1",
			},
			expectedTypes: []models.OperationType{
				models.OperationUpdate,
				models.OperationUpdate,
			},
		},
		{
			name:          "DELETE statements",
			query:         "DELETE FROM temp_logs; DELETE FROM cache WHERE expired = 1",
			expectedCount: 2,
			expectedStmts: []string{
				"DELETE FROM temp_logs",
				"DELETE FROM cache WHERE expired = 1",
			},
			expectedTypes: []models.OperationType{
				models.OperationDelete,
				models.OperationDelete,
			},
		},
		{
			name:          "Mixed operations",
			query:         "SELECT * FROM users; INSERT INTO logs (msg) VALUES ('query executed'); UPDATE stats SET count = count + 1",
			expectedCount: 3,
			expectedStmts: []string{
				"SELECT * FROM users",
				"INSERT INTO logs (msg) VALUES ('query executed')",
				"UPDATE stats SET count = count + 1",
			},
			expectedTypes: []models.OperationType{
				models.OperationSelect,
				models.OperationInsert,
				models.OperationUpdate,
			},
		},
		{
			name:          "SELECT with JOIN",
			query:         "SELECT u.*, o.total FROM users u JOIN orders o ON u.id = o.user_id; SELECT COUNT(*) FROM products",
			expectedCount: 2,
			expectedStmts: []string{
				"SELECT u.*, o.total FROM users u JOIN orders o ON u.id = o.user_id",
				"SELECT COUNT(*) FROM products",
			},
			expectedTypes: []models.OperationType{
				models.OperationSelect,
				models.OperationSelect,
			},
		},
		{
			name:          "Complex UPDATE with subquery",
			query:         "UPDATE users SET status = (SELECT MAX(status) FROM status_types) WHERE id = 1; SELECT 1",
			expectedCount: 2,
			expectedStmts: []string{
				"UPDATE users SET status = (SELECT MAX(status) FROM status_types) WHERE id = 1",
				"SELECT 1",
			},
			expectedTypes: []models.OperationType{
				models.OperationUpdate,
				models.OperationSelect,
			},
		},
		{
			name:          "CREATE TABLE",
			query:         "CREATE TABLE test (id INT); INSERT INTO test VALUES (1)",
			expectedCount: 2,
			expectedStmts: []string{
				"CREATE TABLE test (id INT)",
				"INSERT INTO test VALUES (1)",
			},
			expectedTypes: []models.OperationType{
				models.OperationCreateTable,
				models.OperationInsert,
			},
		},
		{
			name:          "DROP TABLE",
			query:         "DROP TABLE IF EXISTS old_table; SELECT * FROM new_table",
			expectedCount: 2,
			expectedStmts: []string{
				"DROP TABLE IF EXISTS old_table",
				"SELECT * FROM new_table",
			},
			expectedTypes: []models.OperationType{
				models.OperationDropTable,
				models.OperationSelect,
			},
		},
		{
			name:          "ALTER TABLE",
			query:         "ALTER TABLE users ADD COLUMN phone VARCHAR(20); SELECT phone FROM users",
			expectedCount: 2,
			expectedStmts: []string{
				"ALTER TABLE users ADD COLUMN phone VARCHAR(20)",
				"SELECT phone FROM users",
			},
			expectedTypes: []models.OperationType{
				models.OperationAlterTable,
				models.OperationSelect,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMultipleQueries(tt.query)
			assert.Len(t, result.Statements, tt.expectedCount, "Expected %d statements, got %d", tt.expectedCount, len(result.Statements))
			assert.Empty(t, result.Errors, "Expected no errors")

			for i, expectedStmt := range tt.expectedStmts {
				if i < len(result.Statements) {
					assert.Equal(t, expectedStmt, result.Statements[i].QueryText, "Statement %d mismatch", i)
					assert.Equal(t, i, result.Statements[i].Sequence, "Statement %d sequence mismatch", i)
					assert.Equal(t, tt.expectedTypes[i], result.Statements[i].OperationType, "Statement %d operation type mismatch", i)
				}
			}
		})
	}
}

// TestValidateMultiQuery_TransactionControl_Blocked tests that transaction control statements are blocked
func TestValidateMultiQuery_TransactionControl_Blocked(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectBlocked bool
		blockedStmt   string
	}{
		{
			name:          "BEGIN blocked",
			query:         "BEGIN; SELECT * FROM users",
			expectBlocked: true,
			blockedStmt:   "BEGIN",
		},
		{
			name:          "BEGIN TRANSACTION blocked",
			query:         "BEGIN TRANSACTION; UPDATE users SET active = 1",
			expectBlocked: true,
			blockedStmt:   "BEGIN TRANSACTION",
		},
		{
			name:          "COMMIT blocked",
			query:         "SELECT * FROM users; COMMIT",
			expectBlocked: true,
			blockedStmt:   "COMMIT",
		},
		{
			name:          "ROLLBACK blocked",
			query:         "ROLLBACK",
			expectBlocked: true,
			blockedStmt:   "ROLLBACK",
		},
		{
			name:          "START TRANSACTION blocked",
			query:         "START TRANSACTION; INSERT INTO logs (msg) VALUES ('test')",
			expectBlocked: true,
			blockedStmt:   "START TRANSACTION",
		},
		{
			name:          "Multiple transaction statements all blocked",
			query:         "BEGIN; SELECT 1; COMMIT; ROLLBACK",
			expectBlocked: true,
			blockedStmt:   "BEGIN",
		},
		{
			name:          "Mixed case BEGIN blocked",
			query:         "begin; select * from users",
			expectBlocked: true,
			blockedStmt:   "begin",
		},
		{
			name:          "Mixed case COMMIT blocked",
			query:         "select * from users; Commit",
			expectBlocked: true,
			blockedStmt:   "Commit",
		},
		{
			name:          "SAVEPOINT not blocked (not in prefix check)",
			query:         "SAVEPOINT sp1; SELECT * FROM users",
			expectBlocked: false,
		},
		{
			name:          "Valid multi-query not blocked",
			query:         "SELECT * FROM users; UPDATE users SET active = 1",
			expectBlocked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateMultiQuery(tt.query)

			if tt.expectBlocked {
				assert.NotEmpty(t, result.Errors, "Expected errors for transaction control statement")
				found := false
				for _, err := range result.Errors {
					if err.Message == "Transaction control statements (BEGIN, COMMIT, ROLLBACK, START TRANSACTION) are not allowed in multi-query mode" {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected transaction control error message")
			} else {
				// For SAVEPOINT or other non-blocked transaction statements
				// they might be parsed but should not have transaction control errors
				hasTransError := false
				for _, err := range result.Errors {
					if err.Message == "Transaction control statements (BEGIN, COMMIT, ROLLBACK, START TRANSACTION) are not allowed in multi-query mode" {
						hasTransError = true
						break
					}
				}
				assert.False(t, hasTransError, "Should not have transaction control error")
			}
		})
	}
}

// TestIsMultiQuery_Detection_Works tests the IsMultiQuery detection functionality
func TestIsMultiQuery_Detection_Works(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "Empty string is not multi-query",
			query:    "",
			expected: false,
		},
		{
			name:     "Whitespace only is not multi-query",
			query:    "   \n\t  ",
			expected: false,
		},
		{
			name:     "Single SELECT is not multi-query",
			query:    "SELECT * FROM users",
			expected: false,
		},
		{
			name:     "Single SELECT with semicolon is not multi-query",
			query:    "SELECT * FROM users;",
			expected: false,
		},
		{
			name:     "Two SELECT statements is multi-query",
			query:    "SELECT 1; SELECT 2",
			expected: true,
		},
		{
			name:     "SET followed by UPDATE is multi-query",
			query:    "SET @x = 1; UPDATE t SET col = @x",
			expected: true,
		},
		{
			name:     "Three statements is multi-query",
			query:    "SELECT 1; SELECT 2; SELECT 3",
			expected: true,
		},
		{
			name:     "INSERT then UPDATE is multi-query",
			query:    "INSERT INTO logs VALUES ('start'); UPDATE stats SET count = count + 1",
			expected: true,
		},
		{
			name:     "Comment followed by query is multi-query if two statements",
			query:    "-- Comment\nSELECT 1; SELECT 2",
			expected: true,
		},
		{
			name:     "Comment only is not multi-query",
			query:    "-- Just a comment",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMultiQuery(tt.query)
			assert.Equal(t, tt.expected, result, "IsMultiQuery result mismatch for: %s", tt.name)
		})
	}
}

// TestParseMultipleQueries_SemicolonsInStrings_Handled tests handling of semicolons within string literals
func TestParseMultipleQueries_SemicolonsInStrings_Handled(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedStmts []string
	}{
		{
			name:          "Semicolon in single quotes",
			query:         "INSERT INTO test VALUES ('a;b'); SELECT 1",
			expectedCount: 2,
			expectedStmts: []string{
				"INSERT INTO test VALUES ('a;b')",
				"SELECT 1",
			},
		},
		{
			name:          "Semicolon in double quotes",
			query:         `INSERT INTO test VALUES ("x;y;z"); SELECT 2`,
			expectedCount: 2,
			expectedStmts: []string{
				`INSERT INTO test VALUES ("x;y;z")`,
				"SELECT 2",
			},
		},
		{
			name:          "Multiple semicolons in string",
			query:         "INSERT INTO test VALUES ('a;b;c;d'); SELECT 1",
			expectedCount: 2,
			expectedStmts: []string{
				"INSERT INTO test VALUES ('a;b;c;d')",
				"SELECT 1",
			},
		},
		{
			name:          "Escaped quote with semicolon",
			query:         `INSERT INTO test VALUES ('it\'s; working'); SELECT 1`,
			expectedCount: 2,
			expectedStmts: []string{
				`INSERT INTO test VALUES ('it\'s; working')`,
				"SELECT 1",
			},
		},
		{
			name:          "String with semicolon then real separator",
			query:         "SELECT 'value;with;semicolons'; UPDATE users SET name = 'test'",
			expectedCount: 2,
			expectedStmts: []string{
				"SELECT 'value;with;semicolons'",
				"UPDATE users SET name = 'test'",
			},
		},
		{
			name:          "SET with semicolon in value",
			query:         "SET @msg = 'Hello; World'; SELECT @msg",
			expectedCount: 2,
			expectedStmts: []string{
				"SET @msg = 'Hello; World'",
				"SELECT @msg",
			},
		},
		{
			name:          "Mixed quotes with semicolons",
			query:         `SELECT 'single;quote'; SELECT "double;quote"`,
			expectedCount: 2,
			expectedStmts: []string{
				`SELECT 'single;quote'`,
				`SELECT "double;quote"`,
			},
		},
		{
			name:          "Complex nested quotes",
			query:         `INSERT INTO test VALUES ('He said "Hi; there"'); SELECT 1`,
			expectedCount: 2,
			expectedStmts: []string{
				`INSERT INTO test VALUES ('He said "Hi; there"')`,
				"SELECT 1",
			},
		},
		{
			name:          "Empty string then statement",
			query:         "SELECT ''; SELECT 1",
			expectedCount: 2,
			expectedStmts: []string{
				"SELECT ''",
				"SELECT 1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMultipleQueries(tt.query)
			assert.Len(t, result.Statements, tt.expectedCount, "Expected %d statements, got %d", tt.expectedCount, len(result.Statements))
			assert.Empty(t, result.Errors, "Expected no errors")

			for i, expectedStmt := range tt.expectedStmts {
				if i < len(result.Statements) {
					assert.Equal(t, expectedStmt, result.Statements[i].QueryText, "Statement %d mismatch", i)
				}
			}
		})
	}
}

// TestParseMultipleQueries_CommentHandling_Works tests parsing with various comment styles
func TestParseMultipleQueries_CommentHandling_Works(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedStmts []string
	}{
		{
			name:          "Line comment before statement",
			query:         "-- This is a comment\nSELECT * FROM users",
			expectedCount: 1,
			expectedStmts: []string{
				"-- This is a comment\nSELECT * FROM users",
			},
		},
		{
			name:          "Line comment after statement",
			query:         "SELECT * FROM users; -- End of query\nSELECT 1",
			expectedCount: 2,
			expectedStmts: []string{
				"SELECT * FROM users",
				"-- End of query\nSELECT 1",
			},
		},
		{
			name:          "Block comment before statement",
			query:         "/* Start */\nSELECT * FROM users",
			expectedCount: 1,
			expectedStmts: []string{
				"/* Start */\nSELECT * FROM users",
			},
		},
		{
			name:          "Block comment after statement",
			query:         "SELECT * FROM users /* end */; SELECT 1",
			expectedCount: 2,
			expectedStmts: []string{
				"SELECT * FROM users /* end */",
				"SELECT 1",
			},
		},
		{
			name:          "Multi-line block comment",
			query:         "/* Line 1\n   Line 2\n   Line 3 */\nSELECT 1",
			expectedCount: 1,
			expectedStmts: []string{
				"/* Line 1\n   Line 2\n   Line 3 */\nSELECT 1",
			},
		},
		{
			name:          "Mixed line and block comments",
			query:         "-- Line comment\n/* Block */\nSELECT 1; -- Another line\nSELECT 2",
			expectedCount: 2,
			expectedStmts: []string{
				"-- Line comment\n/* Block */\nSELECT 1",
				"-- Another line\nSELECT 2",
			},
		},
		{
			name:          "Comment with semicolon inside",
			query:         "-- Comment with ; semicolon\nSELECT 1",
			expectedCount: 1,
			expectedStmts: []string{
				"-- Comment with ; semicolon\nSELECT 1",
			},
		},
		{
			name:          "Block comment with semicolon inside",
			query:         "/* Comment with ; semicolon */ SELECT 1",
			expectedCount: 1,
			expectedStmts: []string{
				"/* Comment with ; semicolon */ SELECT 1",
			},
		},
		{
			name:          "Multiple line comments",
			query:         "-- Comment 1\n-- Comment 2\nSELECT 1; -- Comment 3",
			expectedCount: 2,
			expectedStmts: []string{
				"-- Comment 1\n-- Comment 2\nSELECT 1",
				"-- Comment 3",
			},
		},
		{
			name:          "Indonesian language comments",
			query:         "-- Update data pengguna\nUPDATE users SET status = 'active'; -- Berhasil diupdate",
			expectedCount: 2,
			expectedStmts: []string{
				"-- Update data pengguna\nUPDATE users SET status = 'active'",
				"-- Berhasil diupdate",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMultipleQueries(tt.query)
			assert.Len(t, result.Statements, tt.expectedCount, "Expected %d statements, got %d", tt.expectedCount, len(result.Statements))

			for i, expectedStmt := range tt.expectedStmts {
				if i < len(result.Statements) {
					assert.Equal(t, expectedStmt, result.Statements[i].QueryText, "Statement %d mismatch", i)
				}
			}
		})
	}
}

// TestParseMultipleQueries_HandlerIntegration_Works tests parser integration with handler flow scenarios
func TestParseMultipleQueries_HandlerIntegration_Works(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectValid    bool
		expectedStmts  int
		expectedErrors int
	}{
		{
			name:           "Real-world workflow: SET variables then UPDATE",
			query:          "SET @wallet_id = 23916; SET @execution_date = '2026-03-03 15:15:15'; UPDATE wallets SET updated_at = @execution_date WHERE id = @wallet_id",
			expectValid:    true,
			expectedStmts:  3,
			expectedErrors: 0,
		},
		{
			name:           "Real-world workflow: SELECT then INSERT log",
			query:          "SELECT * FROM users WHERE id = 1; INSERT INTO audit_log (action, user_id) VALUES ('viewed', 1)",
			expectValid:    true,
			expectedStmts:  2,
			expectedErrors: 0,
		},
		{
			name:           "Blocked: Transaction in multi-query",
			query:          "BEGIN; SELECT * FROM users; COMMIT",
			expectValid:    false,
			expectedStmts:  3,
			expectedErrors: 2, // BEGIN and COMMIT both blocked
		},
		{
			name:           "Valid: Complex reporting query",
			query:          "SELECT COUNT(*) FROM orders; SELECT SUM(total) FROM orders; SELECT AVG(total) FROM orders",
			expectValid:    true,
			expectedStmts:  3,
			expectedErrors: 0,
		},
		{
			name:           "Valid: Data migration script",
			query:          "INSERT INTO new_table SELECT * FROM old_table WHERE id <= 100; INSERT INTO new_table SELECT * FROM old_table WHERE id > 100",
			expectValid:    true,
			expectedStmts:  2,
			expectedErrors: 0,
		},
		{
			name:           "Mixed valid and blocked statements",
			query:          "SELECT * FROM users; START TRANSACTION; UPDATE users SET active = 1",
			expectValid:    false,
			expectedStmts:  3,
			expectedErrors: 1, // START TRANSACTION blocked
		},
		{
			name:           "Workflow with comments",
			query:          "-- Initialize\nSET @batch_size = 100;\n-- Process batch\nUPDATE items SET processed = 1 LIMIT @batch_size",
			expectValid:    true,
			expectedStmts:  2,
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First parse without validation
			parseResult := ParseMultipleQueries(tt.query)
			assert.Len(t, parseResult.Statements, tt.expectedStmts, "Expected %d statements", tt.expectedStmts)

			// Then validate
			validateResult := ValidateMultiQuery(tt.query)
			assert.Len(t, validateResult.Statements, tt.expectedStmts, "Validate should return same statement count")
			assert.Len(t, validateResult.Errors, tt.expectedErrors, "Expected %d validation errors", tt.expectedErrors)

			if tt.expectValid {
				assert.Empty(t, validateResult.Errors, "Expected no validation errors for valid query")
			} else {
				assert.NotEmpty(t, validateResult.Errors, "Expected validation errors for invalid query")
			}
		})
	}
}

// TestParseMultipleQueries_EmptyQuery_Handled tests handling of empty and whitespace-only queries
func TestParseMultipleQueries_EmptyQuery_Handled(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedEmpty bool
	}{
		{
			name:          "Empty string",
			query:         "",
			expectedCount: 0,
			expectedEmpty: true,
		},
		{
			name:          "Whitespace only",
			query:         "   ",
			expectedCount: 0,
			expectedEmpty: true,
		},
		{
			name:          "Newlines and tabs only",
			query:         "\n\n\t\t  \n",
			expectedCount: 0,
			expectedEmpty: true,
		},
		{
			name:          "Empty with semicolons only",
			query:         "; ; ;",
			expectedCount: 0,
			expectedEmpty: true,
		},
		{
			name:          "Single space between semicolons",
			query:         "; ;",
			expectedCount: 0,
			expectedEmpty: true,
		},
		{
			name:          "Empty string with comment only",
			query:         "-- Just a comment",
			expectedCount: 1,
			expectedEmpty: false,
		},
		{
			name:          "Empty string with block comment only",
			query:         "/* Just a block comment */",
			expectedCount: 1,
			expectedEmpty: false,
		},
		{
			name:          "Valid query after empty",
			query:         "; SELECT 1",
			expectedCount: 1,
			expectedEmpty: false,
		},
		{
			name:          "Query before empty",
			query:         "SELECT 1; ; ;",
			expectedCount: 1,
			expectedEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMultipleQueries(tt.query)

			if tt.expectedEmpty {
				assert.Empty(t, result.Statements, "Expected empty statements for empty query")
			} else {
				assert.NotEmpty(t, result.Statements, "Expected non-empty statements")
			}
			assert.Len(t, result.Statements, tt.expectedCount, "Expected %d statements", tt.expectedCount)
			assert.Empty(t, result.Errors, "Empty queries should not produce errors")
		})
	}
}

// TestIsMultiQuery_SingleVsMulti_Detected tests detection of single vs multi-query statements
func TestIsMultiQuery_SingleVsMulti_Detected(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		isMulti     bool
		description string
	}{
		{
			name:        "Simple single SELECT",
			query:       "SELECT * FROM users",
			isMulti:     false,
			description: "Single statement without semicolon",
		},
		{
			name:        "Single SELECT with trailing semicolon",
			query:       "SELECT * FROM users;",
			isMulti:     false,
			description: "Single statement with semicolon is still single",
		},
		{
			name:        "Single INSERT",
			query:       "INSERT INTO users (name) VALUES ('John')",
			isMulti:     false,
			description: "Single INSERT without semicolon",
		},
		{
			name:        "Single UPDATE",
			query:       "UPDATE users SET active = 1 WHERE id = 5",
			isMulti:     false,
			description: "Single UPDATE without semicolon",
		},
		{
			name:        "Single DELETE",
			query:       "DELETE FROM logs WHERE created < '2024-01-01'",
			isMulti:     false,
			description: "Single DELETE without semicolon",
		},
		{
			name:        "Two SELECTs is multi",
			query:       "SELECT 1; SELECT 2",
			isMulti:     true,
			description: "Two statements make it multi-query",
		},
		{
			name:        "Three statements is multi",
			query:       "SELECT 1; SELECT 2; SELECT 3",
			isMulti:     true,
			description: "Three statements definitely multi-query",
		},
		{
			name:        "Mixed operations multi",
			query:       "SELECT * FROM users; UPDATE stats SET count = count + 1",
			isMulti:     true,
			description: "SELECT followed by UPDATE is multi-query",
		},
		{
			name:        "SET and SELECT is multi",
			query:       "SET @x = 1; SELECT @x",
			isMulti:     true,
			description: "Variable set followed by select is multi-query",
		},
		{
			name:        "Single SET is single",
			query:       "SET @x = 1",
			isMulti:     false,
			description: "Single SET statement is not multi-query",
		},
		{
			name:        "Complex single statement",
			query:       "SELECT u.*, o.total FROM users u JOIN orders o ON u.id = o.user_id WHERE o.total > 100 ORDER BY o.total DESC",
			isMulti:     false,
			description: "Complex JOIN with WHERE and ORDER BY is still single statement",
		},
		{
			name:        "Subquery in single statement",
			query:       "SELECT * FROM users WHERE id IN (SELECT user_id FROM orders WHERE total > 100)",
			isMulti:     false,
			description: "Subquery does not make it multi-query",
		},
		{
			name:        "CTE single statement",
			query:       "WITH cte AS (SELECT * FROM users) SELECT * FROM cte",
			isMulti:     false,
			description: "CTE is still a single statement",
		},
		{
			name:        "Union in single statement",
			query:       "SELECT * FROM table1 UNION ALL SELECT * FROM table2",
			isMulti:     false,
			description: "UNION combines into single statement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMultiQuery(tt.query)
			assert.Equal(t, tt.isMulti, result, "%s: %s", tt.description, tt.name)
		})
	}
}
