import { test, expect } from '@playwright/test';

test.describe('Admin Features', () => {
  test.beforeEach(async ({ page }) => {
    // Login as admin before each test
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await page.waitForURL(/\/dashboard/, { timeout: 10000 });
    await page.waitForLoadState('networkidle');
  });

  test('should display admin navigation links', async ({ page }) => {
    await expect(page.locator('a:has-text("Data Sources")')).toBeVisible();
    await expect(page.locator('a:has-text("Users")')).toBeVisible();
    await expect(page.locator('a:has-text("Groups")')).toBeVisible();
  });

  test('should navigate to data sources page', async ({ page }) => {
    await page.click('a:has-text("Data Sources")');
    await expect(page).toHaveURL(/\/admin\/datasources/);
    await expect(page.locator('h1')).toContainText('Data Sources');
  });

  test('should navigate to users page', async ({ page }) => {
    await page.click('a:has-text("Users")');
    await expect(page).toHaveURL(/\/admin\/users/);
    await expect(page.locator('h1')).toContainText('Users');
  });

  test('should navigate to groups page', async ({ page }) => {
    await page.click('a:has-text("Groups")');
    await expect(page).toHaveURL(/\/admin\/groups/);
    await expect(page.locator('h1')).toContainText('Groups');
  });

  test('should display data source list', async ({ page }) => {
    await page.goto('/admin/datasources');
    // Wait for data to load
    await page.waitForSelector('table', { timeout: 5000 });
    await expect(page.locator('table')).toBeVisible();
  });

  test('should display user list', async ({ page }) => {
    await page.goto('/admin/users');
    await page.waitForSelector('table', { timeout: 5000 });
    await expect(page.locator('table')).toBeVisible();
  });

  test('should display group list', async ({ page }) => {
    await page.goto('/admin/groups');
    await page.waitForSelector('table', { timeout: 5000 });
    await expect(page.locator('table')).toBeVisible();
  });
});
