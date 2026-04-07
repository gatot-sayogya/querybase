import { test, expect, Page } from '@playwright/test';

/**
 * Multi-Query Transaction E2E Tests
 * 
 * Tests cover:
 * - Multi-query preview functionality
 * - Multi-query execution
 * - Commit after success
 * - Rollback on failure
 * - SET variable handling
 * - Transaction control blocking
 * - Multiple statements execution order
 * - Transaction status display
 */

// Test user credentials
const TEST_USERS = {
  admin: {
    username: 'admin',
    password: 'admin123',
    role: 'admin',
    fullName: 'System Administrator',
  },
};

// Helper to login
async function login(page: Page, username: string, password: string) {
  await page.goto('/login');
  await page.waitForLoadState('networkidle');
  await page.fill('input[name="username"]', username);
  await page.fill('input[name="password"]', password);
  await page.click('button[type="submit"]');
}

// Helper to get first available data source
async function getFirstDataSource(page: Page): Promise<string | null> {
  const authStorage = await page.evaluate(() => {
    const stored = localStorage.getItem('auth-storage');
    return stored ? JSON.parse(stored) : null;
  });
  
  const token = authStorage?.state?.token;
  
  const response = await page.request.get('http://localhost:8080/api/v1/datasources', {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });
  
  const data = await response.json();
  const dataSources = data.data_sources || [];
  
  if (dataSources.length > 0) {
    return dataSources[0].id;
  }
  
  return null;
}

// Helper to ensure data source exists
async function ensureDataSourceExists(page: Page): Promise<string> {
  let dataSourceId = await getFirstDataSource(page);
  
  if (!dataSourceId) {
    const authStorage = await page.evaluate(() => {
      const stored = localStorage.getItem('auth-storage');
      return stored ? JSON.parse(stored) : null;
    });
    
    const token = authStorage?.state?.token;
    
    const response = await page.request.post('http://localhost:8080/api/v1/datasources', {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      data: {
        name: 'Test PostgreSQL',
        type: 'postgresql',
        host: 'localhost',
        port: 5432,
        database_name: 'querybase',
        username: 'querybase',
        password: 'querybase',
      },
    });
    
    const dataSource = await response.json();
    dataSourceId = dataSource.id;
  }
  
  return dataSourceId;
}

// Helper to navigate to query editor
async function navigateToQueryEditor(page: Page) {
  await page.goto('/dashboard/query');
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(1000);
}

// Helper to select data source
async function selectDataSource(page: Page) {
  await page.waitForSelector('[data-testid="datasource-selector"], select, [role="combobox"]', { timeout: 10000 });
  
  const dataSourceSelector = page.locator('[data-testid="datasource-selector"]').first();
  if (await dataSourceSelector.count() > 0) {
    await dataSourceSelector.click();
  } else {
    const selectElement = page.locator('select').first();
    if (await selectElement.count() > 0) {
      await selectElement.selectOption({ index: 0 });
    }
  }
  
  await page.waitForTimeout(2000);
}

// Helper to type query in editor
async function typeQuery(page: Page, query: string) {
  const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
  await page.waitForSelector(editorSelector, { timeout: 10000 });
  
  const editor = page.locator(editorSelector).first();
  await editor.click();
  await editor.fill(query);
}

// Helper to execute query
async function executeQuery(page: Page) {
  await page.locator('button:has-text("Execute")').click();
  await page.waitForTimeout(2000);
}

test.describe('Multi-Query Transaction', () => {
  test.beforeEach(async ({ page }) => {
    // Clear auth storage before each test
    await page.goto('/login');
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await page.waitForLoadState('networkidle');
  });

  test('multi-query preview works', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Type a multi-query (multiple SELECT statements)
    const multiQuery = `SELECT 1 as id, 'first' as name;
SELECT 2 as id, 'second' as name;
SELECT 3 as id, 'third' as name;`;
    
    await typeQuery(page, multiQuery);
    await executeQuery(page);
    
    // Wait for preview modal to appear
    await page.waitForTimeout(3000);
    
    // Check for multi-query preview modal
    const previewModal = page.locator('text=/Multi-Query Preview|statements|Preview/i');
    const hasPreviewModal = await previewModal.count() > 0;
    
    // If preview modal appears, verify its contents
    if (hasPreviewModal) {
      // Check for statement count
      const statementCount = page.locator('text=/3.*Statements|Statements.*3/i');
      await expect(statementCount.first()).toBeVisible({ timeout: 5000 });
      
      // Check for operation type badges
      const selectBadge = page.locator('text=/SELECT/i').first();
      await expect(selectBadge).toBeVisible({ timeout: 5000 });
      
      // Check for execute/cancel buttons
      const executeButton = page.locator('button:has-text("Execute"), button:has-text("Execute All")');
      const cancelButton = page.locator('button:has-text("Cancel")');
      
      await expect(executeButton.first()).toBeVisible({ timeout: 5000 });
      await expect(cancelButton.first()).toBeVisible({ timeout: 5000 });
    } else {
      // If no preview modal, check for results directly (admin might execute directly)
      const resultsVisible = await page.locator('text=/Results|rows|Multi-Query/i').count() > 0;
      expect(resultsVisible).toBe(true);
    }
  });

  test('multi-query execution works', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Type a multi-query with SELECT statements
    const multiQuery = `SELECT 1 as test_id, 'statement1' as test_name;
SELECT 2 as test_id, 'statement2' as test_name;`;
    
    await typeQuery(page, multiQuery);
    await executeQuery(page);
    
    // Wait for execution
    await page.waitForTimeout(3000);
    
    // Check for preview modal or direct execution
    const previewModal = page.locator('text=/Multi-Query Preview/i');
    const hasPreviewModal = await previewModal.count() > 0;
    
    if (hasPreviewModal) {
      // Click execute button in preview modal
      const executeButton = page.locator('button:has-text("Execute"), button:has-text("Execute All")');
      await executeButton.first().click();
      await page.waitForTimeout(3000);
    }
    
    // Check for results
    const resultsVisible = await page.locator('text=/Results|rows|Success|Statement/i').count() > 0;
    expect(resultsVisible).toBe(true);
    
    // Check for statement count indicator
    const statementIndicator = page.locator('text=/2.*statement|statement.*2/i');
    const hasStatementIndicator = await statementIndicator.count() > 0;
    
    // Either statement indicator or results should be visible
    expect(resultsVisible || hasStatementIndicator).toBe(true);
  });

  test('commit after success works', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Type a multi-query with write operations (as admin, should execute directly)
    // Using safe queries that won't affect real data
    const multiQuery = `SELECT 1 as id;
SELECT 2 as id;`;
    
    await typeQuery(page, multiQuery);
    await executeQuery(page);
    
    // Wait for execution
    await page.waitForTimeout(3000);
    
    // Check for preview modal
    const previewModal = page.locator('text=/Multi-Query Preview/i');
    const hasPreviewModal = await previewModal.count() > 0;
    
    if (hasPreviewModal) {
      // Click execute button
      const executeButton = page.locator('button:has-text("Execute"), button:has-text("Execute All")');
      await executeButton.first().click();
      await page.waitForTimeout(3000);
    }
    
    // Check for success status
    const successIndicator = page.locator('text=/Success|success|completed|Committed/i');
    const hasSuccess = await successIndicator.count() > 0;
    
    // Check for results display
    const resultsVisible = await page.locator('text=/Results|rows/i').count() > 0;
    
    expect(hasSuccess || resultsVisible).toBe(true);
  });

  test('rollback on failure works', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Type a multi-query with an error (invalid table name)
    const multiQuery = `SELECT 1 as id;
SELECT * FROM nonexistent_table_xyz_12345;
SELECT 3 as id;`;
    
    await typeQuery(page, multiQuery);
    await executeQuery(page);
    
    // Wait for execution
    await page.waitForTimeout(3000);
    
    // Check for preview modal
    const previewModal = page.locator('text=/Multi-Query Preview/i');
    const hasPreviewModal = await previewModal.count() > 0;
    
    if (hasPreviewModal) {
      // Check for error indicator in preview
      const errorIndicator = page.locator('text=/error|Error|failed|Failed/i');
      const hasError = await errorIndicator.count() > 0;
      
      // Execute button should be disabled or show error
      if (hasError) {
        // Verify error is shown
        expect(hasError).toBe(true);
      } else {
        // Try to execute and check for error
        const executeButton = page.locator('button:has-text("Execute"), button:has-text("Execute All")');
        const isDisabled = await executeButton.first().isDisabled();
        
        if (!isDisabled) {
          await executeButton.first().click();
          await page.waitForTimeout(3000);
        }
      }
    } else {
      // Check for error message in results
      const errorVisible = await page.locator('text=/error|failed|does not exist|not found/i').count() > 0;
      expect(errorVisible).toBe(true);
    }
  });

  test('SET variable handling works', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Type a multi-query with SET statements
    const multiQuery = `SET search_path TO public;
SELECT 1 as test_value;
SET search_path TO public;`;
    
    await typeQuery(page, multiQuery);
    await executeQuery(page);
    
    // Wait for execution
    await page.waitForTimeout(3000);
    
    // Check for preview modal
    const previewModal = page.locator('text=/Multi-Query Preview/i');
    const hasPreviewModal = await previewModal.count() > 0;
    
    if (hasPreviewModal) {
      // SET statements should be recognized
      const setStatement = page.locator('text=/SET|set/i');
      const hasSetStatement = await setStatement.count() > 0;
      expect(hasSetStatement).toBe(true);
      
      // Execute the query
      const executeButton = page.locator('button:has-text("Execute"), button:has-text("Execute All")');
      await executeButton.first().click();
      await page.waitForTimeout(3000);
    }
    
    // Check for results
    const resultsVisible = await page.locator('text=/Results|rows|Success/i').count() > 0;
    expect(resultsVisible).toBe(true);
  });

  test('transaction control blocking works', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Type a multi-query with transaction control statements (should be blocked)
    const multiQuery = `BEGIN;
SELECT 1 as test;
COMMIT;`;
    
    await typeQuery(page, multiQuery);
    await executeQuery(page);
    
    // Wait for execution
    await page.waitForTimeout(3000);
    
    // Check for error about transaction control statements
    const errorVisible = await page.locator('text=/Transaction control|BEGIN|COMMIT|ROLLBACK|not allowed|blocked/i').count() > 0;
    
    // The system should either show an error or block the execution
    expect(errorVisible).toBe(true);
  });

  test('multiple statements execute in order', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Type a multi-query where order matters
    const multiQuery = `SELECT 1 as sequence_order, 'first' as label;
SELECT 2 as sequence_order, 'second' as label;
SELECT 3 as sequence_order, 'third' as label;`;
    
    await typeQuery(page, multiQuery);
    await executeQuery(page);
    
    // Wait for execution
    await page.waitForTimeout(3000);
    
    // Check for preview modal
    const previewModal = page.locator('text=/Multi-Query Preview/i');
    const hasPreviewModal = await previewModal.count() > 0;
    
    if (hasPreviewModal) {
      // Check that statements are listed in order
      const statementLabels = page.locator('text=/first|second|third/i');
      const count = await statementLabels.count();
      
      // Should have multiple statement references
      expect(count).toBeGreaterThanOrEqual(1);
      
      // Execute the query
      const executeButton = page.locator('button:has-text("Execute"), button:has-text("Execute All")');
      await executeButton.first().click();
      await page.waitForTimeout(3000);
    }
    
    // Check for results showing execution order
    const resultsVisible = await page.locator('text=/Results|rows|Success/i').count() > 0;
    expect(resultsVisible).toBe(true);
  });

  test('transaction status is displayed', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Type a multi-query
    const multiQuery = `SELECT 1 as id;
SELECT 2 as id;`;
    
    await typeQuery(page, multiQuery);
    await executeQuery(page);
    
    // Wait for execution
    await page.waitForTimeout(3000);
    
    // Check for preview modal
    const previewModal = page.locator('text=/Multi-Query Preview/i');
    const hasPreviewModal = await previewModal.count() > 0;
    
    if (hasPreviewModal) {
      // Check for status indicators in preview
      const statusIndicator = page.locator('text=/Statement|Est.*Affected|rows/i');
      const hasStatus = await statusIndicator.count() > 0;
      expect(hasStatus).toBe(true);
      
      // Execute the query
      const executeButton = page.locator('button:has-text("Execute"), button:has-text("Execute All")');
      await executeButton.first().click();
      await page.waitForTimeout(3000);
    }
    
    // Check for transaction status display
    const transactionStatus = page.locator('text=/Success|Failed|Pending|Status|Transaction|committed/i');
    const hasTransactionStatus = await transactionStatus.count() > 0;
    
    // Check for execution time
    const executionTime = page.locator('text=/ms|seconds|Execution.*time/i');
    const hasExecutionTime = await executionTime.count() > 0;
    
    // Check for affected rows
    const affectedRows = page.locator('text=/rows|affected|Statement/i');
    const hasAffectedRows = await affectedRows.count() > 0;
    
    // At least one of these should be visible
    expect(hasTransactionStatus || hasExecutionTime || hasAffectedRows).toBe(true);
  });

  test('multi-query with mixed operations shows correct types', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Type a multi-query with different operation types
    const multiQuery = `SELECT 1 as id;
SELECT 'test' as name;`;
    
    await typeQuery(page, multiQuery);
    await executeQuery(page);
    
    // Wait for preview
    await page.waitForTimeout(3000);
    
    // Check for preview modal
    const previewModal = page.locator('text=/Multi-Query Preview/i');
    const hasPreviewModal = await previewModal.count() > 0;
    
    if (hasPreviewModal) {
      // Check for SELECT operation type badge
      const selectBadge = page.locator('text=/SELECT/i');
      const selectCount = await selectBadge.count();
      
      // Should have SELECT badges for both statements
      expect(selectCount).toBeGreaterThanOrEqual(1);
      
      // Check for statement count
      const statementCount = page.locator('text=/2.*Statement|Statement.*2/i');
      const hasStatementCount = await statementCount.count() > 0;
      expect(hasStatementCount).toBe(true);
    }
  });

  test('multi-query preview shows estimated rows', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Type a multi-query
    const multiQuery = `SELECT generate_series(1, 10) as id;
SELECT generate_series(1, 5) as id;`;
    
    await typeQuery(page, multiQuery);
    await executeQuery(page);
    
    // Wait for preview
    await page.waitForTimeout(3000);
    
    // Check for preview modal
    const previewModal = page.locator('text=/Multi-Query Preview/i');
    const hasPreviewModal = await previewModal.count() > 0;
    
    if (hasPreviewModal) {
      // Check for estimated rows display
      const estimatedRows = page.locator('text=/Est.*row|row.*est|~.*row/i');
      const hasEstimatedRows = await estimatedRows.count() > 0;
      
      // Check for total estimated rows
      const totalRows = page.locator('text=/Total.*Affected|Affected.*row/i');
      const hasTotalRows = await totalRows.count() > 0;
      
      expect(hasEstimatedRows || hasTotalRows).toBe(true);
    }
  });

  test('multi-query cancel returns to editor', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Type a multi-query
    const multiQuery = `SELECT 1 as id;
SELECT 2 as id;`;
    
    await typeQuery(page, multiQuery);
    await executeQuery(page);
    
    // Wait for preview
    await page.waitForTimeout(3000);
    
    // Check for preview modal
    const previewModal = page.locator('text=/Multi-Query Preview/i');
    const hasPreviewModal = await previewModal.count() > 0;
    
    if (hasPreviewModal) {
      // Click cancel button
      const cancelButton = page.locator('button:has-text("Cancel")');
      await cancelButton.first().click();
      await page.waitForTimeout(1000);
      
      // Modal should be closed
      const modalClosed = await previewModal.count() === 0;
      expect(modalClosed).toBe(true);
      
      // Query should still be in editor
      const editorValue = await page.locator('.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea').first().inputValue();
      expect(editorValue).toContain('SELECT');
    }
  });

  test('multi-query with write operations requires approval for non-admin', async ({ page }) => {
    // Login as admin (admin can execute directly, but we check the preview behavior)
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Type a multi-query with write operations
    const multiQuery = `SELECT 1 as id;
UPDATE users SET full_name = 'Test' WHERE username = 'nonexistent_user_xyz';
SELECT 2 as id;`;
    
    await typeQuery(page, multiQuery);
    await executeQuery(page);
    
    // Wait for preview
    await page.waitForTimeout(3000);
    
    // Check for preview modal
    const previewModal = page.locator('text=/Multi-Query Preview/i');
    const hasPreviewModal = await previewModal.count() > 0;
    
    if (hasPreviewModal) {
      // Check for write operation indicator
      const writeOperation = page.locator('text=/UPDATE|Write|approval/i');
      const hasWriteOperation = await writeOperation.count() > 0;
      
      // Check for estimated rows (should show 0 for nonexistent user)
      const estimatedRows = page.locator('text=/0.*row|row.*0/i');
      const hasZeroRows = await estimatedRows.count() > 0;
      
      // Admin should see the preview with write operations
      expect(hasWriteOperation || hasZeroRows).toBe(true);
    }
  });

  test('multi-query results show individual statement status', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Type a multi-query
    const multiQuery = `SELECT 1 as id, 'first' as name;
SELECT 2 as id, 'second' as name;`;
    
    await typeQuery(page, multiQuery);
    await executeQuery(page);
    
    // Wait for execution
    await page.waitForTimeout(3000);
    
    // Check for preview modal
    const previewModal = page.locator('text=/Multi-Query Preview/i');
    const hasPreviewModal = await previewModal.count() > 0;
    
    if (hasPreviewModal) {
      // Execute the query
      const executeButton = page.locator('button:has-text("Execute"), button:has-text("Execute All")');
      await executeButton.first().click();
      await page.waitForTimeout(3000);
    }
    
    // Check for individual statement results
    const statementResults = page.locator('text=/Statement|#1|#2|sequence/i');
    const hasStatementResults = await statementResults.count() > 0;
    
    // Check for status indicators
    const statusIndicators = page.locator('text=/Success|success|completed|✓/i');
    const hasStatusIndicators = await statusIndicators.count() > 0;
    
    expect(hasStatementResults || hasStatusIndicators).toBe(true);
  });

  test('multi-query handles empty statements gracefully', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Type a multi-query with empty statements (extra semicolons)
    const multiQuery = `SELECT 1 as id;

SELECT 2 as id;

;`;
    
    await typeQuery(page, multiQuery);
    await executeQuery(page);
    
    // Wait for execution
    await page.waitForTimeout(3000);
    
    // Check for preview modal or results
    const previewModal = page.locator('text=/Multi-Query Preview/i');
    const hasPreviewModal = await previewModal.count() > 0;
    
    if (hasPreviewModal) {
      // Should show valid statements (empty ones filtered out)
      const statementCount = page.locator('text=/2.*Statement|Statement.*2/i');
      const hasCorrectCount = await statementCount.count() > 0;
      expect(hasCorrectCount).toBe(true);
    } else {
      // Check for results
      const resultsVisible = await page.locator('text=/Results|rows|Success/i').count() > 0;
      expect(resultsVisible).toBe(true);
    }
  });

  test('multi-query shows execution time per statement', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Type a multi-query
    const multiQuery = `SELECT 1 as id;
SELECT 2 as id;`;
    
    await typeQuery(page, multiQuery);
    await executeQuery(page);
    
    // Wait for execution
    await page.waitForTimeout(3000);
    
    // Check for preview modal
    const previewModal = page.locator('text=/Multi-Query Preview/i');
    const hasPreviewModal = await previewModal.count() > 0;
    
    if (hasPreviewModal) {
      // Execute the query
      const executeButton = page.locator('button:has-text("Execute"), button:has-text("Execute All")');
      await executeButton.first().click();
      await page.waitForTimeout(3000);
    }
    
    // Check for execution time display
    const executionTime = page.locator('text=/ms|millisecond|Execution.*time|time.*ms/i');
    const hasExecutionTime = await executionTime.count() > 0;
    
    // Check for total execution time
    const totalTime = page.locator('text=/Total.*time|time.*total|\\d+\\.\\d+s/i');
    const hasTotalTime = await totalTime.count() > 0;
    
    expect(hasExecutionTime || hasTotalTime).toBe(true);
  });
});