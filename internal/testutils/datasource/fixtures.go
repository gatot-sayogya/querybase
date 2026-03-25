package datasource

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/yourorg/querybase/internal/models"
)

// Default encryption key for test data sources (32 bytes for AES-256)
const defaultTestEncryptionKey = "test-encryption-key-32-bytes!!!!"

// CreateTestPostgreSQLDataSource creates a test PostgreSQL data source.
// The data source is automatically cleaned up when the test completes.
func CreateTestPostgreSQLDataSource(t *testing.T, db *gorm.DB, name string) *models.DataSource {
	t.Helper()

	password := fmt.Sprintf("testpass_%s", uuid.New().String()[:8])
	encryptedPassword, err := EncryptPassword(password, defaultTestEncryptionKey)
	if err != nil {
		t.Fatalf("Failed to encrypt password: %v", err)
	}

	dataSource := &models.DataSource{
		ID:                uuid.New(),
		Name:              name,
		Type:              models.DataSourceTypePostgreSQL,
		Host:              "localhost",
		Port:              5432,
		DatabaseName:      fmt.Sprintf("test_db_%s", uuid.New().String()[:8]),
		Username:          "testuser",
		EncryptedPassword: encryptedPassword,
		IsActive:          true,
		IsHealthy:         true,
		AuditRowThreshold: 1000,
		AuditCapability:   models.AuditCapabilityUnknown,
	}

	if err := db.Create(dataSource).Error; err != nil {
		t.Fatalf("Failed to create test data source: %v", err)
	}

	t.Cleanup(func() {
		db.Unscoped().Delete(dataSource)
	})

	return dataSource
}

// CreateTestMySQLDataSource creates a test MySQL data source.
// The data source is automatically cleaned up when the test completes.
func CreateTestMySQLDataSource(t *testing.T, db *gorm.DB, name string) *models.DataSource {
	t.Helper()

	password := fmt.Sprintf("testpass_%s", uuid.New().String()[:8])
	encryptedPassword, err := EncryptPassword(password, defaultTestEncryptionKey)
	if err != nil {
		t.Fatalf("Failed to encrypt password: %v", err)
	}

	dataSource := &models.DataSource{
		ID:                uuid.New(),
		Name:              name,
		Type:              models.DataSourceTypeMySQL,
		Host:              "localhost",
		Port:              3306,
		DatabaseName:      fmt.Sprintf("test_db_%s", uuid.New().String()[:8]),
		Username:          "testuser",
		EncryptedPassword: encryptedPassword,
		IsActive:          true,
		IsHealthy:         true,
		AuditRowThreshold: 1000,
		AuditCapability:   models.AuditCapabilityUnknown,
	}

	if err := db.Create(dataSource).Error; err != nil {
		t.Fatalf("Failed to create test data source: %v", err)
	}

	t.Cleanup(func() {
		db.Unscoped().Delete(dataSource)
	})

	return dataSource
}

// CreateTestDataSourceWithPerms creates a test data source with group permissions.
// The data source and permissions are automatically cleaned up when the test completes.
func CreateTestDataSourceWithPerms(t *testing.T, db *gorm.DB, name string, groupID uuid.UUID, canRead, canWrite, canApprove bool) *models.DataSource {
	t.Helper()

	// Create the data source
	ds := CreateTestPostgreSQLDataSource(t, db, name)

	// Create the permission
	permission := &models.DataSourcePermission{
		ID:           uuid.New(),
		DataSourceID: ds.ID,
		GroupID:      groupID,
		CanRead:      canRead,
		CanWrite:     canWrite,
		CanApprove:   canApprove,
	}

	if err := db.Create(permission).Error; err != nil {
		t.Fatalf("Failed to create data source permission: %v", err)
	}

	t.Cleanup(func() {
		db.Unscoped().Delete(permission)
	})

	return ds
}

// CreateTestDataSourceWithUniqueName creates a test PostgreSQL data source with a unique name.
func CreateTestDataSourceWithUniqueName(t *testing.T, db *gorm.DB) *models.DataSource {
	t.Helper()
	uniqueName := fmt.Sprintf("test-ds-%s", uuid.New().String()[:8])
	return CreateTestPostgreSQLDataSource(t, db, uniqueName)
}

// CreateTestMySQLDataSourceWithUniqueName creates a test MySQL data source with a unique name.
func CreateTestMySQLDataSourceWithUniqueName(t *testing.T, db *gorm.DB) *models.DataSource {
	t.Helper()
	uniqueName := fmt.Sprintf("test-mysql-ds-%s", uuid.New().String()[:8])
	return CreateTestMySQLDataSource(t, db, uniqueName)
}

// CreateTestDataSourceWithType creates a test data source with a specific type.
func CreateTestDataSourceWithType(t *testing.T, db *gorm.DB, name string, dsType models.DataSourceType) *models.DataSource {
	t.Helper()

	switch dsType {
	case models.DataSourceTypeMySQL:
		return CreateTestMySQLDataSource(t, db, name)
	case models.DataSourceTypePostgreSQL:
		fallthrough
	default:
		return CreateTestPostgreSQLDataSource(t, db, name)
	}
}

// CreateInactiveTestDataSource creates an inactive test data source.
func CreateInactiveTestDataSource(t *testing.T, db *gorm.DB, name string) *models.DataSource {
	t.Helper()

	password := fmt.Sprintf("testpass_%s", uuid.New().String()[:8])
	encryptedPassword, err := EncryptPassword(password, defaultTestEncryptionKey)
	if err != nil {
		t.Fatalf("Failed to encrypt password: %v", err)
	}

	dataSource := &models.DataSource{
		ID:                uuid.New(),
		Name:              name,
		Type:              models.DataSourceTypePostgreSQL,
		Host:              "localhost",
		Port:              5432,
		DatabaseName:      fmt.Sprintf("test_db_%s", uuid.New().String()[:8]),
		Username:          "testuser",
		EncryptedPassword: encryptedPassword,
		IsActive:          false,
		IsHealthy:         false,
		AuditRowThreshold: 1000,
		AuditCapability:   models.AuditCapabilityUnknown,
	}

	if err := db.Select("ID", "Name", "Type", "Host", "Port", "DatabaseName", "Username", "EncryptedPassword", "AuditRowThreshold", "AuditCapability").Create(dataSource).Error; err != nil {
		t.Fatalf("Failed to create inactive test data source: %v", err)
	}

	db.Model(dataSource).Updates(map[string]interface{}{
		"is_active":  false,
		"is_healthy": false,
	})

	t.Cleanup(func() {
		db.Unscoped().Delete(dataSource)
	})

	return dataSource
}

// EncryptPassword encrypts a password using AES-256-GCM.
func EncryptPassword(password, encryptionKey string) (string, error) {
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesgcm.Seal(nonce, nonce, []byte(password), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPassword decrypts an encrypted password.
func DecryptPassword(encryptedPassword, encryptionKey string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesgcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("encrypted data too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptTestPassword is a convenience wrapper for EncryptPassword with the default test key.
func EncryptTestPassword(password string) string {
	encrypted, err := EncryptPassword(password, defaultTestEncryptionKey)
	if err != nil {
		// This should never happen with a valid test key
		panic(fmt.Sprintf("failed to encrypt test password: %v", err))
	}
	return encrypted
}

// GetTestEncryptionKey returns the default test encryption key.
func GetTestEncryptionKey() string {
	return defaultTestEncryptionKey
}
