package googlechat

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/yourorg/querybase/internal/models"
	"gorm.io/gorm"
)

// ThreadStore handles ChatThread CRUD operations
type ThreadStore struct {
	db *gorm.DB
}

// NewThreadStore creates a new thread store
func NewThreadStore(db *gorm.DB) *ThreadStore {
	return &ThreadStore{db: db}
}

// Save creates a new thread ↔ approval mapping
func (s *ThreadStore) Save(approvalID uuid.UUID, spaceName, threadName string) error {
	thread := &models.ChatThread{
		ID:         uuid.New(),
		ApprovalID: approvalID,
		SpaceName:  spaceName,
		ThreadName: threadName,
	}

	// Upsert: if thread already exists for this approval, update it
	result := s.db.Where("approval_id = ?", approvalID).FirstOrCreate(thread)
	if result.Error != nil {
		return fmt.Errorf("failed to save chat thread: %w", result.Error)
	}

	// Update thread name if it changed
	if result.RowsAffected == 0 {
		if err := s.db.Model(thread).Where("approval_id = ?", approvalID).Updates(map[string]interface{}{
			"space_name":  spaceName,
			"thread_name": threadName,
		}).Error; err != nil {
			return fmt.Errorf("failed to update chat thread: %w", err)
		}
	}

	return nil
}

// GetByApprovalID retrieves a thread mapping by approval ID
func (s *ThreadStore) GetByApprovalID(approvalID string) (*models.ChatThread, error) {
	parsedID, err := uuid.Parse(approvalID)
	if err != nil {
		return nil, fmt.Errorf("invalid approval ID: %w", err)
	}

	var thread models.ChatThread
	if err := s.db.Where("approval_id = ?", parsedID).First(&thread).Error; err != nil {
		return nil, fmt.Errorf("thread not found for approval %s: %w", approvalID, err)
	}

	return &thread, nil
}

// GetByThreadName retrieves a thread mapping by Google Chat thread name
// Used for comment syncing: message in thread → look up approval ID
func (s *ThreadStore) GetByThreadName(threadName string) (*models.ChatThread, error) {
	var thread models.ChatThread
	if err := s.db.Where("thread_name = ?", threadName).First(&thread).Error; err != nil {
		return nil, fmt.Errorf("no approval found for thread %s: %w", threadName, err)
	}

	return &thread, nil
}
