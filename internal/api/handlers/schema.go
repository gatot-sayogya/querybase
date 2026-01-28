package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

	schema, err := h.schemaService.GetSchema(c, dataSourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schema)
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
