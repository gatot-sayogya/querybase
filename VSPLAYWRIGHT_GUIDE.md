# 🎭 Playwright E2E Testing - VSCode Integration Guide

## Quick Start (3 Methods)

### ✅ Method 1: Playwright UI Mode (Easiest)

Open your VSCode terminal and run:

```bash
cd web
npm run test:e2e:ui
```

This opens a beautiful GUI where you can:
- 📊 See all tests in a list
- ▶️ Click any test to run it
- 👁️ Watch browser actions in real-time
- 🐛 Debug step-by-step
- 📸 View screenshots and traces

### ✅ Method 2: Playwright Extension (Integrated)

1. Install the extension (already done!)
2. Open any test file in `web/e2e/`
3. Click the green ▶ button next to test names
4. See results inline with ✅ or ❌

### ✅ Method 3: VSCode Debugger (Full Control)

1. Open any test file
2. Press `F5` or click "Run and Debug"
3. Choose "Debug Playwright Tests"
4. Set breakpoints by clicking line numbers
5. Tests pause at breakpoints - inspect everything!

## 🎯 Debugging Workflows

### Workflow 1: Find Element Selectors

Run the **Playwright Inspector** (codegen tool):

```bash
cd web
npx playwright codegen http://localhost:3001
```

This opens:
- A browser with your app
- A Playwright window
- Perform actions in browser → Code auto-generates!
- Click "Pick Locator" to get perfect selectors

### Workflow 2: Debug Failing Tests

When a test fails:

1. **Run with headed mode** (see the browser):
   ```bash
   npx playwright test e2e/auth.spec.ts:41 --headed
   ```

2. **View the trace** (post-mortem analysis):
   ```bash
   npx playwright show-trace test-results/auth-*/trace.zip
   ```
   This shows:
   - Timeline of all actions
   - Screenshots at each step
   - Network requests
   - Console logs
   - Complete DOM snapshot

3. **Debug step-by-step**:
   ```bash
   npx playwright test e2e/auth.spec.ts:41 --debug
   ```
   Opens browser with debugger, pauses on each step

### Workflow 3: Record New Tests

```bash
cd web
npx playwright codegen http://localhost:3001
```

1. Browser opens with your app
2. Playwright Inspector appears
3. Click around, fill forms, etc.
4. Code generates automatically!
5. Copy into your test file

## 🎮 VSCode Integration Features

### Inline Test Runners

After installing the extension, test files show:
- Green ▶ buttons next to each test
- Click to run just that test
- See results inline

### Code Lens Actions

Above each test you'll see:
- ▶ Run this test
- 🐛 Debug this test
- Show in Inspector

### Test Explorer Sidebar

Press `Cmd+Shift+P` → "Testing: Show Test Explorer"
See all tests, run groups, filter by status

### Breakpoints

Set breakpoints by clicking line numbers
When debugging with F5, tests pause at breakpoints
Hover over variables to inspect values

## 📊 Viewing Test Results

### HTML Report

After tests run, view the report:

```bash
npx playwright show-report
```

Opens in browser with:
- All test results
- Screenshots
- Timings
- Error details
- Trace files

### Terminal Output

Tests show colored output:
- ✅ Green checkmarks (passing)
- ❌ Red X marks (failing)
- ⚠️ Warnings

## 🔧 Common Commands

```bash
# Run all tests
npm run test:e2e

# Run with UI (best for debugging!)
npm run test:e2e:ui

# Run in visible browser
npm run test:e2e:headed

# Debug step-by-step
npm run test:e2e:debug

# Run specific file
npx playwright test e2e/auth.spec.ts

# Run specific test by line number
npx playwright test e2e/auth.spec.ts:41

# Run tests matching pattern
npx playwright test -g "login"

# View report
npx playwright show-report

# View trace
npx playwright show-trace test-results/path/to/trace.zip

# Generate tests from browser actions
npx playwright codegen http://localhost:3001

# Record and save test
npx playwright codegen --target=javascript http://localhost:3001 > my-test.spec.ts
```

## 🎨 Best Practices for Reliable Tests

### 1. Use data-testid Attributes

In your React components:

```tsx
<input
  data-testid="username-input"
  name="username"
  type="text"
/>
```

In tests:

```typescript
await page.locator('[data-testid="username-input"]').fill('admin');
```

### 2. Wait Properly

```typescript
// ✅ Good - wait for element
await page.waitForSelector('input[name="username"]');
await page.fill('input[name="username"]', 'admin');

// ❌ Bad - arbitrary timeout
await page.waitForTimeout(5000);
await page.fill('input[name="username"]', 'admin');
```

### 3. Use Role-Based Locators

```typescript
// ✅ Most accessible
page.getByRole('button', { name: 'Submit' })

// ✅ Good fallback
page.locator('button[type="submit"]')

// ❌ Brittle CSS selector
page.locator('.btn.btn-primary')
```

### 4. Check Visibility

```typescript
// ✅ Good - checks element is visible
await expect(page.locator('h1')).toBeVisible();

// ❌ Bad - doesn't check if visible
await page.locator('h1'); // just exists, maybe hidden
```

## 🐛 Debugging Checklist

When tests fail:

1. **Run in UI mode** - See what's happening visually
2. **Check the browser** - Is the page rendering?
3. **View the trace** - See exact moment of failure
4. **Use codegen** - Get correct selectors
5. **Add data-testid** - Make selectors more reliable
6. **Increase timeouts** - For slow pages
7. **Check console** - Any JavaScript errors?
8. **Verify network** - API calls succeeding?

## 📁 File Structure

```
web/
├── e2e/                        # Test files
│   ├── auth.spec.ts           # Authentication tests
│   ├── dashboard.spec.ts      # Dashboard tests
│   └── admin.spec.ts          # Admin feature tests
├── playwright.config.ts        # Playwright configuration
├── playwright-report/          # HTML test reports
└── test-results/              # Screenshots, traces, videos
```

## 🚀 Quick Troubleshooting

### Tests timing out

**Problem**: Elements not found
**Solution**:
- Run in UI mode to see what's rendered
- Use codegen to get correct selectors
- Add data-testid attributes
- Increase timeout in config

### Browser not opening

**Problem**: Headless mode
**Solution**:
- Use `npm run test:e2e:headed`
- Or `--ui` flag for GUI

### Tests passing but app not working

**Problem**: Fake success
**Solution**:
- Check if selectors match actual elements
- Verify element is visible (not hidden)
- Test actual user flows, not just DOM

## 💡 Pro Tips

1. **Start Simple**: Test happy path first
2. **Use UI Mode**: Visual debugging is 10x faster
3. **Record Tests**: Use codegen to scaffold
4. **Check Traces**: Post-mortem analysis is powerful
5. **Data-testid**: Most reliable selector strategy
6. **Run Frequently**: Catch bugs early
7. **CI/CD**: Run on every push (already configured!)

## 🎓 Learning Resources

- Official Docs: https://playwright.dev/docs/intro
- VSCode Guide: https://playwright.dev/docs/getting-started-vscode
- Best Practices: https://playwright.dev/docs/best-practices
- API Reference: https://playwright.dev/docs/api/class-page

Happy Testing! 🎭
