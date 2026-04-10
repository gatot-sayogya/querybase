# Google Chat Integration for QueryBase - Mode 1 (Webhook-Only)

## TL;DR

> **Quick Summary**: Implement outgoing Google Chat notifications for approval requests. Users receive rich card notifications in Google Chat with "View in QueryBase" buttons. Quick setup, immediate value!
> 
> **Deliverables**:
> - Rich card notifications sent to Google Chat
> - "View in QueryBase" buttons linking to approval pages
> - Approval status update notifications
> - Async delivery via Asynq queue
> 
> **Estimated Effort**: Small (4 tasks)
> **Parallel Execution**: YES - All 4 tasks can run in parallel
> **Setup Required**: Just webhook URL in database

---

## Context

### What You're Building
A **notification system** that sends approval alerts to Google Chat spaces. When someone submits a query for approval, all approvers get notified in Google Chat with rich cards showing query details.

### What You Have
- ✅ QueryBase HTTPS endpoint
- ✅ Google Chat webhook URL

### What You DON'T Need (for Mode 1)
- ❌ Service account
- ❌ Chat App configuration
- ❌ Admin SDK
- ❌ Google Cloud Console setup

### User Experience

**When query is submitted:**
```
🔔 New Approval Request

👤 John Doe (john@company.com)
📊 Data Source: Production PostgreSQL

📝 Query:
UPDATE users 
SET status = 'active' 
WHERE created_at > '2024-01-01';

⚠️ Estimated impact: 1,245 rows

[View in QueryBase]  ← Click opens QueryBase
```

**When approved/rejected:**
```
✅ Approved by Jane Smith

Query #297 has been approved.

[View in QueryBase]
```

---

## Work Objectives

### Core Objective
Send approval notifications from QueryBase to Google Chat spaces via webhook URLs configured per group.

### Concrete Deliverables

1. **Google Chat DTOs** (`internal/api/dto/googlechat.go`)
   - Card message structures
   - Widget definitions

2. **Queue Handler** (`internal/queue/tasks.go`)
   - Implement `HandleSendNotification()` (currently TODO)
   - Send messages via webhook HTTP POST

3. **Configuration** (`internal/config/config.go`)
   - Add Google Chat settings (minimal for Mode 1)

4. **Approval Workflow Integration** (`internal/service/approval.go`)
   - Call `SendApprovalNotification()` after approval creation
   - Call `SendReviewNotification()` after review submission

### Definition of Done

- [ ] Approval request sends notification to Google Chat
- [ ] Notification card shows: requester, data source, query preview
- [ ] "View in QueryBase" button opens approval page
- [ ] Status updates sent when approved/rejected
- [ ] Notifications sent async via Asynq queue
- [ ] All tests pass

### Must Have

1. Rich card notifications with query details
2. "View in QueryBase" button (opens browser)
3. Status change notifications
4. Queue-based async delivery
5. Error handling for webhook failures

### Must NOT Have (Out of Scope for Mode 1)

1. Interactive Approve/Reject buttons in Chat
2. Processing approval actions from Google Chat
3. User identity resolution from Chat
4. Card updates after actions
5. Service account authentication

---

## Verification Strategy

### Test Approach
- **Unit Tests**: Test notification service methods
- **Integration Tests**: Test queue handler with mocked HTTP client
- **Manual QA**: Add webhook URL, create approval, verify in Chat

### QA Policy
Every task MUST include agent-executed QA scenarios.

---

## Execution Strategy

### Parallel Execution - All 4 Tasks Can Run Simultaneously

```
Wave 1 (Foundation - all can start immediately):
├── Task 1: Create Google Chat DTOs
├── Task 2: Implement HandleSendNotification queue handler
├── Task 3: Add Google Chat configuration
└── Task 4: Wire notifications into approval workflow

Wave FINAL (After ALL tasks):
├── Task F1: Code quality review
└── Task F2: Integration testing
```

### No Dependencies Between Tasks 1-4
All four tasks are independent and can be worked on in parallel!

---

## TODOs

- [x] 1. Create Google Chat DTOs and Card Models

  **What to do**:
  - Create `internal/api/dto/googlechat.go`
  - Define card message structures for Google Chat
  - Define widgets (decoratedText, buttonList with openLink)
  - Follow existing DTO patterns
  
  **Code structure**:
  ```go
  type GoogleChatMessage struct {
      Text  string `json:"text,omitempty"`
      Cards []Card `json:"cards,omitempty"`
  }
  
  type Card struct {
      Header   *CardHeader   `json:"header,omitempty"`
      Sections []CardSection `json:"sections,omitempty"`
  }
  
  type CardSection struct {
      Header  string   `json:"header,omitempty"`
      Widgets []Widget `json:"widgets,omitempty"`
  }
  
  type Widget struct {
      DecoratedText *DecoratedTextWidget `json:"decoratedText,omitempty"`
      ButtonList    *ButtonListWidget    `json:"buttonList,omitempty"`
  }
  
  type DecoratedTextWidget struct {
      TopLabel string `json:"topLabel,omitempty"`
      Text     string `json:"text"`
  }
  
  type ButtonListWidget struct {
      Buttons []Button `json:"buttons"`
  }
  
  type Button struct {
      Text    string   `json:"text"`
      OnClick *OnClick `json:"onClick"`
  }
  
  type OnClick struct {
      OpenLink *OpenLink `json:"openLink,omitempty"`
  }
  
  type OpenLink struct {
      URL string `json:"url"`
  }
  ```

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []
  - Reason: Simple struct definitions

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Blocks**: None (all independent)

  **References**:
  - Pattern: `internal/api/dto/approval.go` - DTO structure
  - External: https://developers.google.com/workspace/chat/api/reference/rest/v1/cards

  **Acceptance Criteria**:
  - [ ] File created: `internal/api/dto/googlechat.go`
  - [ ] All structs compile
  - [ ] `go build ./internal/api/dto/` succeeds

  **QA Scenarios**:
  ```
  Scenario: Verify DTOs compile
    Tool: Bash
    Steps:
      1. Run: go build ./internal/api/dto/
    Expected: Build succeeds
    Evidence: .sisyphus/evidence/task-1-compile.log
  ```

  **Commit**: YES
  - Message: `feat(googlechat): add Google Chat card DTOs`
  - Files: `internal/api/dto/googlechat.go`

- [x] 2. Implement HandleSendNotification Queue Handler

  **What to do**:
  - Open `internal/queue/tasks.go`
  - Find `HandleSendNotification` (currently TODO at line ~174)
  - Implement to send notifications via HTTP POST to webhook URL
  - Load notification config from database
  - Format message as Google Chat card
  - Send via HTTP client
  - Add error handling
  
  **Implementation**:
  ```go
  func HandleSendNotification(ctx context.Context, t *asynq.Task) error {
      var payload SendNotificationPayload
      if err := json.Unmarshal(t.Payload(), &payload); err != nil {
          return fmt.Errorf("failed to unmarshal payload: %w", err)
      }
      
      // Get database from context
      db := ctx.Value("db").(*gorm.DB)
      
      // Get notification config
      var config models.NotificationConfig
      if err := db.First(&config, "id = ?", payload.NotificationID).Error; err != nil {
          return fmt.Errorf("notification config not found: %w", err)
      }
      
      // Get approval with relations
      var approval models.ApprovalRequest
      if err := db.Preload("DataSource").Preload("RequestedByUser").
          First(&approval, "id = ?", payload.ApprovalID).Error; err != nil {
          return fmt.Errorf("approval not found: %w", err)
      }
      
      // Build Google Chat message
      message := buildApprovalNotification(&approval, payload.Type)
      
      // Send to webhook
      jsonData, _ := json.Marshal(message)
      req, _ := http.NewRequest("POST", config.WebhookURL, bytes.NewBuffer(jsonData))
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
      
      return nil
  }
  
  func buildApprovalNotification(approval *models.ApprovalRequest, notifType string) *GoogleChatMessage {
      baseURL := os.Getenv("APP_BASE_URL") // e.g., https://querybase.yourcompany.com
      approvalURL := fmt.Sprintf("%s/approvals/%s", baseURL, approval.ID)
      
      card := &Card{
          Header: &CardHeader{
              Title:    "🔔 New Approval Request",
              Subtitle: fmt.Sprintf("From %s", approval.RequestedByUser.FullName),
          },
          Sections: []CardSection{
              {
                  Widgets: []Widget{
                      {
                          DecoratedText: &DecoratedTextWidget{
                              TopLabel: "Data Source",
                              Text:     approval.DataSource.Name,
                          },
                      },
                      {
                          DecoratedText: &DecoratedTextWidget{
                              TopLabel: "Query",
                              Text:     truncate(approval.QueryText, 200),
                          },
                      },
                  },
              },
              {
                  Widgets: []Widget{
                      {
                          ButtonList: &ButtonListWidget{
                              Buttons: []Button{
                                  {
                                      Text: "View in QueryBase",
                                      OnClick: &OnClick{
                                          OpenLink: &OpenLink{URL: approvalURL},
                                      },
                                  },
                              },
                          },
                      },
                  },
              },
          },
      }
      
      return &GoogleChatMessage{Cards: []Card{*card}}
  }
  ```

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []
  - Reason: Core logic implementation

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Blocks**: Task 4 (can test independently)

  **References**:
  - Pattern: `internal/queue/tasks.go:174` - TODO location
  - Pattern: `internal/queue/tasks.go:207-260` - HandleSyncDataSourceSchema example

  **Acceptance Criteria**:
  - [ ] `HandleSendNotification` implemented
  - [ ] Sends HTTP POST to webhook URL
  - [ ] Proper error handling
  - [ ] Unit tests with mocked HTTP client

  **QA Scenarios**:
  ```
  Scenario: Send notification via queue
    Tool: Bash (go test)
    Steps:
      1. Run: go test -v ./internal/queue/ -run TestHandleSendNotification
    Expected: Test passes, mock webhook called
    Evidence: .sisyphus/evidence/task-2-queue.log
  ```

  **Commit**: YES
  - Message: `feat(queue): implement HandleSendNotification for Google Chat`
  - Files: `internal/queue/tasks.go`, `internal/queue/tasks_test.go`

- [x] 3. Add Google Chat Configuration

  **What to do**:
  - Add minimal Google Chat config to `internal/config/config.go`
  - Just need `Enabled` flag and `BaseURL` for generating links
  
  **Implementation**:
  ```go
  type GoogleChatConfig struct {
      Enabled bool   `mapstructure:"enabled"`
      BaseURL string `mapstructure:"base_url"` // For generating approval links
  }
  
  // Add to main Config struct
  type Config struct {
      Server      ServerConfig      `mapstructure:"server"`
      Database    DatabaseConfig    `mapstructure:"database"`
      Redis       RedisConfig       `mapstructure:"redis"`
      JWT         JWTConfig         `mapstructure:"jwt"`
      CORS        CORSConfig        `mapstructure:"cors"`
      GoogleChat  GoogleChatConfig  `mapstructure:"googlechat"` // ADD THIS
  }
  ```

  **Environment variables**:
  ```bash
  GOOGLECHAT_ENABLED=true
  GOOGLECHAT_BASE_URL=https://querybase.yourcompany.com
  ```

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []
  - Reason: Simple config addition

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Blocks**: Task 2 (uses BaseURL)

  **Acceptance Criteria**:
  - [ ] Config struct added
  - [ ] Default values set
  - [ ] Builds successfully

  **QA Scenarios**:
  ```
  Scenario: Config loads correctly
    Tool: Bash
    Steps:
      1. Run: go build ./internal/config/
    Expected: Build succeeds
    Evidence: .sisyphus/evidence/task-3-config.log
  ```

  **Commit**: YES
  - Message: `feat(config): add Google Chat configuration for Mode 1`
  - Files: `internal/config/config.go`

- [x] 4. Wire Notifications into Approval Workflow

  **What to do**:
  - Modify `internal/service/approval.go`
  - Inject `NotificationService` into `ApprovalService`
  - Call notification after approval creation (line 52)
  - Call notification after review submission (line 214)
  - Update `NewApprovalService` constructor
  - Update service initialization in `cmd/api/main.go`
  
  **Implementation**:
  ```go
  // In ApprovalService struct
  type ApprovalService struct {
      db                  *gorm.DB
      queryService        *QueryService
      statsService        *StatsService
      auditService        *AuditService
      notificationService *NotificationService  // ADD THIS
  }
  
  // Update constructor
  func NewApprovalService(db *gorm.DB, queryService *QueryService, 
      statsService *StatsService, notificationService *NotificationService) *ApprovalService {
      return &ApprovalService{
          db:                  db,
          queryService:        queryService,
          statsService:        statsService,
          auditService:        NewAuditService(db),
          notificationService: notificationService,  // ADD THIS
      }
  }
  
  // In CreateApprovalRequest (after line 52)
  func (s *ApprovalService) CreateApprovalRequest(ctx context.Context, req *ApprovalRequest) (*models.ApprovalRequest, error) {
      // ... existing code ...
      
      if err := s.db.Create(approval).Error; err != nil {
          return nil, fmt.Errorf("failed to create approval request: %w", err)
      }
      
      // Send notification to approvers
      if s.notificationService != nil {
          go func() {
              ctx := context.Background()
              if err := s.notificationService.SendApprovalNotification(ctx, approval); err != nil {
                  log.Printf("[CreateApprovalRequest] Failed to send notification: %v", err)
              }
          }()
      }
      
      // ... rest of existing code ...
  }
  
  // In ReviewApproval (after line 214)
  func (s *ApprovalService) ReviewApproval(ctx context.Context, review *ReviewInput) (*models.ApprovalReview, error) {
      // ... existing code to create review ...
      
      // Send notification to requester
      if s.notificationService != nil {
          go func() {
              ctx := context.Background()
              if err := s.notificationService.SendReviewNotification(ctx, approval, review); err != nil {
                  log.Printf("[ReviewApproval] Failed to send notification: %v", err)
              }
          }()
      }
      
      return approvalReview, nil
  }
  ```

  **Update cmd/api/main.go**:
  ```go
  // Initialize services
  notificationService := service.NewNotificationService(db)
  approvalService := service.NewApprovalService(db, queryService, statsService, notificationService)
  ```

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []
  - Reason: Integration work

  **Parallelization**:
  - **Can Run In Parallel**: YES (can wire after Task 2 is done)
  - **Blocks**: None

  **References**:
  - Pattern: `internal/service/approval.go:52` - After approval creation
  - Pattern: `internal/service/approval.go:214` - After review

  **Acceptance Criteria**:
  - [ ] NotificationService injected
  - [ ] SendApprovalNotification called after creation
  - [ ] SendReviewNotification called after review
  - [ ] All initializations updated

  **QA Scenarios**:
  ```
  Scenario: Notification sent on approval creation
    Tool: Bash (go test)
    Steps:
      1. Create approval request
      2. Verify notification queued
    Expected: Notification task created
    Evidence: .sisyphus/evidence/task-4-wire.log
  ```

  **Commit**: YES
  - Message: `feat(approval): wire notification service into workflow`
  - Files: `internal/service/approval.go`, `cmd/api/main.go`

---

## Setup Instructions (After Implementation)

### 1. Add Webhook URL to Database

```sql
-- Get your group ID first
SELECT id, name FROM groups;

-- Insert webhook configuration
INSERT INTO notification_configs (
    id, 
    group_id, 
    webhook_url, 
    is_active, 
    notification_events,
    created_at,
    updated_at
) VALUES (
    gen_random_uuid(),
    'your-group-id-here',  -- Replace with actual group ID
    'https://chat.googleapis.com/v1/spaces/AAAAxxxxxxxxx/webhooks/...',  -- Your webhook URL
    true,
    ARRAY['approval_request', 'approval_status_change'],
    NOW(),
    NOW()
);
```

### 2. Set Environment Variables

```bash
# Add to .env or export
GOOGLECHAT_ENABLED=true
GOOGLECHAT_BASE_URL=https://querybase.yourcompany.com
```

### 3. Restart QueryBase

```bash
make run-api
make run-worker
```

### 4. Test It!

1. Create a query that requires approval
2. Check your Google Chat space
3. You should see the notification!

---

## Verification

### Test Commands

```bash
# Build
make build

# Test
make test

# Check logs for notifications
tail -f logs/querybase.log | grep "notification"
```

### Manual Testing

1. **Create approval request** via QueryBase UI
2. **Check Google Chat** - should see notification within 5 seconds
3. **Click "View in QueryBase"** - should open approval page
4. **Approve the query** in QueryBase UI
5. **Check Google Chat** - should see status update

---

## Success Criteria

- [ ] All 4 tasks complete
- [ ] All tests pass (`make test`)
- [ ] Notifications sent to Google Chat
- [ ] "View in QueryBase" button works
- [ ] Status updates sent correctly
- [ ] No errors in logs

---

## Next Steps (Mode 2 - Optional Upgrade)

When you want interactive Approve/Reject buttons:
1. Configure Chat App in Google Cloud Console
2. Create service account
3. Enable Admin SDK
4. System automatically upgrades to Mode 2!

**No code changes needed** - just add configuration!
