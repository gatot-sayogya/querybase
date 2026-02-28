package googlechat

import (
	"fmt"

	"github.com/yourorg/querybase/internal/models"
	"gorm.io/gorm"
)

// ResolveUser finds a QueryBase user by their email address
// Used to map Google Chat user (identified by email) to a QueryBase user
func ResolveUser(db *gorm.DB, email string) (*models.User, error) {
	var user models.User
	if err := db.Where("email = ? AND is_active = ?", email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no active QueryBase user found with email: %s", email)
		}
		return nil, fmt.Errorf("failed to look up user by email: %w", err)
	}

	return &user, nil
}
