# 🎭 E2E Testing with Playwright - Quick Start Guide

## ✅ Current Status

**Authentication Tests: 4/4 PASSING** ✅

All authentication tests are now working perfectly:
- ✅ Display login page
- ✅ Show error with invalid credentials
- ✅ Login successfully with valid credentials
- ✅ Redirect to login if not authenticated

## 🚀 How to Run Tests in VSCode

### Option 1: Playwright UI Mode (BEST for debugging)

Open VSCode terminal and run:

```bash
# Make sure you're in the web directory
cd web

# Run with beautiful GUI interface
PLAYWRIGHT_REUSE_EXISTING_SERVER=1 npx playwright test --ui
```

This opens a visual interface where you can:
- See all tests listed
- Click any test to run it
- Watch browser actions in real-time
- Inspect elements
- Debug step-by-step

### Option 2: Run Single Test File

```bash
cd web
PLAYWRIGHT_REUSE_EXISTING_SERVER=1 npx playwright test e2e/auth.spec.ts
```

### Option 3: Run Specific Test

```bash
cd web
PLAYWRIGHT_REUSE_EXISTING_SERVER=1 npx playwright test e2e/auth.spec.ts:30
```

### Option 4: Run Tests in VSCode Extension

1. Install Playwright extension (already installed!)
2. Open any test file in `web/e2e/`
3. Click green ▶ buttons next to tests
4. See results inline

## 🐛 How to Debug Issues

### Step 1: Run in UI Mode

```bash
cd web
PLAYWRIGHT_REUSE_EXISTING_SERVER=1 npx playwright test --ui
```

### Step 2: Pick Locators (Get Perfect Selectors)

Open Playwright Inspector:

```bash
cd web
npx playwright codegen http://localhost:3001
```

1. Browser opens with your app
2. Click "Pick Locator" (crosshair icon)
3. Click any element on the page
4. Perfect selector generates automatically!

### Step 3: View Traces (Post-Mortem Debugging)

When a test fails, Playwright saves a trace:

```bash
npx playwright show-trace test-results/some-test/trace.zip
```

Shows:
- Complete timeline
- Screenshots at each step
- Network requests
- Console logs
- DOM snapshots

## 📝 Test Structure

```
web/e2e/
├── auth.spec.ts       ✅ 4/4 passing
│   ├── Display login page
│   ├── Invalid credentials
│   ├── Valid login
│   └── Redirect to login
├── dashboard.spec.ts  ⏳ Needs fixes
│   ├── Display query editor
│   ├── Navigate to history
│   ├── Navigate to approvals
│   ├── User info display
│   └── Logout
└── admin.spec.ts      ⏳ Needs fixes
    ├── Display admin links
    ├── Navigate to data sources
    ├── Navigate to users
    ├── Navigate to groups
    └── Display lists
```

## ⚠️ Known Issues & Fixes Needed

### Issue 1: Dashboard/Admin Tests Navigate to Wrong URL

**Problem**: Tests go to `/` which shows home page instead of `/login`

**Fix Needed**: Update dashboard and admin tests to navigate to `/login` first

### Issue 2: Page Loading Timing

**Problem**: Tests timeout waiting for elements

**Current Fix**: Using `waitForTimeout()` as temporary solution

**Better Fix**: Use proper waits:
```typescript
// ❌ Temporary fix
await page.waitForTimeout(3000);

// ✅ Better - wait for element
await page.waitForSelector('input[name="username"]');
```

### Issue 3: Dev Server Management

**Problem**: Port conflicts when webServer tries to start server

**Current Fix**: Use `PLAYWRIGHT_REUSE_EXISTING_SERVER=1`

**Better Fix**: Update playwright.config.ts:
```typescript
webServer: {
  command: 'npm run dev',
  url: 'http://localhost:3001',
  reuseExistingServer: true,  // Always reuse
  timeout: 120000,
}
```

## 🎯 Quick Fixes to Apply

### Fix 1: Update Playwright Config

Edit `web/playwright.config.ts`, line 45:

```typescript
reuseExistingServer: true,  // Changed from !process.env.CI
```

### Fix 2: Update Dashboard/Admin Tests

Edit tests to navigate to `/login` instead of `/`:

```typescript
// Before
await page.goto('/');

// After
await page.goto('/login');
```

### Fix 3: Add Better Waits

Replace `waitForTimeout` with proper waits:

```typescript
// Before
await page.waitForTimeout(3000);
await expect(page).toHaveURL(/\/admin\/datasources/);

// After
await page.click('a:has-text("Data Sources")');
await expect(page).toHaveURL(/\/admin\/datasources/);
await page.waitForLoadState('domcontentloaded');
```

## 🚀 Recommended Workflow

### 1. Start Dev Server (Once)

```bash
cd web
npm run dev
```

Leave this running in a terminal

### 2. Run Tests in Another Terminal

```bash
cd web
PLAYWRIGHT_REUSE_EXISTING_SERVER=1 npx playwright test --ui
```

### 3. Debug in VSCode

1. Open test file in VSCode
2. Press F5 or click "Run and Debug"
3. Choose "Debug Playwright Tests"
4. Set breakpoints by clicking line numbers
5. Tests pause at breakpoints - inspect everything!

## 📊 Current Test Results

```
✅ Authentication: 4/4 passing (100%)
⏳ Dashboard: 0/5 passing (needs URL fix)
⏳ Admin: 1/7 passing (needs URL fix)
```

**Total: 5/16 tests passing (31%)**

## 💡 Next Steps

1. ✅ **Auth tests working** - No changes needed!
2. ⏳ **Fix dashboard/admin tests** - Change `/` to `/login` in beforeEach
3. ⏳ **Improve selectors** - Use codegen to get perfect selectors
4. ⏳ **Add data-testid** - Make tests more reliable
5. ⏳ **Remove arbitrary waits** - Use proper waitForSelector

## 🎮 Common Commands

```bash
# Run all tests (GUI mode - BEST!)
cd web && PLAYWRIGHT_REUSE_EXISTING_SERVER=1 npx playwright test --ui

# Run all tests (terminal mode)
cd web && PLAYWRIGHT_REUSE_EXISTING_SERVER=1 npx playwright test

# Run single file
cd web && PLAYWRIGHT_REUSE_EXISTING_SERVER=1 npx playwright test e2e/auth.spec.ts

# Run with visible browser
cd web && PLAYWRIGHT_REUSE_EXISTING_SERVER=1 npx playwright test --headed

# Debug step-by-step
cd web && PLAYWRIGHT_REUSE_EXISTING_SERVER=1 npx playwright test --debug

# Generate tests from browser actions
cd web && npx playwright codegen http://localhost:3001

# View test report
cd web && npx playwright show-report

# View trace of failed test
cd web && npx playwright show-trace test-results/some-test/trace.zip
```

## 🎯 Success Criteria

You'll know tests are fully working when:

1. ✅ All 16 tests pass
2. ✅ No arbitrary `waitForTimeout` calls
3. ✅ All tests use `data-testid` or role-based selectors
4. ✅ Tests run reliably every time
5. ✅ Can debug any failing test in under 30 seconds

## 📚 Resources

- **Guide**: [VSPLAYWRIGHT_GUIDE.md](VSPLAYWRIGHT_GUIDE.md)
- **Docs**: [web/E2E_TESTING.md](web/E2E_TESTING.md)
- **API**: https://playwright.dev/docs/api/class-page
- **VSCode**: https://playwright.dev/docs/getting-started-vscode

---

**Remember**: Always use `PLAYWRIGHT_REUSE_EXISTING_SERVER=1` when running tests locally!
