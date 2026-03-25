package database

import (
	"testing"

	"gorm.io/gorm"
)

// RunTestWithTransaction runs a test function within a database transaction.
// The transaction is automatically rolled back after the test completes,
// ensuring test isolation. This prevents test data from persisting.
func RunTestWithTransaction(t *testing.T, db *gorm.DB, fn func(tx *gorm.DB)) {
	t.Helper()

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin transaction: %v", tx.Error)
	}

	fn(tx)

	if r := tx.Rollback(); r.Error != nil {
		t.Logf("warning: rollback failed: %v", r.Error)
	}
}

// RunTestInTransaction runs a test function within a transaction and returns the transaction.
// The caller is responsible for rolling back the transaction.
// This is useful when you need to perform setup operations within the transaction
// before running the actual test.
func RunTestInTransaction(t *testing.T, db *gorm.DB, fn func(tx *gorm.DB)) *gorm.DB {
	t.Helper()

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin transaction: %v", tx.Error)
	}

	fn(tx)

	return tx
}

// CleanupTransaction rolls back a transaction created by RunTestInTransaction.
func CleanupTransaction(tx *gorm.DB) {
	if tx != nil {
		tx.Rollback()
	}
}
