package database

import (
	"fmt"

	"github.com/yourorg/querybase/internal/auth"
	"github.com/yourorg/querybase/internal/models"
	"gorm.io/gorm"
)

// SeedData seeds the database with initial data
func SeedData(db *gorm.DB) error {
	// Check if admin user already exists
	var count int64
	db.Model(&models.User{}).Where("email = ?", "admin@querybase.local").Count(&count)

	if count > 0 {
		fmt.Println("Seed data already exists, skipping...")
		return nil
	}

	fmt.Println("Seeding database with initial data...")

	// Hash default password
	passwordHash, err := auth.HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user
	adminUser := models.User{
		Email:        "admin@querybase.local",
		Username:     "admin",
		PasswordHash: passwordHash,
		FullName:     "System Administrator",
		Role:         models.RoleAdmin,
		IsActive:     true,
	}

	if err := db.Create(&adminUser).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create default groups
	groups := []models.Group{
		{Name: "Admins", Description: "Full system access"},
		{Name: "Data Analysts", Description: "Read and write access to data sources"},
		{Name: "Data Viewers", Description: "Read-only access to data sources"},
	}

	for _, group := range groups {
		if err := db.Create(&group).Error; err != nil {
			return fmt.Errorf("failed to create group %s: %w", group.Name, err)
		}
	}

	// Assign admin to Admins group
	var adminsGroup models.Group
	if err := db.Where("name = ?", "Admins").First(&adminsGroup).Error; err != nil {
		return fmt.Errorf("failed to find Admins group: %w", err)
	}

	if err := db.Model(&adminUser).Association("Groups").Append(&adminsGroup); err != nil {
		return fmt.Errorf("failed to assign admin to Admins group: %w", err)
	}

	fmt.Println("✅ Seed data created successfully")
	fmt.Println("   Admin email: admin@querybase.local")
	fmt.Println("   Admin password: admin123 (⚠️ CHANGE IN PRODUCTION!)")

	return nil
}
