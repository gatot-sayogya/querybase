package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/yourorg/querybase/internal/config"
	"github.com/yourorg/querybase/internal/database"
	"github.com/yourorg/querybase/internal/models"
	"github.com/yourorg/querybase/internal/queue"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := config.Load("./config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	log.Println("Worker connected to database successfully")

	// Create Redis connection for Asynq
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)

	// Create Asynq server
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			// Number of concurrent workers
			Concurrency: 5,

			// Use custom queue names
			Queues: map[string]int{
				"queries":       6,
				"notifications": 3,
				"maintenance":   1,
			},

			// Retry failed tasks
			RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
				// Exponential backoff: 2^n seconds
				return time.Duration(1<<uint(n)) * time.Second
			},
		},
	)

	// Register task handlers
	mux := asynq.NewServeMux()

	// Query execution handler
	mux.HandleFunc(queue.TypeExecuteQuery, func(ctx context.Context, t *asynq.Task) error {
		return queue.HandleExecuteQuery(ctx, t)
	})

	// Notification handler
	mux.HandleFunc(queue.TypeSendNotification, func(ctx context.Context, t *asynq.Task) error {
		return queue.HandleSendNotification(ctx, t)
	})

	// Cleanup handler
	mux.HandleFunc(queue.TypeCleanupOldResults, func(ctx context.Context, t *asynq.Task) error {
		return queue.HandleCleanupOldResults(ctx, t)
	})

	// Schema sync handler
	mux.HandleFunc(queue.TypeSyncDataSourceSchema, func(ctx context.Context, t *asynq.Task) error {
		// Inject DB and encryption key into context
		ctx = context.WithValue(ctx, "db", db)
		ctx = context.WithValue(ctx, "encryption_key", cfg.JWT.Secret)
		return queue.HandleSyncDataSourceSchema(ctx, t)
	})

	// Start worker in a goroutine
	go func() {
		log.Println("Worker starting...")
		if err := srv.Run(mux); err != nil {
			log.Fatalf("Failed to start worker: %v", err)
		}
	}()

	// Start periodic schema sync scheduler
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		// Initial sync on startup
		log.Println("[Periodic Sync] Running initial schema sync...")
	 enqueueSchemaSyncs(db, redisAddr)

		for range ticker.C {
			log.Println("[Periodic Sync] Running scheduled schema sync...")
			enqueueSchemaSyncs(db, redisAddr)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the worker
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Worker shutting down...")

	// Graceful shutdown
	srv.Shutdown()

	log.Println("Worker stopped")
	os.Exit(0)
}

// enqueueSchemaSyncs enqueues schema sync tasks for all active data sources
func enqueueSchemaSyncs(db *gorm.DB, redisAddr string) {
	// Get all active data sources
	var dataSources []models.DataSource
	if err := db.Where("is_active = ?", true).Find(&dataSources).Error; err != nil {
		log.Printf("[Periodic Sync] Failed to fetch data sources: %v", err)
		return
	}

	// Create Asynq client
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})

	// Enqueue sync task for each data source
	for _, ds := range dataSources {
		_, err := queue.EnqueueSchemaSync(client, ds.ID.String(), false) // forceRefresh = false for periodic sync
		if err != nil {
			log.Printf("[Periodic Sync] Failed to enqueue sync for %s: %v", ds.Name, err)
		} else {
			log.Printf("[Periodic Sync] Enqueued sync for %s", ds.Name)
		}
	}
}
