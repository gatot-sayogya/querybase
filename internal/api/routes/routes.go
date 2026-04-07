package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/querybase/internal/api/handlers"
	"github.com/yourorg/querybase/internal/api/middleware"
	"github.com/yourorg/querybase/internal/auth"
	"github.com/yourorg/querybase/internal/service"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, authHandler *handlers.AuthHandler, queryHandler *handlers.QueryHandler, approvalHandler *handlers.ApprovalHandler, dataSourceHandler *handlers.DataSourceHandler, groupHandler *handlers.GroupHandler, schemaHandler *handlers.SchemaHandler, webSocketHandler *handlers.WebSocketHandler, statsHandler *handlers.StatsHandler, multiQueryHandler *handlers.MultiQueryHandler, jwtManager *auth.JWTManager, blacklist *service.TokenBlacklistService) {
	// Serve static files from the "web/out" directory
	// This assumes the frontend has been built to this directory
	router.Use(func(c *gin.Context) {
		// Skip for API routes
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}

		// Serve static files
		fs := gin.Dir("web/out", true)
		fileServer := http.StripPrefix("/", http.FileServer(fs))

		// Check if file exists
		path := c.Request.URL.Path
		if path == "/" {
			path = "index.html"
		}

		// 1. Try with .html extension (priority for cleaning URLs like /dashboard -> dashboard.html)
		if !strings.HasSuffix(path, ".html") && path != "/" {
			htmlPath := path + ".html"
			if _, err := os.Stat(filepath.Join("web/out", htmlPath)); err == nil {
				c.Request.URL.Path = htmlPath
				fileServer.ServeHTTP(c.Writer, c.Request)
				c.Abort()
				return
			}
		}

		// 2. Try exact path (but avoid directory listing)
		fullPath := filepath.Join("web/out", path)
		stat, err := os.Stat(fullPath)
		if err == nil {
			if !stat.IsDir() {
				// It's a file, serve it
				fileServer.ServeHTTP(c.Writer, c.Request)
				c.Abort()
				return
			}

			// It's a directory, check for index.html
			if _, err := os.Stat(filepath.Join(fullPath, "index.html")); err == nil {
				fileServer.ServeHTTP(c.Writer, c.Request)
				c.Abort()
				return
			}

			// Directory without index.html -> Don't serve (fall through to SPA)
		}

		// Try with .html extension
		if !strings.HasSuffix(path, ".html") {
			htmlPath := path + ".html"
			_, err := os.Stat(filepath.Join("web/out", htmlPath))
			if err == nil {
				c.Request.URL.Path = htmlPath
				fileServer.ServeHTTP(c.Writer, c.Request)
				c.Abort()
				return
			}
		}

		// If file doesn't exist and it's not an API route, serve index.html (SPA Fallback)
		// This handles client-side routing
		if !strings.HasPrefix(path, "/api") {
			c.File("web/out/index.html")
			c.Abort()
			return
		}

		c.Next()
	})

	api := router.Group("/api/v1")
	{
		// Public routes
		authGroup := api.Group("/auth")
		{
			// Strict brute-force protection: 5 req/min per IP, burst of 3
			loginLimiter := middleware.RateLimiterMiddleware(middleware.StrictAuthRateLimitConfig())
			authGroup.POST("/login", loginLimiter, authHandler.Login)
			authGroup.POST("/refresh", authHandler.Refresh)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtManager, blacklist))
		{
			// User routes
			authGroupProtected := protected.Group("/auth")
			{
				authGroupProtected.GET("/me", authHandler.GetMe)
				authGroupProtected.POST("/change-password", authHandler.ChangePassword)
				authGroupProtected.POST("/logout", authHandler.Logout)
			}

			// Dashboard Stats
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/stats", statsHandler.GetDashboardStats)
			}

			// Admin routes
			admin := protected.Group("")
			admin.Use(middleware.RequireAdmin())
			{
				// Admin auth routes
				authAdminGroup := protected.Group("/auth")
				{
					authAdminGroup.POST("/users", authHandler.CreateUser)
					authAdminGroup.GET("/users", authHandler.ListUsers)
					authAdminGroup.GET("/users/:id", authHandler.GetUser)
					authAdminGroup.PUT("/users/:id", authHandler.UpdateUser)
					authAdminGroup.DELETE("/users/:id", authHandler.DeleteUser)
					authAdminGroup.POST("/users/:id/reset-password", authHandler.ResetUserPassword)
					authAdminGroup.GET("/users/:id/groups", authHandler.GetUserGroups)
					authAdminGroup.PUT("/users/:id/groups", authHandler.AssignUserGroups)
				}

				// Group routes
				groups := admin.Group("/groups")
				{
					groups.POST("", groupHandler.CreateGroup)
					groups.GET("", groupHandler.ListGroups)
					groups.GET("/:id", groupHandler.GetGroup)
					groups.PUT("/:id", groupHandler.UpdateGroup)
					groups.DELETE("/:id", groupHandler.DeleteGroup)
					groups.POST("/:id/members", groupHandler.AddUserToGroup)

					groups.DELETE("/:id/members/:uid", groupHandler.RemoveUserFromGroup)
					groups.GET("/:id/members", groupHandler.ListGroupUsers)

					// Group data source permissions
					groups.GET("/:id/datasource_permissions", groupHandler.GetGroupDataSourcePermissions)
					groups.PUT("/:id/datasource_permissions", groupHandler.SetGroupDataSourcePermission)
				}
			}

			// Query routes
			queries := protected.Group("/queries")
			{
				queries.POST("", queryHandler.ExecuteQuery)
				queries.POST("/save", queryHandler.SaveQuery)
				queries.GET("", queryHandler.ListQueries)
				queries.GET("/:id", queryHandler.GetQuery)
				queries.GET("/:id/results", queryHandler.GetQueryResults)
				queries.DELETE("/:id", queryHandler.DeleteQuery)

				// Query history routes
				queries.GET("/history", queryHandler.ListQueryHistory)

				// Query analysis routes
				queries.POST("/explain", queryHandler.ExplainQuery)
				queries.POST("/dry-run", queryHandler.DryRunDelete)

				// Query export routes
				queries.POST("/export", queryHandler.ExportQuery)

				// Query preview route (for DELETE/UPDATE)
				queries.POST("/preview", queryHandler.PreviewWriteQuery)

				// INSERT preview route
				queries.POST("/preview-insert", queryHandler.PreviewInsertQuery)

				// Multi-query routes
				queries.POST("/multi/preview", multiQueryHandler.PreviewMultiQuery)
				queries.POST("/multi/execute", multiQueryHandler.ExecuteMultiQuery)
				queries.GET("/multi/:id/statements", multiQueryHandler.GetMultiQueryStatements)
				queries.POST("/multi/:id/commit", multiQueryHandler.CommitMultiQuery)
				queries.POST("/multi/:id/rollback", multiQueryHandler.RollbackMultiQuery)
			}

			// Approval routes
			approvals := protected.Group("/approvals")
			{
				approvals.GET("", approvalHandler.ListApprovals)
				approvals.GET("/counts", approvalHandler.GetApprovalCounts)
				approvals.GET("/:id", approvalHandler.GetApproval)
				approvals.POST("/:id/review", approvalHandler.ReviewApproval)
				approvals.POST("/:id/transaction-start", approvalHandler.StartTransaction)

				// Comment routes
				approvals.POST("/:id/comments", approvalHandler.AddComment)
				approvals.GET("/:id/comments", approvalHandler.GetComments)
				approvals.DELETE("/:id/comments/:comment_id", approvalHandler.DeleteComment)
			}

			// Transaction routes
			transactions := protected.Group("/transactions")
			{
				transactions.POST("/:id/commit", approvalHandler.CommitTransaction)
				transactions.POST("/:id/rollback", approvalHandler.RollbackTransaction)
				transactions.GET("/:id", approvalHandler.GetTransactionStatus)
			}

			// Query validation route (AST-based, dialect-aware)
			protected.POST("/queries/validate", queryHandler.ValidateQuery)

			// Schema routes
			schemas := protected.Group("/datasources")
			{
				schemas.GET("/:id/schema", schemaHandler.GetDatabaseSchema)
				schemas.POST("/:id/sync", schemaHandler.SyncSchema)
				schemas.GET("/:id/tables", schemaHandler.GetTables)
				schemas.GET("/:id/table", schemaHandler.GetTableDetails)
				schemas.GET("/:id/search", schemaHandler.SearchTables)
			}

			// Data source routes
			datasources := protected.Group("/datasources")
			{
				datasources.GET("", dataSourceHandler.ListDataSources)
				datasources.GET("/:id", dataSourceHandler.GetDataSource)
				datasources.GET("/:id/permissions", dataSourceHandler.GetPermissions)
				datasources.POST("/:id/test", dataSourceHandler.TestConnection)
				datasources.GET("/:id/health", dataSourceHandler.CheckHealth)
				datasources.GET("/:id/approvers", approvalHandler.GetEligibleApprovers)
			}

			// Admin only data source routes
			admin = protected.Group("")
			admin.Use(middleware.RequireAdmin())
			{
				adminDatasources := admin.Group("/datasources")
				{
					adminDatasources.POST("", dataSourceHandler.CreateDataSource)
					adminDatasources.POST("/test", dataSourceHandler.TestConnectionWithParams)
					adminDatasources.PUT("/:id", dataSourceHandler.UpdateDataSource)
					adminDatasources.DELETE("/:id", dataSourceHandler.DeleteDataSource)
					adminDatasources.PUT("/:id/permissions", dataSourceHandler.SetPermissions)
					adminDatasources.POST("/:id/test-audit", dataSourceHandler.TestAuditCapability)
				}
			}

			// // Data source routes (to be implemented)
			// datasources := protected.Group("/datasources")
			// {
			//     datasources.GET("", datasourceHandler.ListDataSources)
			//     datasources.POST("", datasourceHandler.CreateDataSource)
			//     datasources.GET("/:id", datasourceHandler.GetDataSource)
			//     datasources.PUT("/:id", datasourceHandler.UpdateDataSource)
			//     datasources.DELETE("/:id", datasourceHandler.DeleteDataSource)
			//     datasources.POST("/:id/test", datasourceHandler.TestConnection)
			//     datasources.PUT("/:id/permissions", datasourceHandler.SetPermissions)
			// }
		}
	}

	// WebSocket endpoint (no auth required for simplicity, can be added later)
	router.GET("/ws", webSocketHandler.HandleWebSocket)
}
