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

// AddUserToGroupRequest represents an add user to group request
type AddUserToGroupRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

// RemoveUserFromGroupRequest represents a remove user from group request
type RemoveUserFromGroupRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}
