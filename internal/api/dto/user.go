package dto

// ChangePasswordRequest represents a change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// UserDetailResponse represents a detailed user response with groups
type UserDetailResponse struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	Username  string   `json:"username"`
	FullName  string   `json:"full_name"`
	Role      string   `json:"role"`
	AvatarURL string   `json:"avatar_url"`
	IsActive  bool     `json:"is_active"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	Groups    []string `json:"groups,omitempty"` // List of group IDs
}
