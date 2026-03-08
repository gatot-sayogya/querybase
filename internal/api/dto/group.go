package dto

// CreateGroupRequest represents a create group request
type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateGroupRequest represents an update group request
type UpdateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GroupResponse represents a group response
type GroupResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// GroupDetailResponse represents a detailed group response with users
type GroupDetailResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	Users       []UserInGroupResp `json:"users,omitempty"`
}

// UserInGroupResp represents a user in a group
type UserInGroupResp struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

type AddUserToGroupRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

// RemoveUserFromGroupRequest represents a remove user from group request
type RemoveUserFromGroupRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

// GroupDataSourcePermissionRequest represents a request to update a group's permission for a single data source
type GroupDataSourcePermissionRequest struct {
	DataSourceID string `json:"data_source_id" binding:"required,uuid"`
	CanRead      bool   `json:"can_read"`
	CanWrite     bool   `json:"can_write"`
	CanApprove   bool   `json:"can_approve"`
}

// GroupDataSourcePermissionResponse represents a group's permission for a single data source
type GroupDataSourcePermissionResponse struct {
	DataSourceID   string `json:"data_source_id"`
	DataSourceName string `json:"data_source_name"`
	GroupID        string `json:"group_id"`
	CanRead        bool   `json:"can_read"`
	CanWrite       bool   `json:"can_write"`
	CanApprove     bool   `json:"can_approve"`
}
