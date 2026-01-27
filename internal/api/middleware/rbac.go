package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/yourorg/querybase/internal/models"
)

// RequireAdmin checks if user has admin role
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != string(models.RoleAdmin) {
			c.JSON(403, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
