package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourorg/querybase/internal/models"
)

// TestParseMultipleQueries_JSONValues verifies that the parser correctly handles
// SQL statements whose string literals contain JSON data — including embedded
// semicolons, nested braces/brackets, JSONB operators, and multi-line values.
func TestParseMultipleQueries_JSONValues(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedStmts []string
		expectedTypes []models.OperationType
	}{
		{
			name:          "INSERT with JSON object value containing semicolon",
			query:         `INSERT INTO event_logs (event_type, payload) VALUES ('order_created', '{"message":"Hello; World","status":"ok"}')`,
			expectedCount: 1,
			expectedStmts: []string{
				`INSERT INTO event_logs (event_type, payload) VALUES ('order_created', '{"message":"Hello; World","status":"ok"}')`,
			},
			expectedTypes: []models.OperationType{models.OperationInsert},
		},
		{
			name:          "INSERT with JSON value containing multiple semicolons",
			query:         `INSERT INTO notes (body) VALUES ('{"sql":"SELECT 1; SELECT 2; SELECT 3","executed":false}')`,
			expectedCount: 1,
			expectedStmts: []string{
				`INSERT INTO notes (body) VALUES ('{"sql":"SELECT 1; SELECT 2; SELECT 3","executed":false}')`,
			},
			expectedTypes: []models.OperationType{models.OperationInsert},
		},
		{
			name:          "INSERT with JSON array value",
			query:         `INSERT INTO products (name, tags) VALUES ('Laptop Pro', '["electronics","portable","high-performance"]')`,
			expectedCount: 1,
			expectedStmts: []string{
				`INSERT INTO products (name, tags) VALUES ('Laptop Pro', '["electronics","portable","high-performance"]')`,
			},
			expectedTypes: []models.OperationType{models.OperationInsert},
		},
		{
			name:          "INSERT with deeply nested JSON object",
			query:         `INSERT INTO customers (name, metadata) VALUES ('John Doe', '{"tier":"gold","address":{"city":"New York","zip":"10001"},"tags":["vip","loyal"],"credit_limit":5000}')`,
			expectedCount: 1,
			expectedStmts: []string{
				`INSERT INTO customers (name, metadata) VALUES ('John Doe', '{"tier":"gold","address":{"city":"New York","zip":"10001"},"tags":["vip","loyal"],"credit_limit":5000}')`,
			},
			expectedTypes: []models.OperationType{models.OperationInsert},
		},
		{
			name:          "UPDATE with JSONB merge operator and semicolon in value",
			query:         `UPDATE customers SET metadata = metadata || '{"note":"VIP; Priority","updated":true}' WHERE id = 1`,
			expectedCount: 1,
			expectedStmts: []string{
				`UPDATE customers SET metadata = metadata || '{"note":"VIP; Priority","updated":true}' WHERE id = 1`,
			},
			expectedTypes: []models.OperationType{models.OperationUpdate},
		},
		{
			name:          "SELECT with JSONB ->> operator and semicolon in compared value",
			query:         `SELECT * FROM customers WHERE metadata->>'note' = 'VIP; Priority'`,
			expectedCount: 1,
			expectedStmts: []string{
				`SELECT * FROM customers WHERE metadata->>'note' = 'VIP; Priority'`,
			},
			expectedTypes: []models.OperationType{models.OperationSelect},
		},
		{
			name:          "SELECT with JSONB @> containment operator",
			query:         `SELECT * FROM products WHERE tags @> '["laptop","portable"]'`,
			expectedCount: 1,
			expectedStmts: []string{
				`SELECT * FROM products WHERE tags @> '["laptop","portable"]'`,
			},
			expectedTypes: []models.OperationType{models.OperationSelect},
		},
		{
			name: "Two INSERT statements each with a JSON payload",
			query: `INSERT INTO event_logs (event_type, payload) VALUES ('login', '{"user_id":1,"ip":"192.168.1.1"}');
INSERT INTO event_logs (event_type, payload) VALUES ('logout', '{"user_id":1,"session_ms":3600000}')`,
			expectedCount: 2,
			expectedStmts: []string{
				`INSERT INTO event_logs (event_type, payload) VALUES ('login', '{"user_id":1,"ip":"192.168.1.1"}')`,
				`INSERT INTO event_logs (event_type, payload) VALUES ('logout', '{"user_id":1,"session_ms":3600000}')`,
			},
			expectedTypes: []models.OperationType{
				models.OperationInsert,
				models.OperationInsert,
			},
		},
		{
			name: "SELECT followed by INSERT with JSON value",
			query: `SELECT id FROM customers WHERE email = 'test@example.com';
INSERT INTO event_logs (event_type, payload) VALUES ('profile_view', '{"viewer_id":99,"target_email":"test@example.com"}')`,
			expectedCount: 2,
			expectedStmts: []string{
				`SELECT id FROM customers WHERE email = 'test@example.com'`,
				`INSERT INTO event_logs (event_type, payload) VALUES ('profile_view', '{"viewer_id":99,"target_email":"test@example.com"}')`,
			},
			expectedTypes: []models.OperationType{
				models.OperationSelect,
				models.OperationInsert,
			},
		},
		{
			name: "Multi-line INSERT with JSON object spanning lines",
			query: `INSERT INTO customers (
  first_name,
  last_name,
  metadata
) VALUES (
  'Jane',
  'Smith',
  '{"tier":"silver","preferences":{"newsletter":true,"sms":false}}'
)`,
			expectedCount: 1,
			expectedTypes: []models.OperationType{models.OperationInsert},
		},
		{
			name: "Multi-line UPDATE with JSONB merge and JSON value",
			query: `UPDATE customers
SET
  metadata = metadata || '{"tier":"gold","upgraded_at":"2024-02-01"}',
  updated_at = NOW()
WHERE email = 'bob.johnson@example.com'`,
			expectedCount: 1,
			expectedTypes: []models.OperationType{models.OperationUpdate},
		},
		{
			name:          "DELETE using JSONB filter",
			query:         `DELETE FROM event_logs WHERE payload->>'event_type' = 'test_event' AND payload->>'source' = 'querybase-test'`,
			expectedCount: 1,
			expectedStmts: []string{
				`DELETE FROM event_logs WHERE payload->>'event_type' = 'test_event' AND payload->>'source' = 'querybase-test'`,
			},
			expectedTypes: []models.OperationType{models.OperationDelete},
		},
		{
			name: "SET variable then UPDATE with JSON-like string value",
			query: `SET @config = '{"env":"staging","debug":true}';
UPDATE app_settings SET config = @config WHERE id = 1`,
			expectedCount: 2,
			expectedStmts: []string{
				`SET @config = '{"env":"staging","debug":true}'`,
				`UPDATE app_settings SET config = @config WHERE id = 1`,
			},
			expectedTypes: []models.OperationType{
				models.OperationSet,
				models.OperationUpdate,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMultipleQueries(tt.query)
			assert.Len(t, result.Statements, tt.expectedCount,
				"expected %d statements, got %d", tt.expectedCount, len(result.Statements))
			assert.Empty(t, result.Errors, "expected no parse errors")

			for i, expectedStmt := range tt.expectedStmts {
				if i < len(result.Statements) {
					assert.Equal(t, expectedStmt, result.Statements[i].QueryText,
						"statement %d text mismatch", i)
				}
			}

			for i, expectedType := range tt.expectedTypes {
				if i < len(result.Statements) {
					assert.Equal(t, expectedType, result.Statements[i].OperationType,
						"statement %d operationType mismatch", i)
				}
			}
		})
	}
}

// TestDetectOperationType_JSONQueries verifies that operation type detection
// works correctly for queries that use JSONB operators and JSON string literals.
func TestDetectOperationType_JSONQueries(t *testing.T) {
	tests := []struct {
		name         string
		sql          string
		expectedType models.OperationType
	}{
		{
			name:         "INSERT with JSON object literal",
			sql:          `INSERT INTO event_logs (payload) VALUES ('{"type":"login","user_id":1}')`,
			expectedType: models.OperationInsert,
		},
		{
			name:         "INSERT with JSON array literal",
			sql:          `INSERT INTO products (tags) VALUES ('["laptop","portable"]')`,
			expectedType: models.OperationInsert,
		},
		{
			name:         "UPDATE with JSONB merge operator ||",
			sql:          `UPDATE customers SET metadata = metadata || '{"tier":"gold"}' WHERE id = 1`,
			expectedType: models.OperationUpdate,
		},
		{
			name:         "UPDATE with JSONB set path #=",
			sql:          `UPDATE customers SET metadata = jsonb_set(metadata, '{tier}', '"gold"') WHERE id = 1`,
			expectedType: models.OperationUpdate,
		},
		{
			name:         "DELETE with JSONB ->> filter",
			sql:          `DELETE FROM event_logs WHERE payload->>'event_type' = 'test'`,
			expectedType: models.OperationDelete,
		},
		{
			name:         "SELECT with JSONB -> accessor",
			sql:          `SELECT metadata->'address' AS address FROM customers`,
			expectedType: models.OperationSelect,
		},
		{
			name:         "SELECT with JSONB ->> text accessor",
			sql:          `SELECT metadata->>'tier' AS tier FROM customers`,
			expectedType: models.OperationSelect,
		},
		{
			name:         "SELECT with JSONB @> containment",
			sql:          `SELECT * FROM products WHERE tags @> '["laptop"]'`,
			expectedType: models.OperationSelect,
		},
		{
			name:         "SELECT with jsonb_build_object",
			sql:          `SELECT jsonb_build_object('name', first_name, 'tier', metadata->>'tier') FROM customers`,
			expectedType: models.OperationSelect,
		},
		{
			name:         "SELECT with jsonb_agg",
			sql:          `SELECT customer_id, jsonb_agg(order_number) AS orders FROM orders GROUP BY customer_id`,
			expectedType: models.OperationSelect,
		},
		{
			name:         "SELECT with jsonb_array_elements",
			sql:          `SELECT jsonb_array_elements(line_items)->>'name' AS item FROM orders`,
			expectedType: models.OperationSelect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectOperationType(tt.sql)
			assert.Equal(t, tt.expectedType, result,
				"DetectOperationType(%q) = %v, want %v", tt.sql, result, tt.expectedType)
		})
	}
}

// TestValidateSQL_JSONQueries verifies that ValidateSQL does not produce false
// errors for valid queries that contain JSON literals or JSONB operators.
func TestValidateSQL_JSONQueries(t *testing.T) {
	tests := []struct {
		name        string
		sql         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid INSERT with JSON object value",
			sql:         `INSERT INTO event_logs (payload) VALUES ('{"type":"login","user_id":1}')`,
			expectError: false,
		},
		{
			name:        "Valid INSERT with JSON array value",
			sql:         `INSERT INTO products (tags) VALUES ('["laptop","portable"]')`,
			expectError: false,
		},
		{
			name:        "Valid INSERT with nested JSON (braces inside string, not real parens)",
			sql:         `INSERT INTO customers (metadata) VALUES ('{"address":{"city":"NY","zip":"10001"},"tags":["vip"]}')`,
			expectError: false,
		},
		{
			name:        "Valid UPDATE with JSONB merge operator",
			sql:         `UPDATE customers SET metadata = metadata || '{"tier":"gold"}' WHERE id = 1`,
			expectError: false,
		},
		{
			name:        "Valid SELECT with JSONB -> operator",
			sql:         `SELECT metadata->'address' FROM customers WHERE id = 1`,
			expectError: false,
		},
		{
			name:        "Valid SELECT with JSONB ->> operator",
			sql:         `SELECT metadata->>'tier' FROM customers`,
			expectError: false,
		},
		{
			name:        "Valid SELECT with JSONB @> operator",
			sql:         `SELECT * FROM products WHERE tags @> '["laptop"]'`,
			expectError: false,
		},
		{
			name:        "Valid SELECT with jsonb_build_object function",
			sql:         `SELECT jsonb_build_object('name', first_name, 'email', email) FROM customers`,
			expectError: false,
		},
		{
			name:        "Valid SELECT with jsonb_array_elements and subquery",
			sql:         `SELECT id, jsonb_array_elements(line_items)->>'name' AS item FROM orders WHERE status = 'delivered'`,
			expectError: false,
		},
		{
			name:        "Valid SELECT with IN subquery using JSONB filter",
			sql:         `SELECT * FROM customers WHERE id IN (SELECT customer_id FROM orders WHERE shipping_address->>'city' = 'New York')`,
			expectError: false,
		},
		{
			name:        "Valid INSERT with JSON containing semicolons inside string",
			sql:         `INSERT INTO notes (body) VALUES ('{"sql":"SELECT 1; SELECT 2","done":false}')`,
			expectError: false,
		},
		{
			name:        "Valid INSERT with JSON — unmatched braces inside string are fine",
			sql:         `INSERT INTO event_logs (payload) VALUES ('{"fragment":"{incomplete"}')`,
			expectError: false,
		},
		// These should still fail correctly (unchanged behaviour)
		{
			name:        "Invalid INSERT missing VALUES",
			sql:         `INSERT INTO event_logs`,
			expectError: true,
			errorMsg:    "VALUES",
		},
		{
			name:        "Invalid UPDATE missing SET",
			sql:         `UPDATE customers WHERE id = 1`,
			expectError: true,
			errorMsg:    "SET",
		},
		{
			name:        "Unbalanced parentheses with JSON value",
			sql:         `SELECT * FROM customers WHERE (metadata->>'tier' = 'gold'`,
			expectError: true,
			errorMsg:    "unbalanced parentheses",
		},
		{
			name:        "Unterminated string containing JSON",
			sql:         `SELECT * FROM customers WHERE metadata->>'tier' = '{"tier":"gold"`,
			expectError: true,
			errorMsg:    "unterminated string literal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSQL(tt.sql)
			if tt.expectError {
				assert.Error(t, err, "expected ValidateSQL to return an error for: %q", tt.sql)
				if tt.errorMsg != "" && err != nil {
					assert.Contains(t, err.Error(), tt.errorMsg,
						"error message should contain %q", tt.errorMsg)
				}
			} else {
				assert.NoError(t, err, "expected ValidateSQL to return nil for: %q", tt.sql)
			}
		})
	}
}

// TestIsMultiQuery_JSONValues verifies multi-query detection is not confused
// by JSON data that contains semicolons or special characters.
func TestIsMultiQuery_JSONValues(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "Single INSERT with JSON value — not multi-query",
			query:    `INSERT INTO event_logs (payload) VALUES ('{"event":"login","user_id":1}')`,
			expected: false,
		},
		{
			name:     "Single UPDATE with JSONB merge — not multi-query",
			query:    `UPDATE customers SET metadata = metadata || '{"tier":"gold"}' WHERE id = 1`,
			expected: false,
		},
		{
			name:     "Single SELECT with JSONB @> — not multi-query",
			query:    `SELECT * FROM products WHERE tags @> '["laptop"]'`,
			expected: false,
		},
		{
			name:     "JSON value with multiple semicolons — not multi-query",
			query:    `INSERT INTO notes (body) VALUES ('{"steps":"step1; step2; step3","done":false}')`,
			expected: false,
		},
		{
			name:     "Two INSERTs with JSON values — is multi-query",
			query:    `INSERT INTO event_logs (payload) VALUES ('{"type":"a"}'); INSERT INTO event_logs (payload) VALUES ('{"type":"b"}')`,
			expected: true,
		},
		{
			name:     "SELECT then UPDATE with JSON value — is multi-query",
			query:    `SELECT id FROM customers WHERE email = 'x@example.com'; UPDATE customers SET metadata = '{"verified":true}' WHERE email = 'x@example.com'`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMultiQuery(tt.query)
			assert.Equal(t, tt.expected, result, "IsMultiQuery mismatch for %q", tt.query)
		})
	}
}
