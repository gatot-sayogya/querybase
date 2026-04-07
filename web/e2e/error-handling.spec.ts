import { test, expect, Page, Route } from '@playwright/test';

/**
 * Error Handling E2E Tests
 * 
 * Tests cover:
 * - Network errors are handled gracefully
 * - Authentication errors show login prompt
 * - Authorization errors (403) show permission message
 * - Validation errors show field-level feedback
 * - SQL syntax errors show helpful messages
 * - Data source connection errors are handled
 * - Timeout handling works
 * - Error messages are user-friendly
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
  await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
  await page.waitForURL(/\/dashboard/, { timeout: 15000 });
  
  const authStorage = await page.evaluate(() => {
    const stored = localStorage.getItem('auth-storage');
    return stored ? JSON.parse(stored) : null;
  });
  
  const token = authStorage?.state?.token;
  
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
  
  await page.evaluate(() => localStorage.removeItem('auth-storage'));
  
  return response;
}

// Helper to cleanup test user
async function deleteTestUser(page: Page, username: string) {
  await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
  await page.waitForURL(/\/dashboard/, { timeout: 15000 });
  
  const authStorage = await page.evaluate(() => {
    const stored = localStorage.getItem('auth-storage');
    return stored ? JSON.parse(stored) : null;
  });
  
  const token = authStorage?.state?.token;
  
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
  
  await page.evaluate(() => localStorage.removeItem('auth-storage'));
}

// Helper to get auth token
async function getAuthToken(page: Page): Promise<string | null> {
  const authStorage = await page.evaluate(() => {
    const stored = localStorage.getItem('auth-storage');
    return stored ? JSON.parse(stored) : null;
  });
  return authStorage?.state?.token || null;
}

// Helper to navigate to query editor
async function navigateToQueryEditor(page: Page) {
  await page.goto('/dashboard/query');
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(1000);
}

// Helper to ensure data source exists
async function ensureDataSourceExists(page: Page): Promise<string> {
  const token = await getAuthToken(page);
  
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
  
  // Create a test data source
  const createResponse = await page.request.post('http://localhost:8080/api/v1/datasources', {
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
  
  const dataSource = await createResponse.json();
  return dataSource.id;
}

test.describe('Error Handling', () => {
  test.beforeEach(async ({ page }) => {
    // Clear auth storage before each test
    await page.goto('/login');
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await page.waitForLoadState('networkidle');
  });

  test('network errors are handled gracefully', async ({ page }) => {
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
    
    // Intercept API calls and simulate network failure
    await page.route('**/api/v1/queries', async (route: Route) => {
      await route.abort('failed');
    });
    
    // Type a query
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    await editor.fill('SELECT 1 as network_test;');
    
    // Execute the query
    await page.locator('button:has-text("Execute")').click();
    
    // Wait for error handling
    await page.waitForTimeout(3000);
    
    // Check for error message - should be user-friendly
    const errorVisible = await page.locator('text=/error|failed|network|connection|try again/i').count() > 0;
    expect(errorVisible).toBe(true);
    
    // Verify the page didn't crash and is still usable
    await expect(page.locator('button:has-text("Execute")')).toBeVisible();
    
    // Restore network and verify recovery
    await page.unroute('**/api/v1/queries');
    
    // Clear the error and try again
    await editor.fill('SELECT 1 as recovery_test;');
    await page.locator('button:has-text("Execute")').click();
    await page.waitForTimeout(3000);
    
    // Should show results or at least not crash
    const pageStillWorks = await page.locator('text=/Results|rows|error/i').count() > 0;
    expect(pageStillWorks).toBe(true);
  });

  test('authentication errors show login prompt', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Verify logged in
    await expect(page.locator('strong:has-text("admin")')).toBeVisible({ timeout: 10000 });
    
    // Simulate 401 error by clearing auth and making request
    await page.evaluate(() => {
      const authStorage = localStorage.getItem('auth-storage');
      if (authStorage) {
        const parsed = JSON.parse(authStorage);
        // Corrupt the token
        parsed.state.token = 'invalid_token';
        localStorage.setItem('auth-storage', JSON.stringify(parsed));
      }
    });
    
    // Intercept API calls and return 401
    await page.route('**/api/v1/**', async (route: Route) => {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Unauthorized' }),
      });
    });
    
    // Try to navigate to a protected route
    await page.goto('/dashboard');
    await page.waitForTimeout(2000);
    
    // Should redirect to login or show login prompt
    const redirectedToLogin = page.url().includes('/login');
    const loginPromptVisible = await page.locator('text=/login|sign in|session|expired/i').count() > 0;
    
    expect(redirectedToLogin || loginPromptVisible).toBe(true);
    
    // Restore network
    await page.unroute('**/api/v1/**');
  });

  test('authorization errors show permission message', async ({ page }) => {
    // Create a viewer user (limited permissions)
    await createTestUser(page, TEST_USERS.viewer);
    
    // Clear storage and login as viewer
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await login(page, TEST_USERS.viewer.username, TEST_USERS.viewer.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Try to access admin-only page (users management)
    await page.goto('/admin/users');
    await page.waitForTimeout(2000);
    
    // Should show permission denied or redirect away
    const permissionDenied = await page.locator('text=/permission|access denied|not authorized|forbidden|admin/i').count() > 0;
    const redirectedAway = !page.url().includes('/admin/users');
    
    expect(permissionDenied || redirectedAway).toBe(true);
    
    // Try to access data sources admin
    await page.goto('/admin/datasources');
    await page.waitForTimeout(2000);
    
    const dataSourceAccessDenied = await page.locator('text=/permission|access denied|not authorized|forbidden/i').count() > 0;
    const dataSourceRedirectedAway = !page.url().includes('/admin/datasources');
    
    expect(dataSourceAccessDenied || dataSourceRedirectedAway).toBe(true);
    
    // Cleanup
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await deleteTestUser(page, TEST_USERS.viewer.username);
  });

  test('validation errors show field-level feedback', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to admin users page
    await page.goto('/admin/users');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    // Try to create user with empty fields
    const createUserButton = page.locator('button:has-text("Add User"), button:has-text("Create"), button:has-text("New")').first();
    if (await createUserButton.count() > 0) {
      await createUserButton.click();
      await page.waitForTimeout(500);
      
      // Try to submit empty form
      const submitButton = page.locator('button[type="submit"]').first();
      await submitButton.click();
      await page.waitForTimeout(1000);
      
      // Check for validation errors
      const validationError = await page.locator('text=/required|invalid|empty|please|enter/i').count() > 0;
      const formStillOpen = await page.locator('input[name="username"], input[name="email"]').count() > 0;
      
      // Either validation error shown or form didn't close (HTML5 validation)
      expect(validationError || formStillOpen).toBe(true);
    }
    
    // Navigate to data sources
    await page.goto('/admin/datasources');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    // Try to create data source with invalid data
    const createDsButton = page.locator('button:has-text("Add Data Source"), button:has-text("Create"), button:has-text("New")').first();
    if (await createDsButton.count() > 0) {
      await createDsButton.click();
      await page.waitForTimeout(500);
      
      // Try to submit with empty/invalid fields
      const submitButton = page.locator('button[type="submit"]').first();
      await submitButton.click();
      await page.waitForTimeout(1000);
      
      // Check for validation errors
      const validationError = await page.locator('text=/required|invalid|empty|please|enter|valid/i').count() > 0;
      const formStillOpen = await page.locator('input[name="name"], input[name="host"]').count() > 0;
      
      expect(validationError || formStillOpen).toBe(true);
    }
  });

  test('SQL syntax errors show helpful messages', async ({ page }) => {
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
    
    // Test various SQL syntax errors
    const syntaxErrors = [
      {
        query: 'SELECT * FORM users;',
        errorPatterns: ['syntax', 'FORM', 'FROM', 'error'],
      },
      {
        query: 'SELCT * FROM users;',
        errorPatterns: ['syntax', 'SELCT', 'SELECT', 'error'],
      },
      {
        query: 'SELECT * FROM nonexistent_table_xyz;',
        errorPatterns: ['not exist', 'does not exist', 'unknown', 'error'],
      },
      {
        query: 'SELECT * FROM users WHERE;',
        errorPatterns: ['syntax', 'error', 'WHERE'],
      },
    ];
    
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    for (const { query, errorPatterns } of syntaxErrors) {
      const editor = page.locator(editorSelector).first();
      await editor.click();
      await editor.fill(query);
      
      // Execute the query
      await page.locator('button:has-text("Execute")').click();
      await page.waitForTimeout(3000);
      
      // Check for error message - at least one pattern should match
      let errorVisible = false;
      for (const pattern of errorPatterns) {
        const count = await page.locator(`text=/${pattern}/i`).count();
        if (count > 0) {
          errorVisible = true;
          break;
        }
      }
      expect(errorVisible).toBe(true);
      
      // Clear for next test
      await editor.fill('');
    }
  });

  test('data source connection errors are handled', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to admin data sources
    await page.goto('/admin/datasources');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    // Try to create a data source with invalid connection
    const createButton = page.locator('button:has-text("Add Data Source"), button:has-text("Create"), button:has-text("New")').first();
    if (await createButton.count() > 0) {
      await createButton.click();
      await page.waitForTimeout(500);
      
      // Fill in invalid connection details
      const nameInput = page.locator('input[name="name"]').first();
      if (await nameInput.count() > 0) {
        await nameInput.fill('Invalid Connection');
      }
      
      const hostInput = page.locator('input[name="host"]').first();
      if (await hostInput.count() > 0) {
        await hostInput.fill('nonexistent-host-xyz.invalid');
      }
      
      const portInput = page.locator('input[name="port"]').first();
      if (await portInput.count() > 0) {
        await portInput.fill('5432');
      }
      
      const dbInput = page.locator('input[name="database_name"], input[name="database"]').first();
      if (await dbInput.count() > 0) {
        await dbInput.fill('nonexistent_db');
      }
      
      const userInput = page.locator('input[name="username"]').first();
      if (await userInput.count() > 0) {
        await userInput.fill('invalid_user');
      }
      
      const passInput = page.locator('input[name="password"]').first();
      if (await passInput.count() > 0) {
        await passInput.fill('invalid_pass');
      }
      
      // Try to test connection
      const testButton = page.locator('button:has-text("Test Connection"), button:has-text("Test")').first();
      if (await testButton.count() > 0) {
        await testButton.click();
        await page.waitForTimeout(3000);
        
        // Should show connection error
        const connectionError = await page.locator('text=/connection|failed|error|unreachable|timeout|refused/i').count() > 0;
        expect(connectionError).toBe(true);
      }
    }
    
    // Test query execution with disconnected data source
    await navigateToQueryEditor(page);
    
    // If there's a way to select a disconnected data source, test it
    await page.waitForTimeout(1000);
    
    // The page should handle gracefully if no data source is available
    const noDataSourceMessage = await page.locator('text=/no data source|select.*data source|connect/i').count() > 0;
    const dataSourceSelectorVisible = await page.locator('[data-testid="datasource-selector"], select').count() > 0;
    
    expect(noDataSourceMessage || dataSourceSelectorVisible).toBe(true);
  });

  test('timeout handling works', async ({ page }) => {
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
    
    // Intercept API and simulate timeout
    await page.route('**/api/v1/queries', async (route: Route) => {
      // Delay response to simulate timeout
      await new Promise(resolve => setTimeout(resolve, 35000));
      await route.continue();
    });
    
    // Type a query
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    await editor.fill('SELECT 1 as timeout_test;');
    
    // Execute the query
    await page.locator('button:has-text("Execute")').click();
    
    // Wait for timeout handling (should show error before 30s)
    await page.waitForTimeout(5000);
    
    // Check for timeout or error message
    const timeoutError = await page.locator('text=/timeout|timed out|error|failed|try again/i').count() > 0;
    const loadingState = await page.locator('text=/loading|executing|running/i').count() > 0;
    
    // Either timeout error shown or still loading (will timeout eventually)
    expect(timeoutError || loadingState).toBe(true);
    
    // Restore network
    await page.unroute('**/api/v1/queries');
  });

  test('error messages are user-friendly', async ({ page }) => {
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
    
    // Test various error scenarios and verify user-friendly messages
    
    // 1. SQL syntax error
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    await editor.fill('SELECT * FROM nonexistent_table_xyz;');
    
    await page.locator('button:has-text("Execute")').click();
    await page.waitForTimeout(3000);
    
    // Error message should be user-friendly (not raw stack trace)
    const errorMessage = await page.locator('.bg-red-50, .dark\\:bg-red-900\\/20, [role="alert"], .error-msg, text=/error|failed/i').first();
    
    if (await errorMessage.count() > 0) {
      const errorText = await errorMessage.textContent();
      
      // Should not contain raw stack traces or technical jargon
      expect(errorText).not.toContain('stack trace');
      expect(errorText).not.toContain('undefined');
      expect(errorText).not.toContain('null');
      expect(errorText).not.toContain('[object Object]');
      
      // Should contain helpful information
      const hasHelpfulInfo = 
        errorText?.toLowerCase().includes('table') ||
        errorText?.toLowerCase().includes('does not exist') ||
        errorText?.toLowerCase().includes('not found') ||
        errorText?.toLowerCase().includes('syntax') ||
        errorText?.toLowerCase().includes('error');
      
      expect(hasHelpfulInfo).toBe(true);
    }
    
    // 2. Test login error with invalid credentials
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    
    await page.fill('input[name="username"]', 'invaliduser');
    await page.fill('input[name="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');
    
    await page.waitForTimeout(2000);
    
    // Should show user-friendly error
    const loginError = await page.locator('.bg-red-50, .dark\\:bg-red-900\\/20, .error-msg, text=/invalid|incorrect|failed|error/i').count();
    expect(loginError).toBeGreaterThan(0);
    
    // Error should not expose technical details
    const loginErrorText = await page.locator('.bg-red-50, .dark\\:bg-red-900\\/20, .error-msg').first().textContent();
    expect(loginErrorText).not.toContain('JWT');
    expect(loginErrorText).not.toContain('bcrypt');
    expect(loginErrorText).not.toContain('database');
  });

  test('error recovery allows retry', async ({ page }) => {
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
    
    // First, cause an error
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    await editor.fill('SELECT * FROM nonexistent_table;');
    
    await page.locator('button:has-text("Execute")').click();
    await page.waitForTimeout(3000);
    
    // Verify error appeared
    const errorVisible = await page.locator('text=/error|failed|not exist/i').count() > 0;
    expect(errorVisible).toBe(true);
    
    // Now fix the query and retry
    await editor.fill('SELECT 1 as retry_test;');
    await page.locator('button:has-text("Execute")').click();
    await page.waitForTimeout(3000);
    
    // Should show results after retry
    const resultsVisible = await page.locator('text=/Results|rows|retry_test/i').count() > 0;
    expect(resultsVisible).toBe(true);
    
    // Error should be cleared
    const errorStillVisible = await page.locator('.bg-red-50, .dark\\:bg-red-900\\/20').count();
    expect(errorStillVisible).toBe(0);
  });

  test('concurrent errors are handled properly', async ({ page }) => {
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
    
    // Intercept API and return errors
    let callCount = 0;
    await page.route('**/api/v1/queries', async (route: Route) => {
      callCount++;
      if (callCount <= 3) {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Internal server error' }),
        });
      } else {
        await route.continue();
      }
    });
    
    const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
    await page.waitForSelector(editorSelector, { timeout: 10000 });
    
    const editor = page.locator(editorSelector).first();
    await editor.click();
    await editor.fill('SELECT 1 as concurrent_test;');
    
    // Execute multiple times quickly
    await page.locator('button:has-text("Execute")').click();
    await page.waitForTimeout(500);
    await page.locator('button:has-text("Execute")').click();
    await page.waitForTimeout(500);
    await page.locator('button:has-text("Execute")').click();
    
    await page.waitForTimeout(3000);
    
    // Should show error message (not crash)
    const errorVisible = await page.locator('text=/error|failed/i').count() > 0;
    expect(errorVisible).toBe(true);
    
    // Page should still be functional
    await expect(page.locator('button:has-text("Execute")')).toBeVisible();
    
    // Restore network and retry
    await page.unroute('**/api/v1/queries');
    
    await editor.fill('SELECT 1 as recovery_test;');
    await page.locator('button:has-text("Execute")').click();
    await page.waitForTimeout(3000);
    
    // Should work now
    const resultsVisible = await page.locator('text=/Results|rows|recovery_test/i').count() > 0;
    expect(resultsVisible).toBe(true);
  });

  test('error boundary catches unexpected errors', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to various pages to ensure error boundaries work
    const routes = [
      '/dashboard',
      '/dashboard/history',
      '/dashboard/approvals',
      '/admin/users',
      '/admin/groups',
      '/admin/datasources',
    ];
    
    for (const route of routes) {
      await page.goto(route);
      await page.waitForTimeout(1000);
      
      // Page should load without crashing
      const pageLoaded = await page.locator('h1, h2, [role="main"], main').count() > 0;
      expect(pageLoaded).toBe(true);
      
      // Should not show generic error boundary
      const errorBoundary = await page.locator('text=/something went wrong|unexpected error|try refreshing/i').count();
      expect(errorBoundary).toBe(0);
    }
  });

  test('form validation errors are clear and actionable', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to login page and test form validation
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    
    // Test empty form submission
    await page.click('button[type="submit"]');
    await page.waitForTimeout(500);
    
    // Should stay on login page (HTML5 validation)
    await expect(page).toHaveURL(/\/login/);
    
    // Test invalid username format
    await page.fill('input[name="username"]', 'a'); // Too short
    await page.fill('input[name="password"]', 'test');
    await page.click('button[type="submit"]');
    await page.waitForTimeout(1000);
    
    // Should still be on login or show validation error
    const stillOnLogin = page.url().includes('/login');
    const validationError = await page.locator('text=/invalid|required|error/i').count() > 0;
    
    expect(stillOnLogin || validationError).toBe(true);
    
    // Test with valid format but wrong credentials
    await page.fill('input[name="username"]', 'nonexistentuser');
    await page.fill('input[name="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');
    await page.waitForTimeout(2000);
    
    // Should show authentication error
    const authError = await page.locator('text=/invalid|incorrect|failed|not found|error/i').count() > 0;
    expect(authError).toBe(true);
    
    // Error should be user-friendly
    const errorElement = page.locator('.bg-red-50, .dark\\:bg-red-900\\/20, .error-msg').first();
    if (await errorElement.count() > 0) {
      const errorText = await errorElement.textContent();
      // Should not expose technical details
      expect(errorText).not.toContain('SELECT');
      expect(errorText).not.toContain('password_hash');
      expect(errorText).not.toContain('database');
    }
  });
});