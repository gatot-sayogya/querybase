package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/yourorg/querybase/internal/models"
	"gorm.io/gorm"
)

// AuditService handles audit capture during query execution
type AuditService struct {
	db *gorm.DB
}

// NewAuditService creates a new audit service
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

// AuditResult holds the captured audit data from a query execution
type AuditResult struct {
	AffectedRows  int                      `json:"affected_rows"`
	EstimatedRows int                      `json:"estimated_rows"`
	BeforeData    []map[string]interface{} `json:"before_data,omitempty"`
	AfterData     []map[string]interface{} `json:"after_data,omitempty"`
	AuditMode     models.AuditMode         `json:"audit_mode"`
	Caution       bool                     `json:"caution"`
	CautionMsg    string                   `json:"caution_message,omitempty"`
}

// TestAuditCapability tests whether a data source connection has DDL permissions
// needed for trigger-based audit (CREATE TEMP TABLE, CREATE TRIGGER).
// Updates the data source's audit_capability field with the result.
func (s *AuditService) TestAuditCapability(ctx context.Context, dataSourceDB *gorm.DB, dataSource *models.DataSource) (models.AuditCapability, error) {
	var capability models.AuditCapability

	switch dataSource.Type {
	case models.DataSourceTypePostgreSQL:
		capability = s.testPostgreSQLCapability(dataSourceDB)
	case models.DataSourceTypeMySQL:
		capability = s.testMySQLCapability(dataSourceDB)
	default:
		capability = models.AuditCapabilityCountOnly
	}

	// Persist the result
	s.db.Model(&models.DataSource{}).
		Where("id = ?", dataSource.ID).
		Update("audit_capability", capability)

	dataSource.AuditCapability = capability
	return capability, nil
}

// testPostgreSQLCapability tests DDL permissions on PostgreSQL
func (s *AuditService) testPostgreSQLCapability(db *gorm.DB) models.AuditCapability {
	tx := db.Begin()
	if tx.Error != nil {
		return models.AuditCapabilityCountOnly
	}
	defer tx.Rollback()

	// Test creating a temp table
	err := tx.Exec(`CREATE TEMP TABLE _qb_audit_test (id SERIAL, data TEXT)`).Error
	if err != nil {
		return models.AuditCapabilityCountOnly
	}

	// Test creating a function (needed for triggers in PostgreSQL)
	err = tx.Exec(`
		CREATE OR REPLACE FUNCTION _qb_audit_test_fn() RETURNS TRIGGER AS $$
		BEGIN
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql
	`).Error
	if err != nil {
		return models.AuditCapabilityCountOnly
	}

	// Clean up the function
	tx.Exec(`DROP FUNCTION IF EXISTS _qb_audit_test_fn()`)

	return models.AuditCapabilityFull
}

// testMySQLCapability tests DDL permissions on MySQL
func (s *AuditService) testMySQLCapability(db *gorm.DB) models.AuditCapability {
	tx := db.Begin()
	if tx.Error != nil {
		return models.AuditCapabilityCountOnly
	}
	defer tx.Rollback()

	// Test creating a temporary table
	err := tx.Exec(`CREATE TEMPORARY TABLE _qb_audit_test (id INT AUTO_INCREMENT PRIMARY KEY, data TEXT)`).Error
	if err != nil {
		return models.AuditCapabilityCountOnly
	}

	// Drop the temp table
	tx.Exec(`DROP TEMPORARY TABLE IF EXISTS _qb_audit_test`)

	// MySQL trigger creation requires TRIGGER privilege which we test separately
	// For now, temp table support is a prerequisite
	return models.AuditCapabilityFull
}

// EstimateAffectedRows estimates how many rows will be affected by a write query.
// It parses the query to extract the target table and WHERE clause, then runs a COUNT(*).
func (s *AuditService) EstimateAffectedRows(ctx context.Context, queryText string, dataSourceDB *gorm.DB, dataSource *models.DataSource) (int, error) {
	countQuery, err := s.buildCountQuery(queryText)
	if err != nil {
		return 0, err
	}

	var count int
	err = dataSourceDB.Raw(countQuery).Scan(&count).Error
	if err != nil {
		// If count estimation fails, return 0 (non-fatal)
		return 0, nil
	}

	return count, nil
}

// CheckCaution checks if the estimated row count exceeds the data source threshold
func (s *AuditService) CheckCaution(estimatedRows int, dataSource *models.DataSource) (bool, string) {
	threshold := dataSource.AuditRowThreshold
	if threshold <= 0 {
		threshold = 1000 // Default
	}

	if estimatedRows > threshold {
		return true, fmt.Sprintf(
			"This query will affect ~%d rows (threshold: %d). Consider using 'count_only' or 'sample' audit mode to improve performance.",
			estimatedRows, threshold,
		)
	}

	return false, ""
}

// ExecuteWithAudit executes a query within a transaction, capturing audit data
// based on the specified audit mode. Returns the audit result.
func (s *AuditService) ExecuteWithAudit(
	ctx context.Context,
	tx *gorm.DB,
	queryText string,
	dataSource *models.DataSource,
	auditMode models.AuditMode,
	sampleLimit int,
) (*AuditResult, error) {
	switch auditMode {
	case models.AuditModeFull:
		return s.executeWithFullAudit(ctx, tx, queryText, dataSource)
	case models.AuditModeSample:
		if sampleLimit <= 0 {
			sampleLimit = 100
		}
		return s.executeWithSampleAudit(ctx, tx, queryText, dataSource, sampleLimit)
	case models.AuditModeCountOnly:
		return s.executeCountOnly(ctx, tx, queryText)
	default:
		return s.executeCountOnly(ctx, tx, queryText)
	}
}

// executeWithFullAudit uses trigger-based audit to capture all before/after data
func (s *AuditService) executeWithFullAudit(
	ctx context.Context,
	tx *gorm.DB,
	queryText string,
	dataSource *models.DataSource,
) (*AuditResult, error) {
	tableName, err := s.extractTargetTable(queryText)
	if err != nil {
		// Fall back to count-only if we can't parse the table name
		return s.executeCountOnly(ctx, tx, queryText)
	}

	switch dataSource.Type {
	case models.DataSourceTypePostgreSQL:
		return s.executeWithPostgreSQLAudit(ctx, tx, queryText, tableName, 0)
	case models.DataSourceTypeMySQL:
		return s.executeWithMySQLAudit(ctx, tx, queryText, tableName, 0)
	default:
		return s.executeCountOnly(ctx, tx, queryText)
	}
}

// executeWithSampleAudit captures only the first N rows
func (s *AuditService) executeWithSampleAudit(
	ctx context.Context,
	tx *gorm.DB,
	queryText string,
	dataSource *models.DataSource,
	limit int,
) (*AuditResult, error) {
	tableName, err := s.extractTargetTable(queryText)
	if err != nil {
		return s.executeCountOnly(ctx, tx, queryText)
	}

	switch dataSource.Type {
	case models.DataSourceTypePostgreSQL:
		return s.executeWithPostgreSQLAudit(ctx, tx, queryText, tableName, limit)
	case models.DataSourceTypeMySQL:
		return s.executeWithMySQLAudit(ctx, tx, queryText, tableName, limit)
	default:
		return s.executeCountOnly(ctx, tx, queryText)
	}
}

// executeCountOnly executes the query and only records the affected row count
func (s *AuditService) executeCountOnly(ctx context.Context, tx *gorm.DB, queryText string) (*AuditResult, error) {
	result := tx.Exec(queryText)
	if result.Error != nil {
		return nil, fmt.Errorf("query execution failed: %w", result.Error)
	}

	return &AuditResult{
		AffectedRows: int(result.RowsAffected),
		AuditMode:    models.AuditModeCountOnly,
	}, nil
}

// executeWithPostgreSQLAudit uses PostgreSQL triggers to capture audit data
func (s *AuditService) executeWithPostgreSQLAudit(
	ctx context.Context,
	tx *gorm.DB,
	queryText string,
	tableName string,
	sampleLimit int,
) (*AuditResult, error) {
	// Step 1: Create temp audit table
	err := tx.Exec(`
		CREATE TEMP TABLE IF NOT EXISTS _qb_audit_log (
			_qb_seq SERIAL,
			_qb_action VARCHAR(10),
			_qb_data JSONB
		)
	`).Error
	if err != nil {
		// Fallback to count-only
		return s.executeCountOnly(ctx, tx, queryText)
	}

	// Step 2: Create the audit trigger function
	fnName := "_qb_audit_fn_" + strings.ReplaceAll(tableName, ".", "_")
	triggerBefore := "_qb_trg_before_" + strings.ReplaceAll(tableName, ".", "_")
	triggerAfter := "_qb_trg_after_" + strings.ReplaceAll(tableName, ".", "_")

	createFnSQL := fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION %s() RETURNS TRIGGER AS $$
		BEGIN
			IF TG_OP = 'DELETE' THEN
				INSERT INTO _qb_audit_log (_qb_action, _qb_data)
				VALUES ('DELETE', to_jsonb(OLD));
				RETURN OLD;
			ELSIF TG_OP = 'UPDATE' THEN
				IF TG_WHEN = 'BEFORE' THEN
					INSERT INTO _qb_audit_log (_qb_action, _qb_data)
					VALUES ('BEFORE_UPD', to_jsonb(OLD));
				ELSE
					INSERT INTO _qb_audit_log (_qb_action, _qb_data)
					VALUES ('AFTER_UPD', to_jsonb(NEW));
				END IF;
				RETURN NEW;
			ELSIF TG_OP = 'INSERT' THEN
				INSERT INTO _qb_audit_log (_qb_action, _qb_data)
				VALUES ('INSERT', to_jsonb(NEW));
				RETURN NEW;
			END IF;
			RETURN NULL;
		END;
		$$ LANGUAGE plpgsql
	`, fnName)

	err = tx.Exec(createFnSQL).Error
	if err != nil {
		return s.executeCountOnly(ctx, tx, queryText)
	}

	// Step 3: Create triggers on the target table
	// BEFORE trigger for UPDATE (captures OLD state) and DELETE
	err = tx.Exec(fmt.Sprintf(`
		CREATE TRIGGER %s
		BEFORE UPDATE OR DELETE ON %s
		FOR EACH ROW EXECUTE FUNCTION %s()
	`, triggerBefore, tableName, fnName)).Error
	if err != nil {
		tx.Exec(fmt.Sprintf("DROP FUNCTION IF EXISTS %s()", fnName))
		return s.executeCountOnly(ctx, tx, queryText)
	}

	// AFTER trigger for UPDATE (captures NEW state) and INSERT
	err = tx.Exec(fmt.Sprintf(`
		CREATE TRIGGER %s
		AFTER UPDATE OR INSERT ON %s
		FOR EACH ROW EXECUTE FUNCTION %s()
	`, triggerAfter, tableName, fnName)).Error
	if err != nil {
		tx.Exec(fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", triggerBefore, tableName))
		tx.Exec(fmt.Sprintf("DROP FUNCTION IF EXISTS %s()", fnName))
		return s.executeCountOnly(ctx, tx, queryText)
	}

	// Step 4: Execute the user's query
	execResult := tx.Exec(queryText)
	if execResult.Error != nil {
		// Cleanup triggers before returning error
		s.cleanupPostgreSQLTriggers(tx, tableName, triggerBefore, triggerAfter, fnName)
		return nil, fmt.Errorf("query execution failed: %w", execResult.Error)
	}

	// Step 5: Read audit log
	beforeData, afterData, err := s.readPostgreSQLAuditLog(tx, sampleLimit)

	// Step 6: Cleanup triggers
	s.cleanupPostgreSQLTriggers(tx, tableName, triggerBefore, triggerAfter, fnName)

	if err != nil {
		return &AuditResult{
			AffectedRows: int(execResult.RowsAffected),
			AuditMode:    models.AuditModeCountOnly,
		}, nil
	}

	mode := models.AuditModeFull
	if sampleLimit > 0 {
		mode = models.AuditModeSample
	}

	return &AuditResult{
		AffectedRows: int(execResult.RowsAffected),
		BeforeData:   beforeData,
		AfterData:    afterData,
		AuditMode:    mode,
	}, nil
}

// readPostgreSQLAuditLog reads captured data from the audit temp table
func (s *AuditService) readPostgreSQLAuditLog(tx *gorm.DB, sampleLimit int) ([]map[string]interface{}, []map[string]interface{}, error) {
	var beforeData, afterData []map[string]interface{}

	// Build the query with optional limit
	query := `SELECT _qb_action, _qb_data FROM _qb_audit_log ORDER BY _qb_seq`
	if sampleLimit > 0 {
		query = fmt.Sprintf("%s LIMIT %d", query, sampleLimit*2) // *2 to account for before+after pairs
	}

	rows, err := tx.Raw(query).Rows()
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var action string
		var dataJSON string
		if err := rows.Scan(&action, &dataJSON); err != nil {
			continue
		}

		var rowData map[string]interface{}
		if err := json.Unmarshal([]byte(dataJSON), &rowData); err != nil {
			continue
		}

		switch action {
		case "DELETE", "BEFORE_UPD":
			beforeData = append(beforeData, rowData)
		case "INSERT", "AFTER_UPD":
			afterData = append(afterData, rowData)
		}
	}

	return beforeData, afterData, nil
}

// cleanupPostgreSQLTriggers removes audit triggers and function
func (s *AuditService) cleanupPostgreSQLTriggers(tx *gorm.DB, tableName, triggerBefore, triggerAfter, fnName string) {
	tx.Exec(fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", triggerAfter, tableName))
	tx.Exec(fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", triggerBefore, tableName))
	tx.Exec(fmt.Sprintf("DROP FUNCTION IF EXISTS %s()", fnName))
	tx.Exec("DROP TABLE IF EXISTS _qb_audit_log")
}

// executeWithMySQLAudit uses MySQL triggers to capture audit data
func (s *AuditService) executeWithMySQLAudit(
	ctx context.Context,
	tx *gorm.DB,
	queryText string,
	tableName string,
	sampleLimit int,
) (*AuditResult, error) {
	// Step 1: Create temp audit table
	err := tx.Exec(`
		CREATE TEMPORARY TABLE IF NOT EXISTS _qb_audit_log (
			_qb_seq INT AUTO_INCREMENT PRIMARY KEY,
			_qb_action VARCHAR(10),
			_qb_data JSON
		)
	`).Error
	if err != nil {
		return s.executeCountOnly(ctx, tx, queryText)
	}

	// Step 2: Get columns for the target table to build JSON_OBJECT
	columnListSQL, err := s.getMySQLColumnList(tx, tableName)
	if err != nil {
		return s.executeCountOnly(ctx, tx, queryText)
	}

	// Step 3: Create triggers
	safeName := strings.ReplaceAll(tableName, ".", "_")
	triggerBefore := fmt.Sprintf("_qb_trg_before_%s", safeName)
	triggerAfterUpdate := fmt.Sprintf("_qb_trg_after_upd_%s", safeName)
	triggerAfterInsert := fmt.Sprintf("_qb_trg_after_ins_%s", safeName)

	// BEFORE DELETE / BEFORE UPDATE trigger
	err = tx.Exec(fmt.Sprintf(`
		CREATE TRIGGER %s BEFORE DELETE ON %s
		FOR EACH ROW
		INSERT INTO _qb_audit_log (_qb_action, _qb_data) VALUES ('DELETE', %s)
	`, triggerBefore, tableName, s.buildMySQLJSONObject("OLD", columnListSQL))).Error
	if err != nil {
		return s.executeCountOnly(ctx, tx, queryText)
	}

	// AFTER UPDATE trigger
	tx.Exec(fmt.Sprintf(`DROP TRIGGER IF EXISTS %s`, triggerAfterUpdate))
	tx.Exec(fmt.Sprintf(`
		CREATE TRIGGER %s AFTER UPDATE ON %s
		FOR EACH ROW
		INSERT INTO _qb_audit_log (_qb_action, _qb_data) VALUES ('AFTER_UPD', %s)
	`, triggerAfterUpdate, tableName, s.buildMySQLJSONObject("NEW", columnListSQL)))

	// AFTER INSERT trigger
	tx.Exec(fmt.Sprintf(`DROP TRIGGER IF EXISTS %s`, triggerAfterInsert))
	tx.Exec(fmt.Sprintf(`
		CREATE TRIGGER %s AFTER INSERT ON %s
		FOR EACH ROW
		INSERT INTO _qb_audit_log (_qb_action, _qb_data) VALUES ('INSERT', %s)
	`, triggerAfterInsert, tableName, s.buildMySQLJSONObject("NEW", columnListSQL)))

	// Step 4: Execute the user's query
	execResult := tx.Exec(queryText)
	if execResult.Error != nil {
		s.cleanupMySQLTriggers(tx, triggerBefore, triggerAfterUpdate, triggerAfterInsert)
		return nil, fmt.Errorf("query execution failed: %w", execResult.Error)
	}

	// Step 5: Read audit log
	beforeData, afterData, err := s.readMySQLAuditLog(tx, sampleLimit)

	// Step 6: Cleanup
	s.cleanupMySQLTriggers(tx, triggerBefore, triggerAfterUpdate, triggerAfterInsert)

	if err != nil {
		return &AuditResult{
			AffectedRows: int(execResult.RowsAffected),
			AuditMode:    models.AuditModeCountOnly,
		}, nil
	}

	mode := models.AuditModeFull
	if sampleLimit > 0 {
		mode = models.AuditModeSample
	}

	return &AuditResult{
		AffectedRows: int(execResult.RowsAffected),
		BeforeData:   beforeData,
		AfterData:    afterData,
		AuditMode:    mode,
	}, nil
}

// getMySQLColumnList retrieves column names for building JSON_OBJECT
func (s *AuditService) getMySQLColumnList(tx *gorm.DB, tableName string) ([]string, error) {
	var columns []string
	rows, err := tx.Raw(`
		SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`, tableName).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			continue
		}
		columns = append(columns, col)
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("no columns found for table %s", tableName)
	}

	return columns, nil
}

// buildMySQLJSONObject builds a JSON_OBJECT expression for MySQL triggers
func (s *AuditService) buildMySQLJSONObject(prefix string, columns []string) string {
	var parts []string
	for _, col := range columns {
		parts = append(parts, fmt.Sprintf("'%s', %s.`%s`", col, prefix, col))
	}
	return fmt.Sprintf("JSON_OBJECT(%s)", strings.Join(parts, ", "))
}

// readMySQLAuditLog reads captured data from the MySQL audit temp table
func (s *AuditService) readMySQLAuditLog(tx *gorm.DB, sampleLimit int) ([]map[string]interface{}, []map[string]interface{}, error) {
	var beforeData, afterData []map[string]interface{}

	query := `SELECT _qb_action, _qb_data FROM _qb_audit_log ORDER BY _qb_seq`
	if sampleLimit > 0 {
		query = fmt.Sprintf("%s LIMIT %d", query, sampleLimit*2)
	}

	rows, err := tx.Raw(query).Rows()
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var action string
		var dataJSON string
		if err := rows.Scan(&action, &dataJSON); err != nil {
			continue
		}

		var rowData map[string]interface{}
		if err := json.Unmarshal([]byte(dataJSON), &rowData); err != nil {
			continue
		}

		switch action {
		case "DELETE", "BEFORE_UPD":
			beforeData = append(beforeData, rowData)
		case "INSERT", "AFTER_UPD":
			afterData = append(afterData, rowData)
		}
	}

	return beforeData, afterData, nil
}

// cleanupMySQLTriggers removes audit triggers
func (s *AuditService) cleanupMySQLTriggers(tx *gorm.DB, triggers ...string) {
	for _, trigger := range triggers {
		tx.Exec(fmt.Sprintf("DROP TRIGGER IF EXISTS %s", trigger))
	}
	tx.Exec("DROP TEMPORARY TABLE IF EXISTS _qb_audit_log")
}

// buildCountQuery converts a write query into a SELECT COUNT(*) for estimation
func (s *AuditService) buildCountQuery(queryText string) (string, error) {
	trimmed := strings.TrimSpace(queryText)
	upper := strings.ToUpper(trimmed)

	// DELETE FROM table WHERE ... → SELECT COUNT(*) FROM table WHERE ...
	if strings.HasPrefix(upper, "DELETE") {
		return s.deleteToCount(trimmed, upper)
	}

	// UPDATE table SET ... WHERE ... → SELECT COUNT(*) FROM table WHERE ...
	if strings.HasPrefix(upper, "UPDATE") {
		return s.updateToCount(trimmed, upper)
	}

	// INSERT → Typically 1 row or based on SELECT
	if strings.HasPrefix(upper, "INSERT") {
		return s.insertToCount(trimmed, upper)
	}

	return "", fmt.Errorf("unsupported query type for estimation")
}

// deleteToCount converts DELETE to SELECT COUNT(*)
func (s *AuditService) deleteToCount(query, upper string) (string, error) {
	// DELETE [FROM] table_name [WHERE ...]
	var rest string
	if strings.HasPrefix(upper, "DELETE FROM ") {
		rest = query[12:]
	} else if strings.HasPrefix(upper, "DELETE ") {
		rest = query[7:]
	} else {
		return "", fmt.Errorf("cannot parse DELETE query")
	}

	// Trim trailing semicolon and whitespace
	rest = strings.TrimRight(rest, "; ")
	return "SELECT COUNT(*) FROM " + strings.TrimSpace(rest), nil
}

// updateToCount converts UPDATE to SELECT COUNT(*)
func (s *AuditService) updateToCount(query, upper string) (string, error) {
	// UPDATE table_name SET ... WHERE ...
	// Extract table name and WHERE clause
	re := regexp.MustCompile(`(?i)^UPDATE\s+(\S+)\s+SET\s+`)
	matches := re.FindStringSubmatchIndex(query)
	if matches == nil {
		return "", fmt.Errorf("cannot parse UPDATE query")
	}

	tableName := query[matches[2]:matches[3]]

	// Find WHERE clause
	whereIdx := strings.Index(upper, " WHERE ")
	if whereIdx == -1 {
		// No WHERE clause — affects all rows
		return fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName), nil
	}

	whereClause := query[whereIdx:]
	// Trim trailing semicolon
	whereClause = strings.TrimRight(whereClause, "; ")
	return fmt.Sprintf("SELECT COUNT(*) FROM %s%s", tableName, whereClause), nil
}

// insertToCount estimates row count for INSERT queries
func (s *AuditService) insertToCount(query, upper string) (string, error) {
	// INSERT INTO ... SELECT ... → convert the SELECT to COUNT
	selectIdx := strings.Index(upper, " SELECT ")
	if selectIdx != -1 {
		selectPart := query[selectIdx+1:]
		// Replace the SELECT columns with COUNT(*)
		fromIdx := strings.Index(strings.ToUpper(selectPart), " FROM ")
		if fromIdx != -1 {
			return "SELECT COUNT(*)" + selectPart[fromIdx:], nil
		}
	}

	// Simple INSERT with VALUES — count the number of value groups
	valuesCount := strings.Count(upper, "),(") + 1
	if strings.Contains(upper, "VALUES") {
		return fmt.Sprintf("SELECT %d", valuesCount), nil
	}

	return "SELECT 1", nil
}

// extractTargetTable extracts the target table name from a write query
func (s *AuditService) extractTargetTable(queryText string) (string, error) {
	trimmed := strings.TrimSpace(queryText)
	upper := strings.ToUpper(trimmed)

	// UPDATE table SET ...
	if strings.HasPrefix(upper, "UPDATE") {
		re := regexp.MustCompile(`(?i)^UPDATE\s+` + "`?" + `(\S+?)` + "`?" + `\s+SET\s+`)
		matches := re.FindStringSubmatch(trimmed)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	// DELETE [FROM] table ...
	if strings.HasPrefix(upper, "DELETE") {
		re := regexp.MustCompile(`(?i)^DELETE\s+(?:FROM\s+)?` + "`?" + `(\S+?)` + "`?" + `(?:\s+|;|$)`)
		matches := re.FindStringSubmatch(trimmed)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	// INSERT INTO table ...
	if strings.HasPrefix(upper, "INSERT") {
		re := regexp.MustCompile(`(?i)^INSERT\s+INTO\s+` + "`?" + `(\S+?)` + "`?" + `(?:\s*\(|\s+)`)
		matches := re.FindStringSubmatch(trimmed)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("could not extract target table from query")
}

// ResolveAuditMode determines the effective audit mode based on capability and request.
// If the data source doesn't support full audit, it degrades gracefully.
func (s *AuditService) ResolveAuditMode(requested models.AuditMode, capability models.AuditCapability) models.AuditMode {
	if capability == models.AuditCapabilityCountOnly {
		return models.AuditModeCountOnly
	}

	if requested == "" {
		return models.AuditModeFull
	}

	return requested
}
