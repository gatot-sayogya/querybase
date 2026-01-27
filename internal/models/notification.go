package models

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationApprovalRequest      NotificationType = "approval_request"
	NotificationApprovalStatusChange NotificationType = "approval_status_change"
	NotificationQueryResult          NotificationType = "query_result"
	NotificationError                NotificationType = "error"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
)

// Legacy constants for compatibility
const (
	NotificationPending = NotificationStatusPending
	NotificationSent    = NotificationStatusSent
	NotificationFailed  = NotificationStatusFailed
)

// NotificationConfig represents Google Chat webhook configuration
type NotificationConfig struct {
	ID                 uuid.UUID        `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	GroupID            uuid.UUID        `gorm:"type:uuid;not null" json:"group_id"`
	WebhookURL         string           `gorm:"type:text;not null" json:"webhook_url"`
	IsActive           bool             `gorm:"default:true" json:"is_active"`
	NotificationEvents []string         `gorm:"type:text[];not null;default:'{approval_request,approval_status_change,query_result}'" json:"notification_events"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	Group              Group            `gorm:"foreignKey:GroupID" json:"-"`
}

// TableName specifies the table name for NotificationConfig
func (NotificationConfig) TableName() string {
	return "notification_configs"
}

// Notification represents a notification to be sent
type Notification struct {
	ID                  uuid.UUID          `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	NotificationConfigID *uuid.UUID        `gorm:"type:uuid" json:"notification_config_id"`
	ApprovalRequestID   *uuid.UUID         `gorm:"type:uuid" json:"approval_request_id"`
	QueryID             *uuid.UUID         `gorm:"type:uuid" json:"query_id"`
	Type                NotificationType   `gorm:"not null" json:"type"`
	Status              NotificationStatus `gorm:"default:'pending'" json:"status"`
	Payload             string             `gorm:"type:jsonb;not null" json:"payload"`
	RetryCount          int                `gorm:"default:0" json:"retry_count"`
	LastError           string             `json:"last_error"`
	SentAt              *time.Time         `json:"sent_at"`
	CreatedAt           time.Time          `json:"created_at"`
}

// TableName specifies the table name for Notification
func (Notification) TableName() string {
	return "notifications"
}
