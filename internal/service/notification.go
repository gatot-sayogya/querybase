package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yourorg/querybase/internal/models"
	"gorm.io/gorm"
)

// NotificationService handles notification logic
type NotificationService struct {
	db        *gorm.DB
	httpClient *http.Client
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{
		db: db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendApprovalNotification sends a notification for approval requests
func (s *NotificationService) SendApprovalNotification(ctx context.Context, approval *models.ApprovalRequest) error {
	// Get active notification configs
	var configs []models.NotificationConfig
	err := s.db.Where("is_active = ?", true).Find(&configs).Error
	if err != nil {
		return fmt.Errorf("failed to get notification configs: %w", err)
	}

	// For now, we'll use a simple message format
	message := s.formatApprovalMessage(approval)

	// Send to each config
	for _, config := range configs {
		if err := s.sendGoogleChatNotification(&config, message); err != nil {
			// Log error but continue trying other configs
			fmt.Printf("Failed to send notification to %s: %v\n", config.WebhookURL, err)
		}
	}

	return nil
}

// SendReviewNotification sends a notification when an approval is reviewed
func (s *NotificationService) SendReviewNotification(ctx context.Context, approval *models.ApprovalRequest, review *models.ApprovalReview) error {
	// Get active notification configs
	var configs []models.NotificationConfig
	err := s.db.Where("is_active = ?", true).Find(&configs).Error
	if err != nil {
		return fmt.Errorf("failed to get notification configs: %w", err)
	}

	// Create notification message based on review decision
	message := s.formatReviewMessage(approval, review)

	// Send to each config
	for _, config := range configs {
		if err := s.sendGoogleChatNotification(&config, message); err != nil {
			// Log error but continue trying other configs
			fmt.Printf("Failed to send notification to %s: %v\n", config.WebhookURL, err)
		}
	}

	return nil
}

// sendGoogleChatNotification sends a notification to Google Chat webhook
func (s *NotificationService) sendGoogleChatNotification(config *models.NotificationConfig, message *GoogleChatMessage) error {
	// Marshal message to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// formatApprovalMessage formats an approval request message for Google Chat
func (s *NotificationService) formatApprovalMessage(approval *models.ApprovalRequest) *GoogleChatMessage {
	// Get approval URL (this should be configured in the app)
	approvalURL := fmt.Sprintf("https://your-app.com/approvals/%s", approval.ID)

	message := &GoogleChatMessage{
		Text: fmt.Sprintf("🔔 New Approval Request\n\n%s has submitted a query for approval.", approval.RequestedByUser.FullName),
	}

	// Create card widget
	card := Card{
		Header: &CardHeader{
			Title:    "New Approval Request",
			Subtitle: fmt.Sprintf("Data Source: %s", approval.DataSource.Name),
		},
		Sections: []CardSection{
			{
				Widgets: []Widget{
					{
						TextParagraph: &TextWidget{
							Text: fmt.Sprintf("**Query:**\n```sql\n%s\n```", approval.QueryText),
						},
					},
				},
			},
			{
				Widgets: []Widget{
					{
						Buttons: []ButtonWidget{
							{
								TextButton: &TextButton{
									Text:   "View Details",
									OnClick: &OnClick{
										OpenLink: &OpenLink{
											URL: approvalURL,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	message.Cards = []Card{card}

	return message
}

// formatReviewMessage formats a review notification message
func (s *NotificationService) formatReviewMessage(approval *models.ApprovalRequest, review *models.ApprovalReview) *GoogleChatMessage {
	emoji := "✅"
	if review.Decision == models.ApprovalDecisionRejected {
		emoji = "❌"
	}

	status := "Approved"
	if review.Decision == models.ApprovalDecisionRejected {
		status = "Rejected"
	}

	message := &GoogleChatMessage{
		Text: fmt.Sprintf("%s Approval Request %s", emoji, status),
	}

	card := Card{
		Header: &CardHeader{
			Title:    fmt.Sprintf("Approval %s", status),
			Subtitle: fmt.Sprintf("Reviewed by %s", review.Reviewer.FullName),
		},
		Sections: []CardSection{
			{
				Widgets: []Widget{
					{
						TextParagraph: &TextWidget{
							Text: fmt.Sprintf("**Decision:** %s\n\n**Comments:** %s", status, review.Comments),
						},
					},
				},
			},
		},
	}

	message.Cards = []Card{card}

	return message
}

// GoogleChatMessage represents a Google Chat webhook message
type GoogleChatMessage struct {
	Text  string `json:"text,omitempty"`
	Cards []Card `json:"cards,omitempty"`
}

// Card represents a Google Chat card
type Card struct {
	Header  *CardHeader   `json:"header,omitempty"`
	Sections []CardSection `json:"sections,omitempty"`
}

// CardHeader represents a card header
type CardHeader struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
	ImageURL string `json:"imageUrl,omitempty"`
}

// CardSection represents a card section
type CardSection struct {
	Header  string   `json:"header,omitempty"`
	Widgets []Widget `json:"widgets,omitempty"`
}

// Widget represents a card widget
type Widget struct {
	KeyValue        *KeyValueWidget        `json:"keyValue,omitempty"`
	TextParagraph   *TextWidget            `json:"textParagraph,omitempty"`
	Image           *ImageWidget           `json:"image,omitempty"`
	Buttons         []ButtonWidget         `json:"buttons,omitempty"`
	TextButton      *TextButton            `json:"textButton,omitempty"`
	DecoratedText   *DecoratedTextWidget   `json:"decoratedText,omitempty"`
}

// KeyValueWidget represents a key-value widget
type KeyValueWidget struct {
	TopLabel         string `json:"topLabel,omitempty"`
	Content          string `json:"content,omitempty"`
	ContentMultiline string `json:"contentMultiline,omitempty"`
	BottomLabel      string `json:"bottomLabel,omitempty"`
	Icon             string `json:"icon,omitempty"`
}

// TextWidget represents a text paragraph widget
type TextWidget struct {
	Text string `json:"text"`
}

// ImageWidget represents an image widget
type ImageWidget struct {
	ImageURL string `json:"imageUrl,omitempty"`
	OnClick  *OnClick `json:"onClick,omitempty"`
}

// ButtonWidget represents a button widget
type ButtonWidget struct {
	TextButton *TextButton `json:"textButton,omitempty"`
}

// TextButton represents a text button
type TextButton struct {
	Text    string  `json:"text"`
	OnClick *OnClick `json:"onClick,omitempty"`
}

// OnClick represents an onClick action
type OnClick struct {
	OpenLink *OpenLink `json:"openLink,omitempty"`
	Action   *Action   `json:"action,omitempty"`
}

// OpenLink represents a link to open
type OpenLink struct {
	URL string `json:"url"`
}

// Action represents an action
type Action struct {
	Function string                 `json:"function,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// DecoratedTextWidget represents a decorated text widget
type DecoratedTextWidget struct {
	TopLabel    string    `json:"topLabel,omitempty"`
	Text        string    `json:"text,omitempty"`
	BottomLabel string    `json:"bottomLabel,omitempty"`
	OnClick     *OnClick  `json:"onClick,omitempty"`
	StartIcon   string    `json:"startIcon,omitempty"`
}
