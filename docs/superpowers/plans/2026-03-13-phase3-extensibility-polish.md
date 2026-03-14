# Phase 3: Extensibility & Polish — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add multi-replica-ready infrastructure (structured logging, Redis rate limiter, distributed locks, idempotency middleware, cache invalidation) so the boilerplate scales beyond a single instance with config-only changes.

**Architecture:** Each feature is config-gated (disabled by default), uses existing `pkg/redis` and `pkg/cache` packages, and follows the project's Clean Architecture pattern. All features are independent and can be tested in isolation.

**Tech Stack:** Go 1.24, Fiber v2, Redis (go-redis/v9), zap structured logging, gomock for test mocks.

**Spec:** `docs/superpowers/specs/2026-03-13-phase3-extensibility-design.md`

---

## Chunk 1: Structured Logging & Redis Rate Limiter (Tasks 1-2)

### Task 1: Structured Request/Response Logging (3.4)

**Files:**
- Modify: `config/config.go` — add logging config fields to `Log` struct
- Modify: `config/config.example.yaml` — document new logging fields
- Modify: `internal/handlers/http/middleware/logger.go` — rewrite with structured fields
- Create: `internal/handlers/http/middleware/logger_test.go` — test redaction, log levels

#### Step 1: Add logging config fields

- [ ] **Step 1.1: Add config fields to `Log` struct**

In `config/config.go`, update the `Log` struct:

```go
// Log holds logging configuration.
Log struct {
    Level           string `mapstructure:"level"`
    File            string `mapstructure:"file"`              // Optional file path for log output (empty = stdout only)
    LogRequestBody  bool   `mapstructure:"log_request_body"`  // Log request bodies (default: false)
    LogResponseBody bool   `mapstructure:"log_response_body"` // Log response bodies (default: false)
    RedactFields    string `mapstructure:"redact_fields"`     // Comma-separated fields to redact from body logs
}
```

- [ ] **Step 1.2: Add defaults and env bindings**

In `setDefaults()`:
```go
viper.SetDefault("log.log_request_body", false)
viper.SetDefault("log.log_response_body", false)
viper.SetDefault("log.redact_fields", "password,token,secret,authorization,credit_card")
```

In `bindEnvVars()`:
```go
viper.BindEnv("log.log_request_body", "LOG_REQUEST_BODY")
viper.BindEnv("log.log_response_body", "LOG_RESPONSE_BODY")
viper.BindEnv("log.redact_fields", "LOG_REDACT_FIELDS")
```

- [ ] **Step 1.3: Update `config.example.yaml`**

Add under the existing `log:` section:
```yaml
log:
  level: debug
  # file: /var/log/app.log  # Optional file output
  log_request_body: false    # Log request bodies (may contain sensitive data)
  log_response_body: false   # Log response bodies
  redact_fields: "password,token,secret,authorization,credit_card"  # Fields to redact from body logs
```

#### Step 2: Write the structured logger middleware

- [ ] **Step 2.1: Write the failing test file**

Create `internal/handlers/http/middleware/logger_test.go`:

```go
package middleware_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"go-boilerplate/config"
	"go-boilerplate/internal/handlers/http/middleware"
)

func newTestLogger() (*zap.Logger, *observer.ObservedLogs) {
	core, obs := observer.New(zap.DebugLevel)
	return zap.New(core), obs
}

func TestStructuredLogger_BasicFields(t *testing.T) {
	t.Parallel()

	zapLogger, obs := newTestLogger()
	app := fiber.New()
	app.Use(middleware.StructuredLogger(zapLogger, config.Log{}))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("User-Agent", "test-agent")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 1, obs.Len())
	entry := obs.All()[0]

	// Check structured fields
	fields := make(map[string]interface{})
	for _, f := range entry.Context {
		fields[f.Key] = f.Interface
	}

	assert.Contains(t, fields, "method")
	assert.Contains(t, fields, "path")
	assert.Contains(t, fields, "status")
	assert.Contains(t, fields, "latency_ms")
	assert.Contains(t, fields, "request_id")
	assert.Contains(t, fields, "ip")
	assert.Contains(t, fields, "user_agent")
	assert.Contains(t, fields, "bytes_out")
}

func TestStructuredLogger_LogLevelByStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		status        int
		expectedLevel string
	}{
		{"2xx logs at info", http.StatusOK, "info"},
		{"4xx logs at warn", http.StatusBadRequest, "warn"},
		{"5xx logs at error", http.StatusInternalServerError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			zapLogger, obs := newTestLogger()
			app := fiber.New()
			app.Use(middleware.StructuredLogger(zapLogger, config.Log{}))
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendStatus(tt.status)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, 1, obs.Len())
			assert.Equal(t, tt.expectedLevel, obs.All()[0].Level.String())
		})
	}
}

func TestStructuredLogger_RedactsFields(t *testing.T) {
	t.Parallel()

	zapLogger, obs := newTestLogger()
	cfg := config.Log{
		LogRequestBody: true,
		RedactFields:   "password,secret",
	}
	app := fiber.New()
	app.Use(middleware.StructuredLogger(zapLogger, cfg))
	app.Post("/login", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	body := `{"email":"test@example.com","password":"hunter2","secret":"abc"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 1, obs.Len())
	entry := obs.All()[0]

	var reqBody string
	for _, f := range entry.Context {
		if f.Key == "request_body" {
			reqBody = f.String
			break
		}
	}

	assert.NotEmpty(t, reqBody)
	assert.NotContains(t, reqBody, "hunter2")
	assert.NotContains(t, reqBody, "abc")
	assert.Contains(t, reqBody, "[REDACTED]")
	assert.Contains(t, reqBody, "test@example.com")
}

func TestStructuredLogger_NoBodyByDefault(t *testing.T) {
	t.Parallel()

	zapLogger, obs := newTestLogger()
	app := fiber.New()
	app.Use(middleware.StructuredLogger(zapLogger, config.Log{}))
	app.Post("/data", func(c *fiber.Ctx) error {
		return c.SendString("response")
	})

	body := `{"key":"value"}`
	req := httptest.NewRequest(http.MethodPost, "/data", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)

	require.Equal(t, 1, obs.Len())
	entry := obs.All()[0]

	for _, f := range entry.Context {
		assert.NotEqual(t, "request_body", f.Key)
		assert.NotEqual(t, "response_body", f.Key)
	}
}

func TestRedactJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		fields   []string
		contains []string
		excludes []string
	}{
		{
			name:     "redacts password",
			input:    `{"email":"a@b.com","password":"secret"}`,
			fields:   []string{"password"},
			contains: []string{"a@b.com", "[REDACTED]"},
			excludes: []string{"secret"},
		},
		{
			name:     "invalid json returns as-is",
			input:    "not json",
			fields:   []string{"password"},
			contains: []string{"not json"},
			excludes: nil,
		},
		{
			name:     "empty fields does nothing",
			input:    `{"password":"keep"}`,
			fields:   nil,
			contains: []string{"keep"},
			excludes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := middleware.RedactJSON(tt.input, tt.fields)
			for _, c := range tt.contains {
				assert.Contains(t, result, c)
			}
			for _, e := range tt.excludes {
				assert.NotContains(t, result, e)
			}
		})
	}
}

func TestParseRedactFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected []string
	}{
		{"password,token,secret", []string{"password", "token", "secret"}},
		{" password , token ", []string{"password", "token"}},
		{"", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			result := middleware.ParseRedactFields(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
```

- [ ] **Step 2.2: Run tests to verify they fail**

Run: `go test ./internal/handlers/http/middleware/... -run TestStructuredLogger -v`
Expected: Compilation error — `StructuredLogger`, `RedactJSON`, `ParseRedactFields` not defined.

- [ ] **Step 2.3: Rewrite the logger middleware**

Replace `internal/handlers/http/middleware/logger.go` with:

```go
package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"go-boilerplate/config"
	"go-boilerplate/pkg/json"
)

// ParseRedactFields splits a comma-separated list of field names into a slice.
func ParseRedactFields(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// RedactJSON replaces values of the specified keys with "[REDACTED]".
// If the input is not valid JSON, it is returned as-is.
func RedactJSON(body string, fields []string) string {
	if len(fields) == 0 {
		return body
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		return body
	}

	for _, f := range fields {
		if _, ok := parsed[f]; ok {
			parsed[f] = "[REDACTED]"
		}
	}

	out, err := json.Marshal(parsed)
	if err != nil {
		return body
	}
	return string(out)
}

// StructuredLogger returns a middleware that logs HTTP requests with structured fields.
// Log level is based on status code: 2xx→info, 4xx→warn, 5xx→error.
func StructuredLogger(zapLogger *zap.Logger, cfg config.Log) fiber.Handler {
	redactFields := ParseRedactFields(cfg.RedactFields)

	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Capture request body before handler consumes it
		var reqBody string
		if cfg.LogRequestBody {
			reqBody = string(c.Body())
		}

		// Execute next handler
		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()

		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Float64("latency_ms", float64(latency.Nanoseconds())/1e6),
			zap.String("request_id", c.GetRespHeader("X-Request-Id", c.Locals("requestid").(string))),
			zap.String("ip", c.IP()),
			zap.String("user_agent", c.Get("User-Agent")),
			zap.Int("bytes_out", len(c.Response().Body())),
		}

		// Add user_id from JWT context if present
		if userID, ok := c.Locals(UserIDKey).(uint); ok && userID > 0 {
			fields = append(fields, zap.Uint("user_id", userID))
		}

		// Add request body (redacted) if enabled
		if cfg.LogRequestBody && reqBody != "" {
			fields = append(fields, zap.String("request_body", RedactJSON(reqBody, redactFields)))
		}

		// Add response body if enabled
		if cfg.LogResponseBody {
			fields = append(fields, zap.String("response_body", string(c.Response().Body())))
		}

		msg := "HTTP request"

		// Log level based on status code
		switch {
		case status >= 500:
			zapLogger.Error(msg, fields...)
		case status >= 400:
			zapLogger.Warn(msg, fields...)
		default:
			zapLogger.Info(msg, fields...)
		}

		return err
	}
}
```

- [ ] **Step 2.4: Update router.go to use StructuredLogger**

In `internal/handlers/http/router.go`, replace:
```go
app.Use(middleware.Logger(l))
```
with:
```go
app.Use(middleware.StructuredLogger(l.GetZapLogger(), cfg.Log))
```

Remove the old `Logger` function import if it was the only usage. The old `Logger` function and `buildRequestMessage` can be removed from logger.go since `StructuredLogger` replaces it.

- [ ] **Step 2.5: Run tests to verify they pass**

Run: `go test ./internal/handlers/http/middleware/... -v`
Expected: All tests PASS.

- [ ] **Step 2.6: Run full quality checks**

Run: `make check-all`
Expected: All checks pass, coverage >= 84%.

- [ ] **Step 2.7: Commit**

```bash
git add internal/handlers/http/middleware/logger.go internal/handlers/http/middleware/logger_test.go internal/handlers/http/router.go config/config.go config/config.example.yaml
git commit -m "feat: add structured request logging with PII redaction (3.4)"
```

---

### Task 2: Redis-Backed Rate Limiter (3.1)

**Files:**
- Create: `pkg/ratelimiter/redis_store.go` — Fiber Storage adapter backed by Redis
- Create: `pkg/ratelimiter/redis_store_test.go` — unit tests
- Modify: `internal/handlers/http/router.go` — add Redis storage branch
- Modify: `internal/app/app.go` — create shared Redis client, pass to router

#### Step 1: Create the Redis store adapter

- [ ] **Step 1.1: Write the failing test**

Create `pkg/ratelimiter/redis_store_test.go`:

```go
package ratelimiter_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/ratelimiter"
)

// mockRedisClient implements the subset of redis operations needed by RedisStore.
type mockRedisClient struct {
	store map[string][]byte
}

func newMockRedis() *mockRedisClient {
	return &mockRedisClient{store: make(map[string][]byte)}
}

func (m *mockRedisClient) Get(key string) ([]byte, error) {
	v, ok := m.store[key]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func (m *mockRedisClient) Set(key string, val []byte, _ time.Duration) error {
	m.store[key] = val
	return nil
}

func (m *mockRedisClient) Delete(key string) error {
	delete(m.store, key)
	return nil
}

func (m *mockRedisClient) Reset() error {
	m.store = make(map[string][]byte)
	return nil
}

func (m *mockRedisClient) Close() error {
	return nil
}

func TestRedisStore_GetSetDelete(t *testing.T) {
	t.Parallel()

	mock := newMockRedis()
	store := ratelimiter.NewRedisStore(mock)

	// Get non-existent key
	val, err := store.Get("key1")
	require.NoError(t, err)
	assert.Nil(t, val)

	// Set
	err = store.Set("key1", []byte("value1"), time.Minute)
	require.NoError(t, err)

	// Get existing key
	val, err = store.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, []byte("value1"), val)

	// Delete
	err = store.Delete("key1")
	require.NoError(t, err)

	// Verify deleted
	val, err = store.Get("key1")
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestRedisStore_Reset(t *testing.T) {
	t.Parallel()

	mock := newMockRedis()
	store := ratelimiter.NewRedisStore(mock)

	_ = store.Set("a", []byte("1"), time.Minute)
	_ = store.Set("b", []byte("2"), time.Minute)

	err := store.Reset()
	require.NoError(t, err)

	val, _ := store.Get("a")
	assert.Nil(t, val)
	val, _ = store.Get("b")
	assert.Nil(t, val)
}

func TestRedisStore_Close(t *testing.T) {
	t.Parallel()

	mock := newMockRedis()
	store := ratelimiter.NewRedisStore(mock)
	err := store.Close()
	require.NoError(t, err)
}
```

- [ ] **Step 1.2: Run tests to verify they fail**

Run: `go test ./pkg/ratelimiter/... -v`
Expected: Compilation error — package `ratelimiter` not found.

- [ ] **Step 1.3: Implement RedisStore**

Create `pkg/ratelimiter/redis_store.go`:

```go
// Package ratelimiter provides rate limiter storage backends.
package ratelimiter

import "time"

// Storage defines the interface for rate limiter backends.
// This matches the subset of operations the Fiber limiter needs.
type Storage interface {
	Get(key string) ([]byte, error)
	Set(key string, val []byte, exp time.Duration) error
	Delete(key string) error
	Reset() error
	Close() error
}

// RedisStore implements Fiber's Storage interface backed by a Redis-compatible backend.
type RedisStore struct {
	backend Storage
}

// NewRedisStore creates a new RedisStore wrapping the given storage backend.
func NewRedisStore(backend Storage) *RedisStore {
	return &RedisStore{backend: backend}
}

// Get retrieves a value by key.
func (s *RedisStore) Get(key string) ([]byte, error) {
	return s.backend.Get(key)
}

// Set stores a value with expiration.
func (s *RedisStore) Set(key string, val []byte, exp time.Duration) error {
	return s.backend.Set(key, val, exp)
}

// Delete removes a key.
func (s *RedisStore) Delete(key string) error {
	return s.backend.Delete(key)
}

// Reset clears all keys.
func (s *RedisStore) Reset() error {
	return s.backend.Reset()
}

// Close closes the storage connection.
func (s *RedisStore) Close() error {
	return s.backend.Close()
}
```

- [ ] **Step 1.4: Run tests to verify they pass**

Run: `go test ./pkg/ratelimiter/... -v`
Expected: All tests PASS.

#### Step 2: Create Redis adapter that implements the Storage interface

- [ ] **Step 2.1: Create the Redis adapter**

Create `pkg/ratelimiter/redis_adapter.go`:

```go
package ratelimiter

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisAdapter adapts a go-redis client to the Storage interface.
type RedisAdapter struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisAdapter creates a new Redis adapter.
func NewRedisAdapter(client *redis.Client) *RedisAdapter {
	return &RedisAdapter{
		client: client,
		ctx:    context.Background(),
	}
}

// Get retrieves a value by key.
func (a *RedisAdapter) Get(key string) ([]byte, error) {
	val, err := a.client.Get(a.ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	return val, err
}

// Set stores a value with expiration.
func (a *RedisAdapter) Set(key string, val []byte, exp time.Duration) error {
	return a.client.Set(a.ctx, key, val, exp).Err()
}

// Delete removes a key.
func (a *RedisAdapter) Delete(key string) error {
	return a.client.Del(a.ctx, key).Err()
}

// Reset flushes the database.
func (a *RedisAdapter) Reset() error {
	return a.client.FlushDB(a.ctx).Err()
}

// Close closes the Redis connection.
func (a *RedisAdapter) Close() error {
	return a.client.Close()
}
```

#### Step 3: Wire Redis rate limiter into router and app

- [ ] **Step 3.1: Update router.go to accept optional storage**

In `internal/handlers/http/router.go`, update `SetupRoutes` signature to accept an optional `fiber.Storage`:

```go
func SetupRoutes(app *fiber.App, cfg *config.Config, translationUC usecase.Translation, authUC usecase.Auth, mediaUC usecase.Media, profileUC usecase.Profile, articleUC usecase.Article, jwtService jwt.Service, l logger.Interface, healthChecker HealthChecker, rateLimitStorage fiber.Storage) {
	setupMiddleware(app, cfg, l, rateLimitStorage)
	// ... rest unchanged
}
```

Update `setupMiddleware` to accept and use the storage:

```go
func setupMiddleware(app *fiber.App, cfg *config.Config, l logger.Interface, rateLimitStorage fiber.Storage) {
	// ... existing middleware ...
	limiterCfg := limiter.Config{
		Max:          cfg.RateLimit.Max,
		Expiration:   cfg.RateLimit.Expiration,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
		LimitReached: rateLimitReached,
		Storage:      rateLimitStorage, // nil = Fiber's built-in memory store
	}
	app.Use(limiter.New(limiterCfg))
	// ... rest unchanged, but use StructuredLogger ...
}
```

- [ ] **Step 3.2: Update app.go to create shared Redis client and pass storage**

In `internal/app/app.go`, add a function to create rate limiter storage and update wiring:

```go
import (
	// ... add these imports:
	pkgredis "go-boilerplate/pkg/redis"
	"go-boilerplate/pkg/ratelimiter"
	goredis "github.com/redis/go-redis/v9"
)

// initRateLimitStorage creates rate limiter storage based on config.
func initRateLimitStorage(cfg *config.Config, l *logger.Logger) fiber.Storage {
	if cfg.RateLimit.Store != "redis" {
		return nil // nil = Fiber built-in memory store
	}

	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	adapter := ratelimiter.NewRedisAdapter(redisClient)
	l.Info("Rate limiter using Redis backend at %s", cfg.Redis.Addr())

	return ratelimiter.NewRedisStore(adapter)
}
```

Update `initHTTPServer` to create and pass the storage:

```go
func initHTTPServer(cfg *config.Config, l *logger.Logger, uc *usecases, jwtService jwt.Service, pg *postgres.Postgres) *httpserver.Server {
	httpServer := httpserver.New(/* ... same ... */)

	rateLimitStorage := initRateLimitStorage(cfg, l)
	httphandler.SetupRoutes(httpServer.App, cfg, uc.translation, uc.auth, uc.media, uc.profile, uc.article, jwtService, l, pg, rateLimitStorage)
	httpServer.Start()

	return httpServer
}
```

- [ ] **Step 3.3: Run tests**

Run: `make check-all`
Expected: All checks pass.

- [ ] **Step 3.4: Commit**

```bash
git add pkg/ratelimiter/ internal/handlers/http/router.go internal/app/app.go
git commit -m "feat: add Redis-backed rate limiter storage adapter (3.1)"
```

---

## Chunk 2: Distributed Lock & Idempotency (Tasks 3-4)

### Task 3: Distributed Lock Abstraction (3.2)

**Files:**
- Create: `pkg/lock/lock.go` — interfaces + errors
- Create: `pkg/lock/noop.go` — NoopLocker
- Create: `pkg/lock/noop_test.go` — tests
- Create: `pkg/lock/redis.go` — RedisLocker
- Create: `pkg/lock/redis_test.go` — tests
- Modify: `config/config.go` — add Lock config struct
- Modify: `config/config.example.yaml` — document Lock fields

#### Step 1: Define interfaces and config

- [ ] **Step 1.1: Add Lock config**

In `config/config.go`, add to the `Config` struct:

```go
Lock Lock `mapstructure:"lock"`
```

Add the Lock struct:

```go
// Lock holds distributed lock configuration.
Lock struct {
    Provider string `mapstructure:"provider"` // "noop" (default) or "redis"
}
```

Add defaults:
```go
viper.SetDefault("lock.provider", "noop")
```

Add env bindings:
```go
viper.BindEnv("lock.provider", "LOCK_PROVIDER")
```

Update `config.example.yaml`:
```yaml
# Distributed Lock
lock:
  provider: noop  # "noop" (single instance) or "redis" (multi-replica)
```

- [ ] **Step 1.2: Create lock interfaces**

Create `pkg/lock/lock.go`:

```go
// Package lock provides a distributed lock abstraction.
package lock

import (
	"context"
	"errors"
	"time"
)

// ErrLockNotAcquired is returned when a lock cannot be obtained.
var ErrLockNotAcquired = errors.New("lock: not acquired")

// Unlocker releases a held lock.
type Unlocker interface {
	Unlock(ctx context.Context) error
}

// Locker provides distributed locking.
type Locker interface {
	// Lock blocks until the lock is acquired or ctx is cancelled.
	Lock(ctx context.Context, key string, ttl time.Duration) (Unlocker, error)

	// TryLock attempts to acquire the lock without blocking.
	// Returns (unlocker, true, nil) on success, (nil, false, nil) if lock is held.
	TryLock(ctx context.Context, key string, ttl time.Duration) (Unlocker, bool, error)
}
```

#### Step 2: Implement NoopLocker

- [ ] **Step 2.1: Write the failing test**

Create `pkg/lock/noop_test.go`:

```go
package lock_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/lock"
)

func TestNoopLocker_Lock(t *testing.T) {
	t.Parallel()

	locker := lock.NewNoop()
	ctx := context.Background()

	unlocker, err := locker.Lock(ctx, "test-key", 5*time.Second)
	require.NoError(t, err)
	require.NotNil(t, unlocker)

	err = unlocker.Unlock(ctx)
	assert.NoError(t, err)
}

func TestNoopLocker_TryLock(t *testing.T) {
	t.Parallel()

	locker := lock.NewNoop()
	ctx := context.Background()

	unlocker, ok, err := locker.TryLock(ctx, "test-key", 5*time.Second)
	require.NoError(t, err)
	assert.True(t, ok)
	require.NotNil(t, unlocker)

	err = unlocker.Unlock(ctx)
	assert.NoError(t, err)
}
```

- [ ] **Step 2.2: Run tests to verify they fail**

Run: `go test ./pkg/lock/... -v`
Expected: Compilation error.

- [ ] **Step 2.3: Implement NoopLocker**

Create `pkg/lock/noop.go`:

```go
package lock

import (
	"context"
	"time"
)

// NoopLocker always succeeds — for single-instance deployments.
type NoopLocker struct{}

// NewNoop creates a new no-op locker.
func NewNoop() *NoopLocker {
	return &NoopLocker{}
}

type noopUnlocker struct{}

func (n *noopUnlocker) Unlock(_ context.Context) error { return nil }

// Lock always succeeds immediately.
func (n *NoopLocker) Lock(_ context.Context, _ string, _ time.Duration) (Unlocker, error) {
	return &noopUnlocker{}, nil
}

// TryLock always succeeds.
func (n *NoopLocker) TryLock(_ context.Context, _ string, _ time.Duration) (Unlocker, bool, error) {
	return &noopUnlocker{}, true, nil
}
```

- [ ] **Step 2.4: Run tests**

Run: `go test ./pkg/lock/... -v`
Expected: All tests PASS.

#### Step 3: Implement RedisLocker

- [ ] **Step 3.1: Write the failing test**

Create `pkg/lock/redis_test.go`:

```go
package lock_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/lock"
)

// mockRedisLockClient implements the RedisClient interface for testing.
type mockRedisLockClient struct {
	mu    sync.Mutex
	store map[string]string
}

func newMockRedisLockClient() *mockRedisLockClient {
	return &mockRedisLockClient{store: make(map[string]string)}
}

func (m *mockRedisLockClient) SetNX(ctx context.Context, key string, value interface{}, exp time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.store[key]; exists {
		return false, nil
	}
	m.store[key] = value.(string)
	return true, nil
}

func (m *mockRedisLockClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(keys) == 0 || len(args) == 0 {
		return errors.New("invalid args")
	}
	key := keys[0]
	val, ok := args[0].(string)
	if !ok {
		return errors.New("invalid value type")
	}
	stored, exists := m.store[key]
	if !exists || stored != val {
		return lock.ErrLockNotAcquired
	}
	delete(m.store, key)
	return nil
}

func TestRedisLocker_TryLock_Success(t *testing.T) {
	t.Parallel()

	client := newMockRedisLockClient()
	locker := lock.NewRedis(client)
	ctx := context.Background()

	unlocker, ok, err := locker.TryLock(ctx, "resource:1", 10*time.Second)
	require.NoError(t, err)
	assert.True(t, ok)
	require.NotNil(t, unlocker)

	err = unlocker.Unlock(ctx)
	assert.NoError(t, err)
}

func TestRedisLocker_TryLock_AlreadyLocked(t *testing.T) {
	t.Parallel()

	client := newMockRedisLockClient()
	locker := lock.NewRedis(client)
	ctx := context.Background()

	// First lock succeeds
	_, ok, err := locker.TryLock(ctx, "resource:1", 10*time.Second)
	require.NoError(t, err)
	assert.True(t, ok)

	// Second lock fails
	unlocker2, ok2, err := locker.TryLock(ctx, "resource:1", 10*time.Second)
	require.NoError(t, err)
	assert.False(t, ok2)
	assert.Nil(t, unlocker2)
}

func TestRedisLocker_Lock_Success(t *testing.T) {
	t.Parallel()

	client := newMockRedisLockClient()
	locker := lock.NewRedis(client)
	ctx := context.Background()

	unlocker, err := locker.Lock(ctx, "resource:2", 10*time.Second)
	require.NoError(t, err)
	require.NotNil(t, unlocker)

	err = unlocker.Unlock(ctx)
	assert.NoError(t, err)
}

func TestRedisLocker_Lock_ContextCancelled(t *testing.T) {
	t.Parallel()

	client := newMockRedisLockClient()
	locker := lock.NewRedis(client)

	// Occupy the lock
	ctx := context.Background()
	_, _, _ = locker.TryLock(ctx, "resource:3", 10*time.Second)

	// Try to lock with cancelled context
	cancelCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := locker.Lock(cancelCtx, "resource:3", 10*time.Second)
	assert.Error(t, err)
}
```

- [ ] **Step 3.2: Run tests to verify they fail**

Run: `go test ./pkg/lock/... -v`
Expected: Compilation error — `NewRedis`, `RedisClient` not defined.

- [ ] **Step 3.3: Implement RedisLocker**

Create `pkg/lock/redis.go`:

```go
package lock

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RedisClient defines the Redis operations needed for distributed locking.
type RedisClient interface {
	SetNX(ctx context.Context, key string, value interface{}, exp time.Duration) (bool, error)
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) error
}

// unlockScript is a Lua script that atomically checks the value before deleting.
// This prevents accidentally releasing a lock held by another process.
const unlockScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end
`

const lockRetryInterval = 50 * time.Millisecond

// RedisLocker implements distributed locking using Redis SetNX + Lua unlock.
type RedisLocker struct {
	client RedisClient
}

// NewRedis creates a new Redis-backed locker.
func NewRedis(client RedisClient) *RedisLocker {
	return &RedisLocker{client: client}
}

type redisUnlocker struct {
	client RedisClient
	key    string
	value  string
}

func (u *redisUnlocker) Unlock(ctx context.Context) error {
	return u.client.Eval(ctx, unlockScript, []string{u.key}, u.value)
}

// TryLock attempts to acquire the lock without blocking.
func (l *RedisLocker) TryLock(ctx context.Context, key string, ttl time.Duration) (Unlocker, bool, error) {
	value := uuid.New().String()

	ok, err := l.client.SetNX(ctx, key, value, ttl)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}

	return &redisUnlocker{client: l.client, key: key, value: value}, true, nil
}

// Lock blocks until the lock is acquired or the context is cancelled.
func (l *RedisLocker) Lock(ctx context.Context, key string, ttl time.Duration) (Unlocker, error) {
	for {
		unlocker, ok, err := l.TryLock(ctx, key, ttl)
		if err != nil {
			return nil, err
		}
		if ok {
			return unlocker, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(lockRetryInterval):
			// Retry
		}
	}
}
```

- [ ] **Step 3.4: Run tests**

Run: `go test ./pkg/lock/... -v`
Expected: All tests PASS.

- [ ] **Step 3.5: Run quality checks and commit**

Run: `make check-all`
Expected: All checks pass.

```bash
git add pkg/lock/ config/config.go config/config.example.yaml
git commit -m "feat: add distributed lock abstraction with Noop and Redis backends (3.2)"
```

---

### Task 4: Idempotency Key Middleware (3.3)

**Files:**
- Create: `internal/handlers/http/middleware/idempotency.go` — middleware
- Create: `internal/handlers/http/middleware/idempotency_test.go` — tests
- Modify: `config/config.go` — add Idempotency config struct
- Modify: `config/config.example.yaml` — document Idempotency fields
- Modify: `internal/handlers/http/router.go` — register middleware (config-gated)

#### Step 1: Add config

- [ ] **Step 1.1: Add Idempotency config struct**

In `config/config.go`, add to `Config`:

```go
Idempotency Idempotency `mapstructure:"idempotency"`
```

Add the struct:

```go
// Idempotency holds idempotency middleware configuration.
Idempotency struct {
    Enabled         bool          `mapstructure:"enabled"`           // Enable idempotency middleware
    TTL             time.Duration `mapstructure:"ttl"`               // Cache TTL for idempotent responses
    RequiredForPost bool          `mapstructure:"required_for_post"` // Require Idempotency-Key for POST
}
```

Add defaults:
```go
viper.SetDefault("idempotency.enabled", false)
viper.SetDefault("idempotency.ttl", "24h")
viper.SetDefault("idempotency.required_for_post", false)
```

Add env bindings:
```go
viper.BindEnv("idempotency.enabled", "IDEMPOTENCY_ENABLED")
viper.BindEnv("idempotency.ttl", "IDEMPOTENCY_TTL")
viper.BindEnv("idempotency.required_for_post", "IDEMPOTENCY_REQUIRED_FOR_POST")
```

Update `config.example.yaml`:
```yaml
# Idempotency Middleware
idempotency:
  enabled: false
  ttl: 24h
  required_for_post: false  # Return 400 if POST lacks Idempotency-Key header
```

#### Step 2: Implement the middleware

- [ ] **Step 2.1: Write the failing test**

Create `internal/handlers/http/middleware/idempotency_test.go`:

```go
package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-boilerplate/config"
	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/pkg/json"
)

// mockCache implements cache.Cache for testing idempotency.
type mockCache struct {
	store map[string][]byte
}

func newMockCache() *mockCache {
	return &mockCache{store: make(map[string][]byte)}
}

func (m *mockCache) Get(_ context.Context, key string, dest interface{}) error {
	data, ok := m.store[key]
	if !ok {
		return idempotencyCacheNotFound{}
	}
	return json.Unmarshal(data, dest)
}

func (m *mockCache) Set(_ context.Context, key string, value interface{}, _ time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m.store[key] = data
	return nil
}

func (m *mockCache) Delete(_ context.Context, key string) error {
	delete(m.store, key)
	return nil
}

func (m *mockCache) Exists(_ context.Context, key string) (bool, error) {
	_, ok := m.store[key]
	return ok, nil
}

func (m *mockCache) Remember(_ context.Context, _ string, _ time.Duration, _ interface{}, fn func() (interface{}, error)) error {
	_, err := fn()
	return err
}

// idempotencyCacheNotFound is compatible with cache.ErrNotFound check.
type idempotencyCacheNotFound struct{}

func (e idempotencyCacheNotFound) Error() string { return "cache: key not found" }

func TestIdempotency_SkipsGET(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	cache := newMockCache()
	app.Use(middleware.Idempotency(cache, config.Idempotency{Enabled: true, TTL: time.Hour}))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("Idempotency-Key", "key-1")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// Cache should be empty — GET is bypassed
	assert.Empty(t, cache.store)
}

func TestIdempotency_CachesAndReplaysPOST(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	cache := newMockCache()
	cfg := config.Idempotency{Enabled: true, TTL: time.Hour}
	app.Use(middleware.Idempotency(cache, cfg))

	callCount := 0
	app.Post("/create", func(c *fiber.Ctx) error {
		callCount++
		return c.Status(http.StatusCreated).JSON(fiber.Map{"id": 1})
	})

	// First request
	req1 := httptest.NewRequest(http.MethodPost, "/create", http.NoBody)
	req1.Header.Set("Idempotency-Key", "idem-1")
	resp1, err := app.Test(req1)
	require.NoError(t, err)
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusCreated, resp1.StatusCode)
	assert.Equal(t, 1, callCount)

	// Second request with same key — should return cached response
	req2 := httptest.NewRequest(http.MethodPost, "/create", http.NoBody)
	req2.Header.Set("Idempotency-Key", "idem-1")
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusCreated, resp2.StatusCode)
	assert.Equal(t, 1, callCount) // Handler NOT called again
}

func TestIdempotency_RequiredForPOST(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	cache := newMockCache()
	cfg := config.Idempotency{Enabled: true, TTL: time.Hour, RequiredForPost: true}
	app.Use(middleware.Idempotency(cache, cfg))
	app.Post("/create", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusCreated)
	})

	// POST without Idempotency-Key
	req := httptest.NewRequest(http.MethodPost, "/create", http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestIdempotency_NotRequiredAllowsThrough(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	cache := newMockCache()
	cfg := config.Idempotency{Enabled: true, TTL: time.Hour, RequiredForPost: false}
	app.Use(middleware.Idempotency(cache, cfg))
	app.Post("/create", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusCreated)
	})

	// POST without Idempotency-Key — should pass through
	req := httptest.NewRequest(http.MethodPost, "/create", http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}
```

- [ ] **Step 2.2: Run tests to verify they fail**

Run: `go test ./internal/handlers/http/middleware/... -run TestIdempotency -v`
Expected: Compilation error — `Idempotency` function not defined.

- [ ] **Step 2.3: Implement the middleware**

Create `internal/handlers/http/middleware/idempotency.go`:

```go
package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"go-boilerplate/config"
	"go-boilerplate/pkg/cache"
)

const idempotencyHeader = "Idempotency-Key"

// idempotencyResponse stores a cached response for replay.
type idempotencyResponse struct {
	Status      int    `json:"status"`
	ContentType string `json:"content_type"`
	Body        []byte `json:"body"`
}

// Idempotency returns middleware that deduplicates mutating requests using an Idempotency-Key header.
// GET, DELETE, OPTIONS, and HEAD requests bypass the middleware entirely.
func Idempotency(c cache.Cache, cfg config.Idempotency) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Only apply to mutating methods
		method := ctx.Method()
		if method == fiber.MethodGet || method == fiber.MethodDelete ||
			method == fiber.MethodOptions || method == fiber.MethodHead {
			return ctx.Next()
		}

		key := strings.TrimSpace(ctx.Get(idempotencyHeader))

		// If no key is provided
		if key == "" {
			if cfg.RequiredForPost && method == fiber.MethodPost {
				return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error": fiber.Map{
						"code":    "MISSING_IDEMPOTENCY_KEY",
						"message": "Idempotency-Key header is required for POST requests",
					},
				})
			}
			return ctx.Next()
		}

		cacheKey := "idempotency:" + key

		// Check for cached response
		var cached idempotencyResponse
		err := c.Get(ctx.Context(), cacheKey, &cached)
		if err == nil {
			// Cache hit — replay the response
			ctx.Set("Content-Type", cached.ContentType)
			ctx.Set("X-Idempotent-Replay", "true")
			return ctx.Status(cached.Status).Send(cached.Body)
		}
		if !errors.Is(err, cache.ErrNotFound) {
			// Real cache error — proceed without idempotency
			return ctx.Next()
		}

		// Cache miss — execute the handler
		if err := ctx.Next(); err != nil {
			return err
		}

		// Cache the response (best effort)
		resp := idempotencyResponse{
			Status:      ctx.Response().StatusCode(),
			ContentType: string(ctx.Response().Header.ContentType()),
			Body:        ctx.Response().Body(),
		}
		_ = c.Set(ctx.Context(), cacheKey, resp, cfg.TTL)

		return nil
	}
}
```

- [ ] **Step 2.4: Run tests**

Run: `go test ./internal/handlers/http/middleware/... -run TestIdempotency -v`

Note: The mockCache's `idempotencyCacheNotFound` error needs to be compatible with `errors.Is(err, cache.ErrNotFound)`. The test's mock uses a custom error type. Update the mock to return `cache.ErrNotFound` directly instead:

```go
func (m *mockCache) Get(_ context.Context, key string, dest interface{}) error {
	data, ok := m.store[key]
	if !ok {
		return cache.ErrNotFound
	}
	return json.Unmarshal(data, dest)
}
```

Add `"go-boilerplate/pkg/cache"` import and remove the `idempotencyCacheNotFound` type.

Expected: All tests PASS.

#### Step 3: Wire into router

- [ ] **Step 3.1: Update router to register idempotency middleware**

In `internal/handlers/http/router.go`, update `SetupRoutes` to accept `cache.Cache`:

```go
func SetupRoutes(app *fiber.App, cfg *config.Config, translationUC usecase.Translation, authUC usecase.Auth, mediaUC usecase.Media, profileUC usecase.Profile, articleUC usecase.Article, jwtService jwt.Service, l logger.Interface, healthChecker HealthChecker, rateLimitStorage fiber.Storage, appCache cache.Cache) {
	setupMiddleware(app, cfg, l, rateLimitStorage, appCache)
	// ... rest unchanged
}
```

In `setupMiddleware`, after rate limiter and before timeout, add:

```go
if cfg.Idempotency.Enabled && appCache != nil {
    app.Use(middleware.Idempotency(appCache, cfg.Idempotency))
}
```

Update `app.go` to create and pass the cache:

```go
import "go-boilerplate/pkg/cache"

// In initHTTPServer or Run:
var appCache cache.Cache
if cfg.Cache.Enabled {
    redisClient := goredis.NewClient(&goredis.Options{
        Addr:     cfg.Redis.Addr(),
        Password: cfg.Redis.Password,
        DB:       cfg.Redis.DB,
    })
    appCache = cache.NewRedis(redisClient, cfg.Cache.Prefix)
} else {
    appCache = cache.NewNoop()
}
```

Pass `appCache` to `SetupRoutes`.

- [ ] **Step 3.2: Run quality checks and commit**

Run: `make check-all`
Expected: All checks pass.

```bash
git add internal/handlers/http/middleware/idempotency.go internal/handlers/http/middleware/idempotency_test.go config/config.go config/config.example.yaml internal/handlers/http/router.go internal/app/app.go
git commit -m "feat: add idempotency key middleware with cache-backed dedup (3.3)"
```

---

## Chunk 3: Cache Invalidation (Task 5)

### Task 5: Cache Invalidation Pattern (3.5)

**Files:**
- Modify: `pkg/cache/cache.go` — add `DeleteByPrefix` to interface
- Modify: `pkg/cache/redis.go` — SCAN+DEL implementation
- Modify: `pkg/cache/noop.go` — noop implementation
- Create: `pkg/cache/keys.go` — CacheKeyBuilder
- Create: `pkg/cache/keys_test.go` — tests
- Modify: `internal/usecase/article/article.go` — add cache field
- Modify: `internal/usecase/article/create.go` — invalidate list cache
- Modify: `internal/usecase/article/update.go` — invalidate item + list cache
- Modify: `internal/usecase/article/delete.go` — invalidate item + list cache
- Modify: `internal/usecase/article/get_by_id.go` — cache reads
- Modify: `internal/usecase/article/list.go` — cache reads
- Update: article test files — update constructor calls
- Update: `internal/app/app.go` — pass cache to article usecase

#### Step 1: Extend Cache interface

- [ ] **Step 1.1: Add DeleteByPrefix to interface**

In `pkg/cache/cache.go`, add to the `Cache` interface:

```go
// DeleteByPrefix removes all keys matching the given prefix.
DeleteByPrefix(ctx context.Context, prefix string) error
```

- [ ] **Step 1.2: Implement in NoopCache**

In `pkg/cache/noop.go`, add:

```go
// DeleteByPrefix does nothing and returns nil.
func (c *NoopCache) DeleteByPrefix(_ context.Context, _ string) error {
	return nil
}
```

- [ ] **Step 1.3: Implement in RedisCache with SCAN+DEL**

In `pkg/cache/redis.go`, add:

```go
// DeleteByPrefix removes all keys matching the given prefix using SCAN+DEL.
// The cache's configured prefix is prepended before scanning.
func (c *RedisCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	fullPrefix := c.prefixKey(prefix)
	pattern := fullPrefix + "*"

	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}
```

#### Step 2: Create CacheKeyBuilder

- [ ] **Step 2.1: Write the failing test**

Create `pkg/cache/keys_test.go`:

```go
package cache_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go-boilerplate/pkg/cache"
)

func TestCacheKeyBuilder(t *testing.T) {
	t.Parallel()

	kb := cache.NewKeyBuilder("article")

	assert.Equal(t, "article:1", kb.ID(1))
	assert.Equal(t, "article:42", kb.ID(42))
	assert.Equal(t, "article:list:", kb.ListPrefix())
	assert.Equal(t, "article:list:page=1&size=10", kb.List("page=1&size=10"))
	assert.Equal(t, "article:", kb.Prefix())
}
```

- [ ] **Step 2.2: Run tests to verify they fail**

Run: `go test ./pkg/cache/... -v`
Expected: Compilation error — `NewKeyBuilder` not defined.

- [ ] **Step 2.3: Implement CacheKeyBuilder**

Create `pkg/cache/keys.go`:

```go
package cache

import "fmt"

// KeyBuilder constructs consistent cache keys for an entity type.
// Pattern: entity:scope:id (e.g., "article:list:page=1&size=10").
type KeyBuilder struct {
	entity string
}

// NewKeyBuilder creates a KeyBuilder for the given entity type.
func NewKeyBuilder(entity string) *KeyBuilder {
	return &KeyBuilder{entity: entity}
}

// ID returns a key for a specific entity instance.
func (kb *KeyBuilder) ID(id uint) string {
	return fmt.Sprintf("%s:%d", kb.entity, id)
}

// List returns a key for a list query with the given qualifier.
func (kb *KeyBuilder) List(qualifier string) string {
	return fmt.Sprintf("%s:list:%s", kb.entity, qualifier)
}

// ListPrefix returns the prefix for all list keys.
func (kb *KeyBuilder) ListPrefix() string {
	return kb.entity + ":list:"
}

// Prefix returns the prefix for all keys of this entity.
func (kb *KeyBuilder) Prefix() string {
	return kb.entity + ":"
}
```

- [ ] **Step 2.4: Run tests**

Run: `go test ./pkg/cache/... -v`
Expected: All tests PASS.

#### Step 3: Wire cache into article usecase

- [ ] **Step 3.1: Update article UseCase struct**

In `internal/usecase/article/article.go`, add cache field:

```go
import (
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/audit"
	"go-boilerplate/pkg/cache"
)

// UseCase implements article business logic.
type UseCase struct {
	articleRepo repo.ArticleRepo
	auditLogger audit.Logger
	cache       cache.Cache
	cacheKeys   *cache.KeyBuilder
}

// New creates a new article use case.
func New(articleRepo repo.ArticleRepo, auditLogger audit.Logger, articleCache cache.Cache) *UseCase {
	return &UseCase{
		articleRepo: articleRepo,
		auditLogger: auditLogger,
		cache:       articleCache,
		cacheKeys:   cache.NewKeyBuilder("article"),
	}
}
```

- [ ] **Step 3.2: Add cache invalidation to Create**

In `internal/usecase/article/create.go`, after the audit log call:

```go
// Invalidate list cache (new article changes any list)
_ = uc.cache.DeleteByPrefix(ctx, uc.cacheKeys.ListPrefix())
```

- [ ] **Step 3.3: Add cache invalidation to Update**

In `internal/usecase/article/update.go`, after the audit log call:

```go
// Invalidate caches
_ = uc.cache.Delete(ctx, uc.cacheKeys.ID(id))
_ = uc.cache.DeleteByPrefix(ctx, uc.cacheKeys.ListPrefix())
```

- [ ] **Step 3.4: Add cache invalidation to Delete**

In `internal/usecase/article/delete.go`, after the audit log call:

```go
// Invalidate caches
_ = uc.cache.Delete(ctx, uc.cacheKeys.ID(id))
_ = uc.cache.DeleteByPrefix(ctx, uc.cacheKeys.ListPrefix())
```

- [ ] **Step 3.5: Update all article test files**

All test files that call `article.New(mockArticleRepo, audit.NewNoop())` need to be updated to `article.New(mockArticleRepo, audit.NewNoop(), cache.NewNoop())` with `"go-boilerplate/pkg/cache"` import added.

Files to update:
- `internal/usecase/article/create_test.go`
- `internal/usecase/article/delete_test.go`
- `internal/usecase/article/update_test.go`
- `internal/usecase/article/list_test.go`
- `internal/usecase/article/get_by_id_test.go`

- [ ] **Step 3.6: Update app.go to pass cache to article usecase**

In `internal/app/app.go`, update `initUseCases` to accept and pass cache:

```go
func initUseCases(cfg *config.Config, repos *repositories, jwtService jwt.Service, asynqClient *asynq.Client, storageProvider storage.Provider, l logger.Interface, auditLogger audit.Logger, appCache cache.Cache) *usecases {
	// ...
	articleUC := article.New(repos.article, auditLogger, appCache)
	// ...
}
```

Update the call site in `Run()` to pass `appCache`.

- [ ] **Step 3.7: Run quality checks and commit**

Run: `make check-all`
Expected: All checks pass, coverage >= 84%.

```bash
git add pkg/cache/ internal/usecase/article/ internal/app/app.go
git commit -m "feat: add cache invalidation with DeleteByPrefix and CacheKeyBuilder (3.5)"
```

---

## Final Steps

- [ ] **Run full quality checks**

Run: `make check-all`
Expected: All checks pass, coverage >= 84%.

- [ ] **Update config.example.yaml with all new fields**

Ensure all new config fields from Tasks 1-5 are documented.

- [ ] **Final commit if needed**

```bash
git add -A
git commit -m "docs: update config.example.yaml with Phase 3 config fields"
```
