# Phase 1: Security & Correctness — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix all security and correctness bugs so the boilerplate is safe to deploy as a single instance.

**Architecture:** Six independent fixes (except 1.2→1.1 dependency). No new features, no new packages. All changes touch existing files.

**Tech Stack:** Go 1.26, Fiber v2, gomock, Docker

**Spec:** `docs/superpowers/specs/2026-03-13-production-readiness-design.md`

---

## Chunk 1: Article Auth & UserID Security Fixes

### Task 1: Add auth middleware to article write routes (1.2)

This must be done BEFORE Task 2 (UserID removal) so that `middleware.GetUserID(c)` returns a valid value on protected routes.

**Files:**
- Modify: `internal/handlers/http/v1/article/handler.go`
- Modify: `internal/handlers/http/router.go:141`
- Modify: `internal/handlers/http/v1/article/handler_test.go`
- Modify: `internal/handlers/http/v1/article/mocks_test.go`

- [ ] **Step 1: Update article handler struct to accept jwtService**

In `internal/handlers/http/v1/article/handler.go`, add `jwtService` field and update constructor:

```go
package article

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

// Handler handles article endpoints.
type Handler struct {
	articleUC  usecase.Article
	jwtService jwt.Service
	l          logger.Interface
	v          *validator.Validate
}

// New creates a new article handler.
func New(articleUC usecase.Article, jwtService jwt.Service, l logger.Interface) *Handler {
	return &Handler{
		articleUC:  articleUC,
		jwtService: jwtService,
		l:          l,
		v:          validator.New(validator.WithRequiredStructEnabled()),
	}
}

// RegisterRoutes sets up routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	articles := router.Group("/articles")

	// Public routes (no auth required)
	articles.Get("/", h.List)
	articles.Get("/:id", h.GetByID)

	// Protected routes (auth required)
	protected := articles.Group("", middleware.JWTAuth(h.jwtService, h.l))
	protected.Post("/", h.Create)
	protected.Put("/:id", h.Update)
	protected.Delete("/:id", h.Delete)
}
```

- [ ] **Step 2: Update router.go to pass jwtService to article handler**

In `internal/handlers/http/router.go:141`, change:
```go
// Old:
artHandler := articlehandler.New(articleUC, l)
// New:
artHandler := articlehandler.New(articleUC, jwtService, l)
```

- [ ] **Step 3: Update handler test mocks to add MockJWTService**

In `internal/handlers/http/v1/article/mocks_test.go`, add after `MockLogger`:

```go
// MockJWTService is a mock of jwt.Service for testing.
type MockJWTService struct{}

func NewMockJWTService() *MockJWTService {
	return &MockJWTService{}
}

func (m *MockJWTService) GenerateAccessToken(userID uint, email, role string, permissions []string) (string, int64, error) {
	return "mock-access-token", 9999999999, nil
}

func (m *MockJWTService) GenerateRefreshToken() (string, time.Time, error) {
	return "mock-refresh-token", time.Now().Add(24 * time.Hour), nil
}

func (m *MockJWTService) ValidateToken(tokenString string) (*jwt.Claims, error) {
	return &jwt.Claims{
		UserID:      1,
		Email:       "test@example.com",
		Role:        "admin",
		Permissions: []string{"articles:write"},
	}, nil
}

func (m *MockJWTService) GetAccessExpiry() time.Duration  { return 15 * time.Minute }
func (m *MockJWTService) GetRefreshExpiry() time.Duration { return 24 * time.Hour }
```

Add the `jwt` and `time` imports at the top of the file:
```go
import (
	// ... existing imports
	"time"

	"go-boilerplate/pkg/jwt"
)
```

- [ ] **Step 4: Update setupTestApp to pass jwtService**

In `internal/handlers/http/v1/article/handler_test.go`, change `setupTestApp`:

```go
func setupTestApp(t *testing.T, mockArticleUC *MockArticle) *fiber.App {
	t.Helper()

	l := NewMockLogger()
	jwtSvc := NewMockJWTService()
	handler := article.New(mockArticleUC, jwtSvc, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	return app
}
```

- [ ] **Step 5: Add auth token to write test requests and add unauthorized tests**

In `handler_test.go`, add an `addAuth` field to the test table structs for `TestHandler_Create`, `TestHandler_Update`, and `TestHandler_Delete`.

For each test function, add a `addAuth bool` field to the struct and set it to `true` for all existing test cases. Then add an `"unauthorized - no token"` case with `addAuth: false`.

In the test loop, conditionally add the header:

```go
if tt.addAuth {
	req.Header.Set("Authorization", "Bearer mock-access-token")
}
```

**IMPORTANT:** Every existing test case (success, invalid json, invalid id, validation error, usecase error, not found, internal error) must have `addAuth: true`. Without the auth header, those cases would return 401 instead of their expected status codes, since the middleware now runs before the handler.

Add one new case per test function:
```go
{
	name:       "unauthorized - no token",
	body:       validBody, // or id: "1" for Update/Delete
	addAuth:    false,
	setupMock:  func() {},
	wantStatus: fiber.StatusUnauthorized,
},
```

- [ ] **Step 7: Run tests to verify**

Run: `go test -v ./internal/handlers/http/v1/article/...`
Expected: All tests pass, including new unauthorized tests.

- [ ] **Step 8: Commit**

```bash
git add internal/handlers/http/v1/article/ internal/handlers/http/router.go
git commit -m "fix: Add auth middleware to article write routes

POST/PUT/DELETE now require JWT authentication.
GET (list, get by ID) remain public."
```

---

### Task 2: Remove UserID from request DTOs (1.1)

Depends on Task 1 being complete (JWT middleware must be in place).

**Files:**
- Modify: `internal/dto/article/request.go`
- Modify: `internal/handlers/http/v1/article/create.go`
- Modify: `internal/handlers/http/v1/article/update.go`
- Modify: `internal/usecase/article/create.go`
- Modify: `internal/usecase/article/update.go`
- Modify: `internal/usecase/contracts.go:61`
- Modify: `internal/usecase/article/create_test.go`
- Modify: `internal/usecase/article/update_test.go`
- Modify: `internal/handlers/http/v1/article/handler_test.go`
- Modify: `internal/handlers/http/v1/article/mocks_test.go`

- [ ] **Step 1: Update usecase Create signature to accept userID**

In `internal/usecase/contracts.go:61`, change:
```go
// Old:
Create(ctx context.Context, req articledto.CreateRequest) (*articledto.Response, error)
// New:
Create(ctx context.Context, userID uint, req articledto.CreateRequest) (*articledto.Response, error)
```

In `internal/usecase/article/create.go`, change:
```go
// Old:
func (uc *UseCase) Create(ctx context.Context, req articledto.CreateRequest) (*articledto.Response, error) {
// New:
func (uc *UseCase) Create(ctx context.Context, userID uint, req articledto.CreateRequest) (*articledto.Response, error) {
```

- [ ] **Step 2: Remove UserID from DTOs**

In `internal/dto/article/request.go`, remove `UserID` from both structs:

`CreateRequest` — remove line 11: `UserID uint \`json:"user_id" validate:"required"\``

`UpdateRequest` — remove line 24: `UserID *uint \`json:"user_id,omitempty"\``

`ListRequest` — keep `UserID uint \`query:"user_id"\`` (this is a query filter, not a body field — safe).

- [ ] **Step 3: Update Create handler to extract UserID from JWT context**

In `internal/handlers/http/v1/article/create.go`, add middleware import and change:

```go
package article

import (
	"github.com/gofiber/fiber/v2"

	articledto "go-boilerplate/internal/dto/article"
	v1 "go-boilerplate/internal/handlers/http/v1"
	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/pkg/response"
)

// Create godoc
// @Summary     Create article
// @Description Create a new article
// @ID          article-create
// @Tags        articles
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body articledto.CreateRequest true "Create Article request"
// @Success     201 {object} response.Response[articledto.Response]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /articles [post]
func (h *Handler) Create(ctx *fiber.Ctx) error {
	var req articledto.CreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(ctx, v1.ParseValidationErrors(err))
	}

	userID := middleware.GetUserID(ctx)

	result, err := h.articleUC.Create(ctx.UserContext(), userID, req)
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - article - Create")
		return response.InternalError(ctx)
	}

	return response.Created(ctx, result)
}
```

- [ ] **Step 4: Update Update handler to extract UserID from JWT context**

In `internal/handlers/http/v1/article/update.go`, add `middleware` import and extract UserID from context. Even though the current Update usecase doesn't use it, extracting it now establishes the pattern and prevents re-introducing the IDOR if the field is needed later (e.g., ownership checks). Log or store it for future use:

```go
package article

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	articledto "go-boilerplate/internal/dto/article"
	v1 "go-boilerplate/internal/handlers/http/v1"
	"go-boilerplate/internal/handlers/http/middleware"
	articleuc "go-boilerplate/internal/usecase/article"
	"go-boilerplate/pkg/response"
)

// Update godoc
// @Summary     Update article
// @Description Update an existing article
// @ID          article-update
// @Tags        articles
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Article ID"
// @Param       request body articledto.UpdateRequest true "Update Article request"
// @Success     200 {object} response.Response[articledto.Response]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /articles/{id} [put]
func (h *Handler) Update(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		return response.BadRequest(ctx, "INVALID_ID", "Invalid article ID")
	}

	var req articledto.UpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(ctx, v1.ParseValidationErrors(err))
	}

	_ = middleware.GetUserID(ctx) // Extracted for future ownership checks

	result, err := h.articleUC.Update(ctx.UserContext(), uint(id), req)
	if err != nil {
		if errors.Is(err, articleuc.ErrNotFound) {
			return response.NotFound(ctx, "Article not found")
		}
		h.l.Error(err, "handlers - http - v1 - article - Update")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
```

- [ ] **Step 5: Update usecase create implementation**

In `internal/usecase/article/create.go`, use the new `userID` parameter:

```go
func (uc *UseCase) Create(ctx context.Context, userID uint, req articledto.CreateRequest) (*articledto.Response, error) {
	article := &entity.Article{
		UserID: userID,
		// TODO: Map remaining request fields to entity
	}

	if err := uc.articleRepo.Create(ctx, article); err != nil {
		return nil, fmt.Errorf("article - Create - articleRepo.Create: %w", err)
	}

	return articledto.NewResponse(article), nil
}
```

- [ ] **Step 6: Regenerate handler mocks**

The `MockArticle` in `internal/handlers/http/v1/article/mocks_test.go` must match the new `Create` signature.

Update the `Create` mock method:
```go
// Create mocks base method.
func (m *MockArticle) Create(ctx context.Context, userID uint, req articledto.CreateRequest) (*articledto.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, userID, req)
	ret0, _ := ret[0].(*articledto.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockArticleMockRecorder) Create(ctx, userID, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockArticle)(nil).Create), ctx, userID, req)
}
```

- [ ] **Step 7: Update usecase create_test.go**

In `internal/usecase/article/create_test.go`, remove `UserID` from `CreateRequest` and pass `userID` argument:

```go
func TestCreate(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		userID uint
		req    articledto.CreateRequest
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(articleRepo *MockArticleRepo)
		wantErr   error
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				userID: 1,
				req: articledto.CreateRequest{
					Title:   "Test Article",
					Slug:    "test-article",
					Content: "Some content",
					Excerpt: "Some excerpt",
					Status:  "draft",
				},
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "repo error",
			args: args{
				ctx:    context.Background(),
				userID: 1,
				req: articledto.CreateRequest{
					Title:   "Test Article",
					Slug:    "test-article",
					Content: "Some content",
					Excerpt: "Some excerpt",
					Status:  "draft",
				},
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockArticleRepo := NewMockArticleRepo(ctrl)
			tt.setupMock(mockArticleRepo)

			uc := article.New(mockArticleRepo)
			got, err := uc.Create(tt.args.ctx, tt.args.userID, tt.args.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}
```

- [ ] **Step 8: Update handler_test.go for Create**

In `handler_test.go`, update `TestHandler_Create`:
- Remove `"user_id": 1` from `validBody` JSON
- Update mock expectation to match new signature: `Create(gomock.Any(), gomock.Any(), gomock.Any())`

```go
validBody := `{
	"title": "Test Article",
	"slug": "test-article",
	"content": "Article content here",
	"excerpt": "Short excerpt",
	"cover_media_id": 1,
	"status": "draft",
	"published_at": "` + now.Format(time.RFC3339) + `",
	"view_count": 1
}`
```

And update mock expectations from `Create(gomock.Any(), gomock.Any())` to `Create(gomock.Any(), gomock.Any(), gomock.Any())`.

- [ ] **Step 9: Run all tests**

Run: `go test -v ./internal/usecase/article/... ./internal/handlers/http/v1/article/...`
Expected: All tests pass.

- [ ] **Step 10: Run full check**

Run: `make check-all`
Expected: All checks pass.

- [ ] **Step 11: Commit**

```bash
git add internal/dto/article/request.go internal/usecase/contracts.go \
  internal/usecase/article/create.go internal/usecase/article/create_test.go \
  internal/usecase/article/update.go internal/usecase/article/update_test.go \
  internal/handlers/http/v1/article/create.go \
  internal/handlers/http/v1/article/update.go \
  internal/handlers/http/v1/article/handler_test.go \
  internal/handlers/http/v1/article/mocks_test.go
git commit -m "fix: Remove UserID from article request DTOs (IDOR fix)

UserID is now extracted from JWT claims in the handler,
preventing users from creating/updating articles on behalf of others."
```

---

## Chunk 2: CI/CD & Infrastructure Fixes

### Task 3: Fix Dockerfile Go version (1.3)

**Files:**
- Modify: `deployment/docker/Dockerfile`

- [ ] **Step 1: Update both FROM lines**

In `deployment/docker/Dockerfile`, change both occurrences:
```dockerfile
# Old:
FROM golang:1.25-alpine3.21 AS modules
FROM golang:1.25-alpine3.21 AS builder
# New:
FROM golang:1.26-alpine3.21 AS modules
FROM golang:1.26-alpine3.21 AS builder
```

- [ ] **Step 2: Commit**

```bash
git add deployment/docker/Dockerfile
git commit -m "fix: Align Dockerfile Go version with go.mod (1.26)"
```

---

### Task 4: Align Postgres versions in CI (1.4)

**Files:**
- Modify: `.github/workflows/ci.yml:180`

- [ ] **Step 1: Update Postgres service image**

In `.github/workflows/ci.yml`, change line 180:
```yaml
# Old:
image: postgres:16-alpine
# New:
image: postgres:17-alpine
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "fix: Align CI Postgres to v17 (matches local Docker)"
```

---

### Task 5: Make integration tests blocking in CI (1.5)

**Files:**
- Modify: `.github/workflows/ci.yml:217`

- [ ] **Step 1: Remove || true from integration test command**

In `.github/workflows/ci.yml`, change line 217:
```yaml
# Old:
go test -v ./integration-test/... || true
# New:
go test -v ./integration-test/...
```

- [ ] **Step 2: Add integration-test to ci-success needs**

In `.github/workflows/ci.yml`, update the `ci-success` job. Since `integration-test` only runs on PRs (`if: github.event_name == 'pull_request'`), we need to handle it gracefully in the success gate. Add it to `needs` and allow skip:

```yaml
ci-success:
  name: CI Success
  runs-on: ubuntu-latest
  needs: [lint, format, build, test, security, integration-test]
  if: always()
  steps:
    - name: Check all jobs passed
      run: |
        if [[ "${{ needs.lint.result }}" != "success" ]] || \
           [[ "${{ needs.format.result }}" != "success" ]] || \
           [[ "${{ needs.build.result }}" != "success" ]] || \
           [[ "${{ needs.test.result }}" != "success" ]] || \
           [[ "${{ needs.security.result }}" != "success" ]]; then
          echo "One or more required jobs failed"
          exit 1
        fi
        # Integration tests only run on PRs - fail if they ran and failed
        if [[ "${{ needs.integration-test.result }}" == "failure" ]]; then
          echo "Integration tests failed"
          exit 1
        fi
        echo "All required jobs passed!"
```

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "fix: Make integration test failures block PR merges

Removed || true and added integration-test to ci-success gate."
```

---

### Task 6: Add global request body size limit (1.6)

**Files:**
- Modify: `config/config.go` (HTTP struct)
- Modify: `pkg/httpserver/options.go`
- Modify: `pkg/httpserver/server.go`
- Modify: `internal/app/app.go:257-262`

- [ ] **Step 1: Add BodyLimit to HTTP config struct**

In `config/config.go`, add to the `HTTP` struct (after `RequestTimeout`):
```go
// HTTP holds HTTP server configuration.
HTTP struct {
	Port           string        `mapstructure:"port"`
	Timeout        time.Duration `mapstructure:"timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
	BodyLimit      int           `mapstructure:"body_limit"` // Max request body size in bytes (default: 4MB)
}
```

- [ ] **Step 2: Add BodyLimit default, env binding, and config.yaml entry**

In `config/config.go`, find `setDefaults()` and add after the existing HTTP defaults (after line 321):
```go
viper.SetDefault("http.body_limit", 4*1024*1024) // 4MB
```

In `bindEnvVars()`, find the HTTP env bindings (around line 430-433) and add:
```go
viper.BindEnv("http.body_limit", "HTTP_BODY_LIMIT")
```

In `config/config.yaml`, add under the `http:` section:
```yaml
http:
  port: "8080"
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s
  body_limit: 4194304  # 4MB in bytes
```

In `config/config.example.yaml`, add the same entry.

- [ ] **Step 3: Add BodyLimit option to httpserver**

In `pkg/httpserver/options.go`, add after `ShutdownTimeout`:

```go
// BodyLimit sets the maximum allowed request body size in bytes.
func BodyLimit(limit int) Option {
	return func(s *Server) {
		s.bodyLimit = limit
	}
}
```

- [ ] **Step 4: Add bodyLimit field to Server struct**

In `pkg/httpserver/server.go`, add field and default constant:

Add constant:
```go
_defaultBodyLimit = 4 * 1024 * 1024 // 4MB
```

Add field to `Server` struct:
```go
bodyLimit int
```

Set default in `New()`:
```go
bodyLimit: _defaultBodyLimit,
```

Set in `fiber.Config`:
```go
app := fiber.New(fiber.Config{
	Prefork:      s.prefork,
	ReadTimeout:  s.readTimeout,
	WriteTimeout: s.writeTimeout,
	BodyLimit:    s.bodyLimit,
	JSONDecoder:  json.Unmarshal,
	JSONEncoder:  json.Marshal,
})
```

- [ ] **Step 5: Pass BodyLimit from config in app.go**

In `internal/app/app.go:257-262`, add the option:

```go
httpServer := httpserver.New(
	l,
	httpserver.Port(cfg.HTTP.Port),
	httpserver.ReadTimeout(cfg.HTTP.Timeout),
	httpserver.WriteTimeout(cfg.HTTP.Timeout),
	httpserver.BodyLimit(cfg.HTTP.BodyLimit),
)
```

- [ ] **Step 6: Run tests**

Run: `make check-all`
Expected: All checks pass.

- [ ] **Step 7: Commit**

```bash
git add config/config.go config/config.yaml config/config.example.yaml \
  pkg/httpserver/options.go pkg/httpserver/server.go internal/app/app.go
git commit -m "feat: Add configurable request body size limit

Defaults to 4MB. Configurable via HTTP_BODY_LIMIT env var."
```

---

## Final Verification

- [ ] **Run full quality checks**

Run: `make check-all`
Expected: All format, lint, vuln, and test checks pass.

- [ ] **Verify no regressions in existing tests**

Run: `go test -race ./...`
Expected: All tests pass with race detector enabled.
