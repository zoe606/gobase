# Code Patterns & Conventions

This document covers the coding patterns and conventions used throughout the project.

## HTTP Handler APIs

Use the API for the engine recorded in `.gobase.json`.

All engines use the generic envelope `response.Response[T]` for success JSON responses. `T` is the type of `data` in the response. Success helpers such as `response.OK[T]` construct and write this envelope. Error responses use `response.ErrorResponse`.

| Engine | Handler arguments | Response example |
|--------|-------------------|------------------|
| Gin | `c *gin.Context` | `response.OK(c.Writer, c.Request, result)` |
| stdlib | `w http.ResponseWriter, r *http.Request` | `response.OK(w, r, result)` |
| Fiber | `c *fiber.Ctx` | `return response.OK(c, result)` |

The helper's Go return value is separate from the JSON response. Gin and stdlib JSON helpers write to HTTP and return no Go value. Fiber JSON helpers write the same response format and return an `error` from writing the response. Return from the handler after an error response. Request parsing and middleware must use the selected engine's API.

## Error Handling

Errors flow through three layers, each with a distinct responsibility:

```go
// Repository layer — return sentinel errors
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, repo.ErrNotFound
}

// UseCase layer — return domain errors (don't leak implementation details)
if errors.Is(err, repo.ErrNotFound) {
    return nil, ErrInvalidCredentials  // Don't expose "user not found"
}

// Fiber handler
if errors.Is(err, auth.ErrInvalidCredentials) {
    return response.Unauthorized(c, "Invalid email or password")
}

// Gin handler
if errors.Is(err, auth.ErrInvalidCredentials) {
    response.Unauthorized(c.Writer, c.Request, "Invalid email or password")
    return
}

// Standard HTTP handler
if errors.Is(err, auth.ErrInvalidCredentials) {
    response.Unauthorized(w, r, "Invalid email or password")
    return
}
```

## Validation

DTOs use struct tags for validation. Handlers validate before calling usecases.

```go
// DTO with validation tags
type RegisterRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
}

// Fiber handler validates input
if err := h.v.Struct(req); err != nil {
    return response.ValidationError(c, v1.ParseValidationErrors(err))
}
```

Gin and stdlib use the same validator. Pass `c.Writer, c.Request` or `w, r` to their response helpers, then return from the handler.

## Transactions

Use `txHelper.RunInTx` for multi-step operations. All operations within the callback share the same transaction — automatic rollback on error, commit on success.

```go
err := txHelper.RunInTx(ctx, func(txCtx context.Context) error {
    if err := userRepo.Create(txCtx, user); err != nil {
        return err
    }
    if err := profileRepo.Create(txCtx, profile); err != nil {
        return err  // Rolls back user creation too
    }
    return nil  // Commits
})
```

## Response Format

All responses follow a consistent JSON structure:

```json
// Success
{
  "success": true,
  "data": { "id": 1, "email": "user@example.com" },
  "request_id": "abc-123"
}

// Error
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Article not found"
  },
  "request_id": "abc-123"
}

// Validation error
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "email": "must be a valid email",
      "password": "must be at least 8 characters"
    }
  },
  "request_id": "abc-123"
}

// Paginated list
{
  "success": true,
  "data": [{ "id": 1 }, { "id": 2 }],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

## DTO Structure (3-layer)

Each feature has its own DTO package:

- `dto/<feature>/request.go` — Request DTOs with JSON + validation tags
- `dto/<feature>/response.go` — Response DTOs with JSON tags and `New*` constructors

Package naming uses the feature name as prefix to avoid import conflicts: `package articledto` (not `package article`).

## SOLID File Organization

Each usecase method gets its own file with a corresponding test file:

```
internal/usecase/auth/
├── auth.go                # Struct + constructor + shared helpers ONLY
├── errors.go              # Domain error definitions
├── login.go               # Login method
├── login_test.go          # Login tests
├── register.go            # Register method
├── register_test.go       # Register tests
├── logout.go
├── refresh.go
├── get_current_user.go
└── mocks_test.go          # Mock implementations
```

Handler packages follow the same pattern:

```
internal/handlers/http/v1/auth/
├── handler.go             # Struct + constructor + RegisterRoutes ONLY
├── login.go               # POST /login handler
├── register.go            # POST /register handler
├── logout.go              # POST /logout handler
├── refresh.go             # POST /refresh handler
├── me.go                  # GET /me handler
├── handler_test.go        # All handler tests
└── mocks_test.go
```

### Rules

1. Never put multiple methods in one file (except closely related ones like `GetByID` + `GetByAttachable`)
2. Each usecase method must have its own test file
3. Main struct file contains only: struct, constructor, shared helpers
4. `errors.go` contains all error definitions for the package

## Reusable Packages (`pkg/`)

| Package | Purpose |
|---------|---------|
| `apperror` | Standardized error codes |
| `response` | HTTP response helpers |
| `pagination` | Query pagination |
| `tx` | Transaction management |
| `cache` | Caching abstraction |
| `jwt` | JWT token service |
| `hasher` | Password hashing (bcrypt) |
| `logger` | Structured logging (Zap) |
| `resilience` | Circuit breaker |
| `asynctx` | Async job context |
| `audit` | Audit logging |
| `asynq` | Task queue client/server |
| `postgres` | Database connection |
| `redis` | Redis client |
| `httpserver` | HTTP server lifecycle |
| `json` | Fast JSON (goccy/go-json) |
| `codegen` | Code scaffolding |

## Background Workers

Workers use [Asynq](https://github.com/hibiken/asynq) for reliable background job processing.

### Creating a Task

1. Define task type in `internal/worker/tasks/types.go`
2. Create handler in `internal/worker/tasks/`
3. Register in `internal/worker/worker.go`

```go
// 1. Define task type
const TypeMyTask = "my:task"

// 2. Create handler
func HandleMyTask(ctx context.Context, t *asynq.Task) error {
    var payload MyTaskPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return err
    }
    // Process task...
    return nil
}

// 3. Register handler
mux.HandleFunc(tasks.TypeMyTask, tasks.HandleMyTask)
```

### Running the Worker

```bash
make run-worker        # Development
make docker-dev        # Full stack with Docker
```
