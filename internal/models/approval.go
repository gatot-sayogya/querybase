package models

import (
	"time"

	"github.com/google/uuid"
)

// ApprovalStatus represents the status of an approval request
type ApprovalStatus string

const (
	ApprovalStatusPending   ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

// ApprovalDecision represents an approval decision
type ApprovalDecision string

const (
	ApprovalDecisionApproved ApprovalDecision = "approved"
	ApprovalDecisionRejected ApprovalDecision = "rejected"
)

// ApprovalRequest represents a request for query approval
type ApprovalRequest struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	QueryID         *uuid.UUID     `gorm:"type:uuid" json:"query_id"`
	DirectQueryID   *uuid.UUID     `gorm:"type:uuid" json:"direct_query_id"`
	RequestedBy     uuid.UUID      `gorm:"type:uuid;not null" json:"requested_by"`
	OperationType   OperationType  `gorm:"not null" json:"operation_type"`
	QueryText       string         `gorm:"type:text;not null" json:"query_text"`
	DataSourceID    uuid.UUID      `gorm:"type:uuid;not null" json:"data_source_id"`
	Status          ApprovalStatus `gorm:"default:'pending'" json:"status"`
	RejectionReason string         `json:"rejection_reason"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	CompletedAt     *time.Time     `json:"completed_at"`

	// Foreign key relationships
	DataSource       DataSource       `gorm:"foreignKey:DataSourceID" json:"-"`
	RequestedByUser  User             `gorm:"foreignKey:RequestedBy" json:"-"`
	ApprovalReviews  []ApprovalReview `gorm:"foreignKey:ApprovalID;references:ID" json:"reviews"`
}

// TableName specifies the table name for ApprovalRequest
func (ApprovalRequest) TableName() string {
	return "approval_requests"
}

// ApprovalReview represents an approval decision
type ApprovalReview struct {
	ID              uuid.UUID       `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	ApprovalID      uuid.UUID       `gorm:"type:uuid;not null;column:approval_request_id" json:"approval_id"`
	ReviewerID      uuid.UUID       `gorm:"type:uuid;not null;column:reviewed_by" json:"reviewer_id"`
	Decision        ApprovalDecision `gorm:"type:varchar(20);not null" json:"decision"`
	Comments        string          `json:"comments"`
	ReviewedAt      time.Time       `gorm:"default:CURRENT_TIMESTAMP" json:"reviewed_at"`

	// Foreign key relationships
	Reviewer       User            `gorm:"foreignKey:ReviewerID" json:"-"`
	ApprovalRequest ApprovalRequest `gorm:"foreignKey:ApprovalID;references:ID" json:"-"`
}

// TableName specifies the table name for ApprovalReview
func (ApprovalReview) TableName() string {
	return "approval_reviews"
}

// TransactionStatus represents the status of a query transaction
type TransactionStatus string

const (
	TransactionStatusActive     TransactionStatus = "active"
	TransactionStatusCommitted TransactionStatus = "committed"
	TransactionStatusRolledBack TransactionStatus = "rolled_back"
	TransactionStatusFailed     TransactionStatus = "failed"
)

// QueryTransaction represents an active database transaction for preview
type QueryTransaction struct {
	ID               uuid.UUID        `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	ApprovalID       uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex" json:"approval_id"`
	DataSourceID     uuid.UUID        `gorm:"type:uuid;not null" json:"data_source_id"`
	QueryText        string           `gorm:"type:text;not null" json:"query_text"`
	StartedBy        uuid.UUID        `gorm:"type:uuid;not null" json:"started_by"`
	Status           TransactionStatus `gorm:"default:'active'" json:"status"`
	PreviewData      string           `gorm:"type:jsonb" json:"preview_data"` // JSON string of results
	AffectedRows     int              `json:"affected_rows"`
	ErrorMessage     string           `json:"error_message"`
	StartedAt        time.Time        `gorm:"default:CURRENT_TIMESTAMP" json:"started_at"`
	CompletedAt      *time.Time       `json:"completed_at"`

	// Foreign key relationships
	Approval     ApprovalRequest `gorm:"foreignKey:ApprovalID" json:"-"`
	DataSource   DataSource      `gorm:"foreignKey:DataSourceID" json:"-"`
	StartedByUser User            `gorm:"foreignKey:StartedBy" json:"-"`
}

// TableName specifies the table name for QueryTransaction
func (QueryTransaction) TableName() string {
	return "query_transactions"
}
