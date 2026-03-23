package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreviewInsertQuery_VALUES(t *testing.T) {
	// This is an integration test that requires a database
	// For now, test the parse logic only
	parser := &InsertParser{}
	query := "INSERT INTO users (name, email) VALUES ('John', 'john@test.com'), ('Jane', 'jane@test.com')"

	result, err := parser.parseInsertValues(query)

	require.NoError(t, err)
	assert.Equal(t, []string{"name", "email"}, result.Columns)
	assert.Len(t, result.Rows, 2)
	assert.Equal(t, "John", result.Rows[0][0])
	assert.Equal(t, "john@test.com", result.Rows[0][1])
}

func TestPreviewInsertQuery_VALUES_Empty(t *testing.T) {
	parser := &InsertParser{}
	query := "INSERT INTO users (name) VALUES ('test')"

	result, err := parser.parseInsertValues(query)

	require.NoError(t, err)
	assert.Len(t, result.Rows, 1)
}

func TestDetectInsertType_VALUES(t *testing.T) {
	query := "INSERT INTO users (name) VALUES ('John')"
	result := detectInsertType(query)
	assert.Equal(t, InsertPreviewTypeValues, result)
}

func TestDetectInsertType_SELECT(t *testing.T) {
	query := "INSERT INTO archive SELECT * FROM logs WHERE created_at < '2024-01-01'"
	result := detectInsertType(query)
	assert.Equal(t, InsertPreviewTypeSelect, result)
}

func TestExtractSelectFromInsert(t *testing.T) {
	query := "INSERT INTO audit_log SELECT action, user_id FROM events WHERE status = 'pending'"
	selectQuery, err := extractSelectFromInsert(query)

	require.NoError(t, err)
	assert.Equal(t, "SELECT action, user_id FROM events WHERE status = 'pending'", selectQuery)
}

func TestExtractInsertTableName(t *testing.T) {
	tests := []struct {
		query    string
		expected string
	}{
		{"INSERT INTO users (name) VALUES ('John')", "users"},
		{"INSERT INTO \"Users\" VALUES (1)", "Users"},
		{"INSERT INTO public.users VALUES (1)", "public.users"},
	}

	for _, tt := range tests {
		result := extractInsertTableName(tt.query)
		assert.Equal(t, tt.expected, result)
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"NULL", nil},
		{"'hello'", "hello"},
		{"\"world\"", "world"},
		{"123", int64(123)},
		{"45.67", 45.67},
		{"true", true},
		{"false", false},
		{"'O''Brien'", "O'Brien"},
		{"test", "test"},
	}

	for _, tt := range tests {
		result := parseValue(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}
