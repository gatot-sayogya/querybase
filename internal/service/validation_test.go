package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourorg/querybase/internal/models"
)

func TestDetectOperationType_WithSET(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected models.OperationType
	}{
		{
			name:     "SET with user variable",
			query:    "SET @wallet_id = 23916",
			expected: models.OperationSet,
		},
		{
			name:     "SET with multiple variables",
			query:    "SET @a = 1, @b = 2",
			expected: models.OperationSet,
		},
		{
			name:     "SELECT query",
			query:    "SELECT * FROM users",
			expected: models.OperationSelect,
		},
		{
			name:     "UPDATE query",
			query:    "UPDATE users SET name = 'test'",
			expected: models.OperationUpdate,
		},
		{
			name:     "INSERT query",
			query:    "INSERT INTO users VALUES (1)",
			expected: models.OperationInsert,
		},
		{
			name:     "DELETE query",
			query:    "DELETE FROM users WHERE id = 1",
			expected: models.OperationDelete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectOperationType(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}
