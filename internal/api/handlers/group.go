package handlers

import (
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

	role := req.RoleInGroup
	if role == "" {
		role = "viewer" // default
	}

	// Check if already in group
	var count int64
	h.db.Model(&models.UserGroup{}).Where("group_id = ? AND user_id = ?", gID, uID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "User is already in this group"})
		return
	}

	userGroup := models.UserGroup{
		UserID:      uID,
		GroupID:     gID,
		RoleInGroup: role,
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
			ID:          ug.UserID.String(),
			Email:       ug.User.Email,
			Username:    ug.User.Username,
			FullName:    ug.User.FullName,
			RoleInGroup: ug.RoleInGroup,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": len(users),
	})
}

// UpdateGroupMemberRole updates a user's role in a group
func (h *GroupHandler) UpdateGroupMemberRole(c *gin.Context) {
	groupID := c.Param("id")
	userID := c.Param("uid")

	var req dto.UpdateGroupMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Model(&models.UserGroup{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("role_in_group", req.RoleInGroup).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member role updated successfully"})
}

// GetRolePolicies retrieves all role policies for a group
func (h *GroupHandler) GetRolePolicies(c *gin.Context) {
	groupID := c.Param("id")

	var policies []models.GroupRolePolicy
	if err := h.db.Where("group_id = ?", groupID).Find(&policies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch policies"})
		return
	}

	response := make([]dto.GroupRolePolicyResponse, len(policies))
	for i, p := range policies {
		var dsID *string
		if p.DataSourceID != nil {
			id := p.DataSourceID.String()
			dsID = &id
		}
		response[i] = dto.GroupRolePolicyResponse{
			ID:           p.ID.String(),
			GroupID:      p.GroupID.String(),
			DataSourceID: dsID,
			RoleInGroup:  p.RoleInGroup,
			AllowSelect:  p.AllowSelect,
			AllowInsert:  p.AllowInsert,
			AllowUpdate:  p.AllowUpdate,
			AllowDelete:  p.AllowDelete,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"policies": response,
		"total":    len(response),
	})
}

// SetRolePolicy sets a role policy for a group
func (h *GroupHandler) SetRolePolicy(c *gin.Context) {
	groupID := c.Param("id")

	var req dto.GroupRolePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gID, err := uuid.Parse(groupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var dsID *uuid.UUID
	if req.DataSourceID != nil && *req.DataSourceID != "" {
		id, err := uuid.Parse(*req.DataSourceID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data source ID"})
			return
		}
		dsID = &id
	}

	// Upsert policy
	policy := models.GroupRolePolicy{
		ID:           uuid.New(),
		GroupID:      gID,
		DataSourceID: dsID,
		RoleInGroup:  req.RoleInGroup,
		AllowSelect:  req.AllowSelect,
		AllowInsert:  req.AllowInsert,
		AllowUpdate:  req.AllowUpdate,
		AllowDelete:  req.AllowDelete,
	}

	query := h.db.Where("group_id = ? AND role_in_group = ?", gID, req.RoleInGroup)
	if dsID != nil {
		query = query.Where("data_source_id = ?", dsID)
	} else {
		query = query.Where("data_source_id IS NULL")
	}

	var existing models.GroupRolePolicy
	if err := query.First(&existing).Error; err == nil {
		// Update
		policy.ID = existing.ID
		if err := h.db.Save(&policy).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role policy"})
			return
		}
	} else {
		// Create
		if err := h.db.Create(&policy).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role policy"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role policy saved successfully"})
}
