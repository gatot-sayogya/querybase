package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QueryStatus represents the status of a query
type QueryStatus string

const (
	StatusPending   QueryStatus = "pending"
	StatusRunning   QueryStatus = "running"
	StatusCompleted QueryStatus = "completed"
	StatusFailed    QueryStatus = "failed"
)

// Compatibility constants with QueryStatus prefix
const (
	QueryStatusPending   = StatusPending
	QueryStatusRunning   = StatusRunning
	QueryStatusCompleted = StatusCompleted
	QueryStatusFailed    = StatusFailed
)

// OperationType represents the type of SQL operation
type OperationType string

const (
	OperationSelect     OperationType = "select"
	OperationInsert     OperationType = "insert"
	OperationUpdate     OperationType = "update"
	OperationDelete     OperationType = "delete"
	OperationCreateTable OperationType = "create_table"
	OperationDropTable  OperationType = "drop_table"
	OperationAlterTable OperationType = "alter_table"
)

// Compatibility constants with OperationType prefix
const (
	OperationTypeSelect     = OperationSelect
	OperationTypeInsert     = OperationInsert
	OperationTypeUpdate     = OperationUpdate
	OperationTypeDelete     = OperationDelete
	OperationTypeCreateTable = OperationCreateTable
	OperationTypeDropTable  = OperationDropTable
	OperationTypeAlterTable = OperationAlterTable
)

// Query represents a saved query
type Query struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	DataSourceID     uuid.UUID      `gorm:"type:uuid;not null" json:"data_source_id"`
	UserID           uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	QueryText        string         `gorm:"type:text;not null" json:"query_text"`
	OperationType    OperationType  `gorm:"not null" json:"operation_type"`
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	Status           QueryStatus    `gorm:"default:'pending'" json:"status"`
	RowCount         *int           `json:"row_count"`
	ExecutionTimeMs  *int           `json:"execution_time_ms"`
	ErrorMessage     string         `json:"error_message"`
	RequiresApproval bool           `gorm:"default:false" json:"requires_approval"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	DataSource       DataSource    `gorm:"foreignKey:DataSourceID" json:"data_source,omitempty"`
	User             User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Result           *QueryResult  `gorm:"foreignKey:QueryID" json:"result,omitempty"`
}

// TableName specifies the table name for Query
func (Query) TableName() string {
	return "queries"
}

// QueryResult represents stored query results (for result history)
type QueryResult struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	QueryID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"query_id"`
	Data        string    `gorm:"type:jsonb;not null" json:"data"`
	ColumnNames string    `gorm:"type:jsonb;not null" json:"column_names"` // JSON string of []string
	ColumnTypes string    `gorm:"type:jsonb;not null" json:"column_types"` // JSON string of []string
	RowCount    int       `gorm:"not null" json:"row_count"`
	StoredAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"stored_at"`
	SizeBytes   int       `json:"size_bytes"`
	Query       Query     `gorm:"foreignKey:QueryID" json:"query,omitempty"`
}

// TableName specifies the table name for QueryResult
func (QueryResult) TableName() string {
	return "query_results"
}

// QueryHistory represents execution history
type QueryHistory struct {
	ID             uuid.UUID     `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	QueryID        *uuid.UUID    `gorm:"type:uuid" json:"query_id"`
	UserID         uuid.UUID     `gorm:"type:uuid;not null" json:"user_id"`
	DataSourceID   uuid.UUID     `gorm:"type:uuid;not null" json:"data_source_id"`
	QueryText      string        `gorm:"type:text;not null" json:"query_text"`
	OperationType  OperationType `gorm:"not null" json:"operation_type"`
	Status         QueryStatus   `gorm:"not null" json:"status"`
	RowCount       *int          `json:"row_count"`
	ExecutionTimeMs *int         `json:"execution_time_ms"`
	ErrorMessage   string        `json:"error_message"`
	ExecutedAt     time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"executed_at"`
}

// TableName specifies the table name for QueryHistory
func (QueryHistory) TableName() string {
	return "query_history"
}
