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
