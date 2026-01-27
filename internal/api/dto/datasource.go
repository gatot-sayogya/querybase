package dto

// CreateDataSourceRequest represents a create data source request
type CreateDataSourceRequest struct {
	Name             string                 `json:"name" binding:"required"`
	Type             string                 `json:"type" binding:"required,oneof=postgresql mysql"`
	Host             string                 `json:"host" binding:"required"`
	Port             int                    `json:"port" binding:"required,min=1,max=65535"`
	DatabaseName     string                 `json:"database_name" binding:"required"`
	Username         string                 `json:"username" binding:"required"`
	Password         string                 `json:"password" binding:"required"`
	ConnectionParams map[string]interface{} `json:"connection_params"`
}

// UpdateDataSourceRequest represents an update data source request
type UpdateDataSourceRequest struct {
	Name             string                 `json:"name"`
	Host             string                 `json:"host"`
	Port             *int                   `json:"port"`
	DatabaseName     string                 `json:"database_name"`
	Username         string                 `json:"username"`
	Password         string                 `json:"password"`
	ConnectionParams map[string]interface{} `json:"connection_params"`
	IsActive         *bool                  `json:"is_active"`
}

// DataSourceResponse represents a data source response
type DataSourceResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
}

// TestDataSourceRequest represents a test connection request
type TestDataSourceRequest struct {
	Type         string                 `json:"type" binding:"required,oneof=postgresql mysql"`
	Host         string                 `json:"host" binding:"required"`
	Port         int                    `json:"port" binding:"required,min=1,max=65535"`
	DatabaseName string                 `json:"database_name" binding:"required"`
	Username     string                 `json:"username" binding:"required"`
	Password     string                 `json:"password" binding:"required"`
	ConnectionParams map[string]interface{} `json:"connection_params"`
}

// SetDataSourcePermissionsRequest represents set permissions request
type SetDataSourcePermissionsRequest struct {
	GroupID    string `json:"group_id" binding:"required"`
	CanRead    bool   `json:"can_read"`
	CanWrite   bool   `json:"can_write"`
	CanApprove bool   `json:"can_approve"`
}
