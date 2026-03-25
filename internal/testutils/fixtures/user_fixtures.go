package fixtures

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/yourorg/querybase/internal/auth"
	"github.com/yourorg/querybase/internal/models"
)

// CreateTestUser creates a test user with the specified role.
// The user is automatically cleaned up when the test completes.
func CreateTestUser(t *testing.T, db *gorm.DB, role models.UserRole) *models.User {
	t.Helper()

	passwordHash, err := auth.HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
		Username:     fmt.Sprintf("testuser-%s", uuid.New().String()[:8]),
		PasswordHash: passwordHash,
		FullName:     "Test User",
		Role:         role,
		IsActive:     true,
	}

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Cleanup(func() {
		db.Unscoped().Delete(user)
	})

	return user
}

// CreateTestUserWithEmail creates a test user with a specific email.
// The user is automatically cleaned up when the test completes.
func CreateTestUserWithEmail(t *testing.T, db *gorm.DB, email string, role models.UserRole) *models.User {
	t.Helper()

	passwordHash, err := auth.HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		Username:     fmt.Sprintf("testuser-%s", uuid.New().String()[:8]),
		PasswordHash: passwordHash,
		FullName:     "Test User",
		Role:         role,
		IsActive:     true,
	}

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Cleanup(func() {
		db.Unscoped().Delete(user)
	})

	return user
}

// CreateTestAdminUser creates a test user with admin role.
func CreateTestAdminUser(t *testing.T, db *gorm.DB) *models.User {
	return CreateTestUser(t, db, models.RoleAdmin)
}

// CreateTestRegularUser creates a test user with regular user role.
func CreateTestRegularUser(t *testing.T, db *gorm.DB) *models.User {
	return CreateTestUser(t, db, models.RoleUser)
}

// CreateTestViewerUser creates a test user with viewer role.
func CreateTestViewerUser(t *testing.T, db *gorm.DB) *models.User {
	return CreateTestUser(t, db, models.RoleViewer)
}

// CreateInactiveTestUser creates a test user that is inactive.
func CreateInactiveTestUser(t *testing.T, db *gorm.DB) *models.User {
	t.Helper()

	passwordHash, err := auth.HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("inactive-%s@example.com", uuid.New().String()[:8]),
		Username:     fmt.Sprintf("inactiveuser-%s", uuid.New().String()[:8]),
		PasswordHash: passwordHash,
		FullName:     "Inactive Test User",
		Role:         models.RoleUser,
		IsActive:     false,
	}

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create inactive test user: %v", err)
	}

	t.Cleanup(func() {
		db.Unscoped().Delete(user)
	})

	return user
}
