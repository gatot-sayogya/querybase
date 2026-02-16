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
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, authHandler *handlers.AuthHandler, queryHandler *handlers.QueryHandler, approvalHandler *handlers.ApprovalHandler, dataSourceHandler *handlers.DataSourceHandler, groupHandler *handlers.GroupHandler, schemaHandler *handlers.SchemaHandler, webSocketHandler *handlers.WebSocketHandler, jwtManager *auth.JWTManager) {
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

		_, err := os.Stat(filepath.Join("web/out", path))
		if err == nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
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
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			// User routes
			auth := protected.Group("/auth")
			{
				auth.GET("/me", authHandler.GetMe)
				auth.POST("/change-password", authHandler.ChangePassword)
			}

			// Admin routes
			admin := protected.Group("")
			admin.Use(middleware.RequireAdmin())
			{
				auth.POST("/users", authHandler.CreateUser)
				auth.GET("/users", authHandler.ListUsers)
				auth.GET("/users/:id", authHandler.GetUser)
				auth.PUT("/users/:id", authHandler.UpdateUser)
				auth.DELETE("/users/:id", authHandler.DeleteUser)

				// Group routes
				groups := admin.Group("/groups")
				{
					groups.POST("", groupHandler.CreateGroup)
					groups.GET("", groupHandler.ListGroups)
					groups.GET("/:id", groupHandler.GetGroup)
					groups.PUT("/:id", groupHandler.UpdateGroup)
					groups.DELETE("/:id", groupHandler.DeleteGroup)
					groups.POST("/:id/users", groupHandler.AddUserToGroup)
					groups.DELETE("/:id/users", groupHandler.RemoveUserFromGroup)
					groups.GET("/:id/users", groupHandler.ListGroupUsers)
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
			}

			// Approval routes
			approvals := protected.Group("/approvals")
			{
				approvals.GET("", approvalHandler.ListApprovals)
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

			// Query validation route
			protected.POST("/queries/validate", approvalHandler.ValidateQuery)

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
					adminDatasources.PUT("/:id", dataSourceHandler.UpdateDataSource)
					adminDatasources.DELETE("/:id", dataSourceHandler.DeleteDataSource)
					adminDatasources.PUT("/:id/permissions", dataSourceHandler.SetPermissions)
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
