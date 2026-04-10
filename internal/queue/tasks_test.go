package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/querybase/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.NotificationConfig{}, &models.ApprovalRequest{}, &models.DataSource{}, &models.User{})
	require.NoError(t, err)

	return db
}

func createTestUser(t *testing.T, db *gorm.DB, suffix string) *models.User {
	user := &models.User{
		ID:       uuid.New(),
		Email:    fmt.Sprintf("test%s@example.com", suffix),
		Username: fmt.Sprintf("testuser%s", suffix),
		FullName: "Test User",
		Role:     models.RoleUser,
		IsActive: true,
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func createTestDataSource(t *testing.T, db *gorm.DB, suffix string) *models.DataSource {
	ds := &models.DataSource{
		ID:           uuid.New(),
		Name:         fmt.Sprintf("Test Database %s", suffix),
		Type:         models.DataSourceTypePostgreSQL,
		Host:         "localhost",
		Port:         5432,
		DatabaseName: "testdb",
		Username:     "testuser",
		IsActive:     true,
		IsHealthy:    true,
	}
	err := db.Create(ds).Error
	require.NoError(t, err)
	return ds
}

func createTestApproval(t *testing.T, db *gorm.DB, user *models.User, ds *models.DataSource) *models.ApprovalRequest {
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		RequestedBy:   user.ID,
		DataSourceID:  ds.ID,
		OperationType: models.OperationTypeDelete,
		QueryText:     "DELETE FROM users WHERE id = 1",
		Status:        models.ApprovalStatusPending,
	}
	err := db.Create(approval).Error
	require.NoError(t, err)
	return approval
}

func createTestNotificationConfig(t *testing.T, db *gorm.DB, webhookURL string) *models.NotificationConfig {
	config := &models.NotificationConfig{
		ID:         uuid.New(),
		GroupID:    uuid.New(),
		WebhookURL: webhookURL,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err := db.Exec(
		"INSERT INTO notification_configs (id, group_id, webhook_url, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		config.ID, config.GroupID, config.WebhookURL, config.IsActive, config.CreatedAt, config.UpdatedAt,
	).Error
	require.NoError(t, err)
	return config
}

func TestHandleSendNotification_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var message map[string]interface{}
		err = json.Unmarshal(body, &message)
		require.NoError(t, err)

		cards, ok := message["cards"].([]interface{})
		require.True(t, ok)
		assert.Len(t, cards, 1)

		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	db := setupTestDB(t)

	user := createTestUser(t, db, "1")
	ds := createTestDataSource(t, db, "1")
	approval := createTestApproval(t, db, user, ds)
	config := createTestNotificationConfig(t, db, mockServer.URL)

	payload := SendNotificationPayload{
		NotificationID: config.ID,
		ApprovalID:     approval.ID,
		Type:           "approval_request",
		Message:        "Test notification",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeSendNotification, data)
	ctx := context.WithValue(context.Background(), "db", db)

	err = HandleSendNotification(ctx, task)
	assert.NoError(t, err)
}

func TestHandleSendNotification_WebhookFailure(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	db := setupTestDB(t)

	user := createTestUser(t, db, "2")
	ds := createTestDataSource(t, db, "2")
	approval := createTestApproval(t, db, user, ds)
	config := createTestNotificationConfig(t, db, mockServer.URL)

	payload := SendNotificationPayload{
		NotificationID: config.ID,
		ApprovalID:     approval.ID,
		Type:           "approval_request",
		Message:        "Test notification",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeSendNotification, data)
	ctx := context.WithValue(context.Background(), "db", db)

	err = HandleSendNotification(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhook returned status 500")
}

func TestHandleSendNotification_InvalidPayload(t *testing.T) {
	task := asynq.NewTask(TypeSendNotification, []byte("invalid json"))
	ctx := context.Background()

	err := HandleSendNotification(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal payload")
}

func TestHandleSendNotification_NoDBInContext(t *testing.T) {
	payload := SendNotificationPayload{
		NotificationID: uuid.New(),
		ApprovalID:     uuid.New(),
		Type:           "approval_request",
		Message:        "Test notification",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeSendNotification, data)
	ctx := context.Background()

	err = HandleSendNotification(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not found in context")
}

func TestHandleSendNotification_NotificationConfigNotFound(t *testing.T) {
	db := setupTestDB(t)

	user := createTestUser(t, db, "3")
	ds := createTestDataSource(t, db, "3")
	approval := createTestApproval(t, db, user, ds)

	payload := SendNotificationPayload{
		NotificationID: uuid.New(),
		ApprovalID:     approval.ID,
		Type:           "approval_request",
		Message:        "Test notification",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeSendNotification, data)
	ctx := context.WithValue(context.Background(), "db", db)

	err = HandleSendNotification(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notification config not found")
}

func TestHandleSendNotification_ApprovalNotFound(t *testing.T) {
	db := setupTestDB(t)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	config := createTestNotificationConfig(t, db, mockServer.URL)

	payload := SendNotificationPayload{
		NotificationID: config.ID,
		ApprovalID:     uuid.New(),
		Type:           "approval_request",
		Message:        "Test notification",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeSendNotification, data)
	ctx := context.WithValue(context.Background(), "db", db)

	err = HandleSendNotification(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "approval not found")
}

func TestHandleSendNotification_StatusChangeType(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var message map[string]interface{}
		err = json.Unmarshal(body, &message)
		require.NoError(t, err)

		cards, ok := message["cards"].([]interface{})
		require.True(t, ok)
		assert.Len(t, cards, 1)

		card := cards[0].(map[string]interface{})
		header := card["header"].(map[string]interface{})
		title := header["title"].(string)
		assert.Contains(t, title, "Approval Status Updated")

		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	db := setupTestDB(t)

	user := createTestUser(t, db, "4")
	ds := createTestDataSource(t, db, "4")
	approval := createTestApproval(t, db, user, ds)
	config := createTestNotificationConfig(t, db, mockServer.URL)

	payload := SendNotificationPayload{
		NotificationID: config.ID,
		ApprovalID:     approval.ID,
		Type:           "approval_status_change",
		Message:        "Status changed",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeSendNotification, data)
	ctx := context.WithValue(context.Background(), "db", db)

	err = HandleSendNotification(ctx, task)
	assert.NoError(t, err)
}

func TestBuildApprovalNotification(t *testing.T) {
	userID := uuid.New()
	dsID := uuid.New()
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		RequestedBy:   userID,
		DataSourceID:  dsID,
		OperationType: models.OperationTypeDelete,
		QueryText:     "DELETE FROM users WHERE id = 1",
		Status:        models.ApprovalStatusPending,
		RequestedByUser: models.User{
			ID:       userID,
			Username: "testuser",
		},
		DataSource: models.DataSource{
			ID:   dsID,
			Name: "Test DB",
		},
	}

	message := buildApprovalNotification(approval, "approval_request")
	require.NotNil(t, message)

	cards, ok := message["cards"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, cards, 1)

	card := cards[0]
	header, ok := card["header"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, header["title"], "New Approval Request")

	approval.Status = models.ApprovalStatusApproved
	message = buildApprovalNotification(approval, "approval_status_change")
	require.NotNil(t, message)

	cards, ok = message["cards"].([]map[string]interface{})
	require.True(t, ok)
	card = cards[0]
	header, ok = card["header"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, header["title"], "Approval Status Updated")

	message = buildApprovalNotification(approval, "unknown_type")
	require.NotNil(t, message)

	cards, ok = message["cards"].([]map[string]interface{})
	require.True(t, ok)
	card = cards[0]
	header, ok = card["header"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, header["title"], "Notification")
}

func TestBuildApprovalNotification_LongQuery(t *testing.T) {
	userID := uuid.New()
	dsID := uuid.New()
	longQuery := "DELETE FROM users WHERE id = 1 AND " + string(make([]byte, 300))

	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		RequestedBy:   userID,
		DataSourceID:  dsID,
		OperationType: models.OperationTypeDelete,
		QueryText:     longQuery,
		Status:        models.ApprovalStatusPending,
		RequestedByUser: models.User{
			ID:       userID,
			Username: "testuser",
		},
		DataSource: models.DataSource{
			ID:   dsID,
			Name: "Test DB",
		},
	}

	message := buildApprovalNotification(approval, "approval_request")
	require.NotNil(t, message)

	cards := message["cards"].([]map[string]interface{})
	sections := cards[0]["sections"].([]map[string]interface{})
	widgets := sections[0]["widgets"].([]map[string]interface{})

	var queryContent string
	for _, widget := range widgets {
		if kv, ok := widget["keyValue"].(map[string]interface{}); ok {
			if topLabel, ok := kv["topLabel"].(string); ok && topLabel == "Query" {
				queryContent = kv["content"].(string)
				break
			}
		}
	}

	assert.NotNil(t, queryContent)
	assert.LessOrEqual(t, len(queryContent), 200)
	assert.Contains(t, queryContent, "...")
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string no truncation",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length no truncation",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "long string truncation",
			input:    "hello world this is a long string",
			maxLen:   10,
			expected: "hello w...",
		},
		{
			name:     "small maxLen edge case",
			input:    "hello",
			maxLen:   3,
			expected: "hel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandleSendNotification_HTTPClientError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	serverURL := mockServer.URL
	mockServer.Close()

	db := setupTestDB(t)

	user := createTestUser(t, db, "5")
	ds := createTestDataSource(t, db, "5")
	approval := createTestApproval(t, db, user, ds)
	config := createTestNotificationConfig(t, db, serverURL)

	payload := SendNotificationPayload{
		NotificationID: config.ID,
		ApprovalID:     approval.ID,
		Type:           "approval_request",
		Message:        "Test notification",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(TypeSendNotification, data)
	ctx := context.WithValue(context.Background(), "db", db)

	err = HandleSendNotification(ctx, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send webhook")
}

func TestBuildApprovalNotification_MissingRelations(t *testing.T) {
	approval := &models.ApprovalRequest{
		ID:            uuid.New(),
		OperationType: models.OperationTypeDelete,
		QueryText:     "DELETE FROM users WHERE id = 1",
		Status:        models.ApprovalStatusPending,
	}

	message := buildApprovalNotification(approval, "approval_request")
	require.NotNil(t, message)

	cards := message["cards"].([]map[string]interface{})
	sections := cards[0]["sections"].([]map[string]interface{})
	widgets := sections[0]["widgets"].([]map[string]interface{})

	var foundRequestedBy, foundDataSource bool
	for _, widget := range widgets {
		if kv, ok := widget["keyValue"].(map[string]interface{}); ok {
			if topLabel, ok := kv["topLabel"].(string); ok {
				if topLabel == "Requested By" {
					foundRequestedBy = true
					assert.Equal(t, "Unknown", kv["content"])
				}
				if topLabel == "Data Source" {
					foundDataSource = true
					assert.Equal(t, "Unknown", kv["content"])
				}
			}
		}
	}

	assert.True(t, foundRequestedBy)
	assert.True(t, foundDataSource)
}

func TestTruncate_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "zero maxLen",
			input:    "hello",
			maxLen:   0,
			expected: "",
		},
		{
			name:     "unicode string",
			input:    "日本語テキスト",
			maxLen:   10,
			expected: "日本語テ...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}
