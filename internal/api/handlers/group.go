package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/querybase/internal/api/dto"
	"github.com/yourorg/querybase/internal/models"
	"gorm.io/gorm"
)

// GroupHandler handles group endpoints
type GroupHandler struct {
	db *gorm.DB
}

// NewGroupHandler creates a new group handler
func NewGroupHandler(db *gorm.DB) *GroupHandler {
	return &GroupHandler{
		db: db,
	}
}

// CreateGroup creates a new group (admin only)
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req dto.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group := models.Group{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.db.Create(&group).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Group already exists"})
		return
	}

	c.JSON(http.StatusCreated, dto.GroupResponse{
		ID:          group.ID.String(),
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ListGroups returns all groups with pagination
func (h *GroupHandler) ListGroups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	var groups []models.Group
	var total int64

	h.db.Model(&models.Group{}).Count(&total)

	if err := h.db.Offset(offset).Limit(limit).Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}

	response := make([]dto.GroupResponse, len(groups))
	for i, group := range groups {
		response[i] = dto.GroupResponse{
			ID:          group.ID.String(),
			Name:        group.Name,
			Description: group.Description,
			CreatedAt:   group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": response,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

// GetGroup retrieves a group by ID
func (h *GroupHandler) GetGroup(c *gin.Context) {
	groupID := c.Param("id")

	var group models.Group
	if err := h.db.Preload("Users").Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Get users in group
	users := make([]dto.UserInGroupResp, len(group.Users))
	for i, user := range group.Users {
		users[i] = dto.UserInGroupResp{
			ID:       user.ID.String(),
			Email:    user.Email,
			Username: user.Username,
			FullName: user.FullName,
		}
	}

	c.JSON(http.StatusOK, dto.GroupDetailResponse{
		ID:          group.ID.String(),
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Users:       users,
	})
}

// UpdateGroup updates a group (admin only)
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	groupID := c.Param("id")

	var req dto.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var group models.Group
	if err := h.db.Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	if err := h.db.Model(&group).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}

	// Reload to get updated data
	h.db.First(&group, groupID)

	c.JSON(http.StatusOK, dto.GroupResponse{
		ID:          group.ID.String(),
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeleteGroup deletes a group (admin only)
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	groupID := c.Param("id")

	var group models.Group
	if err := h.db.Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Soft delete
	if err := h.db.Delete(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

// AddUserToGroup adds a user to a group
func (h *GroupHandler) AddUserToGroup(c *gin.Context) {
	groupID := c.Param("id")

	var req dto.AddUserToGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if group exists
	var group models.Group
	if err := h.db.Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Check if user exists
	var user models.User
	if err := h.db.Where("id = ?", req.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user is already in group
	var count int64
	h.db.Table("user_groups").
		Where("group_id = ? AND user_id = ?", groupID, req.UserID).
		Count(&count)

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "User is already in this group"})
		return
	}

	// Add user to group
	if err := h.db.Model(&group).Association("Users").Append(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added to group successfully"})
}

// RemoveUserFromGroup removes a user from a group
func (h *GroupHandler) RemoveUserFromGroup(c *gin.Context) {
	groupID := c.Param("id")
	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id query parameter is required"})
		return
	}

	// Check if group exists
	var group models.Group
	if err := h.db.Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Check if user exists
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Remove user from group
	if err := h.db.Model(&group).Association("Users").Delete(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user from group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User removed from group successfully"})
}

// ListGroupUsers lists all users in a group
func (h *GroupHandler) ListGroupUsers(c *gin.Context) {
	groupID := c.Param("id")

	var group models.Group
	if err := h.db.Preload("Users").Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	users := make([]dto.UserInGroupResp, len(group.Users))
	for i, user := range group.Users {
		users[i] = dto.UserInGroupResp{
			ID:       user.ID.String(),
			Email:    user.Email,
			Username: user.Username,
			FullName: user.FullName,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": len(users),
	})
}
