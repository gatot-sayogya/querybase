package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/querybase/internal/api/handlers"
	"github.com/yourorg/querybase/internal/api/middleware"
	"github.com/yourorg/querybase/internal/api/routes"
	"github.com/yourorg/querybase/internal/auth"
	"github.com/yourorg/querybase/internal/config"
	"github.com/yourorg/querybase/internal/database"
	"github.com/yourorg/querybase/internal/service"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := config.Load("./config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database based on dialect
	var db *gorm.DB
	dialect := cfg.Database.Dialect
	if dialect == "" {
		dialect = "postgresql" // Default to PostgreSQL
	}

	switch dialect {
	case "mysql":
		db, err = connectToMySQL(&cfg.Database)
	default:
		db, err = database.NewPostgresConnection(&cfg.Database)
	}

	// Enable SQL logging for debugging
	db = db.Debug()

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	// Run migrations
	// Note: We're using manual SQL migrations instead of GORM AutoMigrate
	// To apply migrations, run: make migrate-up
	// if err := database.AutoMigrate(db); err != nil {
	// 	log.Fatalf("Failed to run migrations: %v", err)
	// }

	// Seed database with initial data
	if err := database.SeedData(db); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpireHours, cfg.JWT.Issuer)

	// Initialize services
	queryService := service.NewQueryService(db, cfg.JWT.Secret)
	approvalService := service.NewApprovalService(db, queryService)
	dataSourceService := service.NewDataSourceService(db, cfg.JWT.Secret)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, jwtManager)
	queryHandler := handlers.NewQueryHandler(db, queryService)
	approvalHandler := handlers.NewApprovalHandler(db, approvalService)
	dataSourceHandler := handlers.NewDataSourceHandler(db, dataSourceService)
	groupHandler := handlers.NewGroupHandler(db)

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create router without default middleware
	router := gin.New()

	// Add custom middleware
	router.Use(middleware.ErrorRecoveryMiddleware())
	router.Use(middleware.LoggingMiddleware())

	// Add CORS middleware
	// Use development config for now, change to ProdConfig for production
	corsConfig := middleware.DefaultConfig()
	// For development, you can also use: middleware.DevelopmentConfig()
	router.Use(middleware.CORSMiddleware(corsConfig))

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "QueryBase API is running",
		})
	})

	// Setup routes
	routes.SetupRoutes(router, authHandler, queryHandler, approvalHandler, dataSourceHandler, groupHandler, jwtManager)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	log.Printf("Starting QueryBase API server on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}

	os.Exit(0)
}

// connectToMySQL creates a MySQL connection using the database config
func connectToMySQL(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	mysqlCfg := &database.MySQLConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Database: cfg.Name,
		SSLMode:  cfg.SSLMode,
	}
	return database.NewMySQLConnection(mysqlCfg)
}

