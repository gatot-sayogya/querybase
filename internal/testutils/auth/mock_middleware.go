package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/yourorg/querybase/internal/auth"
	"github.com/yourorg/querybase/internal/models"
)

func SetMockUser(c *gin.Context, userID, email, role string) {
	c.Set("user_id", userID)
	c.Set("email", email)
	c.Set("role", role)
}

func MockAdminContext(c *gin.Context) {
	SetMockUser(c, TestAdminUserID, TestAdminEmail, string(models.RoleAdmin))
}

func MockUserContext(c *gin.Context) {
	SetMockUser(c, TestUserID, TestUserEmail, string(models.RoleUser))
}

func MockViewerContext(c *gin.Context) {
	SetMockUser(c, TestViewerUserID, TestViewerEmail, string(models.RoleViewer))
}

func MockAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", TestUserID)
		c.Set("email", TestUserEmail)
		c.Set("role", string(models.RoleUser))
		c.Next()
	}
}

func MockAuthMiddlewareWithUser(userID, email, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetMockUser(c, userID, email, role)
		c.Next()
	}
}

func SetupTestRouterWithAuth(jwtManager *auth.JWTManager) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenString := ""
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString = authHeader[7:]
			}

			if tokenString != "" {
				claims, err := jwtManager.ValidateToken(tokenString)
				if err == nil {
					c.Set("user_id", claims.UserID.String())
					c.Set("email", claims.Email)
					c.Set("role", claims.Role)
				}
			}
		}
		c.Next()
	})

	return router
}
