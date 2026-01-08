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

## Adding New Features

1. **Entity**: Add GORM model in `internal/entity/`
2. **Repository**: Add interface in `internal/repo/contracts.go`, implement in `internal/repo/persistent/`
3. **UseCase**: Add interface in `internal/usecase/contracts.go`, implement in `internal/usecase/*/`
4. **DTOs**: Add request/response in `internal/dto/*/`
5. **Handler**: Add handler in `internal/handlers/http/v1/*/`
6. **Routes**: Register in `internal/handlers/http/router.go`
7. **DI**: Wire up in `internal/app/app.go`
8. **Tests**: Add tests following the pattern `<method>_test.go` alongside implementation

## Conventions

- Table names: Plural form (`users`, `translations`, `refresh_tokens`)
- Use `goccy/go-json` via `pkg/json` for consistent JSON handling
- Health endpoints: `/healthz` (liveness), `/readyz` (readiness with DB ping)
- Swagger annotations on handlers for API documentation
