package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DataSourceType represents the type of data source
type DataSourceType string

const (
	DataSourceTypePostgreSQL DataSourceType = "postgresql"
	DataSourceTypeMySQL      DataSourceType = "mysql"
)

// Legacy constants for compatibility
const (
	DataSourcePostgreSQL = DataSourceTypePostgreSQL
	DataSourceMySQL      = DataSourceTypeMySQL
)

// DataSource represents a database connection
type DataSource struct {
	ID                uuid.UUID                `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name              string                   `gorm:"not null" json:"name"`
	Type              DataSourceType           `gorm:"not null" json:"type"`
	Host              string                   `gorm:"not null" json:"host"`
	Port              int                      `gorm:"not null" json:"port"`
	DatabaseName      string                   `gorm:"not null" json:"database_name"`
	Username          string                   `gorm:"not null" json:"username"`
	EncryptedPassword string                   `gorm:"type:text;not null" json:"-"`
	ConnectionParams  string                   `gorm:"type:jsonb;default:'{}'" json:"connection_params"`
	IsActive          bool                     `gorm:"default:true" json:"is_active"`
	IsHealthy         bool                     `gorm:"default:true" json:"is_healthy"`
	LastSchemaSync    *time.Time               `json:"last_schema_sync"`
	LastHealthCheck   *time.Time               `json:"last_health_check"`
	CreatedBy         *uuid.UUID               `gorm:"type:uuid" json:"created_by"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
	DeletedAt         gorm.DeletedAt           `gorm:"index" json:"-"`
	Permissions       []DataSourcePermission   `gorm:"foreignKey:DataSourceID" json:"-"`
}

// TableName specifies the table name for DataSource
func (DataSource) TableName() string {
	return "data_sources"
}

// GetPassword returns the encrypted password
func (ds *DataSource) GetPassword() string {
	return ds.EncryptedPassword
}

// SetPassword sets the encrypted password
func (ds *DataSource) SetPassword(password string) {
	ds.EncryptedPassword = password
}

// GetDatabase returns the database name
func (ds *DataSource) GetDatabase() string {
	return ds.DatabaseName
}

// SetDatabase sets the database name
func (ds *DataSource) SetDatabase(name string) {
	ds.DatabaseName = name
}

// DataSourcePermission represents group permissions for a data source
type DataSourcePermission struct {
	ID           uuid.UUID   `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	DataSourceID uuid.UUID   `gorm:"type:uuid;not null" json:"data_source_id"`
	GroupID      uuid.UUID   `gorm:"type:uuid;not null" json:"group_id"`
	CanRead      bool        `gorm:"default:true" json:"can_read"`
	CanWrite     bool        `gorm:"default:false" json:"can_write"`
	CanApprove   bool        `gorm:"default:false" json:"can_approve"`
	CreatedAt    time.Time   `json:"created_at"`
	DataSource   DataSource  `gorm:"foreignKey:DataSourceID" json:"-"`
	Group        Group       `gorm:"foreignKey:GroupID" json:"-"`
}

// TableName specifies the table name for DataSourcePermission
func (DataSourcePermission) TableName() string {
	return "data_source_permissions"
}
