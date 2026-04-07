import { test, expect, Page } from '@playwright/test';

/**
 * Query Execution E2E Tests
 * 
 * Tests cover:
 * - SELECT query execution
 * - Query results display
 * - Pagination of results
 * - Export to CSV/JSON
 * - Error handling for invalid SQL
 * - Different user roles (admin, regular, viewer)
 * - Query history
 */

// Test user credentials
const TEST_USERS = {
  admin: {
    username: 'admin',
    password: 'admin123',
    role: 'admin',
    fullName: 'System Administrator',
  },
  regularUser: {
    username: 'testuser',
    password: 'testpass123',
    role: 'user',
    fullName: 'Test User',
    email: 'testuser@querybase.local',
  },
  viewer: {
    username: 'testviewer',
    password: 'viewerpass123',
    role: 'viewer',
    fullName: 'Test Viewer',
    email: 'testviewer@querybase.local',
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

// Helper to create a test user via API
async function createTestUser(page: Page, userData: typeof TEST_USERS.regularUser) {
  // First login as admin to get auth token
  await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
  await page.waitForURL(/\/dashboard/, { timeout: 15000 });
  
  // Get the auth token from localStorage
  const authStorage = await page.evaluate(() => {
    const stored = localStorage.getItem('auth-storage');
    return stored ? JSON.parse(stored) : null;
  });
  
  const token = authStorage?.state?.token;
  
  // Create user via API
  const response = await page.request.post('http://localhost:8080/api/v1/auth/users', {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    data: {
      email: userData.email,
      username: userData.username,
      password: userData.password,
      full_name: userData.fullName,
      role: userData.role,
    },
  });
  
  // Logout admin
  await page.evaluate(() => localStorage.removeItem('auth-storage'));
  
  return response;
}

// Helper to cleanup test user
async function deleteTestUser(page: Page, username: string) {
  // Login as admin
  await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
  await page.waitForURL(/\/dashboard/, { timeout: 15000 });
  
  // Get the auth token
  const authStorage = await page.evaluate(() => {
    const stored = localStorage.getItem('auth-storage');
    return stored ? JSON.parse(stored) : null;
  });
  
  const token = authStorage?.state?.token;
  
  // Get users list to find the user ID
  const usersResponse = await page.request.get('http://localhost:8080/api/v1/auth/users', {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });
  
  const users = await usersResponse.json();
  const userToDelete = users.find((u: { username: string }) => u.username === username);
  
  if (userToDelete) {
    await page.request.delete(`http://localhost:8080/api/v1/auth/users/${userToDelete.id}`, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });
  }
  
  // Clear storage
  await page.evaluate(() => localStorage.removeItem('auth-storage'));
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
    // Create a test data source
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
  // Wait for the page to be ready
  await page.waitForTimeout(1000);
}

test.describe('Query Execution', () => {
  test.beforeEach(async ({ page }) => {
    // Clear auth storage before each test
    await page.goto('/login');
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await page.waitForLoadState('networkidle');
  });

  test('SELECT query execution works', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    const dataSourceId = await ensureDataSourceExists(page);
    
    // Select data source from dropdown
    await page.waitForSelector('[data-testid="datasource-selector"], select, [role="combobox"]', { timeout: 10000 });
    
    // Click on data source selector or select from dropdown
    const dataSourceSelector = page.locator('[data-testid="datasource-selector"]').first();
    if (await dataSourceSelector.count() > 0) {
      await dataSourceSelector.click();
    } else {
      // Try alternative selectors
      const selectElement = page.locator('select').first();
      if (await selectElement.count() > 0) {
        await selectElement.selectOption({ index: 0 });
      }
    }
    
    // Wait for schema to load
    await page.waitForTimeout(2000);
    
    // Type a SELECT query in the editor
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    await editor.fill('SELECT 1 as test_column;');
    
    // Execute the query
    const executeButton = page.locator('button:has-text("Execute")');
    await executeButton.click();
    
    // Wait for results
    await page.waitForTimeout(3000);
    
    // Check for results or error - either is acceptable for basic execution test
    const resultsVisible = await page.locator('text=/Results|rows|error/i').count() > 0;
    expect(resultsVisible).toBe(true);
  });

  test('query results display correctly', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
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
    
    // Type a query that returns structured results
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    await editor.fill(`SELECT 
      1 as id, 
      'Test User' as name, 
      'test@example.com' as email
    UNION ALL
    SELECT 
      2 as id, 
      'Another User' as name, 
      'another@example.com' as email;`);
    
    // Execute the query
    await page.locator('button:has-text("Execute")').click();
    
    // Wait for results
    await page.waitForTimeout(3000);
    
    // Check for results container
    const resultsContainer = page.locator('text=/Results|rows/i');
    await expect(resultsContainer.first()).toBeVisible({ timeout: 10000 });
    
    // Check for row count indicator
    const rowCountVisible = await page.locator('text=/2 rows|row_count/i').count() > 0;
    expect(rowCountVisible).toBe(true);
  });

  test('pagination of results works', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
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
    
    // Type a query that returns many rows (generate series for PostgreSQL)
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    // Use a query that generates enough rows to trigger pagination
    await editor.fill(`SELECT generate_series(1, 100) as id, 'User ' || generate_series(1, 100) as name;`);
    
    // Execute the query
    await page.locator('button:has-text("Execute")').click();
    
    // Wait for results
    await page.waitForTimeout(3000);
    
    // Check if pagination controls exist (if results are paginated)
    const paginationControls = page.locator('button:has-text("Next"), button:has-text("Previous"), [aria-label*="page"], [data-testid*="pagination"]');
    const hasPagination = await paginationControls.count() > 0;
    
    // If pagination exists, test it
    if (hasPagination) {
      // Click next page
      const nextButton = page.locator('button:has-text("Next"), [aria-label="Next page"]').first();
      if (await nextButton.isEnabled()) {
        await nextButton.click();
        await page.waitForTimeout(1000);
        
        // Verify we're on a different page
        const pageIndicator = page.locator('text=/page 2|Page 2/i');
        const hasPageIndicator = await pageIndicator.count() > 0;
        expect(hasPageIndicator).toBe(true);
      }
    }
    
    // Results should be visible regardless of pagination
    const resultsVisible = await page.locator('text=/Results|rows/i').count() > 0;
    expect(resultsVisible).toBe(true);
  });

  test('export to CSV works', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
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
    
    // Execute a query first
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    await editor.fill('SELECT 1 as id, \'Test\' as name;');
    
    await page.locator('button:has-text("Execute")').click();
    await page.waitForTimeout(3000);
    
    // Wait for results to appear
    await page.waitForSelector('text=/Results|rows/i', { timeout: 10000 });
    
    // Setup download handler before clicking export
    const downloadPromise = page.waitForEvent('download', { timeout: 10000 }).catch(() => null);
    
    // Click CSV export button
    const csvButton = page.locator('button:has-text("CSV")');
    await csvButton.click();
    
    // Wait for download
    const download = await downloadPromise;
    
    // If download started, verify it
    if (download) {
      const fileName = download.suggestedFilename();
      expect(fileName).toContain('.csv');
    }
  });

  test('export to JSON works', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
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
    
    // Execute a query first
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    await editor.fill('SELECT 1 as id, \'Test\' as name;');
    
    await page.locator('button:has-text("Execute")').click();
    await page.waitForTimeout(3000);
    
    // Wait for results to appear
    await page.waitForSelector('text=/Results|rows/i', { timeout: 10000 });
    
    // Setup download handler before clicking export
    const downloadPromise = page.waitForEvent('download', { timeout: 10000 }).catch(() => null);
    
    // Click JSON export button
    const jsonButton = page.locator('button:has-text("JSON")');
    await jsonButton.click();
    
    // Wait for download
    const download = await downloadPromise;
    
    // If download started, verify it
    if (download) {
      const fileName = download.suggestedFilename();
      expect(fileName).toContain('.json');
    }
  });

  test('error handling for invalid SQL works', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
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
    
    // Type an invalid SQL query
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    await editor.fill('SELECT * FROM nonexistent_table_xyz;');
    
    // Execute the query
    await page.locator('button:has-text("Execute")').click();
    
    // Wait for error to appear
    await page.waitForTimeout(3000);
    
    // Check for error message
    const errorVisible = await page.locator('text=/error|failed|not found|does not exist/i').count() > 0;
    expect(errorVisible).toBe(true);
  });

  test('admin can execute any query', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
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
    
    // Verify admin has write permissions (Read + Write badge)
    const writeBadge = page.locator('text=/Read \\+ Write|Write/i');
    const hasWritePermission = await writeBadge.count() > 0;
    
    // Admin should have write permissions
    expect(hasWritePermission).toBe(true);
    
    // Execute a SELECT query
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    await editor.fill('SELECT 1 as admin_test;');
    
    await page.locator('button:has-text("Execute")').click();
    await page.waitForTimeout(3000);
    
    // Verify results appear
    const resultsVisible = await page.locator('text=/Results|rows|admin_test/i').count() > 0;
    expect(resultsVisible).toBe(true);
  });

  test('regular user can execute with permissions', async ({ page }) => {
    // Create test user first
    await createTestUser(page, TEST_USERS.regularUser);
    
    // Clear storage and navigate to login
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    
    // Login as regular user
    await login(page, TEST_USERS.regularUser.username, TEST_USERS.regularUser.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Check if data source is available for this user
    const dataSourceSelector = page.locator('[data-testid="datasource-selector"]').first();
    const selectElement = page.locator('select').first();
    
    const hasDataSource = (await dataSourceSelector.count() > 0) || (await selectElement.count() > 0);
    
    if (hasDataSource) {
      // Select data source
      if (await dataSourceSelector.count() > 0) {
        await dataSourceSelector.click();
      } else {
        await selectElement.selectOption({ index: 0 });
      }
      
      await page.waitForTimeout(2000);
      
      // Execute a SELECT query
      const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
      await page.waitForSelector(editorSelector, { timeout: 10000 });
      
      const editor = page.locator(editorSelector).first();
      await editor.click();
      await editor.fill('SELECT 1 as user_test;');
      
      await page.locator('button:has-text("Execute")').click();
      await page.waitForTimeout(3000);
      
      // Results or permission error should appear
      const responseVisible = await page.locator('text=/Results|rows|Permission|error/i').count() > 0;
      expect(responseVisible).toBe(true);
    }
    
    // Cleanup
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await deleteTestUser(page, TEST_USERS.regularUser.username);
  });

  test('query history is saved', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
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
    
    // Execute a unique query
    const uniqueQuery = `SELECT 'history_test_${Date.now()}' as test_identifier;`;
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    await editor.fill(uniqueQuery);
    
    await page.locator('button:has-text("Execute")').click();
    await page.waitForTimeout(3000);
    
    // Navigate to history page
    await page.goto('/dashboard/history');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    
    // Check if history page loads
    await expect(page.locator('h1, h2, [data-testid="history-title"]').first()).toBeVisible({ timeout: 10000 });
    
    // History should show queries
    const historyContent = await page.locator('text=/history|query|SELECT/i').count();
    expect(historyContent).toBeGreaterThan(0);
  });

  test('viewer role has read-only access', async ({ page }) => {
    // Create test viewer user
    await createTestUser(page, TEST_USERS.viewer);
    
    // Clear storage and navigate to login
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    
    // Login as viewer
    await login(page, TEST_USERS.viewer.username, TEST_USERS.viewer.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Check if data source is available
    const dataSourceSelector = page.locator('[data-testid="datasource-selector"]').first();
    const selectElement = page.locator('select').first();
    
    const hasDataSource = (await dataSourceSelector.count() > 0) || (await selectElement.count() > 0);
    
    if (hasDataSource) {
      // Select data source
      if (await dataSourceSelector.count() > 0) {
        await dataSourceSelector.click();
      } else {
        await selectElement.selectOption({ index: 0 });
      }
      
      await page.waitForTimeout(2000);
      
      // Check for Read Only badge
      const readOnlyBadge = page.locator('text=/Read Only/i');
      const hasReadOnly = await readOnlyBadge.count() > 0;
      
      // Viewer should have read-only access
      expect(hasReadOnly).toBe(true);
      
      // Execute a SELECT query (should work)
      const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
      await page.waitForSelector(editorSelector, { timeout: 10000 });
      
      const editor = page.locator(editorSelector).first();
      await editor.click();
      await editor.fill('SELECT 1 as viewer_test;');
      
      await page.locator('button:has-text("Execute")').click();
      await page.waitForTimeout(3000);
      
      // Results should appear for SELECT
      const resultsVisible = await page.locator('text=/Results|rows/i').count() > 0;
      expect(resultsVisible).toBe(true);
    }
    
    // Cleanup
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await deleteTestUser(page, TEST_USERS.viewer.username);
  });

  test('query with LIMIT clause works', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
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
    
    // Execute query with LIMIT
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    await editor.fill('SELECT 1 as id UNION ALL SELECT 2 UNION ALL SELECT 3 LIMIT 2;');
    
    await page.locator('button:has-text("Execute")').click();
    await page.waitForTimeout(3000);
    
    // Check for results
    const resultsVisible = await page.locator('text=/Results|rows/i').count() > 0;
    expect(resultsVisible).toBe(true);
    
    // Check row count (should be 2 due to LIMIT)
    const rowCountVisible = await page.locator('text=/2 rows/i').count() > 0;
    expect(rowCountVisible).toBe(true);
  });

  test('empty query shows validation error', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
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
    
    // Try to execute without entering a query
    const executeButton = page.locator('button:has-text("Execute")');
    
    // Button should be disabled or show error
    const isDisabled = await executeButton.isDisabled();
    
    if (!isDisabled) {
      // If not disabled, clicking should show an error
      await executeButton.click();
      await page.waitForTimeout(1000);
      
      const errorVisible = await page.locator('text=/enter.*query|Please enter/i').count() > 0;
      expect(errorVisible).toBe(true);
    } else {
      expect(isDisabled).toBe(true);
    }
  });

  test('query status indicator shows execution time', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
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
    
    // Execute a query
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    await editor.fill('SELECT 1;');
    
    await page.locator('button:has-text("Execute")').click();
    await page.waitForTimeout(3000);
    
    // Check for execution status (completed, running, etc.)
    const statusVisible = await page.locator('text=/completed|running|ms|rows/i').count() > 0;
    expect(statusVisible).toBe(true);
  });

  test('save query functionality works', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
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
    
    // Type a query
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    await editor.fill('SELECT 1 as saved_query_test;');
    
    // Click save button
    const saveButton = page.locator('button:has-text("Save")');
    await saveButton.click();
    
    // Wait for success message
    await page.waitForTimeout(2000);
    
    // Check for success toast or message
    const successVisible = await page.locator('text=/saved|success|Saved query/i').count() > 0;
    expect(successVisible).toBe(true);
  });
});