package models

import (
	"time"

	"github.com/google/uuid"
)

// StatementStatus represents the status of an individual query statement
type StatementStatus string

const (
	StatementStatusPending StatementStatus = "pending"
	StatementStatusSuccess StatementStatus = "success"
	StatementStatusFailed  StatementStatus = "failed"
)

// QueryTransactionStatement represents an individual SQL statement within a multi-query transaction
type QueryTransactionStatement struct {
	ID              uuid.UUID       `gorm:"type:uuid;primary_key" json:"id"`
	TransactionID   uuid.UUID       `gorm:"type:uuid;not null;index" json:"transaction_id"`
	Sequence        int             `gorm:"not null" json:"sequence"`
	QueryText       string          `gorm:"type:text;not null" json:"query_text"`
	OperationType   OperationType   `gorm:"not null" json:"operation_type"`
	Status          StatementStatus `gorm:"default:'pending'" json:"status"`
	AffectedRows    int             `json:"affected_rows"`
	ErrorMessage    string          `json:"error_message"`
	PreviewData     string          `gorm:"type:jsonb" json:"preview_data"`
	BeforeData      string          `gorm:"type:jsonb" json:"before_data"`
	AfterData       string          `gorm:"type:jsonb" json:"after_data"`
	ExecutionTimeMs int             `json:"execution_time_ms"`
	CreatedAt       time.Time       `json:"created_at"`

	// Foreign key relationships
	Transaction QueryTransaction `gorm:"foreignKey:TransactionID" json:"-"`
}

// TableName specifies the table name for QueryTransactionStatement
func (QueryTransactionStatement) TableName() string {
	return "query_transaction_statements"
}
