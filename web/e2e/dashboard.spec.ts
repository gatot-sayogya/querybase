import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await page.waitForURL(/\/dashboard/, { timeout: 10000 });
    await page.waitForLoadState('networkidle');
  });

  test('should display query editor', async ({ page }) => {
    await expect(page.locator('textarea[placeholder*="SELECT"]')).toBeVisible();
    await expect(page.locator('button:has-text("Execute Query")')).toBeVisible();
  });

  test('should navigate to query history', async ({ page }) => {
    await page.click('a:has-text("Query History")');
    await expect(page).toHaveURL(/\/dashboard\/history/);
    await expect(page.locator('h1')).toContainText('Query History');
  });

  test('should navigate to approvals', async ({ page }) => {
    await page.click('a:has-text("Approvals")');
    await expect(page).toHaveURL(/\/dashboard\/approvals/);
    await expect(page.locator('h1')).toContainText('Approval Requests');
  });

  test('should display user info in navigation', async ({ page }) => {
    await expect(page.locator('text=admin')).toBeVisible();
    await expect(page.locator('text=(admin)')).toBeVisible();
  });

  test('should logout successfully', async ({ page }) => {
    await page.click('button:has-text("Logout")');
    await expect(page).toHaveURL(/\/login/);
  });
});
