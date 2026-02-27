package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourorg/querybase/internal/api/dto"
	"github.com/yourorg/querybase/internal/auth"
	"github.com/yourorg/querybase/internal/models"
	"gorm.io/gorm"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	db         *gorm.DB
	jwtManager *auth.JWTManager
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *gorm.DB, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtManager: jwtManager,
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format. Please provide both username and password.",
		})
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username and password are required.",
		})
		return
	}

	var user models.User
	if err := h.db.Preload("Groups").Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		// Don't reveal whether user exists - use generic message for security
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password. Please check your credentials and try again.",
		})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Your account has been deactivated. Please contact your administrator for assistance.",
		})
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		// Use same message as user not found to prevent user enumeration
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password. Please check your credentials and try again.",
		})
		return
	}

	token, err := h.jwtManager.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to complete login. Please try again or contact support if the problem persists.",
		})
		return
	}

	// Extract group IDs
	groupIDs := make([]string, len(user.Groups))
	for i, group := range user.Groups {
		groupIDs[i] = group.ID.String()
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		Token: token,
		User: dto.UserResponse{
			ID:       user.ID.String(),
			Email:    user.Email,
			Username: user.Username,
			FullName: user.FullName,
			Role:     string(user.Role),
			Groups:   groupIDs,
		},
	})
}

// GetMe returns the current user
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := c.GetString("user_id")

	var user models.User
	if err := h.db.Preload("Groups").Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Extract group IDs
	groupIDs := make([]string, len(user.Groups))
	for i, group := range user.Groups {
		groupIDs[i] = group.ID.String()
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		ID:       user.ID.String(),
		Email:    user.Email,
		Username: user.Username,
		FullName: user.FullName,
		Role:     string(user.Role),
		Groups:   groupIDs,
	})
}

// CreateUser creates a new user (admin only)
func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Role:         models.UserRole(req.Role),
		IsActive:     true,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	c.JSON(http.StatusCreated, dto.UserResponse{
		ID:       user.ID.String(),
		Email:    user.Email,
		Username: user.Username,
		FullName: user.FullName,
		Role:     string(user.Role),
	})
}

// ListUsers returns all users
func (h *AuthHandler) ListUsers(c *gin.Context) {
	var users []models.User
	// Exclude users with a zero/nil UUID — these are corrupted records created before
	// the BeforeCreate UUID hook was added.
	if err := h.db.Where("id != ?", "00000000-0000-0000-0000-000000000000").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	response := make([]dto.UserResponse, len(users))
	for i, user := range users {
		response[i] = dto.UserResponse{
			ID:       user.ID.String(),
			Email:    user.Email,
			Username: user.Username,
			FullName: user.FullName,
			Role:     string(user.Role),
		}
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUser updates a user
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.FullName != "" {
		updates["full_name"] = req.FullName
	}
	if req.Role != "" {
		updates["role"] = models.UserRole(req.Role)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	updates["updated_at"] = time.Now()

	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		ID:       user.ID.String(),
		Email:    user.Email,
		Username: user.Username,
		FullName: user.FullName,
		Role:     string(user.Role),
	})
}

// GetUser retrieves a user by ID
func (h *AuthHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := h.db.Preload("Groups").Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get group roles
	var userGroups []models.UserGroup
	h.db.Where("user_id = ?", userID).Find(&userGroups)
	roleMap := make(map[string]string)
	for _, ug := range userGroups {
		roleMap[ug.GroupID.String()] = ug.RoleInGroup
	}

	groupDetails := make([]dto.UserGroupDetail, len(user.Groups))
	for i, group := range user.Groups {
		groupDetails[i] = dto.UserGroupDetail{
			GroupID:     group.ID.String(),
			GroupName:   group.Name,
			RoleInGroup: roleMap[group.ID.String()],
		}
	}

	c.JSON(http.StatusOK, dto.UserDetailResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Username:  user.Username,
		FullName:  user.FullName,
		Role:      string(user.Role),
		AvatarURL: user.AvatarURL,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
		Groups:    groupDetails,
	})
}

// DeleteUser deletes a user (admin only)
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	currentUserID := c.GetString("user_id")

	// Prevent self-deletion
	if userID == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Soft delete
	if err := h.db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ChangePassword changes a user's password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify current password
	if !auth.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password
	if err := h.db.Model(&user).Update("password_hash", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// ResetUserPassword allows admins to reset any user's password
func (h *AuthHandler) ResetUserPassword(c *gin.Context) {
	userID := c.Param("id")
	currentUserID := c.GetString("user_id")

	// Prevent resetting own password (use change password instead)
	if userID == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Use change password endpoint to update your own password"})
		return
	}

	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password
	if err := h.db.Model(&user).Update("password_hash", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// GetUserGroups retrieves groups for a specific user
func (h *AuthHandler) GetUserGroups(c *gin.Context) {
	userID := c.Param("id")

	var userGroups []models.UserGroup
	if err := h.db.Preload("Group").Where("user_id = ?", userID).Find(&userGroups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user groups"})
		return
	}

	response := make([]dto.UserGroupDetail, len(userGroups))
	for i, ug := range userGroups {
		response[i] = dto.UserGroupDetail{
			GroupID:     ug.GroupID.String(),
			GroupName:   ug.Group.Name,
			RoleInGroup: ug.RoleInGroup,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": response,
		"total":  len(response),
	})
}

// AssignUserGroups updates the groups for a user
func (h *AuthHandler) AssignUserGroups(c *gin.Context) {
	userID := c.Param("id")

	var req dto.AssignUserGroupsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Begin transaction
	tx := h.db.Begin()

	// 1. Delete existing memberships
	if err := tx.Where("user_id = ?", uID).Delete(&models.UserGroup{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear existing groups"})
		return
	}

	// 2. Insert new memberships
	if len(req.Groups) > 0 {
		userGroups := make([]models.UserGroup, len(req.Groups))
		for i, gDetail := range req.Groups {
			gID, err := uuid.Parse(gDetail.GroupID)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID: " + gDetail.GroupID})
				return
			}
			userGroups[i] = models.UserGroup{
				UserID:      uID,
				GroupID:     gID,
				RoleInGroup: gDetail.RoleInGroup,
			}
		}

		if err := tx.Create(&userGroups).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign groups"})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "User groups assigned successfully"})
}
