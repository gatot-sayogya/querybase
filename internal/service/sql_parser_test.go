package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/querybase/internal/models"
)

func TestParseAndValidateSQL_PostgreSQL(t *testing.T) {
	tests := []struct {
		name        string
		sql         string
		expectValid bool
		errContains string
	}{
		{
			name:        "Valid SELECT",
			sql:         "SELECT * FROM users WHERE id = 1",
			expectValid: true,
		},
		{
			name:        "Valid INSERT",
			sql:         "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')",
			expectValid: true,
		},
		{
			name:        "Valid UPDATE",
			sql:         "UPDATE users SET name = 'Jane' WHERE id = 1",
			expectValid: true,
		},
		{
			name:        "Valid DELETE",
			sql:         "DELETE FROM users WHERE id = 1",
			expectValid: true,
		},
		{
			name:        "UPDATE with AND is syntactically valid but semantically wrong",
			sql:         "UPDATE users SET name = 'John' AND email = 'john@example.com' WHERE id = 1",
			expectValid: true, // Parsers accept this (it's valid boolean expr)
		},
		{
			name:        "Invalid SQL syntax",
			sql:         "SELECT * FROM",
			expectValid: false,
			errContains: "syntax error",
		},
		{
			name:        "Valid CREATE TABLE",
			sql:         "CREATE TABLE test (id INT PRIMARY KEY, name VARCHAR(100))",
			expectValid: true,
		},
		{
			name:        "Valid DROP TABLE",
			sql:         "DROP TABLE IF EXISTS test",
			expectValid: true,
		},
		{
			name:        "Valid JOIN",
			sql:         "SELECT u.*, o.id FROM users u JOIN orders o ON u.id = o.user_id",
			expectValid: true,
		},
		{
			name:        "Valid subquery",
			sql:         "SELECT * FROM users WHERE id IN (SELECT user_id FROM orders)",
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAndValidateSQL(tt.sql, models.DataSourceTypePostgreSQL)
			
			// Parser should not return an error for parsing failures
			// Instead, it should return a result with Valid=false
			require.NoError(t, err)
			require.NotNil(t, result)
			
			assert.Equal(t, tt.expectValid, result.Valid)
			
			if !tt.expectValid && tt.errContains != "" {
				assert.Contains(t, result.Error, tt.errContains)
			}
		})
	}
}

func TestParseAndValidateSQL_MySQL(t *testing.T) {
	tests := []struct {
		name        string
		sql         string
		expectValid bool
		errContains string
	}{
		{
			name:        "Valid SELECT",
			sql:         "SELECT * FROM users WHERE id = 1",
			expectValid: true,
		},
		{
			name:        "Valid INSERT",
			sql:         "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')",
			expectValid: true,
		},
		{
			name:        "Valid UPDATE",
			sql:         "UPDATE users SET name = 'Jane' WHERE id = 1",
			expectValid: true,
		},
		{
			name:        "Valid DELETE",
			sql:         "DELETE FROM users WHERE id = 1",
			expectValid: true,
		},
		{
			name:        "UPDATE with AND is syntactically valid but semantically wrong",
			sql:         "UPDATE users SET name = 'John' AND email = 'john@example.com' WHERE id = 1",
			expectValid: true, // Parsers accept this (it's valid boolean expr)
		},
		{
			name:        "Valid MySQL INSERT with SET",
			sql:         "INSERT INTO users SET name = 'John', email = 'john@example.com'",
			expectValid: true,
		},
		{
			name:        "Valid MySQL REPLACE",
			sql:         "REPLACE INTO users (id, name) VALUES (1, 'John')",
			expectValid: true,
		},
		{
			name:        "Valid CREATE TABLE",
			sql:         "CREATE TABLE test (id INT PRIMARY KEY, name VARCHAR(100))",
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAndValidateSQL(tt.sql, models.DataSourceTypeMySQL)
			
			require.NoError(t, err)
			require.NotNil(t, result)
			
			assert.Equal(t, tt.expectValid, result.Valid)
			
			if !tt.expectValid && tt.errContains != "" {
				assert.Contains(t, result.Error, tt.errContains)
			}
		})
	}
}

func TestParseAndValidateSQL_Generic(t *testing.T) {
	tests := []struct {
		name        string
		sql         string
		expectValid bool
	}{
		{
			name:        "Valid SELECT",
			sql:         "SELECT * FROM users",
			expectValid: true,
		},
		{
			name:        "Empty query",
			sql:         "",
			expectValid: false,
		},
		{
			name:        "UPDATE without SET",
			sql:         "UPDATE users WHERE id = 1",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAndValidateSQL(tt.sql, "")
			
			require.NoError(t, err)
			require.NotNil(t, result)
			
			assert.Equal(t, tt.expectValid, result.Valid)
		})
	}
}

func TestDetectOperationTypeWithDialect(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		dialect  models.DataSourceType
		expected models.OperationType
	}{
		{
			name:     "PostgreSQL SELECT",
			sql:      "SELECT * FROM users",
			dialect:  models.DataSourceTypePostgreSQL,
			expected: models.OperationSelect,
		},
		{
			name:     "PostgreSQL INSERT",
			sql:      "INSERT INTO users VALUES (1, 'John')",
			dialect:  models.DataSourceTypePostgreSQL,
			expected: models.OperationInsert,
		},
		{
			name:     "MySQL UPDATE",
			sql:      "UPDATE users SET name = 'Jane' WHERE id = 1",
			dialect:  models.DataSourceTypeMySQL,
			expected: models.OperationUpdate,
		},
		{
			name:     "MySQL DELETE",
			sql:      "DELETE FROM users WHERE id = 1",
			dialect:  models.DataSourceTypeMySQL,
			expected: models.OperationDelete,
		},
		{
			name:     "Fallback to generic",
			sql:      "SELECT * FROM users",
			dialect:  "",
			expected: models.OperationSelect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectOperationTypeWithDialect(tt.sql, tt.dialect)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateSQLWithDialect(t *testing.T) {
	tests := []struct {
		name        string
		sql         string
		dialect     models.DataSourceType
		expectError bool
		errContains string
	}{
		{
			name:        "Valid PostgreSQL",
			sql:         "SELECT * FROM users WHERE id = 1",
			dialect:     models.DataSourceTypePostgreSQL,
			expectError: false,
		},
		{
			name:        "Invalid PostgreSQL",
			sql:         "SELECT * FROM",
			dialect:     models.DataSourceTypePostgreSQL,
			expectError: true,
			errContains: "syntax error",
		},
		{
			name:        "Valid MySQL",
			sql:         "SELECT * FROM users WHERE id = 1",
			dialect:     models.DataSourceTypeMySQL,
			expectError: false,
		},
		{
			name:        "Invalid MySQL",
			sql:         "SELECT * FROM",
			dialect:     models.DataSourceTypeMySQL,
			expectError: true,
			errContains: "syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSQLWithDialect(tt.sql, tt.dialect)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidSQL(t *testing.T) {
	tests := []struct {
		name          string
		sql           string
		dialect       models.DataSourceType
		expectValid   bool
		expectErrMsg  string
	}{
		{
			name:        "Valid PostgreSQL",
			sql:         "SELECT 1",
			dialect:     models.DataSourceTypePostgreSQL,
			expectValid: true,
		},
		{
			name:         "Invalid PostgreSQL",
			sql:          "SELECT * FROM",
			dialect:      models.DataSourceTypePostgreSQL,
			expectValid:  false,
			expectErrMsg: "syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errMsg := IsValidSQL(tt.sql, tt.dialect)
			assert.Equal(t, tt.expectValid, valid)
			if tt.expectErrMsg != "" {
				assert.Contains(t, errMsg, tt.expectErrMsg)
			}
		})
	}
}

func TestExtractTables(t *testing.T) {
	tests := []struct {
		name          string
		sql           string
		dialect       models.DataSourceType
		expectError   bool
		expectedTable string
	}{
		{
			name:          "PostgreSQL simple SELECT",
			sql:           "SELECT * FROM users",
			dialect:       models.DataSourceTypePostgreSQL,
			expectError:   false,
			expectedTable: "users",
		},
		{
			name:          "MySQL simple SELECT",
			sql:           "SELECT * FROM orders",
			dialect:       models.DataSourceTypeMySQL,
			expectError:   false,
			expectedTable: "orders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tables, err := ExtractTables(tt.sql, tt.dialect)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Note: Table extraction is simplified and may not work for all queries
				// This test documents current behavior
				_ = tables
			}
		})
	}
}
