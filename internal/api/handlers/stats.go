package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/querybase/internal/service"
)

// StatsHandler handles HTTP requests for dashboard statistics
type StatsHandler struct {
	statsService *service.StatsService
}

// NewStatsHandler creates a new StatsHandler
func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
	}
}

// GetDashboardStats returns dashboard statistics for the current user
func (h *StatsHandler) GetDashboardStats(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	role := c.GetString("role")
	isAdmin := role == "admin"

	stats, err := h.statsService.GetDashboardStats(c.Request.Context(), userID, isAdmin)
	if err != nil {
		log.Printf("Error fetching dashboard stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
