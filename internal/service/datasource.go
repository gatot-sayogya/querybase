package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/querybase/internal/api/dto"
	"github.com/yourorg/querybase/internal/models"
	gormmysql "gorm.io/driver/mysql"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DataSourceService handles data source business logic
type DataSourceService struct {
	db            *gorm.DB
	encryptionKey []byte
}

// NewDataSourceService creates a new data source service
func NewDataSourceService(db *gorm.DB, encryptionKey string) *DataSourceService {
	return &DataSourceService{
		db:            db,
		encryptionKey: []byte(encryptionKey),
	}
}

// CreateDataSource creates a new data source
func (s *DataSourceService) CreateDataSource(ctx context.Context, req *CreateDataSourceInput) (*models.DataSource, error) {
	// Encrypt password
	encryptedPassword, err := s.encryptPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}

	dataSource := &models.DataSource{
		ID:                uuid.New(),
		Name:              req.Name,
		Type:              models.DataSourceType(req.Type),
		Host:              req.Host,
		Port:              req.Port,
		DatabaseName:      req.DatabaseName,
		Username:          req.Username,
		EncryptedPassword: encryptedPassword,
		IsActive:          true,
	}

	if err := s.db.Create(dataSource).Error; err != nil {
		return nil, fmt.Errorf("failed to create data source: %w", err)
	}

	return dataSource, nil
}

// GetDataSource retrieves a data source by ID
func (s *DataSourceService) GetDataSource(ctx context.Context, dataSourceID string) (*models.DataSource, error) {
	var dataSource models.DataSource
	err := s.db.First(&dataSource, "id = ?", dataSourceID).Error
	if err != nil {
		return nil, err
	}
	return &dataSource, nil
}

// ListDataSources retrieves a list of data sources
func (s *DataSourceService) ListDataSources(ctx context.Context, limit, offset int) ([]models.DataSource, int64, error) {
	var dataSources []models.DataSource
	var total int64

	query := s.db.Model(&models.DataSource{})

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dataSources).Error

	return dataSources, total, err
}

// ListDataSourcesWithPermissions retrieves a list of data sources with their permissions
func (s *DataSourceService) ListDataSourcesWithPermissions(ctx context.Context, limit, offset int) ([]models.DataSource, int64, error) {
	var dataSources []models.DataSource
	var total int64

	query := s.db.Model(&models.DataSource{})

	// Get total count
	query.Count(&total)

	// Get paginated results with permissions preloaded
	err := query.Preload("Permissions.Group").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dataSources).Error

	return dataSources, total, err
}

// UpdateDataSource updates a data source
func (s *DataSourceService) UpdateDataSource(ctx context.Context, dataSourceID string, req *UpdateDataSourceInput) (*models.DataSource, error) {
	var dataSource models.DataSource
	if err := s.db.First(&dataSource, "id = ?", dataSourceID).Error; err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Host != "" {
		updates["host"] = req.Host
	}
	if req.Port != 0 {
		updates["port"] = req.Port
	}
	if req.DatabaseName != "" {
		updates["database_name"] = req.DatabaseName
	}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Password != "" {
		encryptedPassword, err := s.encryptPassword(req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}
		updates["encrypted_password"] = encryptedPassword
	}
	if req.Type != "" {
		updates["type"] = models.DataSourceType(req.Type)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := s.db.Model(&dataSource).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update data source: %w", err)
	}

	// Reload to get updated data
	if err := s.db.First(&dataSource, "id = ?", dataSourceID).Error; err != nil {
		return nil, err
	}

	return &dataSource, nil
}

// DeleteDataSource deletes a data source
func (s *DataSourceService) DeleteDataSource(ctx context.Context, dataSourceID string) error {
	// Delete related permissions
	if err := s.db.Where("data_source_id = ?", dataSourceID).Delete(&models.DataSourcePermission{}).Error; err != nil {
		return fmt.Errorf("failed to delete permissions: %w", err)
	}

	// Delete data source
	if err := s.db.Delete(&models.DataSource{}, "id = ?", dataSourceID).Error; err != nil {
		return fmt.Errorf("failed to delete data source: %w", err)
	}

	return nil
}

// TestConnection tests the connection to a data source
func (s *DataSourceService) TestConnection(ctx context.Context, dataSourceID string) error {
	dataSource, err := s.GetDataSource(ctx, dataSourceID)
	if err != nil {
		return fmt.Errorf("data source not found: %w", err)
	}

	// Decrypt password
	password, err := s.decryptPassword(dataSource.EncryptedPassword)
	if err != nil {
		return fmt.Errorf("failed to decrypt password: %w", err)
	}

	// Test connection based on type
	switch dataSource.Type {
	case models.DataSourceTypePostgreSQL:
		return s.testPostgreSQLConnection(dataSource, password)
	case models.DataSourceTypeMySQL:
		return s.testMySQLConnection(dataSource, password)
	default:
		return fmt.Errorf("unsupported data source type: %s", dataSource.Type)
	}
}

// TestConnectionWithParams tests connection with raw parameters
func (s *DataSourceService) TestConnectionWithParams(ctx context.Context, input *TestConnectionInput) error {
	dataSource := &models.DataSource{
		Type:         models.DataSourceType(input.Type),
		Host:         input.Host,
		Port:         input.Port,
		DatabaseName: input.DatabaseName,
		Username:     input.Username,
	}

	// Test connection based on type
	switch dataSource.Type {
	case models.DataSourceTypePostgreSQL:
		return s.testPostgreSQLConnection(dataSource, input.Password)
	case models.DataSourceTypeMySQL:
		return s.testMySQLConnection(dataSource, input.Password)
	default:
		return fmt.Errorf("unsupported data source type: %s", dataSource.Type)
	}
}

// SetPermissions sets permissions for a group on a data source
func (s *DataSourceService) SetPermissions(ctx context.Context, dataSourceID string, groupID string, permissions *PermissionInput) error {
	var perm models.DataSourcePermission

	// Check if permission exists
	err := s.db.Where("data_source_id = ? AND group_id = ?", dataSourceID, groupID).First(&perm).Error

	if err == gorm.ErrRecordNotFound {
		// Create new permission
		perm = models.DataSourcePermission{
			DataSourceID: uuid.MustParse(dataSourceID),
			GroupID:      uuid.MustParse(groupID),
			CanRead:      permissions.CanRead,
			CanWrite:     permissions.CanWrite,
			CanApprove:   permissions.CanApprove,
		}
		if err := s.db.Create(&perm).Error; err != nil {
			return fmt.Errorf("failed to create permissions: %w", err)
		}
	} else if err == nil {
		// Update existing permission
		perm.CanRead = permissions.CanRead
		perm.CanWrite = permissions.CanWrite
		perm.CanApprove = permissions.CanApprove
		if err := s.db.Save(&perm).Error; err != nil {
			return fmt.Errorf("failed to update permissions: %w", err)
		}
	} else {
		return fmt.Errorf("failed to query permissions: %w", err)
	}

	return nil
}

// GetPermissions retrieves all permissions for a data source
func (s *DataSourceService) GetPermissions(ctx context.Context, dataSourceID string) ([]models.DataSourcePermission, error) {
	var permissions []models.DataSourcePermission
	err := s.db.Preload("Group").
		Where("data_source_id = ?", dataSourceID).
		Find(&permissions).Error

	return permissions, err
}

// encryptPassword encrypts a password using AES-256-GCM
func (s *DataSourceService) encryptPassword(password string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate nonce
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := aesgcm.Seal(nonce, nonce, []byte(password), nil)

	// Encode to base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptPassword decrypts an encrypted password
func (s *DataSourceService) decryptPassword(encryptedPassword string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Extract nonce
	nonceSize := aesgcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("encrypted data too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// testPostgreSQLConnection tests a PostgreSQL connection
func (s *DataSourceService) testPostgreSQLConnection(dataSource *models.DataSource, password string) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dataSource.Host,
		dataSource.Port,
		dataSource.Username,
		password,
		dataSource.DatabaseName,
	)

	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	return sqlDB.Ping()
}

// testMySQLConnection tests a MySQL connection
func (s *DataSourceService) testMySQLConnection(dataSource *models.DataSource, password string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dataSource.Username,
		password,
		dataSource.Host,
		dataSource.Port,
		dataSource.DatabaseName,
	)

	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	return sqlDB.Ping()
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status    dto.HealthStatus
	LatencyMs int64
	Error     string
	Message   string
}

// CheckHealth performs a health check on a data source
func (s *DataSourceService) CheckHealth(ctx context.Context, dataSourceID string) (*HealthCheckResult, error) {
	// Get data source
	var dataSource models.DataSource
	if err := s.db.First(&dataSource, "id = ?", dataSourceID).Error; err != nil {
		return nil, fmt.Errorf("data source not found: %w", err)
	}

	// Check if data source is active
	if !dataSource.IsActive {
		return &HealthCheckResult{
			Status:  dto.HealthStatusUnhealthy,
			Message: "Data source is inactive",
		}, nil
	}

	// Decrypt password
	password, err := s.decryptPassword(dataSource.EncryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}

	// Measure connection latency
	start := time.Now()

	var connectionErr error
	switch dataSource.Type {
	case models.DataSourceTypePostgreSQL:
		connectionErr = s.testPostgreSQLConnection(&dataSource, password)
	case models.DataSourceTypeMySQL:
		connectionErr = s.testMySQLConnection(&dataSource, password)
	default:
		return nil, fmt.Errorf("unsupported data source type: %s", dataSource.Type)
	}

	latency := time.Since(start).Milliseconds()

	// Determine health status based on connection result
	if connectionErr != nil {
		return &HealthCheckResult{
			Status:    dto.HealthStatusUnhealthy,
			LatencyMs: latency,
			Error:     connectionErr.Error(),
			Message:   "Failed to connect to data source",
		}, nil
	}

	// Determine health based on latency
	status := dto.HealthStatusHealthy
	message := "Data source is healthy"
	if latency > 1000 {
		status = dto.HealthStatusDegraded
		message = "Data source is responding slowly"
	}

	return &HealthCheckResult{
		Status:    status,
		LatencyMs: latency,
		Message:   message,
	}, nil
}

// CreateDataSourceInput represents input for creating a data source
type CreateDataSourceInput struct {
	Name         string
	Type         string
	Host         string
	Port         int
	DatabaseName string
	Username     string
	Password     string
}

// UpdateDataSourceInput represents input for updating a data source
type UpdateDataSourceInput struct {
	Name         string
	Type         string
	Host         string
	Port         int
	DatabaseName string
	Username     string
	Password     string
	IsActive     *bool
}

// PermissionInput represents permission settings
type PermissionInput struct {
	CanRead    bool
	CanWrite   bool
	CanApprove bool
}

// TestConnectionInput represents input for testing a connection
type TestConnectionInput struct {
	Type         string `json:"type" binding:"required,oneof=postgresql mysql"`
	Host         string `json:"host" binding:"required"`
	Port         int    `json:"port" binding:"required,min=1,max=65535"`
	DatabaseName string `json:"database_name" binding:"required"`
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
}
