# Contributing Guide

Thank you for your interest in contributing to this project!

## Getting Started

1. **Fork and clone** the repository
2. **Install dependencies**: `make deps`
3. **Install development tools**: `make tools`
4. **Install git hooks**: `make install-hooks`
5. **Start services**: `make docker-services`

## Development Workflow

### Before You Code

1. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Ensure git hooks are installed:
   ```bash
   make install-hooks
   ```

### While Coding

- Follow the existing code patterns and architecture
- Add tests for new functionality
- Keep commits focused and atomic
- Write clear commit messages

### Before Committing

The pre-commit hook will automatically run:
- Format check (`gofmt`)
- Syntax validation (`go vet`)
- Fast lint check (`golangci-lint --fast`)

You can also run manually:
```bash
make pre-commit  # fmt + lint + test
```

### Before Pushing

The pre-push hook will automatically run:
- Full build
- All tests
- Full lint
- Vulnerability check

You can also run manually:
```bash
make check-all   # fmt + lint + vuln + test
make ci          # Full CI pipeline locally
```

### Submitting a Pull Request

1. Ensure all checks pass locally
2. Push your branch
3. Open a PR against `main`
4. Fill out the PR template
5. Wait for CI checks to pass
6. Request review

## Code Quality Requirements

All PRs must pass:

| Check | Command | Description |
|-------|---------|-------------|
| Format | `make fmt` | Code formatting |
| Lint | `make lint` | Static analysis |
| Test | `make test` | Unit tests |
| Build | `make build` | Compilation |
| Vuln | `make vuln` | Security scan |

## Project Structure

```
cmd/app/          - Application entrypoint
config/           - Configuration
internal/
  app/            - Dependency injection
  dto/            - Data Transfer Objects
  entity/         - Domain entities
  handlers/http/  - HTTP handlers
  repo/           - Repository layer
  usecase/        - Business logic
  worker/         - Background workers
pkg/              - Reusable packages
migrations/       - Database migrations
```

## Conventions

### File Naming (SOLID Architecture)

Each usecase method gets its own file:
```
internal/usecase/auth/
├── auth.go           # Struct + constructor only
├── errors.go         # Error definitions
├── login.go          # Login method
├── login_test.go     # Login tests
├── register.go       # Register method
├── register_test.go  # Register tests
```

### Commit Messages

Use conventional commits:
```
feat: Add user authentication
fix: Resolve login redirect issue
docs: Update API documentation
refactor: Simplify error handling
test: Add integration tests for auth
chore: Update dependencies
```

### Migration Files

Use sequential numbering:
```
000001_create_users.up.sql
000001_create_users.down.sql
000002_add_roles.up.sql
000002_add_roles.down.sql
```

## Getting Help

- Check existing issues and PRs
- Open an issue for bugs or feature requests
- Join discussions for questions

## Branch Protection

The `main` branch is protected:
- All CI checks must pass
- At least one approval required
- No direct pushes allowed
