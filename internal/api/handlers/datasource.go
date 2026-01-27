package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/querybase/internal/models"
	"github.com/yourorg/querybase/internal/service"
	"gorm.io/gorm"
)

// DataSourceHandler handles data source endpoints
type DataSourceHandler struct {
	db                 *gorm.DB
	dataSourceService  *service.DataSourceService
}

// NewDataSourceHandler creates a new data source handler
func NewDataSourceHandler(db *gorm.DB, dataSourceService *service.DataSourceService) *DataSourceHandler {
	return &DataSourceHandler{
		db:                db,
		dataSourceService: dataSourceService,
	}
}

// CreateDataSource creates a new data source (admin only)
func (h *DataSourceHandler) CreateDataSource(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Type     string `json:"type" binding:"required,oneof=postgresql mysql"`
		Host     string `json:"host" binding:"required"`
		Port     int    `json:"port" binding:"required,min=1,max=65535"`
		Database string `json:"database" binding:"required"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := &service.CreateDataSourceInput{
		Name:     req.Name,
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		Database: req.Database,
		Username: req.Username,
		Password: req.Password,
	}

	dataSource, err := h.dataSourceService.CreateDataSource(c, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create data source"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         dataSource.ID.String(),
		"name":       dataSource.Name,
		"type":       string(dataSource.Type),
		"host":       dataSource.Host,
		"port":       dataSource.Port,
		"database":   dataSource.GetDatabase(),
		"username":   dataSource.Username,
		"is_active":  dataSource.IsActive,
		"created_at": dataSource.CreatedAt,
	})
}

// ListDataSources returns a list of data sources
func (h *DataSourceHandler) ListDataSources(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	dataSources, total, err := h.dataSourceService.ListDataSources(c, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data sources"})
		return
	}

	response := make([]gin.H, len(dataSources))
	for i, ds := range dataSources {
		response[i] = gin.H{
			"id":         ds.ID.String(),
			"name":       ds.Name,
			"type":       string(ds.Type),
			"host":       ds.Host,
			"port":       ds.Port,
			"database":   ds.GetDatabase(),
			"username":   ds.Username,
			"is_active":  ds.IsActive,
			"created_at": ds.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data_sources": response,
		"total":        total,
		"page":         page,
		"limit":        limit,
	})
}

// GetDataSource retrieves a single data source
func (h *DataSourceHandler) GetDataSource(c *gin.Context) {
	dataSourceID := c.Param("id")

	dataSource, err := h.dataSourceService.GetDataSource(c, dataSourceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Data source not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data source"})
		}
		return
	}

	// Get permissions
	permissions, _ := h.dataSourceService.GetPermissions(c, dataSourceID)

	perms := make([]gin.H, len(permissions))
	for i, perm := range permissions {
		perms[i] = gin.H{
			"group_id":   perm.GroupID.String(),
			"group_name": perm.Group.Name,
			"can_read":   perm.CanRead,
			"can_write":  perm.CanWrite,
			"can_approve": perm.CanApprove,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              dataSource.ID.String(),
		"name":            dataSource.Name,
		"type":            string(dataSource.Type),
		"host":            dataSource.Host,
		"port":            dataSource.Port,
		"database":        dataSource.GetDatabase(),
		"username":        dataSource.Username,
		"is_active":       dataSource.IsActive,
		"created_at":      dataSource.CreatedAt,
		"updated_at":      dataSource.UpdatedAt,
		"permissions":     perms,
	})
}

// UpdateDataSource updates a data source (admin only)
func (h *DataSourceHandler) UpdateDataSource(c *gin.Context) {
	dataSourceID := c.Param("id")

	var req struct {
		Name     string `json:"name"`
		Type     string `json:"type" binding:"omitempty,oneof=postgresql mysql"`
		Host     string `json:"host"`
		Port     int    `json:"port" binding:"omitempty,min=1,max=65535"`
		Database string `json:"database"`
		Username string `json:"username"`
		Password string `json:"password"`
		IsActive *bool  `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := &service.UpdateDataSourceInput{
		Name:     req.Name,
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		Database: req.Database,
		Username: req.Username,
		Password: req.Password,
		IsActive: req.IsActive,
	}

	dataSource, err := h.dataSourceService.UpdateDataSource(c, dataSourceID, input)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Data source not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update data source"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         dataSource.ID.String(),
		"name":       dataSource.Name,
		"type":       string(dataSource.Type),
		"host":       dataSource.Host,
		"port":       dataSource.Port,
		"database":   dataSource.GetDatabase(),
		"username":   dataSource.Username,
		"is_active":  dataSource.IsActive,
		"updated_at": dataSource.UpdatedAt,
	})
}

// DeleteDataSource deletes a data source (admin only)
func (h *DataSourceHandler) DeleteDataSource(c *gin.Context) {
	dataSourceID := c.Param("id")

	if err := h.dataSourceService.DeleteDataSource(c, dataSourceID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Data source not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data source"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data source deleted successfully"})
}

// TestConnection tests the connection to a data source
func (h *DataSourceHandler) TestConnection(c *gin.Context) {
	dataSourceID := c.Param("id")

	if err := h.dataSourceService.TestConnection(c, dataSourceID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Connection successful",
	})
}

// SetPermissions sets permissions for a group on a data source (admin only)
func (h *DataSourceHandler) SetPermissions(c *gin.Context) {
	dataSourceID := c.Param("id")

	var req struct {
		GroupID    string `json:"group_id" binding:"required"`
		CanRead    bool   `json:"can_read"`
		CanWrite   bool   `json:"can_write"`
		CanApprove bool   `json:"can_approve"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify group exists
	var group models.Group
	if err := h.db.First(&group, "id = ?", req.GroupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	permissions := &service.PermissionInput{
		CanRead:    req.CanRead,
		CanWrite:   req.CanWrite,
		CanApprove: req.CanApprove,
	}

	if err := h.dataSourceService.SetPermissions(c, dataSourceID, req.GroupID, permissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permissions updated successfully"})
}

// GetPermissions retrieves permissions for a data source
func (h *DataSourceHandler) GetPermissions(c *gin.Context) {
	dataSourceID := c.Param("id")

	permissions, err := h.dataSourceService.GetPermissions(c, dataSourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch permissions"})
		return
	}

	response := make([]gin.H, len(permissions))
	for i, perm := range permissions {
		response[i] = gin.H{
			"group_id":    perm.GroupID.String(),
			"group_name":  perm.Group.Name,
			"can_read":    perm.CanRead,
			"can_write":   perm.CanWrite,
			"can_approve": perm.CanApprove,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": response,
		"count":       len(response),
	})
}
