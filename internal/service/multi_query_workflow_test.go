package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/querybase/internal/models"
)

// ---------------------------------------------------------------------------
// ParseMultipleQueries
// ---------------------------------------------------------------------------

func TestParseMultipleQueries_SingleStatement(t *testing.T) {
	result := ParseMultipleQueries("SELECT * FROM users")
	require.Len(t, result.Statements, 1)
	assert.Equal(t, "SELECT * FROM users", result.Statements[0].QueryText)
	assert.Equal(t, models.OperationSelect, result.Statements[0].OperationType)
	assert.Equal(t, 0, result.Statements[0].Sequence)
}

func TestParseMultipleQueries_MultipleStatements(t *testing.T) {
	sql := "SELECT * FROM users; UPDATE users SET active=1 WHERE id=1; DELETE FROM logs WHERE id=2"
	result := ParseMultipleQueries(sql)
	require.Len(t, result.Statements, 3)
	assert.Equal(t, models.OperationSelect, result.Statements[0].OperationType)
	assert.Equal(t, models.OperationUpdate, result.Statements[1].OperationType)
	assert.Equal(t, models.OperationDelete, result.Statements[2].OperationType)
	assert.Equal(t, 0, result.Statements[0].Sequence)
	assert.Equal(t, 1, result.Statements[1].Sequence)
	assert.Equal(t, 2, result.Statements[2].Sequence)
}

func TestParseMultipleQueries_SemicolonInStringLiteral(t *testing.T) {
	// Semicolons inside string literals must NOT split statements
	sql := "INSERT INTO logs (msg) VALUES ('hello; world'); SELECT 1"
	result := ParseMultipleQueries(sql)
	require.Len(t, result.Statements, 2)
	assert.Contains(t, result.Statements[0].QueryText, "hello; world")
	assert.Equal(t, models.OperationSelect, result.Statements[1].OperationType)
}

func TestParseMultipleQueries_EmptyInput(t *testing.T) {
	result := ParseMultipleQueries("")
	assert.Len(t, result.Statements, 0)
	result2 := ParseMultipleQueries("   ")
	assert.Len(t, result2.Statements, 0)
}

func TestParseMultipleQueries_TrailingSemicolon(t *testing.T) {
	result := ParseMultipleQueries("SELECT 1;")
	require.Len(t, result.Statements, 1)
	assert.Equal(t, "SELECT 1", result.Statements[0].QueryText)
}

// ---------------------------------------------------------------------------
// ValidateMultiQuery — transaction control blocking
// ---------------------------------------------------------------------------

func TestValidateMultiQuery_BlocksTransactionControl(t *testing.T) {
	blocked := []struct {
		name string
		sql  string
	}{
		{"BEGIN", "BEGIN; SELECT 1"},
		{"COMMIT", "SELECT 1; COMMIT"},
		{"ROLLBACK", "SELECT 1; ROLLBACK"},
		{"START TRANSACTION", "START TRANSACTION; SELECT 1"},
		{"BEGIN alone", "BEGIN"},
	}

	for _, tt := range blocked {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateMultiQuery(tt.sql)
			assert.NotEmpty(t, result.Errors, "expected validation errors for %q", tt.sql)
		})
	}
}

func TestValidateMultiQuery_AllowsNormalStatements(t *testing.T) {
	sql := "SELECT 1; INSERT INTO t VALUES (1); UPDATE t SET x=1; DELETE FROM t WHERE id=1"
	result := ValidateMultiQuery(sql)
	assert.Empty(t, result.Errors)
	assert.Len(t, result.Statements, 4)
}

// ---------------------------------------------------------------------------
// RequiresApproval helper
// ---------------------------------------------------------------------------

func TestRequiresApproval_OperationCoverage(t *testing.T) {
	tests := []struct {
		op       models.OperationType
		expected bool
	}{
		{models.OperationSelect, false},
		{models.OperationSet, false},
		{models.OperationInsert, true},
		{models.OperationUpdate, true},
		{models.OperationDelete, true},
		{models.OperationCreateTable, true},
		{models.OperationDropTable, true},
		{models.OperationAlterTable, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.op), func(t *testing.T) {
			assert.Equal(t, tt.expected, RequiresApproval(tt.op))
		})
	}
}

// ---------------------------------------------------------------------------
// CreateMultiQueryTransaction — statement record creation
// ---------------------------------------------------------------------------

func TestCreateMultiQueryTransaction_CreatesStatements(t *testing.T) {
	db := setupWorkflowDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	svc := NewMultiQueryService(db, queryService, nil, nil)

	user := createTestUser(t, db, models.RoleAdmin)
	ds := createTestDataSource(t, db)

	queryTexts := []string{
		"SELECT * FROM users",
		"UPDATE users SET active=1 WHERE id=1",
		"DELETE FROM logs WHERE created_at < '2024-01-01'",
	}

	txn, err := svc.CreateMultiQueryTransaction(nil, nil, ds.ID, queryTexts, user.ID)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, txn.ID)
	assert.Equal(t, models.TransactionStatusActive, txn.Status)
	assert.True(t, txn.IsMultiQuery)
	assert.Equal(t, 3, txn.StatementCount)

	// Verify statement records were persisted
	stmts, err := svc.GetMultiQueryStatements(nil, txn.ID)
	require.NoError(t, err)
	require.Len(t, stmts, 3)

	assert.Equal(t, 0, stmts[0].Sequence)
	assert.Equal(t, models.OperationSelect, stmts[0].OperationType)
	assert.Equal(t, models.StatementStatusPending, stmts[0].Status)

	assert.Equal(t, 1, stmts[1].Sequence)
	assert.Equal(t, models.OperationUpdate, stmts[1].OperationType)

	assert.Equal(t, 2, stmts[2].Sequence)
	assert.Equal(t, models.OperationDelete, stmts[2].OperationType)
}

func TestCreateMultiQueryTransaction_SingleStatement(t *testing.T) {
	db := setupWorkflowDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	svc := NewMultiQueryService(db, queryService, nil, nil)

	user := createTestUser(t, db, models.RoleAdmin)
	ds := createTestDataSource(t, db)

	txn, err := svc.CreateMultiQueryTransaction(nil, nil, ds.ID, []string{"SELECT 1"}, user.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, txn.StatementCount)

	stmts, err := svc.GetMultiQueryStatements(nil, txn.ID)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// ---------------------------------------------------------------------------
// RollbackMultiQuery
// ---------------------------------------------------------------------------

func TestRollbackMultiQuery_ActiveTransaction(t *testing.T) {
	db := setupWorkflowDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	svc := NewMultiQueryService(db, queryService, nil, nil)

	user := createTestUser(t, db, models.RoleAdmin)
	ds := createTestDataSource(t, db)

	txn, err := svc.CreateMultiQueryTransaction(nil, nil, ds.ID, []string{"UPDATE t SET x=1"}, user.ID)
	require.NoError(t, err)

	err = svc.RollbackMultiQuery(nil, txn.ID)
	require.NoError(t, err)

	var updated models.QueryTransaction
	require.NoError(t, db.First(&updated, "id = ?", txn.ID).Error)
	assert.Equal(t, models.TransactionStatusRolledBack, updated.Status)
	assert.NotNil(t, updated.CompletedAt)
}

func TestRollbackMultiQuery_NonActiveTransaction(t *testing.T) {
	db := setupWorkflowDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	svc := NewMultiQueryService(db, queryService, nil, nil)

	user := createTestUser(t, db, models.RoleAdmin)
	ds := createTestDataSource(t, db)

	// Create then immediately rollback to get a rolled-back transaction
	txn, err := svc.CreateMultiQueryTransaction(nil, nil, ds.ID, []string{"UPDATE t SET x=1"}, user.ID)
	require.NoError(t, err)
	require.NoError(t, svc.RollbackMultiQuery(nil, txn.ID))

	// Rollback again should fail
	err = svc.RollbackMultiQuery(nil, txn.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
}

func TestRollbackMultiQuery_NonExistentTransaction(t *testing.T) {
	db := setupWorkflowDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	svc := NewMultiQueryService(db, queryService, nil, nil)

	err := svc.RollbackMultiQuery(nil, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ---------------------------------------------------------------------------
// GetMultiQueryStatements
// ---------------------------------------------------------------------------

func TestGetMultiQueryStatements_OrderedBySequence(t *testing.T) {
	db := setupWorkflowDB(t)
	queryService := NewQueryService(db, "test-encryption-key-32-chars-long!", nil, nil)
	svc := NewMultiQueryService(db, queryService, nil, nil)

	user := createTestUser(t, db, models.RoleAdmin)
	ds := createTestDataSource(t, db)

	queries := []string{"SELECT 1", "SELECT 2", "SELECT 3", "SELECT 4", "SELECT 5"}
	txn, err := svc.CreateMultiQueryTransaction(nil, nil, ds.ID, queries, user.ID)
	require.NoError(t, err)

	stmts, err := svc.GetMultiQueryStatements(nil, txn.ID)
	require.NoError(t, err)
	require.Len(t, stmts, 5)

	for i, stmt := range stmts {
		assert.Equal(t, i, stmt.Sequence, "statements must be returned in sequence order")
	}
}
