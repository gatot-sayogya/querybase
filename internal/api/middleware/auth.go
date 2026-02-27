package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/querybase/internal/auth"
	"github.com/yourorg/querybase/internal/service"
)

// AuthMiddleware validates JWT tokens and checks blacklist
func AuthMiddleware(jwtManager *auth.JWTManager, blacklist *service.TokenBlacklistService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Check if token is blacklisted
		if blacklist != nil {
			isBlacklisted, _ := blacklist.IsBlacklisted(c.Request.Context(), claims.ID)
			if isBlacklisted {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
				c.Abort()
				return
			}
		}

		// Set user info in context (convert UUID to string)
		c.Set("user_id", claims.UserID.String())
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}
