package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// AddUserToGroup adds a user to a group with an optional role
func (h *GroupHandler) AddUserToGroup(c *gin.Context) {
	groupID := c.Param("id")

	var req dto.AddUserToGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gID, err := uuid.Parse(groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}
	uID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if already in group
	var count int64
	h.db.Model(&models.UserGroup{}).Where("group_id = ? AND user_id = ?", gID, uID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "User is already in this group"})
		return
	}

	userGroup := models.UserGroup{
		UserID:  uID,
		GroupID: gID,
	}

	if err := h.db.Create(&userGroup).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to group: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added to group successfully"})
}

// RemoveUserFromGroup removes a user from a group
func (h *GroupHandler) RemoveUserFromGroup(c *gin.Context) {
	groupID := c.Param("id")
	userID := c.Query("user_id")
	// Support parameter extraction for cleanly mapped routes
	if userIDFromParam := c.Param("uid"); userIDFromParam != "" {
		userID = userIDFromParam
	}

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	if err := h.db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.UserGroup{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user from group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User removed from group successfully"})
}

// ListGroupUsers lists all users in a group
func (h *GroupHandler) ListGroupUsers(c *gin.Context) {
	groupID := c.Param("id")

	var userGroups []models.UserGroup
	if err := h.db.Preload("User").Where("group_id = ?", groupID).Find(&userGroups).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group members not found"})
		return
	}

	users := make([]dto.UserInGroupResp, len(userGroups))
	for i, ug := range userGroups {
		users[i] = dto.UserInGroupResp{
			ID:       ug.UserID.String(),
			Email:    ug.User.Email,
			Username: ug.User.Username,
			FullName: ug.User.FullName,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": len(users),
	})
}

// GetGroupDataSourcePermissions retrieves all data source permissions for a specific group
func (h *GroupHandler) GetGroupDataSourcePermissions(c *gin.Context) {
	groupID := c.Param("id")

	var permissions []models.DataSourcePermission
	if err := h.db.Preload("DataSource").Where("group_id = ?", groupID).Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch permissions"})
		return
	}

	response := make([]dto.GroupDataSourcePermissionResponse, len(permissions))
	for i, p := range permissions {
		response[i] = dto.GroupDataSourcePermissionResponse{
			DataSourceID:   p.DataSourceID.String(),
			DataSourceName: p.DataSource.Name,
			GroupID:        p.GroupID.String(),
			CanRead:        p.CanRead,
			CanWrite:       p.CanWrite,
			CanApprove:     p.CanApprove,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": response,
		"count":       len(response),
	})
}

// SetGroupDataSourcePermission sets a specific data source permission for a group
func (h *GroupHandler) SetGroupDataSourcePermission(c *gin.Context) {
	groupID := c.Param("id")

	// ... (We will start replacing from the gin context parsing)
	var req dto.GroupDataSourcePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("[DEBUG] SetGroupDataSourcePermission HTTP PUT received for group %s and ds %s\n", groupID, req.DataSourceID)
	fmt.Printf("[DEBUG] Payload: Read=%v, Write=%v, Approve=%v\n", req.CanRead, req.CanWrite, req.CanApprove)

	gID, err := uuid.Parse(groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	dsID, err := uuid.Parse(req.DataSourceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data source ID"})
		return
	}

	permission := models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: dsID,
		GroupID:      gID,
		CanRead:      req.CanRead,
		CanWrite:     req.CanWrite,
		CanApprove:   req.CanApprove,
	}

	var existing models.DataSourcePermission
	if err := h.db.Where("group_id = ? AND data_source_id = ?", gID, dsID).First(&existing).Error; err == nil {
		fmt.Printf("[DEBUG] Found existing permission. ID: %s. Read=%v\n", existing.ID, existing.CanRead)
		// Record exists, explicitly update using map to avoid zero-value omission
		updateMap := map[string]interface{}{
			"can_read":    req.CanRead,
			"can_write":   req.CanWrite,
			"can_approve": req.CanApprove,
		}
		if err := h.db.Model(&existing).Updates(updateMap).Error; err != nil {
			fmt.Printf("[DEBUG] Failed to update: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update permission"})
			return
		}
		fmt.Printf("[DEBUG] Successfully ran Updates(map). New Read=%v\n", req.CanRead)
	} else {
		fmt.Printf("[DEBUG] Record does not exist, creating new one.\n")
		// Record does not exist, create new one
		if err := h.db.Create(&permission).Error; err != nil {
			fmt.Printf("[DEBUG] Failed to create: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create permission"})
			return
		}
		fmt.Printf("[DEBUG] Successfully ran Create(&permission).\n")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data source permission saved successfully"})
}
