package service

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"

	"crypto/aes"
	"crypto/cipher"

	"github.com/yourorg/querybase/internal/models"
	_ "github.com/lib/pq"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// SchemaService handles database schema inspection
type SchemaService struct {
	db            *gorm.DB
	encryptionKey []byte
}

// NewSchemaService creates a new schema service
func NewSchemaService(db *gorm.DB, encryptionKey string) *SchemaService {
	return &SchemaService{
		db:            db,
		encryptionKey: []byte(encryptionKey),
	}
}

// decryptPassword decrypts an encrypted password
func (s *SchemaService) decryptPassword(encryptedPassword string) (string, error) {
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

// TableInfo represents information about a database table
type TableInfo struct {
	TableName  string        `json:"table_name"`
	Schema     string        `json:"schema"`
	Columns    []ColumnInfo `json:"columns"`
	Indexes    []IndexInfo   `json:"indexes,omitempty"`
}

// ColumnInfo represents information about a column
type ColumnInfo struct {
	ColumnName     string  `json:"column_name"`
	DataType       string  `json:"data_type"`
	IsNullable     bool    `json:"is_nullable"`
	ColumnDefault  *string `json:"column_default,omitempty"`
	IsPrimaryKey   bool    `json:"is_primary_key"`
	IsForeignKey   bool    `json:"is_foreign_key"`
}

// IndexInfo represents information about an index
type IndexInfo struct {
	IndexName  string   `json:"index_name"`
	Columns    []string `json:"columns"`
	IsUnique   bool     `json:"is_unique"`
	IsPrimary  bool     `json:"is_primary"`
}

// DatabaseSchema represents the complete schema of a database
type DatabaseSchema struct {
	DataSourceID string       `json:"data_source_id"`
	DataSourceName string     `json:"data_source_name"`
	DatabaseType string       `json:"database_type"`
	DatabaseName string       `json:"database_name"`
	Tables       []TableInfo  `json:"tables"`
	Schemas      []string     `json:"schemas,omitempty"`
}

// GetSchema fetches the complete schema for a data source
func (s *SchemaService) GetSchema(ctx context.Context, dataSourceID string) (*DatabaseSchema, error) {
	// Fetch data source
	var dataSource models.DataSource
	if err := s.db.First(&dataSource, "id = ?", dataSourceID).Error; err != nil {
		return nil, fmt.Errorf("data source not found: %w", err)
	}

	// Connect to data source
	conn, err := s.connectToDataSource(&dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to data source: %w", err)
	}
	defer conn.Close()

	schema := &DatabaseSchema{
		DataSourceID:   dataSourceID,
		DataSourceName: dataSource.Name,
		DatabaseType:   string(dataSource.Type),
		DatabaseName:   dataSource.DatabaseName,
		Tables:         []TableInfo{},
	}

	// Get schema based on database type
	var schemas []string
	switch dataSource.Type {
	case models.DataSourceTypePostgreSQL:
		schema.Tables, schemas, err = s.getPostgreSQLSchema(conn, dataSource.DatabaseName)
	case models.DataSourceTypeMySQL:
		schema.Tables, err = s.getMySQLSchema(conn)
	default:
		return nil, fmt.Errorf("unsupported data source type: %s", dataSource.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	schema.Schemas = schemas
	return schema, nil
}

// GetTables returns a list of tables for a data source
func (s *SchemaService) GetTables(ctx context.Context, dataSourceID string) ([]TableInfo, error) {
	var dataSource models.DataSource
	if err := s.db.First(&dataSource, "id = ?", dataSourceID).Error; err != nil {
		return nil, err
	}

	conn, err := s.connectToDataSource(&dataSource)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var tables []TableInfo

	switch dataSource.Type {
	case models.DataSourceTypePostgreSQL:
		tables, _, err = s.getPostgreSQLSchema(conn, dataSource.DatabaseName)
	case models.DataSourceTypeMySQL:
		tables, err = s.getMySQLSchema(conn)
	default:
		return nil, fmt.Errorf("unsupported data source type: %s", dataSource.Type)
	}

	return tables, err
}

// GetTableColumns returns detailed column information for a specific table
func (s *SchemaService) GetTableColumns(ctx context.Context, dataSourceID, tableName string) (*TableInfo, error) {
	var dataSource models.DataSource
	if err := s.db.First(&dataSource, "id = ?", dataSourceID).Error; err != nil {
		return nil, err
	}

	conn, err := s.connectToDataSource(&dataSource)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	switch dataSource.Type {
	case models.DataSourceTypePostgreSQL:
		return s.getPostgreSQLTableDetails(conn, tableName)
	case models.DataSourceTypeMySQL:
		return s.getMySQLTableDetails(conn, tableName)
	default:
		return nil, fmt.Errorf("unsupported data source type: %s", dataSource.Type)
	}
}

// getPostgreSQLSchema fetches schema from PostgreSQL database
func (s *SchemaService) getPostgreSQLSchema(db *sql.DB, databaseName string) ([]TableInfo, []string, error) {
	// Get all schemas
	schemas, err := s.getPostgreSQLSchemas(db)
	if err != nil {
		return nil, nil, err
	}

	// Get all tables with columns
	query := `
		SELECT
			t.table_schema,
			t.table_name,
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			COALESCE(pk.constraint_type, '') as is_primary_key
		FROM information_schema.tables t
		LEFT JOIN information_schema.columns c ON t.table_name = c.table_name AND t.table_schema = c.table_schema
		LEFT JOIN (
			SELECT ku.table_name, ku.column_name, tc.constraint_type
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
			WHERE tc.constraint_type = 'PRIMARY KEY'
		) pk ON c.table_name = pk.table_name AND c.column_name = pk.column_name
		WHERE t.table_schema NOT IN ('pg_catalog', 'information_schema')
			AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_schema, t.table_name, c.ordinal_position
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	tableMap := make(map[string]*TableInfo)

	for rows.Next() {
		var schemaName, tableName, columnName, dataType, isNullable, isPrimaryKey string
		var columnDefault sql.NullString
		if err := rows.Scan(&schemaName, &tableName, &columnName, &dataType, &isNullable, &columnDefault, &isPrimaryKey); err != nil {
			return nil, nil, err
		}

		key := schemaName + "." + tableName
		if _, exists := tableMap[key]; !exists {
			tableMap[key] = &TableInfo{
				TableName: tableName,
				Schema:    schemaName,
				Columns:   []ColumnInfo{},
			}
		}

		column := ColumnInfo{
			ColumnName:    columnName,
			DataType:      dataType,
			IsNullable:    isNullable == "YES",
			IsPrimaryKey:  isPrimaryKey == "PRIMARY KEY",
		}

		if columnDefault.Valid {
			column.ColumnDefault = &columnDefault.String
		}

		tableMap[key].Columns = append(tableMap[key].Columns, column)
	}

	// Convert map to slice
	tables := make([]TableInfo, 0, len(tableMap))
	for _, table := range tableMap {
		tables = append(tables, *table)
	}

	return tables, schemas, nil
}

// getPostgreSQLTableDetails fetches detailed information for a specific table
func (s *SchemaService) getPostgreSQLTableDetails(db *sql.DB, tableName string) (*TableInfo, error) {
	query := `
		SELECT
			c.table_name,
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			COALESCE(pk.constraint_type, '') as is_primary_key
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT ku.table_name, ku.column_name, tc.constraint_type
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
			WHERE tc.constraint_type = 'PRIMARY KEY'
		) pk ON c.table_name = pk.table_name AND c.column_name = pk.column_name
		WHERE c.table_name = $1
			AND c.table_schema NOT IN ('pg_catalog', 'information_schema')
		ORDER BY c.ordinal_position
	`

	rows, err := db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tableInfo := &TableInfo{
		TableName: tableName,
		Schema:    "public",
		Columns:   []ColumnInfo{},
	}

	for rows.Next() {
		var name, dataType, isNullable, isPrimaryKey string
		var columnDefault sql.NullString
		if err := rows.Scan(&name, &dataType, &isNullable, &columnDefault, &isPrimaryKey); err != nil {
			return nil, err
		}

		column := ColumnInfo{
			ColumnName:    name,
			DataType:      dataType,
			IsNullable:    isNullable == "YES",
			IsPrimaryKey:  isPrimaryKey == "PRIMARY KEY",
		}

		if columnDefault.Valid {
			column.ColumnDefault = &columnDefault.String
		}

		tableInfo.Columns = append(tableInfo.Columns, column)
	}

	return tableInfo, nil
}

// getPostgreSQLSchemas returns all non-system schemas
func (s *SchemaService) getPostgreSQLSchemas(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		ORDER BY schema_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, err
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

// getMySQLSchema fetches schema from MySQL database
func (s *SchemaService) getMySQLSchema(db *sql.DB) ([]TableInfo, error) {
	// Get the current database name
	var dbName string
	err := db.QueryRow("SELECT DATABASE()").Scan(&dbName)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			TABLE_NAME,
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			COLUMN_KEY = 'PRI' as is_primary_key
		FROM information_schema.columns
		WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME, ORDINAL_POSITION
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tableMap := make(map[string]*TableInfo)

	for rows.Next() {
		var tableName, columnName, dataType, isNullable string
		var columnDefault, isPrimaryKey sql.NullString

		if err := rows.Scan(&tableName, &columnName, &dataType, &isNullable, &columnDefault, &isPrimaryKey); err != nil {
			return nil, err
		}

		if _, exists := tableMap[tableName]; !exists {
			tableMap[tableName] = &TableInfo{
				TableName: tableName,
				Schema:    dbName, // Use the actual database name
				Columns:   []ColumnInfo{},
			}
		}

		column := ColumnInfo{
			ColumnName:   columnName,
			DataType:     dataType,
			IsNullable:   isNullable == "YES",
			IsPrimaryKey: isPrimaryKey.String == "PRI",
		}

		if columnDefault.Valid {
			column.ColumnDefault = &columnDefault.String
		}

		tableMap[tableName].Columns = append(tableMap[tableName].Columns, column)
	}

	tables := make([]TableInfo, 0, len(tableMap))
	for _, table := range tableMap {
		tables = append(tables, *table)
	}

	return tables, nil
}

// getMySQLTableDetails fetches detailed information for a specific table in MySQL
func (s *SchemaService) getMySQLTableDetails(db *sql.DB, tableName string) (*TableInfo, error) {
	query := `
		SELECT
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			COLUMN_KEY = 'PRI' as is_primary_key
		FROM information_schema.columns
		WHERE TABLE_NAME = ? AND TABLE_SCHEMA = DATABASE()
		ORDER BY ORDINAL_POSITION
	`

	rows, err := db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tableInfo := &TableInfo{
		TableName: tableName,
		Schema:    "", // MySQL doesn't use schemas the same way
		Columns:   []ColumnInfo{},
	}

	for rows.Next() {
		var name, dataType, isNullable string
		var columnDefault, isPrimaryKey sql.NullString

		if err := rows.Scan(&name, &dataType, &isNullable, &columnDefault, &isPrimaryKey); err != nil {
			return nil, err
		}

		column := ColumnInfo{
			ColumnName:   name,
			DataType:     dataType,
			IsNullable:   isNullable == "YES",
			IsPrimaryKey: isPrimaryKey.String == "PRI",
		}

		if columnDefault.Valid {
			column.ColumnDefault = &columnDefault.String
		}

		tableInfo.Columns = append(tableInfo.Columns, column)
	}

	return tableInfo, nil
}

// SearchTables searches for tables by name
func (s *SchemaService) SearchTables(ctx context.Context, dataSourceID, searchTerm string) ([]TableInfo, error) {
	var dataSource models.DataSource
	if err := s.db.First(&dataSource, "id = ?", dataSourceID).Error; err != nil {
		return nil, err
	}

	conn, err := s.connectToDataSource(&dataSource)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var tables []TableInfo
	searchPattern := "%" + strings.ToLower(searchTerm) + "%"

	switch dataSource.Type {
	case models.DataSourceTypePostgreSQL:
		query := `
			SELECT DISTINCT
				t.table_name,
				t.table_schema
			FROM information_schema.tables t
			WHERE t.table_schema NOT IN ('pg_catalog', 'information_schema')
				AND t.table_type = 'BASE TABLE'
				AND LOWER(t.table_name) LIKE $1
			ORDER BY t.table_name
			LIMIT 50
		`

		rows, err := conn.Query(query, searchPattern)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var name, schema string
			if err := rows.Scan(&name, &schema); err != nil {
				return nil, err
			}

			tableInfo, err := s.getPostgreSQLTableDetails(conn, name)
			if err != nil {
				continue
			}

			tables = append(tables, *tableInfo)
		}

	case models.DataSourceTypeMySQL:
		query := `
			SELECT DISTINCT TABLE_NAME
			FROM information_schema.tables
			WHERE TABLE_SCHEMA = DATABASE()
				AND TABLE_TYPE = 'BASE TABLE'
				AND LOWER(TABLE_NAME) LIKE ?
			ORDER BY TABLE_NAME
			LIMIT 50
		`

		rows, err := conn.Query(query, searchPattern)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return nil, err
			}

			tableInfo, err := s.getMySQLTableDetails(conn, name)
			if err != nil {
				continue
			}

			tables = append(tables, *tableInfo)
		}
	}

	return tables, nil
}

// connectToDataSource establishes a connection to a data source
func (s *SchemaService) connectToDataSource(dataSource *models.DataSource) (*sql.DB, error) {
	var driverName string
	var dsn string

	switch dataSource.Type {
	case models.DataSourceTypePostgreSQL:
		driverName = "postgres"
		// Decrypt password before using it
		password, err := s.decryptPassword(dataSource.EncryptedPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password: %w", err)
		}
		dsn = fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
			dataSource.Host,
			dataSource.Port,
			dataSource.DatabaseName,
			dataSource.Username,
			password)

	case models.DataSourceTypeMySQL:
		driverName = "mysql"
		// Decrypt password before using it
		password, err := s.decryptPassword(dataSource.EncryptedPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password: %w", err)
		}
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			dataSource.Username,
			password,
			dataSource.Host,
			dataSource.Port,
			dataSource.DatabaseName)

	default:
		return nil, fmt.Errorf("unsupported data source type: %s", dataSource.Type)
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
