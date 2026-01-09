# Go Boilerplate - Claude Code Guidelines

## Project Overview

Clean Architecture Go boilerplate with Fiber, GORM, PostgreSQL, Redis, and Asynq workers.

## Architecture

```
cmd/app/          - Application entrypoint
config/           - Configuration loading (Viper)
internal/
  app/            - DI container & bootstrap
  dto/            - Data Transfer Objects (request/response)
  entity/         - Domain entities (GORM models)
  handlers/http/  - HTTP handlers (Fiber)
  repo/           - Repository interfaces & implementations
  usecase/        - Business logic
  worker/         - Asynq background workers
pkg/              - Reusable packages
integration-test/ - Docker-based E2E tests
```

## Key Patterns

### Layer Flow
```
Handler → UseCase → Repository → Entity/External API
```

### DTO Structure (3-layer)
- `dto/*/request.go` - Request DTOs with JSON + validation tags
- `dto/*/response.go` - Response DTOs with JSON tags and `New*` helpers
- No output.go layer - usecases return response DTOs directly

### Error Handling
- `repo.ErrNotFound` - Standard sentinel error for "not found"
- Usecases check `errors.Is(err, repo.ErrNotFound)` and return domain errors
- Handlers map domain errors to HTTP status codes

### Dependency Injection
- `internal/app/app.go` uses interface types (DIP)
- Allows swapping implementations without changing DI container

## Commands

```bash
# Development
make run              # Run with hot reload (air)
make build            # Build binary

# Testing
make test             # Run unit tests
make integration-test # Run Docker-based E2E tests
make lint             # Run golangci-lint

# Code Generation
go generate ./...     # Regenerate mocks (requires mockgen)
make swag             # Regenerate Swagger docs

# Docker
make compose-up       # Start all services
make compose-down     # Stop all services
```

## Configuration

Environment variables (see `config/config.go`):
- `APP_ENV` - Environment (development/production)
- `HTTP_PORT` - HTTP server port
- `POSTGRES_*` - Database connection
- `REDIS_*` - Redis connection
- `JWT_SECRET_KEY` - JWT signing key (required in production)

## Testing

- Unit tests: Mock repositories with `mockgen`
- Integration tests: Docker Compose with testcontainers
- Handler tests: Use Fiber's built-in `app.Test()`

Generate mocks:
```bash
go install go.uber.org/mock/mockgen@latest
go generate ./...
```

## Code Quality Requirements

**IMPORTANT**: Before running, building, adding new features, or fixing bugs, you MUST:

1. **Run tests**: `make test` - Ensure all unit tests pass
2. **Run linter**: `make lint` - Ensure no linting errors
3. **Run vulnerability check**: `make vuln` - Check for security vulnerabilities
4. **Or run all checks**: `make check-all` - Runs fmt, lint, vuln, and test together

This ensures code quality and catches issues early. Never skip these checks.

## SOLID Architecture - File Organization (IMPORTANT)

**This project follows strict SOLID principles. Each method gets its own file.**

### Usecase Structure (1 file per method + 1 test file per method)
```
internal/usecase/<feature>/
├── <feature>.go           # UseCase struct + New() + shared helpers ONLY
├── errors.go              # Error definitions (ErrNotFound, ErrInvalid, etc.)
├── <method1>.go           # Single method implementation
├── <method1>_test.go      # Tests for that method
├── <method2>.go           # Single method implementation
├── <method2>_test.go      # Tests for that method
├── ...
└── mocks_test.go          # Mock implementations for testing
```

**Example - Auth Usecase:**
```
internal/usecase/auth/
├── auth.go                # UseCase struct + New() + generateTokens helper
├── errors.go              # ErrInvalidCredentials, ErrUserExists, etc.
├── login.go + login_test.go
├── logout.go + logout_test.go
├── register.go + register_test.go
├── refresh.go + refresh_test.go
├── get_current_user.go + get_current_user_test.go
└── mocks_*.go
```

**Example - Media Usecase:**
```
internal/usecase/media/
├── media.go               # UseCase struct + New() + detectMediaType helper
├── errors.go              # ErrFileTooLarge, ErrInvalidMimeType
├── upload.go + upload_test.go
├── get.go + get_test.go   # GetByID + GetByAttachable
├── url.go + url_test.go   # GetURL + GetPresignedUploadURL
├── delete.go + delete_test.go
└── mocks_test.go
```

### Handler Structure (1 file per handler method)
```
internal/handlers/http/v1/<feature>/
├── handler.go             # Handler struct + New() + RegisterRoutes ONLY
├── <method1>.go           # Single handler method with Swagger annotations
├── <method2>.go           # Single handler method with Swagger annotations
├── ...
├── handler_test.go        # All handler tests (can be in one file)
└── mocks_test.go          # Mock implementations
```

**Example - Auth Handler:**
```
internal/handlers/http/v1/auth/
├── handler.go             # Handler struct + New() + RegisterRoutes
├── login.go               # Login handler
├── logout.go              # Logout handler
├── register.go            # Register handler
├── refresh.go             # Refresh handler
├── me.go                  # GetCurrentUser handler
├── handler_test.go
└── mocks_test.go
```

**Example - Media Handler:**
```
internal/handlers/http/v1/media/
├── handler.go             # Handler struct + New() + RegisterRoutes + parseUint helper
├── upload.go              # Upload handler
├── get_by_id.go           # GetByID handler
├── get_url.go             # GetURL handler
├── get_presigned_url.go   # GetPresignedURL handler
├── delete.go              # Delete handler
├── get_by_attachable.go   # GetByAttachable handler
├── handler_test.go
└── mocks_test.go
```

### Rules
1. **NEVER put multiple methods in one file** (except closely related like GetByID + GetByAttachable)
2. **Each usecase method MUST have its own test file** (`<method>_test.go`)
3. **main struct file** (`<feature>.go` or `handler.go`) contains ONLY: struct, constructor, shared helpers
4. **errors.go** contains all error definitions for the package
5. **Follow existing patterns** - look at `auth` package as reference

## Adding New Features

1. **Entity**: Add GORM model in `internal/entity/`
2. **Repository**: Add interface in `internal/repo/contracts.go`, implement in `internal/repo/persistent/`
3. **UseCase**: Add interface in `internal/usecase/contracts.go`, implement in `internal/usecase/*/` **(follow SOLID file structure above)**
4. **DTOs**: Add request/response in `internal/dto/*/`
5. **Handler**: Add handler in `internal/handlers/http/v1/*/` **(follow SOLID file structure above)**
6. **Routes**: Register in `internal/handlers/http/router.go`
7. **DI**: Wire up in `internal/app/app.go`
8. **Tests**: Add tests following the pattern `<method>_test.go` alongside implementation

## Conventions

- **Migration files**: Use sequential format `000XXX_<descriptive_name>.up.sql` / `.down.sql`
  - Example: `000001_create_history.up.sql`, `000006_create_media.down.sql`
  - NEVER use timestamp format like `20210221023242_name.sql`
- Table names: Plural form (`users`, `translations`, `refresh_tokens`)
- Use `goccy/go-json` via `pkg/json` for consistent JSON handling
- Health endpoints: `/healthz` (liveness), `/readyz` (readiness with DB ping)
- Swagger annotations on handlers for API documentation
