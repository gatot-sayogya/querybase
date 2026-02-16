# Contributing to QueryBase

Thank you for your interest in contributing to QueryBase! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Style](#code-style)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## Getting Started

### Prerequisites

- **Go** 1.21 or higher
- **Node.js** 18 or higher
- **Docker** and **Docker Compose**
- **Make** (optional, but recommended)

### Setup Development Environment

1. **Fork the repository** on GitHub

2. **Clone your fork**:

   ```bash
   git clone https://github.com/YOUR_USERNAME/querybase.git
   cd querybase
   ```

3. **Add upstream remote**:

   ```bash
   git remote add upstream https://github.com/ORIGINAL_OWNER/querybase.git
   ```

4. **Start infrastructure**:

   ```bash
   make docker-up
   ```

5. **Run migrations**:

   ```bash
   make migrate-up
   ```

6. **Start the backend**:

   ```bash
   make build-api
   make run-api
   ```

7. **Start the worker** (in a new terminal):

   ```bash
   make build-worker
   make run-worker
   ```

8. **Start the frontend** (in a new terminal):

   ```bash
   cd web
   npm install
   npm run dev
   ```

9. **Access the application**:
   - Frontend: http://localhost:3000
   - API: http://localhost:8080
   - Default admin: `admin@querybase.local` / `admin123`

For detailed setup instructions, see [docs/getting-started/](docs/getting-started/).

## Development Workflow

### Creating a Feature Branch

```bash
# Update your main branch
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/your-feature-name
```

### Making Changes

1. Make your changes in your feature branch
2. Write or update tests as needed
3. Ensure all tests pass
4. Update documentation if needed

### Keeping Your Branch Updated

```bash
# Fetch latest changes from upstream
git fetch upstream

# Rebase your branch on upstream/main
git rebase upstream/main
```

## Code Style

### Go Code

- Follow standard Go conventions
- Use `gofmt` to format code
- Run `golangci-lint` before committing:
  ```bash
  make lint
  ```

**Key conventions:**

- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions small and focused
- Handle errors explicitly

### TypeScript/React Code

- Follow the existing code style
- Use ESLint and Prettier:
  ```bash
  cd web
  npm run lint
  npm run format
  ```

**Key conventions:**

- Use functional components with hooks
- Use TypeScript types (avoid `any`)
- Keep components small and reusable
- Use meaningful prop names

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**

```
feat(api): add password reset endpoint

Implement admin-initiated password reset functionality.
Includes validation and security checks.

Closes #123
```

```
fix(frontend): resolve query editor autocomplete issue

The autocomplete was not showing table names correctly.
Fixed by updating the schema fetching logic.
```

## Testing

### Backend Tests

Run all Go tests:

```bash
make test
```

Run specific package tests:

```bash
go test ./internal/api/handlers -v
```

Run with coverage:

```bash
make test-coverage
```

### Frontend Tests

```bash
cd web
npm test
```

### Integration Tests

```bash
make test-integration
```

### Manual Testing

Use the test scripts in `scripts/`:

```bash
./scripts/test-backend-api.sh
```

## Submitting Changes

### Before Submitting

1. **Run all tests**:

   ```bash
   make test
   cd web && npm test
   ```

2. **Run linters**:

   ```bash
   make lint
   cd web && npm run lint
   ```

3. **Update documentation** if you've changed:
   - API endpoints
   - Configuration options
   - User-facing features

4. **Add tests** for new features or bug fixes

### Pull Request Process

1. **Push your changes** to your fork:

   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a Pull Request** on GitHub:
   - Use a clear, descriptive title
   - Reference any related issues
   - Describe what changed and why
   - Include screenshots for UI changes

3. **PR Template**:

   ```markdown
   ## Description

   Brief description of changes

   ## Type of Change

   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update

   ## Testing

   - [ ] Tests pass locally
   - [ ] Added new tests
   - [ ] Manual testing completed

   ## Checklist

   - [ ] Code follows style guidelines
   - [ ] Documentation updated
   - [ ] No new warnings
   ```

4. **Address review feedback**:
   - Make requested changes
   - Push updates to your branch
   - Respond to comments

5. **Squash commits** if requested:
   ```bash
   git rebase -i HEAD~n  # n = number of commits
   ```

## Reporting Issues

### Bug Reports

Use the bug report template and include:

- **Description**: Clear description of the bug
- **Steps to reproduce**: Detailed steps
- **Expected behavior**: What should happen
- **Actual behavior**: What actually happens
- **Environment**: OS, Go version, Node version
- **Screenshots**: If applicable
- **Logs**: Relevant error messages

### Feature Requests

Use the feature request template and include:

- **Problem**: What problem does this solve?
- **Proposed solution**: How should it work?
- **Alternatives**: Other solutions considered
- **Additional context**: Any other relevant information

## Project Structure

```
querybase/
├── cmd/              # Application entry points
├── internal/         # Private application code
│   ├── api/         # HTTP handlers, routes, middleware
│   ├── auth/        # Authentication & authorization
│   ├── models/      # Data models
│   ├── service/     # Business logic
│   └── ...
├── web/             # Frontend Next.js application
├── docs/            # Documentation
├── migrations/      # Database migrations
├── scripts/         # Utility scripts
└── tests/           # Integration tests
```

For detailed architecture, see [docs/architecture/](docs/architecture/).

## Getting Help

- **Documentation**: Check [docs/](docs/) directory
- **Issues**: Search existing issues on GitHub
- **Discussions**: Use GitHub Discussions for questions

## License

By contributing to QueryBase, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to QueryBase! 🚀
