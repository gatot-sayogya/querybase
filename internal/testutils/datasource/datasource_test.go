package datasource

import (
	"testing"

	"gorm.io/gorm"

	"github.com/yourorg/querybase/internal/models"
	"github.com/yourorg/querybase/internal/testutils/database"
	"github.com/yourorg/querybase/internal/testutils/fixtures"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return database.SetupTestDB(t)
}

func TestCreateTestPostgreSQLDataSource(t *testing.T) {
	db := setupTestDB(t)

	ds := CreateTestPostgreSQLDataSource(t, db, "test-pg-ds")

	if ds.Name != "test-pg-ds" {
		t.Errorf("Expected name 'test-pg-ds', got %s", ds.Name)
	}
	if ds.Type != models.DataSourceTypePostgreSQL {
		t.Errorf("Expected type PostgreSQL, got %s", ds.Type)
	}
	if ds.Port != 5432 {
		t.Errorf("Expected port 5432, got %d", ds.Port)
	}
	if ds.EncryptedPassword == "" {
		t.Error("Expected encrypted password to be set")
	}
	if !ds.IsActive {
		t.Error("Expected data source to be active")
	}
}

func TestCreateTestMySQLDataSource(t *testing.T) {
	db := setupTestDB(t)

	ds := CreateTestMySQLDataSource(t, db, "test-mysql-ds")

	if ds.Name != "test-mysql-ds" {
		t.Errorf("Expected name 'test-mysql-ds', got %s", ds.Name)
	}
	if ds.Type != models.DataSourceTypeMySQL {
		t.Errorf("Expected type MySQL, got %s", ds.Type)
	}
	if ds.Port != 3306 {
		t.Errorf("Expected port 3306, got %d", ds.Port)
	}
	if ds.EncryptedPassword == "" {
		t.Error("Expected encrypted password to be set")
	}
}

func TestCreateTestDataSourceWithUniqueName(t *testing.T) {
	db := setupTestDB(t)

	ds1 := CreateTestDataSourceWithUniqueName(t, db)
	ds2 := CreateTestDataSourceWithUniqueName(t, db)

	if ds1.Name == ds2.Name {
		t.Error("Expected unique names for each data source")
	}
	if ds1.ID == ds2.ID {
		t.Error("Expected unique IDs for each data source")
	}
}

func TestCreateTestDataSourceWithPerms(t *testing.T) {
	db := setupTestDB(t)

	group := fixtures.CreateTestGroup(t, db, "test-group")

	ds := CreateTestDataSourceWithPerms(t, db, "test-ds-with-perms", group.ID, true, true, false)

	if ds.Name != "test-ds-with-perms" {
		t.Errorf("Expected name 'test-ds-with-perms', got %s", ds.Name)
	}

	var permission models.DataSourcePermission
	if err := db.Where("data_source_id = ? AND group_id = ?", ds.ID, group.ID).First(&permission).Error; err != nil {
		t.Fatalf("Expected permission to be created: %v", err)
	}
	if !permission.CanRead {
		t.Error("Expected CanRead to be true")
	}
	if !permission.CanWrite {
		t.Error("Expected CanWrite to be true")
	}
	if permission.CanApprove {
		t.Error("Expected CanApprove to be false")
	}
}

func TestCreateInactiveTestDataSource(t *testing.T) {
	db := setupTestDB(t)

	ds := CreateInactiveTestDataSource(t, db, "inactive-ds")

	var fresh models.DataSource
	db.First(&fresh, "id = ?", ds.ID)

	if fresh.IsActive {
		t.Errorf("Expected IsActive=false, got IsActive=%v", fresh.IsActive)
	}
}

func TestCreateTestDataSourceWithType(t *testing.T) {
	db := setupTestDB(t)

	tests := []struct {
		name     string
		dsType   models.DataSourceType
		expected int
	}{
		{"PostgreSQL", models.DataSourceTypePostgreSQL, 5432},
		{"MySQL", models.DataSourceTypeMySQL, 3306},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := CreateTestDataSourceWithType(t, db, tt.name+"-ds", tt.dsType)
			if ds.Type != tt.dsType {
				t.Errorf("Expected type %s, got %s", tt.dsType, ds.Type)
			}
			if ds.Port != tt.expected {
				t.Errorf("Expected port %d, got %d", tt.expected, ds.Port)
			}
		})
	}
}

func TestEncryptDecryptPassword(t *testing.T) {
	testPassword := "my-secret-password-123"
	encryptionKey := "test-encryption-key-32-bytes!!!!"

	encrypted, err := EncryptPassword(testPassword, encryptionKey)
	if err != nil {
		t.Fatalf("Failed to encrypt password: %v", err)
	}

	if encrypted == testPassword {
		t.Error("Encrypted password should not equal plaintext")
	}

	decrypted, err := DecryptPassword(encrypted, encryptionKey)
	if err != nil {
		t.Fatalf("Failed to decrypt password: %v", err)
	}

	if decrypted != testPassword {
		t.Errorf("Decrypted password %s does not match original %s", decrypted, testPassword)
	}
}

func TestEncryptPasswordDifferentOutputs(t *testing.T) {
	password := "test-password"
	key := "test-encryption-key-32-bytes!!!!"

	encrypted1, err := EncryptPassword(password, key)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	encrypted2, err := EncryptPassword(password, key)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	if encrypted1 == encrypted2 {
		t.Error("Same password should produce different encrypted outputs (due to random nonce)")
	}
}

func TestDecryptPasswordWrongKey(t *testing.T) {
	password := "test-password"
	key1 := "test-encryption-key-32-bytes!!!!"
	key2 := "another-encryption-key-32-bytes!!"

	encrypted, err := EncryptPassword(password, key1)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	_, err = DecryptPassword(encrypted, key2)
	if err == nil {
		t.Error("Expected error when decrypting with wrong key")
	}
}

func TestEncryptTestPassword(t *testing.T) {
	password := "test-password"

	encrypted := EncryptTestPassword(password)

	if encrypted == "" {
		t.Error("Expected non-empty encrypted password")
	}

	decrypted, err := DecryptPassword(encrypted, defaultTestEncryptionKey)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if decrypted != password {
		t.Errorf("Expected %s, got %s", password, decrypted)
	}
}

func TestGetTestEncryptionKey(t *testing.T) {
	key := GetTestEncryptionKey()
	if key != defaultTestEncryptionKey {
		t.Errorf("Expected %s, got %s", defaultTestEncryptionKey, key)
	}
}

func TestGetTestPostgreSQLConfig(t *testing.T) {
	config := GetTestPostgreSQLConfig()

	if config.Type != models.DataSourceTypePostgreSQL {
		t.Errorf("Expected type PostgreSQL, got %s", config.Type)
	}
	if config.Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", config.Host)
	}
	if config.Port != 5432 {
		t.Errorf("Expected port 5432, got %d", config.Port)
	}
}

func TestGetTestMySQLConfig(t *testing.T) {
	config := GetTestMySQLConfig()

	if config.Type != models.DataSourceTypeMySQL {
		t.Errorf("Expected type MySQL, got %s", config.Type)
	}
	if config.Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", config.Host)
	}
	if config.Port != 3306 {
		t.Errorf("Expected port 3306, got %d", config.Port)
	}
}

func TestGetTestPostgreSQLDSN(t *testing.T) {
	ds := &models.DataSource{
		Host:         "testhost",
		Port:         5432,
		Username:     "testuser",
		DatabaseName: "testdb",
	}

	dsn := GetTestPostgreSQLDSN(ds, "testpass")

	expected := "host=testhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	if dsn != expected {
		t.Errorf("Expected DSN %s, got %s", expected, dsn)
	}
}

func TestGetTestMySQLDSN(t *testing.T) {
	ds := &models.DataSource{
		Host:         "testhost",
		Port:         3306,
		Username:     "testuser",
		DatabaseName: "testdb",
	}

	dsn := GetTestMySQLDSN(ds, "testpass")

	expected := "testuser:testpass@tcp(testhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	if dsn != expected {
		t.Errorf("Expected DSN %s, got %s", expected, dsn)
	}
}

func TestDataSourceCleanup(t *testing.T) {
	db := setupTestDB(t)

	ds := CreateTestPostgreSQLDataSource(t, db, "cleanup-test-ds")
	dsID := ds.ID

	db.Unscoped().Delete(ds)

	var count int64
	db.Unscoped().Model(&models.DataSource{}).Where("id = ?", dsID).Count(&count)
	if count != 0 {
		t.Error("Expected data source to be deleted after cleanup")
	}
}

func TestDataSourcePermissionCleanup(t *testing.T) {
	db := setupTestDB(t)

	group := fixtures.CreateTestGroup(t, db, "perm-cleanup-group")
	ds := CreateTestDataSourceWithPerms(t, db, "perm-cleanup-ds", group.ID, true, false, false)

	var permCount int64
	db.Model(&models.DataSourcePermission{}).Where("data_source_id = ?", ds.ID).Count(&permCount)
	if permCount == 0 {
		t.Error("Expected permission to be created")
	}
}

func TestSetupDataSourceWithTestTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := database.SetupTestDBWithPostgres(t, &database.GetTestConfig().Database)
	if db == nil {
		t.Skip("PostgreSQL not available")
	}

	ds := CreateTestPostgreSQLDataSource(t, db, "table-test-ds")

	tableName := "test_custom_table"
	columns := "id SERIAL PRIMARY KEY, name VARCHAR(100)"

	SetupDataSourceWithTestTable(t, db, ds, tableName, columns)

	var count int64
	db.Raw(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = ?
	`, tableName).Scan(&count)

	if count == 0 {
		t.Error("Expected table to be created")
	}
}

func TestDropTestTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := database.SetupTestDBWithPostgres(t, &database.GetTestConfig().Database)
	if db == nil {
		t.Skip("PostgreSQL not available")
	}

	ds := CreateTestPostgreSQLDataSource(t, db, "drop-test-ds")

	tableName := "test_drop_table"
	db.Exec("CREATE TABLE " + tableName + " (id INT)")

	DropTestTable(t, db, ds, tableName)

	var count int64
	db.Raw(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = ?
	`, tableName).Scan(&count)

	if count > 0 {
		t.Error("Expected table to be dropped")
	}
}

func TestInsertTestUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := database.SetupTestDBWithPostgres(t, &database.GetTestConfig().Database)
	if db == nil {
		t.Skip("PostgreSQL not available")
	}

	ds := CreateTestPostgreSQLDataSource(t, db, "insert-test-ds")
	conn := ConnectToDataSource(t, ds, "testpass")
	SetupUserTestTable(t, conn, ds)

	users := []map[string]any{
		{"name": "Alice", "email": "alice@example.com", "age": 30},
		{"name": "Bob", "email": "bob@example.com", "age": 25},
	}

	InsertTestUsers(t, conn, users)

	count := GetRowCount(t, conn, "test_users")
	if count != 2 {
		t.Errorf("Expected 2 users, got %d", count)
	}
}

func TestTableExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := database.SetupTestDBWithPostgres(t, &database.GetTestConfig().Database)
	if db == nil {
		t.Skip("PostgreSQL not available")
	}

	ds := CreateTestPostgreSQLDataSource(t, db, "exists-test-ds")
	conn := ConnectToDataSource(t, ds, "testpass")

	dbName := "test_exists_table"
	SetupDataSourceWithTestTable(t, conn, ds, dbName, "id INT")

	if !TableExists(t, conn, ds, dbName) {
		t.Error("Expected table to exist")
	}

	if TableExists(t, conn, ds, "nonexistent_table_xyz") {
		t.Error("Expected non-existent table to return false")
	}
}

func TestTruncateTable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := database.SetupTestDBWithPostgres(t, &database.GetTestConfig().Database)
	if db == nil {
		t.Skip("PostgreSQL not available")
	}

	ds := CreateTestPostgreSQLDataSource(t, db, "truncate-test-ds")
	conn := ConnectToDataSource(t, ds, "testpass")
	SetupUserTestTable(t, conn, ds)

	users := []map[string]any{
		{"name": "Alice", "email": "alice@example.com"},
		{"name": "Bob", "email": "bob@example.com"},
	}
	InsertTestUsers(t, conn, users)

	TruncateTable(t, conn, "test_users")

	count := GetRowCount(t, conn, "test_users")
	if count != 0 {
		t.Errorf("Expected 0 rows after truncate, got %d", count)
	}
}

func TestCreateTableWithForeignKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-dependent test in short mode")
	}

	db := database.SetupTestDBWithPostgres(t, &database.GetTestConfig().Database)
	if db == nil {
		t.Skip("PostgreSQL not available")
	}
	ds := CreateTestPostgreSQLDataSource(t, db, "fk-test-ds")
	conn := ConnectToDataSource(t, ds, "testpass")
	SetupUserTestTable(t, conn, ds)

	CreateTableWithForeignKey(t, conn, ds, "test_fk_table",
		"id SERIAL PRIMARY KEY, user_id INT",
		"FOREIGN KEY (user_id) REFERENCES test_users(id)",
	)

	if !TableExists(t, conn, ds, "test_fk_table") {
		t.Error("Expected foreign key table to exist")
	}
}
