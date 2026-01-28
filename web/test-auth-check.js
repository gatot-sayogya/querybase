const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext();
  const page = await context.newPage();
  
  // Login first
  await page.goto('http://localhost:3001/login');
  await page.waitForLoadState('networkidle');
  await page.fill('input[name="username"]', 'admin');
  await page.fill('input[name="password"]', 'admin123');
  await page.click('button[type="submit"]');
  await page.waitForURL(/\/dashboard/, { timeout: 10000 });
  
  console.log('Logged in, cookies:', await context.cookies());
  
  // Now navigate to admin page
  await page.goto('http://localhost:3001/admin/datasources');
  await page.waitForTimeout(3000);
  
  // Check what's on the page
  const h1Text = await page.locator('h1').allTextContents();
  const h2Text = await page.locator('h2').allTextContents();
  console.log('H1 elements:', h1Text);
  console.log('H2 elements:', h2Text);
  console.log('Current URL:', page.url());
  
  await page.screenshot({ path: '/tmp/admin-page-debug.png' });
  console.log('Screenshot saved to /tmp/admin-page-debug.png');
  
  await browser.close();
})();
