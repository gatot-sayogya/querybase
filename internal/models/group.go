package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserGroup represents the join table for user groups with roles
type UserGroup struct {
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	GroupID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"group_id"`
	JoinedAt time.Time `gorm:"autoCreateTime"`

	User  User  `gorm:"foreignKey:UserID"`
	Group Group `gorm:"foreignKey:GroupID"`
}

// TableName specifies the table name for UserGroup
func (UserGroup) TableName() string {
	return "user_groups"
}

// Group represents a group in the system
type Group struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null" json:"name"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Members     []UserGroup    `gorm:"foreignKey:GroupID" json:"members,omitempty"`
	Users       []User         `gorm:"many2many:user_groups;joinForeignKey:GroupID;joinReferences:UserID" json:"users,omitempty"`
}

// TableName specifies the table name for Group
func (Group) TableName() string {
	return "groups"
}

// BeforeCreate will set a UUID rather than numeric ID.
func (g *Group) BeforeCreate(tx *gorm.DB) (err error) {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return
}
