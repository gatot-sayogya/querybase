package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/yourorg/querybase/internal/models"
	"github.com/yourorg/querybase/internal/queue"
	"github.com/yourorg/querybase/internal/service"
	"gorm.io/gorm"
)

// SchemaHandler handles schema inspection endpoints
type SchemaHandler struct {
	db           *gorm.DB
	schemaService *service.SchemaService
}

// NewSchemaHandler creates a new schema handler
func NewSchemaHandler(db *gorm.DB, schemaService *service.SchemaService) *SchemaHandler {
	return &SchemaHandler{
		db:           db,
		schemaService: schemaService,
	}
}

// GetDatabaseSchema returns the complete schema for a data source
func (h *SchemaHandler) GetDatabaseSchema(c *gin.Context) {
	dataSourceID := c.Param("id")

	// Get data source to check last sync time
	var dataSource models.DataSource
	if err := h.db.Where("id = ?", dataSourceID).First(&dataSource).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data source not found"})
		return
	}

	// Check if schema is fresh (synced within last 5 minutes)
	isCached := false
	if dataSource.LastSchemaSync != nil {
		isCached = time.Now().Sub(*dataSource.LastSchemaSync) < 5*time.Minute
	}

	// Fetch schema (may use cache if fresh)
	schema, err := h.schemaService.GetSchema(c.Request.Context(), dataSourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add cache headers
	if isCached {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}
	if dataSource.LastSchemaSync != nil {
		c.Header("X-Last-Sync", dataSource.LastSchemaSync.Format(time.RFC3339))
	}

	c.JSON(http.StatusOK, gin.H{
		"schema":    schema,
		"last_sync": dataSource.LastSchemaSync,
		"is_cached": isCached,
		"is_healthy": dataSource.IsHealthy,
		"data_source": gin.H{
			"id":   dataSource.ID,
			"name": dataSource.Name,
			"type": dataSource.Type,
		},
	})
}

// GetTables returns a list of tables for a data source
func (h *SchemaHandler) GetTables(c *gin.Context) {
	dataSourceID := c.Param("id")

	tables, err := h.schemaService.GetTables(c, dataSourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tables": tables,
		"total":  len(tables),
	})
}

// GetTableDetails returns detailed information about a specific table
func (h *SchemaHandler) GetTableDetails(c *gin.Context) {
	dataSourceID := c.Param("id")
	tableName := c.Query("table")

	if tableName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table parameter is required"})
		return
	}

	table, err := h.schemaService.GetTableColumns(c, dataSourceID, tableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, table)
}

// SearchTables searches for tables by name
func (h *SchemaHandler) SearchTables(c *gin.Context) {
	dataSourceID := c.Param("id")
	searchTerm := c.Query("q")

	if searchTerm == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search term 'q' parameter is required"})
		return
	}

	tables, err := h.schemaService.SearchTables(c, dataSourceID, searchTerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tables": tables,
		"total":  len(tables),
	})
}

// SyncSchema forces an immediate schema refresh for a data source
func (h *SchemaHandler) SyncSchema(c *gin.Context) {
	dataSourceID := c.Param("id")

	// Get user from context
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Create Redis client for Asynq
	redisAddr := fmt.Sprintf("%s:%d", "localhost", 6379) // TODO: from config
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})

	// Enqueue immediate schema sync task
	info, err := queue.EnqueueSchemaSync(asynqClient, dataSourceID, true) // forceRefresh = true
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue sync task"})
		return
	}

	// Return the current schema immediately (will be updated shortly by worker)
	schema, err := h.schemaService.GetSchema(c.Request.Context(), dataSourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get data source for metadata
	var dataSource models.DataSource
	if err := h.db.Where("id = ?", dataSourceID).First(&dataSource).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Schema sync initiated",
			"task_id": info.ID,
			"schema":  schema,
			"data_source": gin.H{
				"id":   dataSource.ID,
				"name": dataSource.Name,
				"type": dataSource.Type,
			},
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Schema sync initiated",
			"task_id": info.ID,
			"schema":  schema,
		})
	}
}
