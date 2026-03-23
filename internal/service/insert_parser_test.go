package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseInsertValues_SingleRow(t *testing.T) {
	parser := &InsertParser{}
	query := "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')"

	result, err := parser.parseInsertValues(query)

	assert.NoError(t, err)
	assert.Equal(t, []string{"name", "email"}, result.Columns)
	assert.Len(t, result.Rows, 1)
	assert.Equal(t, "John", result.Rows[0][0])
	assert.Equal(t, "john@example.com", result.Rows[0][1])
}

func TestParseInsertValues_MultiRow(t *testing.T) {
	parser := &InsertParser{}
	query := "INSERT INTO users (name) VALUES ('John'), ('Jane'), ('Bob')"

	result, err := parser.parseInsertValues(query)

	assert.NoError(t, err)
	assert.Len(t, result.Rows, 3)
	assert.Equal(t, "John", result.Rows[0][0])
	assert.Equal(t, "Jane", result.Rows[1][0])
	assert.Equal(t, "Bob", result.Rows[2][0])
}

func TestParseInsertValues_NoColumns(t *testing.T) {
	parser := &InsertParser{}
	query := "INSERT INTO users VALUES (1, 'John', 'john@example.com')"

	result, err := parser.parseInsertValues(query)

	assert.NoError(t, err)
	assert.Nil(t, result.Columns) // Columns not specified
	assert.Len(t, result.Rows[0], 3)
}

func TestParseInsertValues_EscapedQuotes(t *testing.T) {
	parser := &InsertParser{}
	query := "INSERT INTO users (name) VALUES ('O''Brien')"

	result, err := parser.parseInsertValues(query)

	assert.NoError(t, err)
	assert.Equal(t, "O'Brien", result.Rows[0][0])
}

func TestParseInsertValues_JSONData(t *testing.T) {
	parser := &InsertParser{}
	query := `INSERT INTO logs (data) VALUES ('{"type": "login", "user_id": 123}')`

	result, err := parser.parseInsertValues(query)

	assert.NoError(t, err)
	assert.Contains(t, result.Rows[0][0], `"type"`)
	assert.Contains(t, result.Rows[0][0], `"user_id"`)
}

func TestParseInsertValues_NULLValues(t *testing.T) {
	parser := &InsertParser{}
	query := "INSERT INTO users (name, email) VALUES ('John', NULL)"

	result, err := parser.parseInsertValues(query)

	assert.NoError(t, err)
	assert.Equal(t, "NULL", result.Rows[0][1])
}

func TestParseInsertValues_MixedTypes(t *testing.T) {
	parser := &InsertParser{}
	query := "INSERT INTO products (name, price, active) VALUES ('Widget', 9.99, true)"

	result, err := parser.parseInsertValues(query)

	assert.NoError(t, err)
	assert.Equal(t, "Widget", result.Rows[0][0])
	assert.Equal(t, "9.99", result.Rows[0][1])
	assert.Equal(t, "true", result.Rows[0][2])
}

func TestParseInsertValues_InvalidSyntax(t *testing.T) {
	parser := &InsertParser{}
	query := "SELECT * FROM users"

	_, err := parser.parseInsertValues(query)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}
