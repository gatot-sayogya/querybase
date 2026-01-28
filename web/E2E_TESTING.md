# E2E Testing with Playwright

This directory contains end-to-end tests for QueryBase using [Playwright](https://playwright.dev/).

## Setup

Install dependencies:
```bash
npm install
```

Install Playwright browsers:
```bash
npx playwright install chromium
```

## Running Tests

### Run all tests (headless)
```bash
npm run test:e2e
```

### Run tests with UI (watch mode)
```bash
npm run test:e2e:ui
```
This opens the Playwright UI where you can:
- See tests running in real-time
- Click on any step to inspect the page
- View network requests, console logs, and more
- Time-travel through test execution

### Run tests in headed mode (visible browser)
```bash
npm run test:e2e:headed
```

### Debug tests
```bash
npm run test:e2e:debug
```
This opens a browser window and pauses execution, allowing you to:
- Step through tests one command at a time
- Inspect elements
- Execute Playwright commands in the console

### Run specific test file
```bash
npx playwright test auth.spec.ts
```

### Run tests matching a pattern
```bash
npx playwright test -g "login"
```

## Test Structure

```
e2e/
├── auth.spec.ts       # Authentication tests
├── dashboard.spec.ts  # Dashboard and query editor tests
└── admin.spec.ts      # Admin features tests
```

## Writing Tests

### Basic test example

```typescript
import { test, expect } from '@playwright/test';

test('my test', async ({ page }) => {
  await page.goto('/');
  await page.fill('input[name="username"]', 'admin');
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL(/\/dashboard/);
});
```

### Best Practices

1. **Wait for elements**: Use `waitForSelector` or `waitForLoadState` instead of fixed timeouts
2. **Use data-testid**: Add `data-testid` attributes to elements for more reliable selectors
3. **Page Object Model**: For complex flows, extract page logic into reusable classes
4. **Avoid hard-coded waits**: Never use `page.waitForTimeout()` unless absolutely necessary

### Locators

Playwright supports multiple locator strategies:

```typescript
// By CSS selector
page.locator('button.submit')

// By text
page.locator('text=Login')

// By test ID (recommended)
page.locator('data-testid=submit-button')

// By role (accessible)
page.getByRole('button', { name: 'Submit' })

// Combined
page.locator('button').filter({ hasText: 'Submit' })
```

## Viewing Test Results

After running tests, view the HTML report:
```bash
npx playwright show-report
```

## Troubleshooting

### Tests are timing out
- Increase timeout in `playwright.config.ts`
- Check if the dev server is running
- Run with `--debug` flag to see what's happening

### Can't find elements
- Run in headed mode to see the page
- Use Playwright Inspector: `npx playwright codegen http://localhost:3001`
- Check if elements are inside iframes
- Verify element selectors using browser DevTools

### Flaky tests
- Use `waitForLoadState('networkidle')` to ensure page is fully loaded
- Add explicit waits for dynamic content
- Increase retries in config
- Avoid race conditions with proper synchronization

## CI/CD

Tests run automatically on GitHub Actions. See `.github/workflows/e2e.yml` for configuration.

## Resources

- [Playwright Documentation](https://playwright.dev/docs/intro)
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
- [Locator Best Practices](https://playwright.dev/docs/locators)
- [Test Generator](https://playwright.dev/docs/codegen)
