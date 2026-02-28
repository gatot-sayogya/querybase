package googlechat

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/querybase/internal/config"
	"github.com/yourorg/querybase/internal/models"
	"github.com/yourorg/querybase/internal/service"
	"google.golang.org/api/chat/v1"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

// ChatPlugin handles Google Chat App integration for interactive approvals
type ChatPlugin struct {
	db              *gorm.DB
	approvalService *service.ApprovalService
	chatService     *chat.Service
	config          *config.GoogleChatConfig
	threadStore     *ThreadStore
}

// NewChatPlugin creates a new Google Chat plugin
func NewChatPlugin(db *gorm.DB, approvalService *service.ApprovalService, cfg *config.GoogleChatConfig) (*ChatPlugin, error) {
	plugin := &ChatPlugin{
		db:              db,
		approvalService: approvalService,
		config:          cfg,
		threadStore:     NewThreadStore(db),
	}

	// Initialize Google Chat API client with service account
	if cfg.ServiceAccountFile != "" {
		ctx := context.Background()
		svc, err := chat.NewService(ctx,
			option.WithCredentialsFile(cfg.ServiceAccountFile),
			option.WithScopes(chat.ChatBotScope),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to init Google Chat service: %w", err)
		}
		plugin.chatService = svc
		log.Println("[GoogleChat] Chat API client initialized with service account")
	} else {
		log.Println("[GoogleChat] No service account file configured, Chat API calls will be disabled")
	}

	return plugin, nil
}

// RegisterRoutes registers the Google Chat event handler routes
func (p *ChatPlugin) RegisterRoutes(group *gin.RouterGroup) {
	gchat := group.Group("/googlechat")
	{
		handler := NewEventHandler(p)
		gchat.POST("/events", handler.HandleEvent)
	}
	log.Printf("[GoogleChat] Routes registered at %s/googlechat/events", group.BasePath())
}

// SendApprovalCard sends an approval request card to Google Chat and tracks the thread
func (p *ChatPlugin) SendApprovalCard(ctx context.Context, approval *models.ApprovalRequest) error {
	if p.chatService == nil {
		return fmt.Errorf("chat service not initialized (no service account configured)")
	}

	// Build the card
	card := BuildSubmittedCard(approval, p.config.AppURL)

	// Create message in the target space
	msg := &chat.Message{
		CardsV2: []*chat.CardWithId{card},
		Text:    fmt.Sprintf("🔔 New Approval Request from %s", approval.RequestedByUser.FullName),
	}

	// Use approval ID as thread key for consistent threading
	threadKey := approval.ID.String()

	call := p.chatService.Spaces.Messages.Create(p.config.SpaceID, msg)
	call.ThreadKey(threadKey)
	call.MessageReplyOption("REPLY_MESSAGE_FALLBACK_TO_NEW_THREAD")

	result, err := call.Do()
	if err != nil {
		return fmt.Errorf("failed to send approval card: %w", err)
	}

	// Store thread mapping for future replies and comment syncing
	if result.Thread != nil {
		if err := p.threadStore.Save(approval.ID, p.config.SpaceID, result.Thread.Name); err != nil {
			log.Printf("[GoogleChat] Warning: failed to save thread mapping: %v", err)
		}
	}

	log.Printf("[GoogleChat] Approval card sent for %s, thread: %s", approval.ID, result.Thread.Name)
	return nil
}

// ReplyInThread posts a reply in the existing thread for an approval
func (p *ChatPlugin) ReplyInThread(ctx context.Context, approvalID string, cards []*chat.CardWithId, text string) error {
	if p.chatService == nil {
		return fmt.Errorf("chat service not initialized")
	}

	// Look up thread mapping
	thread, err := p.threadStore.GetByApprovalID(approvalID)
	if err != nil {
		return fmt.Errorf("thread not found for approval %s: %w", approvalID, err)
	}

	msg := &chat.Message{
		CardsV2: cards,
		Text:    text,
		Thread: &chat.Thread{
			Name: thread.ThreadName,
		},
	}

	call := p.chatService.Spaces.Messages.Create(thread.SpaceName, msg)
	call.MessageReplyOption("REPLY_MESSAGE_FALLBACK_TO_NEW_THREAD")

	if _, err := call.Do(); err != nil {
		return fmt.Errorf("failed to reply in thread: %w", err)
	}

	return nil
}

// GetDB returns the database connection (used by handler)
func (p *ChatPlugin) GetDB() *gorm.DB {
	return p.db
}

// GetConfig returns the plugin config (used by handler)
func (p *ChatPlugin) GetConfig() *config.GoogleChatConfig {
	return p.config
}

// GetApprovalService returns the approval service (used by handler)
func (p *ChatPlugin) GetApprovalService() *service.ApprovalService {
	return p.approvalService
}

// GetThreadStore returns the thread store (used by handler)
func (p *ChatPlugin) GetThreadStore() *ThreadStore {
	return p.threadStore
}

// NewHTTPClient creates a basic HTTP client for webhook fallback
func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}
