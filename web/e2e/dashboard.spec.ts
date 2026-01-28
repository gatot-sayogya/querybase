import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await page.waitForURL(/\/dashboard/, { timeout: 10000 });
    await page.waitForLoadState('networkidle');
  });

  test('should display query editor', async ({ page }) => {
    // Query editor should be visible on dashboard
    await page.waitForTimeout(2000);
    await expect(page.locator('a:has-text("Query Editor")')).toBeVisible();
  });

  test('should navigate to query history', async ({ page }) => {
    await page.click('a:has-text("Query History")');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/\/dashboard\/history/);
  });

  test('should navigate to approvals', async ({ page }) => {
    await page.click('a:has-text("Approvals")');
    await page.waitForTimeout(2000);
    await expect(page).toHaveURL(/\/dashboard\/approvals/);
  });

  test('should display user info in navigation', async ({ page }) => {
    await page.waitForTimeout(1000);
    // Use more specific selector - look for the <strong> tag with admin username
    await expect(page.locator('strong:has-text("admin")')).toBeVisible();
  });

  test('should logout successfully', async ({ page }) => {
    await page.click('button:has-text("Logout")');
    await expect(page).toHaveURL(/\/login/);
  });
});
