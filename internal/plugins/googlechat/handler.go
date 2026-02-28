package googlechat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourorg/querybase/internal/models"
	"github.com/yourorg/querybase/internal/service"
	"google.golang.org/api/chat/v1"
)

// ChatEvent represents a Google Chat interaction event
type ChatEvent struct {
	Type              string       `json:"type"`
	EventTime         string       `json:"eventTime"`
	Space             *SpaceInfo   `json:"space"`
	Message           *MsgInfo     `json:"message"`
	User              *UserInfo    `json:"user"`
	Action            *ActionInfo  `json:"action"`
	CommonEventObject *CommonEvent `json:"commonEventObject"`
}

// SpaceInfo contains space information
type SpaceInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// MsgInfo contains message information
type MsgInfo struct {
	Name   string      `json:"name"`
	Text   string      `json:"text"`
	Thread *ThreadInfo `json:"thread"`
}

// ThreadInfo contains thread information
type ThreadInfo struct {
	Name string `json:"name"`
}

// UserInfo contains Google Chat user information
type UserInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Type        string `json:"type"`
}

// ActionInfo contains card action information
type ActionInfo struct {
	ActionMethodName string         `json:"actionMethodName"`
	Parameters       []*ActionParam `json:"parameters"`
}

// ActionParam represents an action parameter
type ActionParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CommonEvent contains common event object
type CommonEvent struct {
	InvokedFunction string            `json:"invokedFunction"`
	Parameters      map[string]string `json:"parameters"`
}

// EventHandler handles Google Chat events
type EventHandler struct {
	plugin *ChatPlugin
}

// NewEventHandler creates a new event handler
func NewEventHandler(plugin *ChatPlugin) *EventHandler {
	return &EventHandler{plugin: plugin}
}

// HandleEvent is the main HTTP handler for Google Chat events
func (h *EventHandler) HandleEvent(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[GoogleChat] Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var event ChatEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("[GoogleChat] Failed to parse event: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event"})
		return
	}

	log.Printf("[GoogleChat] Event received: type=%s, user=%s", event.Type, getUserEmail(&event))

	switch event.Type {
	case "CARD_CLICKED":
		h.handleCardClicked(c, &event)
	case "MESSAGE":
		h.handleMessage(c, &event)
	case "ADDED_TO_SPACE":
		h.handleAddedToSpace(c, &event)
	default:
		log.Printf("[GoogleChat] Unknown event type: %s", event.Type)
		c.JSON(http.StatusOK, gin.H{})
	}
}

// handleCardClicked processes button click events
func (h *EventHandler) handleCardClicked(c *gin.Context, event *ChatEvent) {
	actionName := getActionName(event)
	paramID := getActionParam(event, "id")

	log.Printf("[GoogleChat] Card clicked: action=%s, id=%s", actionName, paramID)

	switch actionName {
	case "action_approve":
		h.handleApprove(c, event, paramID)
	case "action_reject":
		h.handleReject(c, event, paramID)
	case "action_commit":
		h.handleCommit(c, event, paramID)
	case "action_rollback":
		h.handleRollback(c, event, paramID)
	default:
		log.Printf("[GoogleChat] Unknown action: %s", actionName)
		c.JSON(http.StatusOK, gin.H{})
	}
}

// handleApprove processes approval button clicks
func (h *EventHandler) handleApprove(c *gin.Context, event *ChatEvent, approvalID string) {
	ctx := context.Background()

	// Resolve the Google Chat user to a QueryBase user
	user, err := ResolveUser(h.plugin.GetDB(), getUserEmail(event))
	if err != nil {
		log.Printf("[GoogleChat] User resolution failed: %v", err)
		respondWithCard(c, BuildErrorCard("User Not Found",
			"Your Google Chat email is not registered in QueryBase. Please contact your administrator."))
		return
	}

	// Parse approval ID
	parsedID, err := uuid.Parse(approvalID)
	if err != nil {
		respondWithCard(c, BuildErrorCard("Invalid Request", "Invalid approval ID."))
		return
	}

	// Submit the approval review
	review := &service.ReviewInput{
		ApprovalID: parsedID,
		ReviewerID: user.ID.String(),
		Decision:   models.ApprovalDecisionApproved,
		Comments:   fmt.Sprintf("Approved via Google Chat by %s", user.FullName),
	}

	_, err = h.plugin.GetApprovalService().ReviewApproval(ctx, review)
	if err != nil {
		log.Printf("[GoogleChat] Review failed: %v", err)
		respondWithCard(c, BuildErrorCard("Approval Failed", err.Error()))
		return
	}

	// Get the full approval for card display
	approval, err := h.plugin.GetApprovalService().GetApproval(ctx, approvalID)
	if err != nil {
		log.Printf("[GoogleChat] Failed to get approval: %v", err)
		respondWithCard(c, BuildErrorCard("Error", "Approval processed but failed to load details."))
		return
	}

	// Synchronous response: approved card
	respondWithCard(c, BuildApprovedCard(approval, user))

	// Async: start transaction and post preview in thread
	go h.startTransactionAsync(ctx, approvalID, user)
}

// handleReject processes rejection button clicks
func (h *EventHandler) handleReject(c *gin.Context, event *ChatEvent, approvalID string) {
	ctx := context.Background()

	user, err := ResolveUser(h.plugin.GetDB(), getUserEmail(event))
	if err != nil {
		respondWithCard(c, BuildErrorCard("User Not Found",
			"Your Google Chat email is not registered in QueryBase."))
		return
	}

	parsedID, err := uuid.Parse(approvalID)
	if err != nil {
		respondWithCard(c, BuildErrorCard("Invalid Request", "Invalid approval ID."))
		return
	}

	review := &service.ReviewInput{
		ApprovalID: parsedID,
		ReviewerID: user.ID.String(),
		Decision:   models.ApprovalDecisionRejected,
		Comments:   fmt.Sprintf("Rejected via Google Chat by %s", user.FullName),
	}

	_, err = h.plugin.GetApprovalService().ReviewApproval(ctx, review)
	if err != nil {
		respondWithCard(c, BuildErrorCard("Rejection Failed", err.Error()))
		return
	}

	approval, err := h.plugin.GetApprovalService().GetApproval(ctx, approvalID)
	if err != nil {
		respondWithCard(c, BuildErrorCard("Error", "Rejection processed but failed to load details."))
		return
	}

	respondWithCard(c, BuildRejectedCard(approval, user, ""))
}

// handleCommit processes transaction commit button clicks
func (h *EventHandler) handleCommit(c *gin.Context, event *ChatEvent, transactionID string) {
	ctx := context.Background()

	_, err := ResolveUser(h.plugin.GetDB(), getUserEmail(event))
	if err != nil {
		respondWithCard(c, BuildErrorCard("User Not Found",
			"Your Google Chat email is not registered in QueryBase."))
		return
	}

	err = h.plugin.GetApprovalService().CommitTransaction(ctx, transactionID)
	if err != nil {
		respondWithCard(c, BuildErrorCard("Commit Failed", err.Error()))
		return
	}

	// Get transaction for display
	var txn models.QueryTransaction
	if err := h.plugin.GetDB().First(&txn, "id = ?", transactionID).Error; err != nil {
		respondWithCard(c, BuildErrorCard("Error", "Transaction committed but failed to load details."))
		return
	}

	respondWithCard(c, BuildCommittedCard(&txn))
}

// handleRollback processes transaction rollback button clicks
func (h *EventHandler) handleRollback(c *gin.Context, event *ChatEvent, transactionID string) {
	ctx := context.Background()

	_, err := ResolveUser(h.plugin.GetDB(), getUserEmail(event))
	if err != nil {
		respondWithCard(c, BuildErrorCard("User Not Found",
			"Your Google Chat email is not registered in QueryBase."))
		return
	}

	err = h.plugin.GetApprovalService().RollbackTransaction(ctx, transactionID)
	if err != nil {
		respondWithCard(c, BuildErrorCard("Rollback Failed", err.Error()))
		return
	}

	var txn models.QueryTransaction
	if err := h.plugin.GetDB().First(&txn, "id = ?", transactionID).Error; err != nil {
		respondWithCard(c, BuildErrorCard("Error", "Rollback completed but failed to load details."))
		return
	}

	respondWithCard(c, BuildRolledBackCard(&txn))
}

// handleMessage processes text messages — syncs thread messages as approval comments
func (h *EventHandler) handleMessage(c *gin.Context, event *ChatEvent) {
	// Only process messages in threads we're tracking
	if event.Message == nil || event.Message.Thread == nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	threadName := event.Message.Thread.Name
	chatThread, err := h.plugin.GetThreadStore().GetByThreadName(threadName)
	if err != nil {
		// Not a tracked thread, ignore
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// Skip bot messages
	if event.User != nil && event.User.Type == "BOT" {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// Resolve user
	email := getUserEmail(event)
	user, err := ResolveUser(h.plugin.GetDB(), email)
	if err != nil {
		log.Printf("[GoogleChat] Can't sync message as comment — user %s not found: %v", email, err)
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// Save as approval comment
	messageText := ""
	if event.Message != nil {
		messageText = event.Message.Text
	}

	if messageText == "" {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	ctx := context.Background()
	comment := fmt.Sprintf("[Google Chat] %s", messageText)
	_, err = h.plugin.GetApprovalService().AddComment(ctx, chatThread.ApprovalID.String(), user.ID.String(), comment)
	if err != nil {
		log.Printf("[GoogleChat] Failed to save comment: %v", err)
	} else {
		log.Printf("[GoogleChat] Saved thread message as comment for approval %s", chatThread.ApprovalID)
	}

	c.JSON(http.StatusOK, gin.H{})
}

// handleAddedToSpace responds when the bot is added to a space
func (h *EventHandler) handleAddedToSpace(c *gin.Context, event *ChatEvent) {
	c.JSON(http.StatusOK, gin.H{
		"text": "👋 QueryBase connected! I'll send approval notifications and handle interactive approvals here.",
	})
}

// startTransactionAsync starts a transaction and posts the preview in the thread
func (h *EventHandler) startTransactionAsync(ctx context.Context, approvalID string, user *models.User) {
	txn, err := h.plugin.GetApprovalService().StartTransaction(ctx, approvalID, user.ID.String())
	if err != nil {
		log.Printf("[GoogleChat] Failed to start transaction for %s: %v", approvalID, err)
		// Post error in thread
		errCard := BuildErrorCard("Transaction Failed",
			fmt.Sprintf("Failed to start transaction: %s", err.Error()))
		if replyErr := h.plugin.ReplyInThread(ctx, approvalID, []*chat.CardWithId{errCard}, "⚠️ Transaction Error"); replyErr != nil {
			log.Printf("[GoogleChat] Failed to reply with error: %v", replyErr)
		}
		return
	}

	// Get the approval for display
	approval, err := h.plugin.GetApprovalService().GetApproval(ctx, approvalID)
	if err != nil {
		log.Printf("[GoogleChat] Failed to get approval for preview: %v", err)
		return
	}

	// Post preview card with Commit/Rollback buttons
	previewCard := BuildPreviewCard(approval, txn)
	text := fmt.Sprintf("📊 Transaction preview ready — %d rows affected", txn.AffectedRows)
	if err := h.plugin.ReplyInThread(ctx, approvalID, []*chat.CardWithId{previewCard}, text); err != nil {
		log.Printf("[GoogleChat] Failed to post preview: %v", err)
	}
}

// --- Helper functions ---

func getUserEmail(event *ChatEvent) string {
	if event.User != nil {
		return event.User.Email
	}
	return ""
}

func getActionName(event *ChatEvent) string {
	// Try commonEventObject first (newer format)
	if event.CommonEventObject != nil && event.CommonEventObject.InvokedFunction != "" {
		return event.CommonEventObject.InvokedFunction
	}
	// Fall back to action
	if event.Action != nil {
		return event.Action.ActionMethodName
	}
	return ""
}

func getActionParam(event *ChatEvent, key string) string {
	// Try commonEventObject parameters first
	if event.CommonEventObject != nil && event.CommonEventObject.Parameters != nil {
		if val, ok := event.CommonEventObject.Parameters[key]; ok {
			return val
		}
	}
	// Fall back to action parameters
	if event.Action != nil {
		for _, p := range event.Action.Parameters {
			if p.Key == key {
				return p.Value
			}
		}
	}
	return ""
}

// respondWithCard sends a synchronous card response to Google Chat
func respondWithCard(c *gin.Context, card *chat.CardWithId) {
	c.JSON(http.StatusOK, gin.H{
		"cardsV2": []*chat.CardWithId{card},
	})
}
