package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/yourorg/querybase/internal/models"
	"github.com/yourorg/querybase/internal/service"
)

// Task types
const (
	TypeExecuteQuery         = "query:execute"
	TypeSendNotification     = "notification:send"
	TypeCleanupOldResults    = "query:cleanup_results"
	TypeSyncDataSourceSchema = "datasource:sync_schema"
)

// ExecuteQueryPayload represents the payload for query execution task
type ExecuteQueryPayload struct {
	QueryID     string `json:"query_id"`
	ApprovalID  string `json:"approval_id,omitempty"`
	DataSourceID string `json:"data_source_id"`
	SQL         string `json:"sql"`
	UserID      string `json:"user_id"`
}

// SendNotificationPayload represents the payload for sending notifications
type SendNotificationPayload struct {
	NotificationID uuid.UUID `json:"notification_id"`
	ApprovalID     uuid.UUID `json:"approval_id"`
	Type           string    `json:"type"`
	Message        string    `json:"message"`
}

// SyncDataSourceSchemaPayload represents the payload for schema sync task
type SyncDataSourceSchemaPayload struct {
	DataSourceID string `json:"data_source_id"`
	ForceRefresh bool   `json:"force_refresh"` // true for manual sync
}

// EnqueueQueryExecution enqueues a query execution task
func EnqueueQueryExecution(client *asynq.Client, payload *ExecuteQueryPayload) (*asynq.TaskInfo, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeExecuteQuery, data)

	// Enqueue with options
	info, err := client.Enqueue(
		task,
		asynq.Queue("queries"),
		asynq.MaxRetry(3),
		asynq.Timeout(300), // 5 minutes timeout
	)

	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	return info, nil
}

// EnqueueNotification enqueues a notification task
func EnqueueNotification(client *asynq.Client, payload *SendNotificationPayload) (*asynq.TaskInfo, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeSendNotification, data)

	info, err := client.Enqueue(
		task,
		asynq.Queue("notifications"),
		asynq.MaxRetry(5),
		asynq.Timeout(30),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	return info, nil
}

// EnqueueCleanupTask enqueues a task to cleanup old query results
func EnqueueCleanupTask(client *asynq.Client) (*asynq.TaskInfo, error) {
	task := asynq.NewTask(TypeCleanupOldResults, nil)

	info, err := client.Enqueue(
		task,
		asynq.Queue("maintenance"),
		asynq.MaxRetry(1),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	return info, nil
}

// EnqueueSchemaSync enqueues a schema sync task
func EnqueueSchemaSync(client *asynq.Client, dataSourceID string, forceRefresh bool) (*asynq.TaskInfo, error) {
	payload := &SyncDataSourceSchemaPayload{
		DataSourceID: dataSourceID,
		ForceRefresh: forceRefresh,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeSyncDataSourceSchema, data)

	// For manual sync, use default queue (higher priority)
	// For periodic sync, use maintenance queue
	queueName := "maintenance"
	if forceRefresh {
		queueName = "default"
	}

	info, err := client.Enqueue(
		task,
		asynq.Queue(queueName),
		asynq.MaxRetry(2),
		asynq.Timeout(300), // 5 minutes timeout
	)

	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	return info, nil
}

// HandleExecuteQuery handles query execution tasks
func HandleExecuteQuery(ctx context.Context, t *asynq.Task) error {
	var payload ExecuteQueryPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[Task] Executing query %s for data source %s", payload.QueryID, payload.DataSourceID)

	// TODO: Implement actual query execution
	// This would call the query service to execute the query
	// For now, we just log it

	log.Printf("[Task] Query execution completed for query %s", payload.QueryID)
	return nil
}

// HandleSendNotification handles notification tasks
func HandleSendNotification(ctx context.Context, t *asynq.Task) error {
	var payload SendNotificationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[Task] Sending notification %s for approval %s", payload.NotificationID, payload.ApprovalID)

	// TODO: Implement actual Google Chat webhook sending
	// This would call the notification service to send the webhook

	log.Printf("[Task] Notification sent successfully")
	return nil
}

// HandleCleanupOldResults handles cleanup of old query results
func HandleCleanupOldResults(ctx context.Context, t *asynq.Task) error {
	log.Printf("[Task] Cleaning up old query results")

	// TODO: Implement actual cleanup logic
	// Delete query results older than retention period

	log.Printf("[Task] Cleanup completed")
	return nil
}

// HandleSyncDataSourceSchema handles schema synchronization task
func HandleSyncDataSourceSchema(ctx context.Context, t *asynq.Task) error {
	var payload SyncDataSourceSchemaPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Get DB from context (injected by worker)
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return errors.New("database not found in context")
	}

	// Get data source
	var dataSource models.DataSource
	if err := db.Where("id = ?", payload.DataSourceID).First(&dataSource).Error; err != nil {
		return fmt.Errorf("data source not found: %w", err)
	}

	log.Printf("[Schema Sync] Syncing schema for data source: %s (%s)", dataSource.Name, payload.DataSourceID)

	// Get encryption key from context (injected by worker)
	encryptionKey, ok := ctx.Value("encryption_key").(string)
	if !ok || encryptionKey == "" {
		// Fallback to empty string if not in context (shouldn't happen)
		encryptionKey = ""
		log.Printf("[Schema Sync] Warning: encryption_key not found in context")
	}

	// Create schema service
	schemaService := service.NewSchemaService(db, encryptionKey)

	// Fetch schema from database
	_, err := schemaService.GetSchema(ctx, payload.DataSourceID)
	if err != nil {
		log.Printf("[Schema Sync] Failed to fetch schema: %v", err)

		// Update health status to unhealthy
		now := time.Now()
		db.Model(&dataSource).Updates(map[string]interface{}{
			"is_healthy":         false,
			"last_health_check":  now,
			"last_schema_sync":   now,
		})
		return err
	}

	// Update health status to healthy and sync time
	now := time.Now()
	db.Model(&dataSource).Updates(map[string]interface{}{
		"is_healthy":         true,
		"last_health_check":  now,
		"last_schema_sync":   now,
	})

	log.Printf("[Schema Sync] Successfully synced schema for %s", dataSource.Name)

	return nil
}
