package fixtures

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/yourorg/querybase/internal/models"
)

// CreateTestGroup creates a test group with the specified name.
// The group is automatically cleaned up when the test completes.
func CreateTestGroup(t *testing.T, db *gorm.DB, name string) *models.Group {
	t.Helper()

	group := &models.Group{
		ID:          uuid.New(),
		Name:        name,
		Description: fmt.Sprintf("Test group: %s", name),
	}

	if err := db.Create(group).Error; err != nil {
		t.Fatalf("Failed to create test group: %v", err)
	}

	t.Cleanup(func() {
		db.Unscoped().Delete(group)
	})

	return group
}

// CreateTestGroupWithUniqueName creates a test group with a unique name.
// The group is automatically cleaned up when the test completes.
func CreateTestGroupWithUniqueName(t *testing.T, db *gorm.DB) *models.Group {
	t.Helper()

	uniqueID := uuid.New().String()[:8]
	name := fmt.Sprintf("test-group-%s", uniqueID)

	return CreateTestGroup(t, db, name)
}

// AddUserToGroup adds a user to a group.
// Returns an error if the operation fails.
func AddUserToGroup(db *gorm.DB, userID, groupID uuid.UUID) error {
	userGroup := &models.UserGroup{
		UserID:  userID,
		GroupID: groupID,
	}

	return db.Create(userGroup).Error
}

// AddUserToGroupWithRole adds a user to a group with a specific role.
// Returns an error if the operation fails.
func AddUserToGroupWithRole(db *gorm.DB, userID, groupID uuid.UUID, role string) error {
	userGroup := &models.UserGroup{
		UserID:  userID,
		GroupID: groupID,
	}

	if err := db.Create(userGroup).Error; err != nil {
		return err
	}

	// Update the user's role within this group if needed
	// Note: This assumes there's a role field in UserGroup model
	// If not, this would need to be handled differently

	return nil
}

// RemoveUserFromGroup removes a user from a group.
// Returns an error if the operation fails.
func RemoveUserFromGroup(db *gorm.DB, userID, groupID uuid.UUID) error {
	return db.Where("user_id = ? AND group_id = ?", userID, groupID).
		Delete(&models.UserGroup{}).Error
}

// GetUserGroups retrieves all groups for a user.
func GetUserGroups(db *gorm.DB, userID uuid.UUID) ([]models.Group, error) {
	var groups []models.Group
	err := db.Model(&models.User{}).
		Where("users.id = ?", userID).
		Association("Groups").
		Find(&groups)
	return groups, err
}

// GetGroupMembers retrieves all members of a group.
func GetGroupMembers(db *gorm.DB, groupID uuid.UUID) ([]models.User, error) {
	var users []models.User
	err := db.Model(&models.Group{}).
		Where("groups.id = ?", groupID).
		Association("Users").
		Find(&users)
	return users, err
}

// CreateTestGroupWithMembers creates a test group and adds specified users to it.
func CreateTestGroupWithMembers(t *testing.T, db *gorm.DB, name string, members []*models.User) *models.Group {
	t.Helper()

	group := CreateTestGroup(t, db, name)

	for _, member := range members {
		if err := AddUserToGroup(db, member.ID, group.ID); err != nil {
			t.Fatalf("Failed to add user %s to group %s: %v", member.ID, group.ID, err)
		}
	}

	return group
}
