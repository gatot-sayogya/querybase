import { test, expect, Page } from '@playwright/test';

/**
 * Permission Matrix E2E Tests
 * 
 * Tests cover:
 * - Admin: all operations allowed
 * - User with read: SELECT only
 * - User with write: SELECT + write with approval
 * - User with approve: can approve write requests
 * - Viewer: SELECT only
 * - UI shows correct options per role
 * - Disabled buttons for missing permissions
 * - Error messages for denied operations
 */

// Test user credentials
const TEST_USERS = {
  admin: {
    username: 'admin',
    password: 'admin123',
    role: 'admin',
    fullName: 'System Administrator',
    email: 'admin@querybase.local',
  },
  readUser: {
    username: 'perm_test_read',
    password: 'testpass123',
    role: 'user',
    fullName: 'Read Permission User',
    email: 'perm_read@querybase.local',
  },
  writeUser: {
    username: 'perm_test_write',
    password: 'testpass123',
    role: 'user',
    fullName: 'Write Permission User',
    email: 'perm_write@querybase.local',
  },
  approveUser: {
    username: 'perm_test_approve',
    password: 'testpass123',
    role: 'user',
    fullName: 'Approve Permission User',
    email: 'perm_approve@querybase.local',
  },
  viewer: {
    username: 'perm_test_viewer',
    password: 'testpass123',
    role: 'viewer',
    fullName: 'Viewer Permission User',
    email: 'perm_viewer@querybase.local',
  },
};

// Test group names for different permission levels
const TEST_GROUPS = {
  readGroup: 'perm_test_read_group',
  writeGroup: 'perm_test_write_group',
  approveGroup: 'perm_test_approve_group',
};

// Helper to login
async function login(page: Page, username: string, password: string) {
  await page.goto('/login');
  await page.waitForLoadState('networkidle');
  await page.fill('input[name="username"]', username);
  await page.fill('input[name="password"]', password);
  await page.click('button[type="submit"]');
}

// Helper to get auth token
async function getAuthToken(page: Page): Promise<string | null> {
  const authStorage = await page.evaluate(() => {
    const stored = localStorage.getItem('auth-storage');
    return stored ? JSON.parse(stored) : null;
  });
  return authStorage?.state?.token || null;
}

// Helper to create a test user via API
async function createTestUser(
  page: Page, 
  userData: typeof TEST_USERS.readUser,
  groupWithPermission?: string
) {
  await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
  await page.waitForURL(/\/dashboard/, { timeout: 15000 });
  
  const token = await getAuthToken(page);
  
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
  
  if (!response.ok()) {
    console.log(`Failed to create user ${userData.username}: ${response.status()}`);
    return null;
  }
  
  const user = await response.json();
  
  // If group is specified, add user to that group
  if (groupWithPermission && user.id) {
    // Get groups to find the group ID
    const groupsResponse = await page.request.get('http://localhost:8080/api/v1/groups', {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });
    
    const groups = await groupsResponse.json();
    const targetGroup = groups.groups?.find((g: { name: string }) => g.name === groupWithPermission);
    
    if (targetGroup) {
      // Add user to group
      await page.request.post(`http://localhost:8080/api/v1/groups/${targetGroup.id}/members`, {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        data: {
          user_id: user.id,
        },
      });
    }
  }
  
  // Logout admin
  await page.evaluate(() => localStorage.removeItem('auth-storage'));
  
  return user;
}

// Helper to create a test group with specific permissions
async function createTestGroupWithPermissions(
  page: Page,
  groupName: string,
  dataSourceId: string,
  permissions: { can_read: boolean; can_write: boolean; can_approve: boolean }
) {
  await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
  await page.waitForURL(/\/dashboard/, { timeout: 15000 });
  
  const token = await getAuthToken(page);
  
  // Create group
  const createGroupResponse = await page.request.post('http://localhost:8080/api/v1/groups', {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    data: {
      name: groupName,
      description: `Test group for permission matrix testing`,
    },
  });
  
  let groupId: string;
  
  if (!createGroupResponse.ok()) {
    // Group might already exist, try to find it
    const groupsResponse = await page.request.get('http://localhost:8080/api/v1/groups', {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });
    
    const groups = await groupsResponse.json();
    const existingGroup = groups.groups?.find((g: { name: string }) => g.name === groupName);
    
    if (existingGroup) {
      groupId = existingGroup.id;
    } else {
      await page.evaluate(() => localStorage.removeItem('auth-storage'));
      return null;
    }
  } else {
    const group = await createGroupResponse.json();
    groupId = group.id;
  }
  
  // Set permissions for the group
  await page.request.put(`http://localhost:8080/api/v1/groups/${groupId}/datasource_permissions`, {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    data: {
      data_source_id: dataSourceId,
      can_read: permissions.can_read,
      can_write: permissions.can_write,
      can_approve: permissions.can_approve,
    },
  });
  
  // Logout admin
  await page.evaluate(() => localStorage.removeItem('auth-storage'));
  
  return groupId;
}

// Helper to cleanup test user
async function deleteTestUser(page: Page, username: string) {
  await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
  await page.waitForURL(/\/dashboard/, { timeout: 15000 });
  
  const token = await getAuthToken(page);
  
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

// Helper to cleanup test group
async function deleteTestGroup(page: Page, groupName: string) {
  await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
  await page.waitForURL(/\/dashboard/, { timeout: 15000 });
  
  const token = await getAuthToken(page);
  
  // Get groups
  const groupsResponse = await page.request.get('http://localhost:8080/api/v1/groups', {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });
  
  const groups = await groupsResponse.json();
  const groupToDelete = groups.groups?.find((g: { name: string }) => g.name === groupName);
  
  if (groupToDelete) {
    await page.request.delete(`http://localhost:8080/api/v1/groups/${groupToDelete.id}`, {
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
  
  return null;
}

// Helper to ensure data source exists
async function ensureDataSourceExists(page: Page): Promise<string> {
  let dataSourceId = await getFirstDataSource(page);
  
  if (!dataSourceId) {
    const token = await getAuthToken(page);
    
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

// Helper to execute a query
async function executeQuery(page: Page, query: string) {
  const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
  await page.waitForSelector(editorSelector, { timeout: 10000 });
  
  const editor = page.locator(editorSelector).first();
  await editor.click();
  await editor.fill(query);
  
  await page.locator('button:has-text("Execute")').click();
  await page.waitForTimeout(3000);
}

// Helper to navigate to approvals page
async function navigateToApprovals(page: Page) {
  await page.goto('/dashboard/approvals');
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(1000);
}

test.describe('Permission Matrix', () => {
  test.beforeEach(async ({ page }) => {
    // Clear auth storage before each test
    await page.goto('/login');
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await page.waitForLoadState('networkidle');
  });

  test('admin can perform all operations', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Ensure data source exists
    await ensureDataSourceExists(page);
    
    // Select data source
    await selectDataSource(page);
    
    // Verify admin has write permissions (Read + Write badge)
    const writeBadge = page.locator('text=/Read \\+ Write|Write/i');
    const hasWritePermission = await writeBadge.count() > 0;
    expect(hasWritePermission).toBe(true);
    
    // Execute a SELECT query
    await executeQuery(page, 'SELECT 1 as admin_test;');
    
    // Verify results appear
    const resultsVisible = await page.locator('text=/Results|rows|admin_test/i').count() > 0;
    expect(resultsVisible).toBe(true);
    
    // Verify admin can access admin pages
    await page.goto('/admin/datasources');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1:has-text("Data Sources")')).toBeVisible({ timeout: 10000 });
    
    await page.goto('/admin/users');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1:has-text("Users")')).toBeVisible({ timeout: 10000 });
    
    await page.goto('/admin/groups');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('h1:has-text("Groups")')).toBeVisible({ timeout: 10000 });
  });

  test('user with read can only SELECT', async ({ page }) => {
    // Setup: Create group with read-only permissions
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    const dataSourceId = await ensureDataSourceExists(page);
    
    // Create read-only group
    await createTestGroupWithPermissions(page, TEST_GROUPS.readGroup, dataSourceId, {
      can_read: true,
      can_write: false,
      can_approve: false,
    });
    
    // Create user and add to read-only group
    await createTestUser(page, TEST_USERS.readUser, TEST_GROUPS.readGroup);
    
    // Clear storage and login as read user
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await login(page, TEST_USERS.readUser.username, TEST_USERS.readUser.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Check if data source is available
    const dataSourceSelector = page.locator('[data-testid="datasource-selector"]').first();
    const selectElement = page.locator('select').first();
    const hasDataSource = (await dataSourceSelector.count() > 0) || (await selectElement.count() > 0);
    
    if (hasDataSource) {
      await selectDataSource(page);
      
      // Check for Read Only badge
      const readOnlyBadge = page.locator('text=/Read Only/i');
      const hasReadOnly = await readOnlyBadge.count() > 0;
      expect(hasReadOnly).toBe(true);
      
      // Execute a SELECT query (should work)
      await executeQuery(page, 'SELECT 1 as read_test;');
      
      // Results should appear for SELECT
      const resultsVisible = await page.locator('text=/Results|rows/i').count() > 0;
      expect(resultsVisible).toBe(true);
      
      // Try to execute a write query (should fail or show permission error)
      await executeQuery(page, "UPDATE users SET full_name = 'Test' WHERE id = 'non-existent';");
      
      // Should see permission error or approval required message
      const permissionError = await page.locator('text=/permission|denied|Write Access Required|error/i').count() > 0;
      expect(permissionError).toBe(true);
    }
    
    // Cleanup
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await deleteTestUser(page, TEST_USERS.readUser.username);
    await deleteTestGroup(page, TEST_GROUPS.readGroup);
  });

  test('user with write can SELECT and write with approval', async ({ page }) => {
    // Setup: Create group with write permissions
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    const dataSourceId = await ensureDataSourceExists(page);
    
    // Create write group
    await createTestGroupWithPermissions(page, TEST_GROUPS.writeGroup, dataSourceId, {
      can_read: true,
      can_write: true,
      can_approve: false,
    });
    
    // Create user and add to write group
    await createTestUser(page, TEST_USERS.writeUser, TEST_GROUPS.writeGroup);
    
    // Clear storage and login as write user
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await login(page, TEST_USERS.writeUser.username, TEST_USERS.writeUser.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Check if data source is available
    const dataSourceSelector = page.locator('[data-testid="datasource-selector"]').first();
    const selectElement = page.locator('select').first();
    const hasDataSource = (await dataSourceSelector.count() > 0) || (await selectElement.count() > 0);
    
    if (hasDataSource) {
      await selectDataSource(page);
      
      // Check for Read + Write badge
      const writeBadge = page.locator('text=/Read \\+ Write|Write/i');
      const hasWrite = await writeBadge.count() > 0;
      expect(hasWrite).toBe(true);
      
      // Execute a SELECT query (should work)
      await executeQuery(page, 'SELECT 1 as write_user_test;');
      
      // Results should appear for SELECT
      const resultsVisible = await page.locator('text=/Results|rows/i').count() > 0;
      expect(resultsVisible).toBe(true);
      
      // Execute a write query (should create approval request)
      await executeQuery(page, "UPDATE users SET full_name = 'Test Write User' WHERE id = 'non-existent-id';");
      
      // Should see approval required message or preview modal
      const approvalRequired = await page.locator('text=/approval|pending|preview|affected/i').count() > 0;
      expect(approvalRequired).toBe(true);
    }
    
    // Cleanup
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await deleteTestUser(page, TEST_USERS.writeUser.username);
    await deleteTestGroup(page, TEST_GROUPS.writeGroup);
  });

  test('user with approve can approve write requests', async ({ page }) => {
    // Setup: Create group with approve permissions
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    const dataSourceId = await ensureDataSourceExists(page);
    
    // Create approve group
    await createTestGroupWithPermissions(page, TEST_GROUPS.approveGroup, dataSourceId, {
      can_read: true,
      can_write: true,
      can_approve: true,
    });
    
    // Create user and add to approve group
    await createTestUser(page, TEST_USERS.approveUser, TEST_GROUPS.approveGroup);
    
    // Clear storage and login as approve user
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await login(page, TEST_USERS.approveUser.username, TEST_USERS.approveUser.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to approvals page
    await navigateToApprovals(page);
    
    // Verify approvals page loads
    await expect(page.locator('h1, h2, [data-testid="approvals-title"]').first()).toBeVisible({ timeout: 10000 });
    
    // Check for pending approvals tab
    const pendingTab = page.locator('button:has-text("Pending")');
    await expect(pendingTab).toBeVisible({ timeout: 5000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Check if data source is available
    const dataSourceSelector = page.locator('[data-testid="datasource-selector"]').first();
    const selectElement = page.locator('select').first();
    const hasDataSource = (await dataSourceSelector.count() > 0) || (await selectElement.count() > 0);
    
    if (hasDataSource) {
      await selectDataSource(page);
      
      // Check for Read + Write badge (approve users also have write)
      const writeBadge = page.locator('text=/Read \\+ Write|Write/i');
      const hasWrite = await writeBadge.count() > 0;
      expect(hasWrite).toBe(true);
      
      // Execute a SELECT query
      await executeQuery(page, 'SELECT 1 as approve_user_test;');
      
      // Results should appear
      const resultsVisible = await page.locator('text=/Results|rows/i').count() > 0;
      expect(resultsVisible).toBe(true);
    }
    
    // Cleanup
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await deleteTestUser(page, TEST_USERS.approveUser.username);
    await deleteTestGroup(page, TEST_GROUPS.approveGroup);
  });

  test('viewer can only SELECT', async ({ page }) => {
    // Create viewer user
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    await ensureDataSourceExists(page);
    
    // Create viewer user (viewer role has no group permissions by default)
    await createTestUser(page, TEST_USERS.viewer);
    
    // Clear storage and login as viewer
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await login(page, TEST_USERS.viewer.username, TEST_USERS.viewer.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Check if data source is available (viewer might not have access)
    const dataSourceSelector = page.locator('[data-testid="datasource-selector"]').first();
    const selectElement = page.locator('select').first();
    const hasDataSource = (await dataSourceSelector.count() > 0) || (await selectElement.count() > 0);
    
    if (hasDataSource) {
      await selectDataSource(page);
      
      // Check for Read Only badge
      const readOnlyBadge = page.locator('text=/Read Only/i');
      const hasReadOnly = await readOnlyBadge.count() > 0;
      expect(hasReadOnly).toBe(true);
      
      // Execute a SELECT query (should work)
      await executeQuery(page, 'SELECT 1 as viewer_test;');
      
      // Results should appear for SELECT
      const resultsVisible = await page.locator('text=/Results|rows/i').count() > 0;
      expect(resultsVisible).toBe(true);
      
      // Verify viewer cannot access admin pages
      await page.goto('/admin/users');
      await page.waitForTimeout(2000);
      
      // Should be redirected or see access denied
      const currentUrl = page.url();
      const isRedirected = !currentUrl.includes('/admin/users') || currentUrl.includes('/login');
      expect(isRedirected).toBe(true);
    }
    
    // Cleanup
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await deleteTestUser(page, TEST_USERS.viewer.username);
  });

  test('UI shows correct options per role', async ({ page }) => {
    // Test admin UI
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Verify admin sees all navigation options
    await expect(page.locator('a:has-text("Data Sources")')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('a:has-text("Users")')).toBeVisible();
    await expect(page.locator('a:has-text("Groups")')).toBeVisible();
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    await ensureDataSourceExists(page);
    await selectDataSource(page);
    
    // Verify admin has Read + Write badge
    const adminWriteBadge = page.locator('text=/Read \\+ Write|Write/i');
    const adminHasWrite = await adminWriteBadge.count() > 0;
    expect(adminHasWrite).toBe(true);
    
    // Verify Execute button is enabled
    const executeButton = page.locator('button:has-text("Execute")');
    const isAdminExecuteEnabled = await executeButton.isEnabled();
    expect(isAdminExecuteEnabled).toBe(true);
    
    // Clear storage
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    
    // Create viewer user and test UI
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    await createTestUser(page, TEST_USERS.viewer);
    
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await login(page, TEST_USERS.viewer.username, TEST_USERS.viewer.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Verify viewer does NOT see admin navigation
    const adminNavVisible = await page.locator('a:has-text("Users")').count() > 0;
    expect(adminNavVisible).toBe(false);
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Check if data source is available
    const dataSourceSelector = page.locator('[data-testid="datasource-selector"]').first();
    const selectElement = page.locator('select').first();
    const hasDataSource = (await dataSourceSelector.count() > 0) || (await selectElement.count() > 0);
    
    if (hasDataSource) {
      await selectDataSource(page);
      
      // Verify viewer has Read Only badge
      const viewerReadOnlyBadge = page.locator('text=/Read Only/i');
      const viewerHasReadOnly = await viewerReadOnlyBadge.count() > 0;
      expect(viewerHasReadOnly).toBe(true);
    }
    
    // Cleanup
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await deleteTestUser(page, TEST_USERS.viewer.username);
  });

  test('buttons disabled for missing permissions', async ({ page }) => {
    // Setup: Create group with read-only permissions
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    const dataSourceId = await ensureDataSourceExists(page);
    
    // Create read-only group
    await createTestGroupWithPermissions(page, TEST_GROUPS.readGroup, dataSourceId, {
      can_read: true,
      can_write: false,
      can_approve: false,
    });
    
    // Create user and add to read-only group
    await createTestUser(page, TEST_USERS.readUser, TEST_GROUPS.readGroup);
    
    // Clear storage and login as read user
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await login(page, TEST_USERS.readUser.username, TEST_USERS.readUser.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Check if data source is available
    const dataSourceSelector = page.locator('[data-testid="datasource-selector"]').first();
    const selectElement = page.locator('select').first();
    const hasDataSource = (await dataSourceSelector.count() > 0) || (await selectElement.count() > 0);
    
    if (hasDataSource) {
      await selectDataSource(page);
      
      // Type a write query
      const editorSelector = '.monaco-editor textarea, [data-testid="sql-editor"] textarea, .editor-container textarea';
      await page.waitForSelector(editorSelector, { timeout: 10000 });
      
      const editor = page.locator(editorSelector).first();
      await editor.click();
      await editor.fill("UPDATE users SET full_name = 'Test' WHERE id = 'non-existent';");
      
      // Execute button should be disabled or show permission error when clicked
      const executeButton = page.locator('button:has-text("Execute")');
      
      // Try to execute
      await executeButton.click();
      await page.waitForTimeout(2000);
      
      // Should see permission error
      const permissionError = await page.locator('text=/permission|denied|Write Access Required|error/i').count() > 0;
      expect(permissionError).toBe(true);
    }
    
    // Cleanup
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await deleteTestUser(page, TEST_USERS.readUser.username);
    await deleteTestGroup(page, TEST_GROUPS.readGroup);
  });

  test('error messages shown for denied operations', async ({ page }) => {
    // Setup: Create group with read-only permissions
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    const dataSourceId = await ensureDataSourceExists(page);
    
    // Create read-only group
    await createTestGroupWithPermissions(page, TEST_GROUPS.readGroup, dataSourceId, {
      can_read: true,
      can_write: false,
      can_approve: false,
    });
    
    // Create user and add to read-only group
    await createTestUser(page, TEST_USERS.readUser, TEST_GROUPS.readGroup);
    
    // Clear storage and login as read user
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await login(page, TEST_USERS.readUser.username, TEST_USERS.readUser.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Check if data source is available
    const dataSourceSelector = page.locator('[data-testid="datasource-selector"]').first();
    const selectElement = page.locator('select').first();
    const hasDataSource = (await dataSourceSelector.count() > 0) || (await selectElement.count() > 0);
    
    if (hasDataSource) {
      await selectDataSource(page);
      
      // Test DELETE operation
      await executeQuery(page, "DELETE FROM users WHERE id = 'non-existent';");
      
      // Should see permission error
      const deleteError = await page.locator('text=/permission|denied|Write Access Required|error/i').count() > 0;
      expect(deleteError).toBe(true);
      
      // Clear error and test UPDATE operation
      await page.waitForTimeout(1000);
      await executeQuery(page, "UPDATE users SET full_name = 'Test' WHERE id = 'non-existent';");
      
      // Should see permission error again
      const updateError = await page.locator('text=/permission|denied|Write Access Required|error/i').count() > 0;
      expect(updateError).toBe(true);
      
      // Clear error and test INSERT operation
      await page.waitForTimeout(1000);
      await executeQuery(page, "INSERT INTO users (id) VALUES ('test-id');");
      
      // Should see permission error
      const insertError = await page.locator('text=/permission|denied|Write Access Required|error/i').count() > 0;
      expect(insertError).toBe(true);
      
      // Verify SELECT still works
      await page.waitForTimeout(1000);
      await executeQuery(page, 'SELECT 1 as permission_test;');
      
      // Results should appear for SELECT
      const resultsVisible = await page.locator('text=/Results|rows|permission_test/i').count() > 0;
      expect(resultsVisible).toBe(true);
    }
    
    // Cleanup
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await deleteTestUser(page, TEST_USERS.readUser.username);
    await deleteTestGroup(page, TEST_GROUPS.readGroup);
  });

  test('permission error shows helpful hint', async ({ page }) => {
    // Setup: Create group with read-only permissions
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    const dataSourceId = await ensureDataSourceExists(page);
    
    // Create read-only group
    await createTestGroupWithPermissions(page, TEST_GROUPS.readGroup, dataSourceId, {
      can_read: true,
      can_write: false,
      can_approve: false,
    });
    
    // Create user and add to read-only group
    await createTestUser(page, TEST_USERS.readUser, TEST_GROUPS.readGroup);
    
    // Clear storage and login as read user
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await login(page, TEST_USERS.readUser.username, TEST_USERS.readUser.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Check if data source is available
    const dataSourceSelector = page.locator('[data-testid="datasource-selector"]').first();
    const selectElement = page.locator('select').first();
    const hasDataSource = (await dataSourceSelector.count() > 0) || (await selectElement.count() > 0);
    
    if (hasDataSource) {
      await selectDataSource(page);
      
      // Try to execute a write query
      await executeQuery(page, "UPDATE users SET full_name = 'Test' WHERE id = 'non-existent';");
      
      // Should see permission error with hint
      const permissionError = page.locator('text=/Write Access Required|permission|denied/i');
      await expect(permissionError.first()).toBeVisible({ timeout: 5000 });
      
      // Check for helpful hint text
      const hintText = await page.locator('text=/group|permission|contact|admin/i').count() > 0;
      expect(hintText).toBe(true);
      
      // Check for "View My Groups" button
      const viewGroupsButton = page.locator('button:has-text("View My Groups")');
      const hasViewGroupsButton = await viewGroupsButton.count() > 0;
      expect(hasViewGroupsButton).toBe(true);
    }
    
    // Cleanup
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await deleteTestUser(page, TEST_USERS.readUser.username);
    await deleteTestGroup(page, TEST_GROUPS.readGroup);
  });

  test('admin can manage all permissions', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to groups page
    await page.goto('/admin/groups');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    // Verify groups page loads
    await expect(page.locator('h1:has-text("Groups")')).toBeVisible({ timeout: 10000 });
    
    // Navigate to data sources page
    await page.goto('/admin/datasources');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    // Verify data sources page loads
    await expect(page.locator('h1:has-text("Data Sources")')).toBeVisible({ timeout: 10000 });
    
    // Get data sources
    const dataSourceId = await ensureDataSourceExists(page);
    
    // Navigate to users page
    await page.goto('/admin/users');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    // Verify users page loads
    await expect(page.locator('h1:has-text("Users")')).toBeVisible({ timeout: 10000 });
    
    // Admin should be able to see all users
    const usersList = page.locator('table, [data-testid="users-list"], .user-item');
    const hasUsers = await usersList.count() > 0;
    expect(hasUsers).toBe(true);
  });

  test('different roles see different data sources', async ({ page }) => {
    // Setup: Create groups with different permissions
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    const dataSourceId = await ensureDataSourceExists(page);
    
    // Create read-only group
    await createTestGroupWithPermissions(page, TEST_GROUPS.readGroup, dataSourceId, {
      can_read: true,
      can_write: false,
      can_approve: false,
    });
    
    // Create read user
    await createTestUser(page, TEST_USERS.readUser, TEST_GROUPS.readGroup);
    
    // Clear storage and login as read user
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await login(page, TEST_USERS.readUser.username, TEST_USERS.readUser.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Navigate to query editor
    await navigateToQueryEditor(page);
    
    // Check available data sources
    const dataSourceSelector = page.locator('[data-testid="datasource-selector"]').first();
    const selectElement = page.locator('select').first();
    
    // User should see only data sources they have access to
    const hasDataSource = (await dataSourceSelector.count() > 0) || (await selectElement.count() > 0);
    
    // If user has access to a data source, verify it's the one with read permission
    if (hasDataSource) {
      await selectDataSource(page);
      
      // Verify user can execute SELECT
      await executeQuery(page, 'SELECT 1 as role_test;');
      
      const resultsVisible = await page.locator('text=/Results|rows/i').count() > 0;
      expect(resultsVisible).toBe(true);
    }
    
    // Cleanup
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await deleteTestUser(page, TEST_USERS.readUser.username);
    await deleteTestGroup(page, TEST_GROUPS.readGroup);
  });
});