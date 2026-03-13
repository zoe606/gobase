# Phase 3: Extensibility & Polish — Design Spec

**Goal:** Add multi-replica-ready infrastructure (Redis rate limiter, distributed locks, idempotency, structured logging, cache invalidation) so the boilerplate scales beyond a single instance with config-only changes.

**Parent spec:** `docs/superpowers/specs/2026-03-13-production-readiness-design.md` (Phase 3 section)

---

## 3.1 Redis-Backed Rate Limiter

**Problem:** Phase 2 wired `cfg.RateLimit.Store` and extracted `limiterCfg` in `router.go`, but only the in-memory backend exists. Multi-replica deploys share no state.

**Solution:**
- Create `pkg/ratelimiter/redis_store.go` implementing Fiber's `fiber.Storage` interface (`Get`, `Set`, `Delete`, `Reset`, `Close`) backed by `pkg/redis.Client`.
- In `router.go`, read `cfg.RateLimit.Store`; when `"redis"`, instantiate `RedisStore` and assign to `limiterCfg.Storage`.
- `redis.Client` must be created in `app.go` (currently only Asynq uses Redis). Create a shared `*redis.Client` when any feature needs it (`RateLimit.Store=redis`, `Lock.Provider=redis`, `Cache.Enabled=true`). Pass through to `SetupRoutes` and `initUseCases` as needed.

**Files:**
- Create: `pkg/ratelimiter/redis_store.go`
- Create: `pkg/ratelimiter/redis_store_test.go`
- Modify: `internal/handlers/http/router.go` — add Redis storage branch
- Modify: `internal/app/app.go` — create shared Redis client, pass to `initHTTPServer`

**Fiber Storage interface (from `github.com/gofiber/fiber/v2`):**
```go
type Storage interface {
    Get(key string) ([]byte, error)
    Set(key string, val []byte, exp time.Duration) error
    Delete(key string) error
    Reset() error
    Close() error
}
```

---

## 3.2 Distributed Lock Abstraction

**Problem:** `pkg/redis` has `SetNX` but no lock abstraction. Concurrent writes (e.g., two replicas processing the same webhook) can cause data corruption.

**Solution:**
- `pkg/lock/lock.go` — `Locker` interface: `Lock(ctx, key, ttl) (Unlocker, error)`, `TryLock(ctx, key, ttl) (Unlocker, bool, error)`
- `Unlocker` interface: `Unlock(ctx) error`
- `pkg/lock/noop.go` — `NoopLocker` (always succeeds, for single-instance mode)
- `pkg/lock/redis.go` — `RedisLocker` using `SetNX` + unique value (UUID) + Lua-script unlock for safety (check value matches before DEL)
- Config: `LOCK_PROVIDER=noop` (default) or `redis`

**Files:**
- Create: `pkg/lock/lock.go` — interface + errors
- Create: `pkg/lock/noop.go` — NoopLocker
- Create: `pkg/lock/noop_test.go`
- Create: `pkg/lock/redis.go` — RedisLocker
- Create: `pkg/lock/redis_test.go`
- Modify: `config/config.go` — add `Lock` config struct

---

## 3.3 Idempotency Key Middleware

**Problem:** No protection against duplicate POST/PUT from client retries (network timeouts, load balancer retries).

**Solution:**
- Middleware reads `Idempotency-Key` header on mutating requests (POST, PUT, PATCH).
- First request: execute handler, cache `{status, content_type, body}` in `pkg/cache` with configurable TTL (default 24h).
- Duplicate key: return cached response, skip handler execution.
- Missing key on POST: return 400 (configurable via `IDEMPOTENCY_REQUIRED_FOR_POST`).
- GET/DELETE/OPTIONS bypass entirely.
- Uses `pkg/cache.Cache` interface — works with both `NoopCache` (disabled) and `RedisCache`.

**Files:**
- Create: `internal/handlers/http/middleware/idempotency.go`
- Create: `internal/handlers/http/middleware/idempotency_test.go`
- Modify: `config/config.go` — add `Idempotency` config struct
- Modify: `internal/handlers/http/router.go` — register middleware (config-gated)

---

## 3.4 Structured Request/Response Logging

**Problem:** Current `middleware/logger.go` logs basic request info but lacks structured fields for production debugging (no latency, no request ID correlation, no log-level differentiation by status).

**Solution:**
- Rewrite `middleware/logger.go` to emit structured fields: `method`, `path`, `status`, `latency_ms`, `request_id`, `user_id` (from JWT context if present), `ip`, `user_agent`, `bytes_in`, `bytes_out`.
- Log level based on status code: 2xx→info, 4xx→warn, 5xx→error.
- Optional request body logging with PII redaction: strip JSON fields matching configurable patterns (default: `password`, `token`, `secret`, `authorization`, `credit_card`).
- Response body logging disabled by default.

**Files:**
- Modify: `internal/handlers/http/middleware/logger.go` — rewrite with structured fields
- Create: `internal/handlers/http/middleware/logger_test.go` — tests for redaction, log levels
- Modify: `config/config.go` — add logging config fields

---

## 3.5 Cache Invalidation Pattern

**Problem:** `pkg/cache` has get/set/delete but no pattern for prefix-based invalidation. Write operations leave stale cached data.

**Solution:**
- Add `DeleteByPrefix(ctx, prefix) error` to `Cache` interface. Note: `RedisCache.prefixKey()` prepends the configured prefix — `DeleteByPrefix` must account for this (prepend the cache prefix before scanning).
- `RedisCache`: implement with `SCAN` + `DEL` (cursor-based, safe for production).
- `NoopCache`: return nil.
- Add `CacheKeyBuilder` for consistent key construction: `entity:scope:id` pattern.
- Reference implementation in article usecase: invalidate `article:list:*` on create/update/delete, invalidate `article:{id}` on update/delete.
- **Interface change:** Update all implementations and mocks simultaneously.

**Files:**
- Modify: `pkg/cache/cache.go` — add `DeleteByPrefix` to interface
- Modify: `pkg/cache/redis.go` — SCAN+DEL implementation
- Modify: `pkg/cache/noop.go` — noop implementation
- Create: `pkg/cache/keys.go` — CacheKeyBuilder
- Create: `pkg/cache/keys_test.go`
- Modify: `internal/usecase/article/article.go` — add cache field
- Modify: `internal/usecase/article/create.go`, `update.go`, `delete.go` — invalidation calls
- Update any mock Cache implementations in test files

---

## Dependency Order

```
3.4 Structured logging     (independent, do first — improves debugging for rest)
3.1 Redis rate limiter      (independent)
3.2 Distributed lock        (independent)
3.3 Idempotency middleware  (independent, but benefits from cache being ready)
3.5 Cache invalidation      (independent, do last — touches cache interface)
```

Recommended execution order: **3.4 → 3.1 → 3.2 → 3.3 → 3.5**

---

## Config Additions Summary

```yaml
# Logging (3.4)
log:
  log_request_body: false      # LOG_REQUEST_BODY
  log_response_body: false     # LOG_RESPONSE_BODY
  redact_fields: "password,token,secret,authorization,credit_card"  # LOG_REDACT_FIELDS

# Lock (3.2)
lock:
  provider: "noop"             # LOCK_PROVIDER (noop|redis)

# Idempotency (3.3)
idempotency:
  enabled: false               # IDEMPOTENCY_ENABLED
  ttl: "24h"                   # IDEMPOTENCY_TTL
  required_for_post: false     # IDEMPOTENCY_REQUIRED_FOR_POST
```

---

## Exit Criteria

- All new features have unit tests
- `make check-all` passes (coverage >= 84%)
- New config fields documented in `config.example.yaml`
- Each feature is config-gated (disabled by default, zero behavior change for existing users)
