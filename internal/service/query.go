package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/querybase/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ActiveTransaction represents an active database transaction
type ActiveTransaction struct {
	DB             *gorm.DB
	DataSourceID   uuid.UUID
	StartedAt      time.Time
	LastActivityAt time.Time
}

// QueryService handles query execution logic
type QueryService struct {
	db            *gorm.DB
	encryptionKey []byte
	// Map to store active transactions per data source
	// Key: data_source_id
	activeTransactions map[uuid.UUID]*ActiveTransaction
	txMutex            sync.RWMutex
}

// NewQueryService creates a new query service
func NewQueryService(db *gorm.DB, encryptionKey string) *QueryService {
	return &QueryService{
		db:                 db,
		encryptionKey:      []byte(encryptionKey),
		activeTransactions: make(map[uuid.UUID]*ActiveTransaction),
	}
}

// ExecuteQuery executes a SQL query on a data source
func (s *QueryService) ExecuteQuery(ctx context.Context, query *models.Query, dataSource *models.DataSource) (*models.QueryResult, error) {
	// Detect operation type
	operationType := DetectOperationType(query.QueryText)

	// For write operations, we should not execute directly (should go through approval)
	if operationType != models.OperationSelect {
		return nil, fmt.Errorf("write operations require approval")
	}

	// Get database connection
	dataSourceDB, err := s.connectToDataSource(dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to data source: %w", err)
	}

	// Execute the query
	rows, err := dataSourceDB.Raw(query.QueryText).Rows()
	if err != nil {
		// Update query status to failed
		s.db.Model(query).Updates(map[string]interface{}{
			"status":      models.StatusFailed,
			"error_message": err.Error(),
		})

		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	// Parse results
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		// Create a slice of interface{} to hold each column value
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			row[col] = v
		}
		results = append(results, row)
	}

	// Serialize results to JSON
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize results: %w", err)
	}

	// Serialize column names and types to JSON strings for storage
	columnNamesJSON, err := json.Marshal(columns)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize column names: %w", err)
	}

	// Note: ColumnTypes is set to empty strings as we don't have type information without additional schema query
	columnTypes := make([]string, len(columns))
	columnTypesJSON, err := json.Marshal(columnTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize column types: %w", err)
	}

	// Create query result
	queryResult := &models.QueryResult{
		QueryID:     query.ID,
		RowCount:    len(results),
		ColumnNames: string(columnNamesJSON), // Store as JSON string
		ColumnTypes: string(columnTypesJSON), // Store as JSON string
		Data:        string(resultsJSON),
		StoredAt:    time.Now(),
	}

	// Save result to database
	if err := s.db.Create(queryResult).Error; err != nil {
		return nil, fmt.Errorf("failed to save query result: %w", err)
	}

	// Update query status
	s.db.Model(query).Updates(map[string]interface{}{
		"status": models.StatusCompleted,
	})

	// Create query history entry
	rowCount := len(results)
	queryHistory := &models.QueryHistory{
		QueryID:       &query.ID,
		UserID:        query.UserID,
		DataSourceID:  query.DataSourceID,
		QueryText:     query.QueryText,
		OperationType: query.OperationType,
		Status:        models.StatusCompleted,
		RowCount:      &rowCount,
		ExecutedAt:    time.Now(),
	}
	s.db.Create(queryHistory)

	return queryResult, nil
}

// SaveQuery saves a new query
func (s *QueryService) SaveQuery(ctx context.Context, query *models.Query) error {
	return s.db.Create(query).Error
}

// GetQuery retrieves a query by ID
func (s *QueryService) GetQuery(ctx context.Context, queryID string) (*models.Query, error) {
	var query models.Query
	err := s.db.Preload("User").First(&query, "id = ?", queryID).Error
	if err != nil {
		return nil, err
	}
	return &query, nil
}

// ListQueries retrieves a list of queries with pagination
func (s *QueryService) ListQueries(ctx context.Context, userID string, limit, offset int) ([]models.Query, int64, error) {
	var queries []models.Query
	var total int64

	query := s.db.Model(&models.Query{})

	// If not admin, only show user's own queries
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err == nil {
		if user.Role != models.RoleAdmin {
			query = query.Where("user_id = ?", userID)
		}
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&queries).Error

	return queries, total, err
}

// ExplainQueryResult represents the result of an EXPLAIN query
type ExplainQueryResult struct {
	Plan      []map[string]interface{} `json:"plan"`
	RawOutput string                    `json:"raw_output"`
}

// DryRunResult represents the result of a DELETE dry run
type DryRunResult struct {
	AffectedRows int                      `json:"affected_rows"`
	Rows         []map[string]interface{} `json:"rows"`
	Query        string                   `json:"query"`
}

// ExplainQuery executes an EXPLAIN or EXPLAIN ANALYZE query
func (s *QueryService) ExplainQuery(ctx context.Context, queryText string, dataSource *models.DataSource, analyze bool) (*ExplainQueryResult, error) {
	// Get database connection
	dataSourceDB, err := s.connectToDataSource(dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to data source: %w", err)
	}

	// Build EXPLAIN query
	explainQuery := "EXPLAIN"
	if analyze {
		explainQuery += " ANALYZE"
	}
	explainQuery += " " + queryText

	// Execute EXPLAIN query
	rows, err := dataSourceDB.Raw(explainQuery).Rows()
	if err != nil {
		return nil, fmt.Errorf("EXPLAIN query failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Parse results
	var plan []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			row[col] = v
		}
		plan = append(plan, row)
	}

	// Serialize to JSON for raw output
	planJSON, _ := json.MarshalIndent(plan, "", "  ")

	return &ExplainQueryResult{
		Plan:      plan,
		RawOutput: string(planJSON),
	}, nil
}

// DryRunDelete converts a DELETE query to SELECT and shows affected rows
func (s *QueryService) DryRunDelete(ctx context.Context, queryText string, dataSource *models.DataSource) (*DryRunResult, error) {
	// Detect operation type
	operationType := DetectOperationType(queryText)
	if operationType != models.OperationDelete {
		return nil, fmt.Errorf("dry run is only supported for DELETE queries")
	}

	// Convert DELETE to SELECT
	// Pattern: DELETE FROM table WHERE ... -> SELECT * FROM table WHERE ...
	selectQuery := convertDeleteToSelect(queryText)

	// Get database connection
	dataSourceDB, err := s.connectToDataSource(dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to data source: %w", err)
	}

	// Execute SELECT query
	rows, err := dataSourceDB.Raw(selectQuery).Rows()
	if err != nil {
		return nil, fmt.Errorf("dry run query failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Parse results
	var resultRows []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			row[col] = v
		}
		resultRows = append(resultRows, row)
	}

	return &DryRunResult{
		AffectedRows: len(resultRows),
		Rows:         resultRows,
		Query:        selectQuery,
	}, nil
}

// convertDeleteToSelect converts a DELETE query to a SELECT query
func convertDeleteToSelect(queryText string) string {
	trimmedSQL := strings.TrimSpace(queryText)
	upperSQL := strings.ToUpper(trimmedSQL)

	// Remove comments
	trimmedSQL = SanitizeSQL(trimmedSQL)
	upperSQL = strings.ToUpper(trimmedSQL)

	// Pattern: DELETE FROM table_name [WHERE ...]
	// Replace with: SELECT * FROM table_name [WHERE ...]

	// Find the DELETE FROM keyword
	deleteFromIndex := strings.Index(upperSQL, "DELETE FROM")
	if deleteFromIndex == -1 {
		return queryText // Not a DELETE query, return as-is
	}

	// Extract the part after "DELETE FROM"
	afterDeleteFrom := trimmedSQL[deleteFromIndex+11:] // len("DELETE FROM") = 11

	// Trim leading whitespace
	afterDeleteFrom = strings.TrimLeft(afterDeleteFrom, " \t\n\r")

	// Build SELECT query
	selectQuery := "SELECT * FROM " + afterDeleteFrom

	return selectQuery
}

// ListQueryHistory retrieves query history with pagination
func (s *QueryService) ListQueryHistory(ctx context.Context, userID string, limit, offset int) ([]models.QueryHistory, int64, error) {
	var history []models.QueryHistory
	var total int64

	query := s.db.Model(&models.QueryHistory{})

	// If not admin, only show user's own history
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err == nil {
		if user.Role != models.RoleAdmin {
			query = query.Where("user_id = ?", userID)
		}
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("executed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&history).Error

	return history, total, err
}

// connectToDataSource establishes a connection to a data source
func (s *QueryService) connectToDataSource(dataSource *models.DataSource) (*gorm.DB, error) {
	// Decrypt password
	password, err := s.decryptPassword(dataSource.GetPassword())
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}

	switch dataSource.Type {
	case models.DataSourceTypePostgreSQL:
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			dataSource.Host,
			dataSource.Port,
			dataSource.Username,
			password,
			dataSource.GetDatabase(),
		)
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})

	case models.DataSourceTypeMySQL:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dataSource.Username,
			password,
			dataSource.Host,
			dataSource.Port,
			dataSource.GetDatabase(),
		)
		return gorm.Open(mysql.Open(dsn), &gorm.Config{})

	default:
		return nil, fmt.Errorf("unsupported data source type: %s", dataSource.Type)
	}
}

// decryptPassword decrypts an encrypted password using AES-256-GCM
func (s *QueryService) decryptPassword(encryptedPassword string) (string, error) {
	// The encrypted password should be in format: base64(nonce + ciphertext)
	data, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", err
	}

	// Create cipher block
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	// Extract nonce (first 12 bytes for GCM)
	nonceSize := 12
	if len(data) < nonceSize {
		return "", fmt.Errorf("encrypted data too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Create GCM mode
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Decrypt
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// ValidateQuerySchema validates that tables and columns referenced in the query exist in the data source
func (s *QueryService) ValidateQuerySchema(ctx context.Context, queryText string, dataSource *models.DataSource) error {
	// Get database connection
	dataSourceDB, err := s.connectToDataSource(dataSource)
	if err != nil {
		return fmt.Errorf("failed to connect to data source for validation: %w", err)
	}

	// Extract table names from the query (simple approach)
	tableNames, err := s.extractTableNames(queryText)
	if err != nil {
		return fmt.Errorf("failed to extract table names: %w", err)
	}

	// Check if each table exists in the database
	for _, tableName := range tableNames {
		if !s.tableExists(dataSourceDB, dataSource, tableName) {
			return fmt.Errorf("table '%s' does not exist in data source", tableName)
		}
	}

	return nil
}

// extractTableNames extracts table names from a SQL query
// This is a simplified implementation - a full SQL parser would be more accurate
func (s *QueryService) extractTableNames(queryText string) ([]string, error) {
	var tables []string
	trimmedSQL := strings.TrimSpace(queryText)
	upperSQL := strings.ToUpper(trimmedSQL)

	// Remove SQL comments to avoid false matches
	trimmedSQL = SanitizeSQL(trimmedSQL)
	upperSQL = strings.ToUpper(trimmedSQL)

	// Common patterns for table references
	// This is a simplified approach - for production, consider using a proper SQL parser

	// Pattern 1: FROM table_name or FROM schema.table_name or FROM "table_name"
	fromRegex := regexp.MustCompile(`\bFROM\s+(?:"([^"]+)"|([\w.]+))(?:\s+AS\s+\w+)?(?:\s|,|;|$)`)
	matches := fromRegex.FindAllStringSubmatch(upperSQL, -1)
	for _, match := range matches {
		if len(match) > 1 {
			// Use quoted group if present, otherwise use unquoted group
			tableName := match[1]
			if tableName == "" {
				tableName = match[2]
			}
			tables = append(tables, strings.ToLower(tableName))
		}
	}

	// Pattern 2: JOIN table_name (also handles quoted)
	joinRegex := regexp.MustCompile(`\b(?:JOIN|INNER\s+JOIN|LEFT\s+JOIN|RIGHT\s+JOIN|FULL\s+OUTER\s+JOIN)\s+(?:"([^"]+)"|([\w.]+))(?:\s+AS\s+\w+)?(?:\s+ON)`)
	matches = joinRegex.FindAllStringSubmatch(upperSQL, -1)
	for _, match := range matches {
		if len(match) > 1 {
			// Use quoted group if present, otherwise use unquoted group
			tableName := match[1]
			if tableName == "" {
				tableName = match[2]
			}
			tables = append(tables, strings.ToLower(tableName))
		}
	}

	// Pattern 3: INSERT INTO table_name
	if strings.HasPrefix(upperSQL, "INSERT") {
		insertRegex := regexp.MustCompile(`INSERT\s+INTO\s+(?:"([^"]+)"|([\w.]+))(?:\s*\(|\s+)`)
		matches := insertRegex.FindStringSubmatch(upperSQL)
		if len(matches) > 1 {
			tableName := matches[1]
			if tableName == "" {
				tableName = matches[2]
			}
			tables = append(tables, strings.ToLower(tableName))
		}
	}

	// Pattern 4: UPDATE table_name
	if strings.HasPrefix(upperSQL, "UPDATE") {
		updateRegex := regexp.MustCompile(`UPDATE\s+(?:"([^"]+)"|([\w.]+))(?:\s+SET)`)
		matches := updateRegex.FindStringSubmatch(upperSQL)
		if len(matches) > 1 {
			tableName := matches[1]
			if tableName == "" {
				tableName = matches[2]
			}
			tables = append(tables, strings.ToLower(tableName))
		}
	}

	// Pattern 5: DELETE FROM table_name
	if strings.HasPrefix(upperSQL, "DELETE") {
		deleteRegex := regexp.MustCompile(`DELETE\s+FROM\s+(?:"([^"]+)"|([\w.]+))(?:\s+WHERE|;|$)`)
		matches := deleteRegex.FindStringSubmatch(upperSQL)
		if len(matches) > 1 {
			tableName := matches[1]
			if tableName == "" {
				tableName = matches[2]
			}
			tables = append(tables, strings.ToLower(tableName))
		}
	}

	// Pattern 6: CREATE TABLE table_name
	if strings.HasPrefix(upperSQL, "CREATE") {
		createRegex := regexp.MustCompile(`CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:"([^"]+)"|([\w.]+))(?:\s*\(|\s+)`)
		matches := createRegex.FindStringSubmatch(upperSQL)
		if len(matches) > 1 {
			tableName := matches[1]
			if tableName == "" {
				tableName = matches[2]
			}
			tables = append(tables, strings.ToLower(tableName))
		}
	}

	// Pattern 7: DROP TABLE table_name
	if strings.HasPrefix(upperSQL, "DROP") {
		dropRegex := regexp.MustCompile(`DROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?(?:"([^"]+)"|([\w.]+))(?:\s|;|,|$)`)
		matches := dropRegex.FindStringSubmatch(upperSQL)
		if len(matches) > 1 {
			tableName := matches[1]
			if tableName == "" {
				tableName = matches[2]
			}
			tables = append(tables, strings.ToLower(tableName))
		}
	}

	// Pattern 8: ALTER TABLE table_name
	if strings.HasPrefix(upperSQL, "ALTER") {
		alterRegex := regexp.MustCompile(`ALTER\s+TABLE\s+(?:"([^"]+)"|([\w.]+))(?:\s+ADD|DROP|ALTER|RENAME)`)
		matches := alterRegex.FindStringSubmatch(upperSQL)
		if len(matches) > 1 {
			tableName := matches[1]
			if tableName == "" {
				tableName = matches[2]
			}
			tables = append(tables, strings.ToLower(tableName))
		}
	}

	// Pattern 9: FROM clauses in subqueries (handles nested queries)
	// This pattern finds FROM clauses anywhere in the query, including inside parentheses
	subqueryRegex := regexp.MustCompile(`\bFROM\s+(?:"([^"]+)"|([\w.]+))(?:\s+AS\s+\w+)?`)
	// Find all matches, including those in nested subqueries
	matches = subqueryRegex.FindAllStringSubmatch(upperSQL, -1)
	for _, match := range matches {
		if len(match) > 1 {
			// Use quoted group if present, otherwise use unquoted group
			tableName := match[1]
			if tableName == "" {
				tableName = match[2]
			}
			// Avoid duplicates by checking if we already have this table
			tableName = strings.ToLower(tableName)
			if !contains(tables, tableName) {
				tables = append(tables, tableName)
			}
		}
	}

	// Remove duplicates
	uniqueTables := make(map[string]bool)
	var result []string
	for _, table := range tables {
		table = strings.ToLower(table)
		if !uniqueTables[table] {
			uniqueTables[table] = true
			result = append(result, table)
		}
	}

	return result, nil
}

// tableExists checks if a table exists in the database
func (s *QueryService) tableExists(db *gorm.DB, dataSource *models.DataSource, tableName string) bool {
	tableName = strings.ToLower(tableName)

	switch dataSource.Type {
	case models.DataSourceTypePostgreSQL:
		// Query PostgreSQL's information_schema
		var count int64
		err := db.Raw(`
			SELECT COUNT(*)
			FROM information_schema.tables
			WHERE table_schema = 'public'
			AND LOWER(table_name) = ?
		`, tableName).Scan(&count).Error
		return err == nil && count > 0

	case models.DataSourceTypeMySQL:
		// Query MySQL's information_schema
		var count int64
		err := db.Raw(`
			SELECT COUNT(*)
			FROM information_schema.tables
			WHERE table_schema = DATABASE()
			AND LOWER(table_name) = ?
		`, tableName).Scan(&count).Error
		return err == nil && count > 0

	default:
		// For unknown database types, assume table exists (don't block)
		return true
	}
}


// ExecuteQueryInTransaction executes a query in a transaction and keeps it open for preview
func (s *QueryService) ExecuteQueryInTransaction(ctx context.Context, approval *models.ApprovalRequest, dataSource *models.DataSource) (*models.QueryResult, error) {
	// Get database connection
	dataSourceDB, err := s.connectToDataSource(dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to data source: %w", err)
	}

	// Begin transaction
	tx := dataSourceDB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Execute the query in transaction
	rows, err := tx.Raw(approval.QueryText).Rows()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	// Parse results
	columns, err := rows.Columns()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		// Create a slice of interface{} to hold each column value
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			row[col] = v
		}
		results = append(results, row)
	}

	// Serialize results to JSON
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to serialize results: %w", err)
	}

	// Serialize column names and types to JSON strings for storage
	columnNamesJSON, err := json.Marshal(columns)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to serialize column names: %w", err)
	}

	// Note: ColumnTypes is set to empty strings as we don't have type information without additional schema query
	columnTypes := make([]string, len(columns))
	columnTypesJSON, err := json.Marshal(columnTypes)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to serialize column types: %w", err)
	}

	// Create query result (with temp ID for preview)
	queryResult := &models.QueryResult{
		ID:          uuid.New(),
		QueryID:     uuid.New(), // Temporary ID
		RowCount:    len(results),
		ColumnNames: string(columnNamesJSON),
		ColumnTypes: string(columnTypesJSON),
		Data:        string(resultsJSON),
		StoredAt:    time.Now(),
	}

	// Store the active transaction
	s.txMutex.Lock()
	s.activeTransactions[dataSource.ID] = &ActiveTransaction{
		DB:             tx,
		DataSourceID:   dataSource.ID,
		StartedAt:      time.Now(),
		LastActivityAt: time.Now(),
	}
	s.txMutex.Unlock()

	return queryResult, nil
}

// CommitTransaction commits an active transaction for a data source
func (s *QueryService) CommitTransaction(ctx context.Context, dataSource *models.DataSource) error {
	s.txMutex.Lock()
	defer s.txMutex.Unlock()

	activeTx, exists := s.activeTransactions[dataSource.ID]
	if !exists {
		return fmt.Errorf("no active transaction found for data source")
	}

	// Commit the transaction
	if err := activeTx.DB.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Remove from active transactions
	delete(s.activeTransactions, dataSource.ID)

	return nil
}

// RollbackTransaction rolls back an active transaction for a data source
func (s *QueryService) RollbackTransaction(ctx context.Context, dataSource *models.DataSource) error {
	s.txMutex.Lock()
	defer s.txMutex.Unlock()

	activeTx, exists := s.activeTransactions[dataSource.ID]
	if !exists {
		return fmt.Errorf("no active transaction found for data source")
	}

	// Rollback the transaction
	if err := activeTx.DB.Rollback().Error; err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	// Remove from active transactions
	delete(s.activeTransactions, dataSource.ID)

	return nil
}

// CleanupOldTransactions cleans up transactions that have been idle for too long
func (s *QueryService) CleanupOldTransactions(timeout time.Duration) {
	s.txMutex.Lock()
	defer s.txMutex.Unlock()

	now := time.Now()
	for dataSourceID, activeTx := range s.activeTransactions {
		if now.Sub(activeTx.LastActivityAt) > timeout {
			// Auto-rollback old transactions
			activeTx.DB.Rollback()
			delete(s.activeTransactions, dataSourceID)
		}
	}
}

// contains checks if a string exists in a slice (case-insensitive)
func contains(slice []string, item string) bool {
	item = strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == item {
			return true
		}
	}
	return false
}
