package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GroupRole constants
const (
	GroupRoleViewer  = "viewer"
	GroupRoleMember  = "member"
	GroupRoleAnalyst = "analyst"
)

// GroupRolePolicy defines query type permissions for a role in a group
type GroupRolePolicy struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	GroupID      uuid.UUID  `gorm:"type:uuid;not null" json:"group_id"`
	DataSourceID *uuid.UUID `gorm:"type:uuid" json:"data_source_id"` // nil means it applies to all datasources for this group
	RoleInGroup  string     `gorm:"type:varchar(20);not null" json:"role_in_group"`
	AllowSelect  bool       `gorm:"default:true;not null" json:"allow_select"`
	AllowInsert  bool       `gorm:"default:false;not null" json:"allow_insert"`
	AllowUpdate  bool       `gorm:"default:false;not null" json:"allow_update"`
	AllowDelete  bool       `gorm:"default:false;not null" json:"allow_delete"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// TableName specifies the table name
func (GroupRolePolicy) TableName() string {
	return "group_role_policies"
}

// BeforeCreate will set a UUID rather than numeric ID.
func (g *GroupRolePolicy) BeforeCreate(tx *gorm.DB) (err error) {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return
}

// EffectivePermissions represents the flattened active permissions for a user across their groups
type EffectivePermissions struct {
	CanSelect  bool `json:"can_select"`
	CanInsert  bool `json:"can_insert"`
	CanUpdate  bool `json:"can_update"`
	CanDelete  bool `json:"can_delete"`
	CanRead    bool `json:"can_read"`
	CanWrite   bool `json:"can_write"`
	CanApprove bool `json:"can_approve"`
}
