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
	"github.com/yourorg/querybase/internal/queue"
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

	// Start worker in a goroutine
	go func() {
		log.Println("Worker starting...")
		if err := srv.Run(mux); err != nil {
			log.Fatalf("Failed to start worker: %v", err)
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
