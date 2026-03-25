# Task Completion Checklist for QueryBase

## When Completing a Task

### 1. Quality Gates (MANDATORY)

#### Go Backend
```bash
# Run tests
make test

# Format code
make fmt

# Run linter
make lint

# Build to verify no compile errors
make build
```

#### Frontend
```bash
cd web

# Run tests
npm test

# Run linter
npm run lint

# Build to verify no compile errors
npm run build
```

### 2. Issue Tracking Update (beads)

Update issue status using `bd`:

```bash
# Close completed issue
bd close <id> --reason "Done: implemented feature X" --json

# Or update status if more work needed
bd update <id> --status in_progress --notes "Progress update..." --json
```

### 3. Session Completion Workflow

**CRITICAL: Work is NOT complete until `git push` succeeds!**

```bash
# 1. Check what changed
git status

# 2. Stage changes
git add <files>

# 3. Commit with descriptive message
git commit -m "feat: add feature X

- Implemented Y
- Fixed Z
- Added tests"

# 4. Push beads first
bd dolt push

# 5. Push code changes
git pull --rebase
git push

# 6. Verify everything is pushed
git status  # MUST show "up to date with origin"
```

### 4. Create Issues for Remaining Work

If there are follow-up tasks, create issues before ending session:

```bash
bd create "Follow-up: optimize query performance" \
  --description="Need to add indexes for better performance" \
  --type=task \
  --priority=2 \
  --json
```

## Important Reminders

- ✅ Always run tests before committing
- ✅ Always format code with `make fmt` or `npm run lint`
- ✅ Always check `git status` before ending session
- ✅ Always push beads with `bd dolt push`
- ✅ Always push code with `git push`
- ❌ Never skip quality gates
- ❌ Never leave work uncommitted
- ❌ Never end session before pushing

## Beads Workflow Rules

- Use `bd` for ALL task tracking
- Check `bd ready` for available work
- Claim work with `bd update <id> --claim`
- Link discovered work with `discovered-from` dependencies
- Do NOT use markdown TODO lists
- Do NOT duplicate tracking systems
