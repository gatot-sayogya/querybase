package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Group represents a group in the system
type Group struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null" json:"name"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Users       []User         `gorm:"many2many:user_groups;" json:"users,omitempty"`
}

// TableName specifies the table name for Group
func (Group) TableName() string {
	return "groups"
}
