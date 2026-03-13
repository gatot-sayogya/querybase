package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourorg/querybase/internal/models"
)

// TestAuditService_BuildCountQuery tests count query generation from write queries
func TestAuditService_BuildCountQuery(t *testing.T) {
	auditSvc := &AuditService{}

	tests := []struct {
		name        string
		query       string
		expected    string
		expectError bool
	}{
		{
			name:     "Simple DELETE with WHERE",
			query:    "DELETE FROM users WHERE id = 5",
			expected: "SELECT COUNT(*) FROM users WHERE id = 5",
		},
		{
			name:     "DELETE without WHERE",
			query:    "DELETE FROM users",
			expected: "SELECT COUNT(*) FROM users",
		},
		{
			name:     "UPDATE with WHERE",
			query:    "UPDATE users SET name = 'test' WHERE id > 10",
			expected: "SELECT COUNT(*) FROM users WHERE id > 10",
		},
		{
			name:     "UPDATE without WHERE",
			query:    "UPDATE users SET name = 'test'",
			expected: "SELECT COUNT(*) FROM users",
		},
		{
			name:     "Simple INSERT with VALUES",
			query:    "INSERT INTO users (name) VALUES ('test')",
			expected: "SELECT 1",
		},
		{
			name:     "INSERT with multiple VALUES",
			query:    "INSERT INTO users (name) VALUES ('a'),('b'),('c')",
			expected: "SELECT 3",
		},
		{
			name:     "INSERT from SELECT",
			query:    "INSERT INTO users_backup SELECT id, name FROM users WHERE active = true",
			expected: "SELECT COUNT(*) FROM users WHERE active = true",
		},
		{
			name:        "SELECT query (unsupported)",
			query:       "SELECT * FROM users",
			expectError: true,
		},
		{
			name:     "DELETE with trailing semicolon",
			query:    "DELETE FROM users WHERE id = 5;",
			expected: "SELECT COUNT(*) FROM users WHERE id = 5",
		},
		{
			name:     "UPDATE with trailing semicolon",
			query:    "UPDATE users SET name = 'test' WHERE id > 10;",
			expected: "SELECT COUNT(*) FROM users WHERE id > 10",
		},
		{
			name:     "Complex UPDATE with IN clause and semicolon",
			query:    "UPDATE product_masters SET conversion = 1 WHERE id IN (902594,711472);",
			expected: "SELECT COUNT(*) FROM product_masters WHERE id IN (902594,711472)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := auditSvc.buildCountQuery(tt.query)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestAuditService_ExtractTargetTable tests table name extraction from write queries
func TestAuditService_ExtractTargetTable(t *testing.T) {
	auditSvc := &AuditService{}

	tests := []struct {
		name        string
		query       string
		expected    string
		expectError bool
	}{
		{
			name:     "UPDATE table",
			query:    "UPDATE users SET name = 'test' WHERE id = 1",
			expected: "users",
		},
		{
			name:     "DELETE FROM table",
			query:    "DELETE FROM orders WHERE id = 5",
			expected: "orders",
		},
		{
			name:     "INSERT INTO table",
			query:    "INSERT INTO products (name, price) VALUES ('Widget', 9.99)",
			expected: "products",
		},
		{
			name:     "UPDATE with schema prefix",
			query:    "UPDATE public.users SET name = 'test' WHERE id = 1",
			expected: "public.users",
		},
		{
			name:        "SELECT query (unsupported)",
			query:       "SELECT * FROM users",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := auditSvc.extractTargetTable(tt.query)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestAuditService_ResolveAuditMode tests audit mode resolution based on capability
func TestAuditService_ResolveAuditMode(t *testing.T) {
	auditSvc := &AuditService{}

	tests := []struct {
		name       string
		requested  models.AuditMode
		capability models.AuditCapability
		expected   models.AuditMode
	}{
		{
			name:       "Full requested, full capability",
			requested:  models.AuditModeFull,
			capability: models.AuditCapabilityFull,
			expected:   models.AuditModeFull,
		},
		{
			name:       "Full requested, count_only capability -> degrades",
			requested:  models.AuditModeFull,
			capability: models.AuditCapabilityCountOnly,
			expected:   models.AuditModeCountOnly,
		},
		{
			name:       "Sample requested, full capability",
			requested:  models.AuditModeSample,
			capability: models.AuditCapabilityFull,
			expected:   models.AuditModeSample,
		},
		{
			name:       "Sample requested, count_only capability -> degrades",
			requested:  models.AuditModeSample,
			capability: models.AuditCapabilityCountOnly,
			expected:   models.AuditModeCountOnly,
		},
		{
			name:       "Empty requested, full capability -> defaults to full",
			requested:  "",
			capability: models.AuditCapabilityFull,
			expected:   models.AuditModeFull,
		},
		{
			name:       "CountOnly requested, full capability -> respects request",
			requested:  models.AuditModeCountOnly,
			capability: models.AuditCapabilityFull,
			expected:   models.AuditModeCountOnly,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auditSvc.ResolveAuditMode(tt.requested, tt.capability)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestAuditService_CheckCaution tests caution threshold checking
func TestAuditService_CheckCaution(t *testing.T) {
	auditSvc := &AuditService{}

	tests := []struct {
		name          string
		estimatedRows int
		threshold     int
		expectCaution bool
	}{
		{
			name:          "Below threshold",
			estimatedRows: 50,
			threshold:     1000,
			expectCaution: false,
		},
		{
			name:          "At threshold",
			estimatedRows: 1000,
			threshold:     1000,
			expectCaution: false,
		},
		{
			name:          "Above threshold",
			estimatedRows: 1500,
			threshold:     1000,
			expectCaution: true,
		},
		{
			name:          "Zero threshold defaults to 1000",
			estimatedRows: 1500,
			threshold:     0,
			expectCaution: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &models.DataSource{
				AuditRowThreshold: tt.threshold,
			}
			caution, msg := auditSvc.CheckCaution(tt.estimatedRows, ds)
			assert.Equal(t, tt.expectCaution, caution)
			if caution {
				assert.NotEmpty(t, msg)
			}
		})
	}
}
