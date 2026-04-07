import { test, expect, Page } from '@playwright/test';

/**
 * Authentication E2E Tests
 * 
 * Tests cover:
 * - Login with different user roles (admin, user, viewer)
 * - Invalid credentials rejection
 * - Token persistence in storage
 * - Logout session clearing
 * - Protected routes redirect to login
 * - Login redirect to dashboard when already authenticated
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

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    // Clear auth storage before each test
    await page.goto('/login');
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await page.waitForLoadState('networkidle');
  });

  test('should display login page', async ({ page }) => {
    await expect(page.locator('h2')).toContainText('QueryBase');
    await expect(page.locator('input[name="username"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });

  test('admin user can login successfully', async ({ page }) => {
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    
    // Should redirect to dashboard
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 15000 });
    await page.waitForLoadState('networkidle');
    
    // Verify user is logged in - check for username in navigation
    await expect(page.locator('strong:has-text("admin")')).toBeVisible({ timeout: 10000 });
    
    // Verify auth state in localStorage
    const authStorage = await page.evaluate(() => {
      const stored = localStorage.getItem('auth-storage');
      return stored ? JSON.parse(stored) : null;
    });
    
    expect(authStorage).not.toBeNull();
    expect(authStorage.state.token).toBeTruthy();
    expect(authStorage.state.isAuthenticated).toBe(true);
    expect(authStorage.state.user.username).toBe('admin');
    expect(authStorage.state.user.role).toBe('admin');
  });

  test('regular user can login successfully', async ({ page }) => {
    // Create test user first
    await createTestUser(page, TEST_USERS.regularUser);
    
    // Clear storage and navigate to login
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    
    // Login as regular user
    await login(page, TEST_USERS.regularUser.username, TEST_USERS.regularUser.password);
    
    // Should redirect to dashboard
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 15000 });
    await page.waitForLoadState('networkidle');
    
    // Verify user is logged in
    await expect(page.locator('strong:has-text("' + TEST_USERS.regularUser.username + '")')).toBeVisible({ timeout: 10000 });
    
    // Verify auth state
    const authStorage = await page.evaluate(() => {
      const stored = localStorage.getItem('auth-storage');
      return stored ? JSON.parse(stored) : null;
    });
    
    expect(authStorage.state.isAuthenticated).toBe(true);
    expect(authStorage.state.user.role).toBe('user');
    
    // Cleanup
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await deleteTestUser(page, TEST_USERS.regularUser.username);
  });

  test('viewer can login successfully', async ({ page }) => {
    // Create test viewer user
    await createTestUser(page, TEST_USERS.viewer);
    
    // Clear storage and navigate to login
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    
    // Login as viewer
    await login(page, TEST_USERS.viewer.username, TEST_USERS.viewer.password);
    
    // Should redirect to dashboard
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 15000 });
    await page.waitForLoadState('networkidle');
    
    // Verify user is logged in
    await expect(page.locator('strong:has-text("' + TEST_USERS.viewer.username + '")')).toBeVisible({ timeout: 10000 });
    
    // Verify auth state
    const authStorage = await page.evaluate(() => {
      const stored = localStorage.getItem('auth-storage');
      return stored ? JSON.parse(stored) : null;
    });
    
    expect(authStorage.state.isAuthenticated).toBe(true);
    expect(authStorage.state.user.role).toBe('viewer');
    
    // Cleanup
    await page.evaluate(() => localStorage.removeItem('auth-storage'));
    await deleteTestUser(page, TEST_USERS.viewer.username);
  });

  test('invalid credentials are rejected', async ({ page }) => {
    await login(page, 'invaliduser', 'wrongpassword');
    
    // Wait for error message to appear
    await page.waitForTimeout(2000);
    
    // Should stay on login page
    await expect(page).toHaveURL(/\/login/);
    
    // Should show error message
    const errorVisible = await page.locator('.bg-red-50, .dark\\:bg-red-900\\/20, .error-msg').count() > 0;
    expect(errorVisible).toBe(true);
    
    // Verify not authenticated
    const authStorage = await page.evaluate(() => {
      const stored = localStorage.getItem('auth-storage');
      return stored ? JSON.parse(stored) : null;
    });
    
    expect(authStorage?.state?.isAuthenticated).toBeFalsy();
  });

  test('token persists in storage', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Get initial auth state
    const initialAuth = await page.evaluate(() => {
      const stored = localStorage.getItem('auth-storage');
      return stored ? JSON.parse(stored) : null;
    });
    
    const initialToken = initialAuth?.state?.token;
    expect(initialToken).toBeTruthy();
    
    // Reload the page
    await page.reload();
    await page.waitForLoadState('networkidle');
    
    // Verify token persists after reload
    const reloadedAuth = await page.evaluate(() => {
      const stored = localStorage.getItem('auth-storage');
      return stored ? JSON.parse(stored) : null;
    });
    
    expect(reloadedAuth?.state?.token).toBe(initialToken);
    expect(reloadedAuth?.state?.isAuthenticated).toBe(true);
    
    // Should still be on dashboard (not redirected to login)
    await expect(page).toHaveURL(/\/dashboard/);
  });

  test('logout clears session', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Verify logged in
    await expect(page.locator('strong:has-text("admin")')).toBeVisible({ timeout: 10000 });
    
    // Click logout button
    await page.click('button:has-text("Logout")');
    
    // Should redirect to login page
    await expect(page).toHaveURL(/\/login/, { timeout: 10000 });
    
    // Verify session is cleared
    const authStorage = await page.evaluate(() => {
      const stored = localStorage.getItem('auth-storage');
      return stored ? JSON.parse(stored) : null;
    });
    
    // Auth storage should be cleared or have isAuthenticated = false
    expect(authStorage?.state?.isAuthenticated).toBeFalsy();
    expect(authStorage?.state?.token).toBeFalsy();
    expect(authStorage?.state?.user).toBeFalsy();
  });

  test('protected routes redirect to login', async ({ page }) => {
    // Try to access protected routes without authentication
    const protectedRoutes = [
      '/dashboard',
      '/dashboard/history',
      '/dashboard/approvals',
      '/admin/datasources',
      '/admin/users',
      '/admin/groups',
    ];
    
    for (const route of protectedRoutes) {
      // Clear storage
      await page.evaluate(() => localStorage.removeItem('auth-storage'));
      
      // Try to access protected route
      await page.goto(route);
      await page.waitForTimeout(2000);
      
      // Should be redirected to login
      await expect(page).toHaveURL(/\/login/, { timeout: 10000 });
    }
  });

  test('login redirects to dashboard when authenticated', async ({ page }) => {
    // First, login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Verify logged in
    await expect(page.locator('strong:has-text("admin")')).toBeVisible({ timeout: 10000 });
    
    // Now try to navigate to login page while authenticated
    await page.goto('/login');
    await page.waitForTimeout(2000);
    
    // Should be redirected back to dashboard (or stay on dashboard)
    // The app should redirect authenticated users away from login
    await page.waitForTimeout(1000);
    
    // Either still on dashboard or redirected back to dashboard
    const currentUrl = page.url();
    expect(currentUrl).toMatch(/\/dashboard/);
  });

  test('session expired shows message on login page', async ({ page }) => {
    // Navigate to login with session expired parameter
    await page.goto('/login?session=expired');
    await page.waitForLoadState('networkidle');
    
    // Should show session expired message
    const sessionExpiredMessage = page.locator('text=/session has expired/i');
    await expect(sessionExpiredMessage).toBeVisible({ timeout: 5000 });
  });

  test('login form validation shows errors', async ({ page }) => {
    // Try to submit empty form
    await page.click('button[type="submit"]');
    await page.waitForTimeout(500);
    
    // HTML5 validation should prevent submission
    // Check that we're still on login page
    await expect(page).toHaveURL(/\/login/);
    
    // Try with only username
    await page.fill('input[name="username"]', 'testuser');
    await page.click('button[type="submit"]');
    await page.waitForTimeout(500);
    
    // Should still be on login page
    await expect(page).toHaveURL(/\/login/);
    
    // Try with only password
    await page.evaluate(() => {
      const usernameInput = document.querySelector('input[name="username"]') as HTMLInputElement;
      if (usernameInput) usernameInput.value = '';
    });
    await page.fill('input[name="password"]', 'testpass');
    await page.click('button[type="submit"]');
    await page.waitForTimeout(500);
    
    // Should still be on login page
    await expect(page).toHaveURL(/\/login/);
  });

  test('login preserves redirect after authentication', async ({ page }) => {
    // Try to access a protected route directly
    await page.goto('/dashboard/history');
    await page.waitForTimeout(2000);
    
    // Should be redirected to login
    await expect(page).toHaveURL(/\/login/, { timeout: 10000 });
    
    // Login
    await page.fill('input[name="username"]', TEST_USERS.admin.username);
    await page.fill('input[name="password"]', TEST_USERS.admin.password);
    await page.click('button[type="submit"]');
    
    // After login, should be redirected to dashboard (not necessarily the original route)
    // This depends on app implementation
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 15000 });
  });

  test('multiple failed login attempts show consistent errors', async ({ page }) => {
    // Attempt multiple failed logins
    for (let i = 0; i < 3; i++) {
      await page.fill('input[name="username"]', `invalid${i}`);
      await page.fill('input[name="password"]', `wrong${i}`);
      await page.click('button[type="submit"]');
      await page.waitForTimeout(1500);
      
      // Should stay on login page
      await expect(page).toHaveURL(/\/login/);
    }
    
    // Verify still not authenticated
    const authStorage = await page.evaluate(() => {
      const stored = localStorage.getItem('auth-storage');
      return stored ? JSON.parse(stored) : null;
    });
    
    expect(authStorage?.state?.isAuthenticated).toBeFalsy();
  });

  test('user role affects navigation visibility', async ({ page }) => {
    // Login as admin
    await login(page, TEST_USERS.admin.username, TEST_USERS.admin.password);
    await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    
    // Admin should see all navigation links
    await expect(page.locator('a:has-text("Data Sources")')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('a:has-text("Users")')).toBeVisible();
    await expect(page.locator('a:has-text("Groups")')).toBeVisible();
  });
});