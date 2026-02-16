package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole represents user roles
type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleUser   UserRole = "user"
	RoleViewer UserRole = "viewer"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email"`
	Username     string         `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"not null" json:"-"`
	FullName     string         `json:"full_name"`
	Role         UserRole       `gorm:"not null;default:'user'" json:"role"`
	AvatarURL    string         `json:"avatar_url"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Groups       []Group        `gorm:"many2many:user_groups;" json:"groups,omitempty"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}
