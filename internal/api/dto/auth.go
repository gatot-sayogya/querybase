package dto

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse represents a user response
type UserResponse struct {
	ID       string   `json:"id"`
	Email    string   `json:"email"`
	Username string   `json:"username"`
	FullName string   `json:"full_name"`
	Role     string   `json:"role"`
	Groups   []string `json:"groups,omitempty"`
}

// CreateUserRequest represents a create user request
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name"`
	Role     string `json:"role" binding:"required,oneof=admin user viewer"`
}

// UpdateUserRequest represents an update user request
type UpdateUserRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	FullName string `json:"full_name"`
	Role     string `json:"role" binding:"omitempty,oneof=admin user viewer"`
	IsActive *bool  `json:"is_active"`
}

// ResetPasswordRequest for admin-initiated password reset
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ForgotPasswordRequest for self-service reset (future)
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordTokenRequest for completing self-service reset (future)
type ResetPasswordTokenRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
