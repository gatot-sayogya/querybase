package dto

// ChangePasswordRequest represents a change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// UserDetailResponse represents a detailed user response with groups
type UserDetailResponse struct {
	ID        string            `json:"id"`
	Email     string            `json:"email"`
	Username  string            `json:"username"`
	FullName  string            `json:"full_name"`
	Role      string            `json:"role"`
	AvatarURL string            `json:"avatar_url"`
	IsActive  bool              `json:"is_active"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
	Groups    []UserGroupDetail `json:"groups,omitempty"`
}

// UserGroupDetail represents a group a user belongs to
type UserGroupDetail struct {
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
}

// AssignUserGroupsRequest is used when saving user groups selection
type AssignUserGroupsRequest struct {
	Groups []UserGroupDetail `json:"groups"`
}
