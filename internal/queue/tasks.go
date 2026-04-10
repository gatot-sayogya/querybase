package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/yourorg/querybase/internal/models"
	"github.com/yourorg/querybase/internal/service"
	"gorm.io/gorm"
)

const (
	TypeExecuteQuery         = "query:execute"
	TypeSendNotification     = "notification:send"
	TypeCleanupOldResults    = "query:cleanup_results"
	TypeSyncDataSourceSchema = "datasource:sync_schema"
)

type ExecuteQueryPayload struct {
	QueryID      string `json:"query_id"`
	ApprovalID   string `json:"approval_id,omitempty"`
	DataSourceID string `json:"data_source_id"`
	SQL          string `json:"sql"`
	UserID       string `json:"user_id"`
}

type SendNotificationPayload struct {
	NotificationID uuid.UUID `json:"notification_id"`
	ApprovalID     uuid.UUID `json:"approval_id"`
	Type           string    `json:"type"`
	Message        string    `json:"message"`
}

type SyncDataSourceSchemaPayload struct {
	DataSourceID string `json:"data_source_id"`
	ForceRefresh bool   `json:"force_refresh"`
}

func EnqueueQueryExecution(client *asynq.Client, payload *ExecuteQueryPayload) (*asynq.TaskInfo, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeExecuteQuery, data)

	info, err := client.Enqueue(
		task,
		asynq.Queue("queries"),
		asynq.MaxRetry(3),
		asynq.Timeout(300),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	return info, nil
}

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

	queueName := "maintenance"
	if forceRefresh {
		queueName = "default"
	}

	info, err := client.Enqueue(
		task,
		asynq.Queue(queueName),
		asynq.MaxRetry(2),
		asynq.Timeout(300),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	return info, nil
}

func HandleExecuteQuery(ctx context.Context, t *asynq.Task) error {
	var payload ExecuteQueryPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[Task] Executing query %s for data source %s", payload.QueryID, payload.DataSourceID)

	log.Printf("[Task] Query execution completed for query %s", payload.QueryID)
	return nil
}

func HandleSendNotification(ctx context.Context, t *asynq.Task) error {
	var payload SendNotificationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[Task] Sending notification %s for approval %s", payload.NotificationID, payload.ApprovalID)

	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return errors.New("database not found in context")
	}

	var config models.NotificationConfig
	if err := db.Select("id", "group_id", "webhook_url", "is_active", "created_at", "updated_at").
		First(&config, "id = ?", payload.NotificationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("notification config not found: %w", err)
		}
		return fmt.Errorf("failed to load notification config: %w", err)
	}

	var approval models.ApprovalRequest
	if err := db.Preload("DataSource").Preload("RequestedByUser").
		First(&approval, "id = ?", payload.ApprovalID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("approval not found: %w", err)
		}
		return fmt.Errorf("failed to load approval: %w", err)
	}

	message := buildApprovalNotification(&approval, payload.Type)

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	log.Printf("[Task] Notification sent successfully")
	return nil
}

func HandleCleanupOldResults(ctx context.Context, t *asynq.Task) error {
	log.Printf("[Task] Cleaning up old query results")

	log.Printf("[Task] Cleanup completed")
	return nil
}

func HandleSyncDataSourceSchema(ctx context.Context, t *asynq.Task) error {
	var payload SyncDataSourceSchemaPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok || db == nil {
		return errors.New("database not found in context")
	}

	var dataSource models.DataSource
	if err := db.Where("id = ?", payload.DataSourceID).First(&dataSource).Error; err != nil {
		return fmt.Errorf("data source not found: %w", err)
	}

	log.Printf("[Schema Sync] Syncing schema for data source: %s (%s)", dataSource.Name, payload.DataSourceID)

	encryptionKey, ok := ctx.Value("encryption_key").(string)
	if !ok || encryptionKey == "" {
		encryptionKey = ""
		log.Printf("[Schema Sync] Warning: encryption_key not found in context")
	}

	schemaService := service.NewSchemaService(db, encryptionKey)

	_, err := schemaService.GetSchema(ctx, payload.DataSourceID)
	if err != nil {
		log.Printf("[Schema Sync] Failed to fetch schema: %v", err)

		now := time.Now()
		db.Model(&dataSource).Updates(map[string]interface{}{
			"is_healthy":        false,
			"last_health_check": now,
			"last_schema_sync":  now,
		})
		return err
	}

	now := time.Now()
	db.Model(&dataSource).Updates(map[string]interface{}{
		"is_healthy":        true,
		"last_health_check": now,
		"last_schema_sync":  now,
	})

	log.Printf("[Schema Sync] Successfully synced schema for %s", dataSource.Name)

	return nil
}

func buildApprovalNotification(approval *models.ApprovalRequest, notificationType string) map[string]interface{} {
	var title, subtitle string
	var color string

	switch notificationType {
	case "approval_request":
		title = "🔔 New Approval Request"
		subtitle = "A query execution requires your approval"
		color = "#4285F4"
	case "approval_status_change":
		title = "✅ Approval Status Updated"
		subtitle = fmt.Sprintf("Your approval request has been %s", approval.Status)
		color = "#34A853"
	default:
		title = "📋 Notification"
		subtitle = "You have a new notification"
		color = "#757575"
	}

	requestedBy := "Unknown"
	if approval.RequestedByUser.Username != "" {
		requestedBy = approval.RequestedByUser.Username
	}

	dataSourceName := "Unknown"
	if approval.DataSource.Name != "" {
		dataSourceName = approval.DataSource.Name
	}

	return map[string]interface{}{
		"cards": []map[string]interface{}{
			{
				"header": map[string]interface{}{
					"title":      title,
					"subtitle":   subtitle,
					"imageUrl":   "https://www.gstatic.com/images/icons/material/system/1x/message_black_48dp.png",
					"imageStyle": "AVATAR",
				},
				"sections": []map[string]interface{}{
					{
						"widgets": []map[string]interface{}{
							{
								"keyValue": map[string]interface{}{
									"topLabel": "Requested By",
									"content":  requestedBy,
								},
							},
							{
								"keyValue": map[string]interface{}{
									"topLabel": "Data Source",
									"content":  dataSourceName,
								},
							},
							{
								"keyValue": map[string]interface{}{
									"topLabel": "Operation Type",
									"content":  string(approval.OperationType),
								},
							},
							{
								"keyValue": map[string]interface{}{
									"topLabel": "Status",
									"content":  string(approval.Status),
								},
							},
							{
								"keyValue": map[string]interface{}{
									"topLabel": "Query",
									"content":  truncate(approval.QueryText, 200),
								},
							},
						},
					},
					{
						"widgets": []map[string]interface{}{
							{
								"buttons": []map[string]interface{}{
									{
										"textButton": map[string]interface{}{
											"text": "View in Dashboard",
											"onClick": map[string]interface{}{
												"openLink": map[string]interface{}{
													"url": fmt.Sprintf("/dashboard/approvals/%s", approval.ID),
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"cardColor": map[string]interface{}{
					"backgroundColor": color,
				},
			},
		},
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
