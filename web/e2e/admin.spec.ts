import { test, expect } from '@playwright/test';

test.describe('Admin Features', () => {
  test.beforeEach(async ({ page }) => {
    // Login as admin before each test
    await page.goto('/login');
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
    await page.waitForTimeout(3000);
    await expect(page).toHaveURL(/\/admin\/datasources/);
  });

  test('should navigate to users page', async ({ page }) => {
    await page.click('a:has-text("Users")');
    await page.waitForTimeout(3000);
    await expect(page).toHaveURL(/\/admin\/users/);
  });

  test('should navigate to groups page', async ({ page }) => {
    await page.click('a:has-text("Groups")');
    await page.waitForTimeout(3000);
    await expect(page).toHaveURL(/\/admin\/groups/);
  });

  test('should display data source list', async ({ page }) => {
    await page.goto('/admin/datasources');
    // Wait for page to load - data sources use cards, not tables
    await expect(page.locator('h1:has-text("Data Sources")')).toBeVisible({ timeout: 10000 });
  });

  test('should display user list', async ({ page }) => {
    await page.goto('/admin/users');
    // Wait for page to load
    await expect(page.locator('h1:has-text("Users")')).toBeVisible({ timeout: 10000 });
  });

  test('should display group list', async ({ page }) => {
    await page.goto('/admin/groups');
    // Wait for page to load
    await expect(page.locator('h1:has-text("Groups")')).toBeVisible({ timeout: 10000 });
  });
});
