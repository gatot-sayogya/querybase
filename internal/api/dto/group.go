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
	ID          string `json:"id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	FullName    string `json:"full_name"`
	RoleInGroup string `json:"role_in_group"`
}

type AddUserToGroupRequest struct {
	UserID      string `json:"user_id" binding:"required,uuid"`
	RoleInGroup string `json:"role_in_group" binding:"omitempty,oneof=viewer member analyst"`
}

// RemoveUserFromGroupRequest represents a remove user from group request
type RemoveUserFromGroupRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

// UpdateGroupMemberRoleRequest represents updating a user's role in a group
type UpdateGroupMemberRoleRequest struct {
	RoleInGroup string `json:"role_in_group" binding:"required,oneof=viewer member analyst"`
}

// GroupRolePolicyRequest represents request to set a group role policy
type GroupRolePolicyRequest struct {
	DataSourceID *string `json:"data_source_id" binding:"omitempty,uuid"`
	RoleInGroup  string  `json:"role_in_group" binding:"required,oneof=viewer member analyst"`
	AllowSelect  bool    `json:"allow_select"`
	AllowInsert  bool    `json:"allow_insert"`
	AllowUpdate  bool    `json:"allow_update"`
	AllowDelete  bool    `json:"allow_delete"`
}

// GroupRolePolicyResponse represents response for a group role policy
type GroupRolePolicyResponse struct {
	ID           string  `json:"id"`
	GroupID      string  `json:"group_id"`
	DataSourceID *string `json:"data_source_id"`
	RoleInGroup  string  `json:"role_in_group"`
	AllowSelect  bool    `json:"allow_select"`
	AllowInsert  bool    `json:"allow_insert"`
	AllowUpdate  bool    `json:"allow_update"`
	AllowDelete  bool    `json:"allow_delete"`
}
