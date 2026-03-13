# Production Readiness Plan — Go Boilerplate

**Date:** 2026-03-13
**Status:** Approved
**Goal:** Make the gobase boilerplate production-ready for single-instance deployment and open-source publication, with plug-and-play horizontal scaling.

## Context

The gobase boilerplate scores 8.2/10 in a comprehensive review. Architecture, testing (85% coverage), observability (full OTel), and developer tooling (custom codegen/wire) are strong. However, 18 issues across security, production hardening, and extensibility must be addressed before real deployment or public release.

**Constraints:**
- Single-instance deployment first; horizontal scaling must be plug-and-play (interface-first, swap implementations via config)
- Both personal production use and open-source community template
- Risk-first phasing: fix security/correctness first, then harden, then polish

## Phase 1: Security & Correctness

Fix bugs that would cause real damage in production. No new features.

### 1.1 UserID from JWT Context

**Problem:** `articledto.CreateRequest` has `UserID` as a required body field (`validate:"required"`), and `articledto.UpdateRequest` has `UserID *uint`. Both are IDOR vulnerabilities — any authenticated user can create or update articles on behalf of others.

**Solution:**
- Remove `UserID` field from both `articledto.CreateRequest` and `articledto.UpdateRequest`
- In `article/create.go` and `article/update.go` handlers, extract `UserID` from JWT claims via `middleware.GetUserID(c)`
- Pass `UserID` as a separate parameter to the usecase, or set it on the entity in the handler before calling the usecase
- Update validation tags accordingly
- Update article create/update handler tests and usecase tests

**Files:**
- `internal/dto/article/request.go` — remove `UserID` field from both `CreateRequest` and `UpdateRequest`
- `internal/handlers/http/v1/article/create.go` — extract UserID from context
- `internal/handlers/http/v1/article/update.go` — extract UserID from context
- `internal/usecase/article/create.go` — adjust signature if needed
- `internal/usecase/article/update.go` — adjust signature if needed
- Tests for both handlers and usecases

### 1.2 Auth Middleware on Article Write Routes

**Problem:** `article/handler.go` registers all CRUD routes without `JWTAuth`. Writes are publicly accessible.

**Solution:**
- Update `articlehandler.New()` to accept `jwtService jwt.Service` and `l logger.Interface` as parameters, store both on the struct
- Split article route registration: public group for `GET` (list, get by ID), protected group for `POST/PUT/DELETE`
- Apply `middleware.JWTAuth(h.jwtService, h.l)` to the protected group (the middleware requires both arguments)
- Keep read endpoints public (suitable for a blog/CMS API)

**Note:** 1.2 should be implemented before 1.1 so that the JWT middleware is in place before the handler stops reading `UserID` from the request body. After 1.2, `middleware.GetUserID(c)` will reliably return a non-zero value for protected routes.

**Files:**
- `internal/handlers/http/v1/article/handler.go` — update `New()` signature, split route groups
- `internal/app/app.go` — pass `jwtService` and `l` when constructing the article handler
- `internal/handlers/http/router.go` — update article handler construction if routes are registered there
- `internal/handlers/http/v1/article/handler_test.go` — add tests for unauthenticated write rejection

### 1.3 Fix Dockerfile Go Version

**Problem:** Dockerfile pins `golang:1.25-alpine3.21` but `go.mod` declares `go 1.26.1`. CI uses 1.26.1 correctly.

**Solution:**
- Update Dockerfile base image to match the Go toolchain version in `go.mod`. Use `golang:1.26-alpine` (Go does not publish patch-level Docker images for all versions — verify the exact available tag on Docker Hub before committing)
- Verify the alpine version supports the target Go version

**Files:**
- `deployment/docker/Dockerfile` — update `FROM` line

### 1.4 Align Postgres Versions

**Problem:** CI uses Postgres 16-alpine, local Docker uses 17-alpine.

**Solution:**
- Pin both to Postgres 17-alpine (the newer version)
- Update CI workflow service image

**Files:**
- `.github/workflows/ci.yml` — update Postgres service image to `postgres:17-alpine`

### 1.5 Make Integration Tests Blocking

**Problem:** Integration test CI job uses `|| true`, meaning failures never block merges.

**Solution:**
- Remove `|| true` from the integration test command
- Keep integration tests in a separate job that is required by the `ci-success` aggregator

**Files:**
- `.github/workflows/ci.yml` — remove `|| true` from integration test step

### 1.6 Global Request Body Size Limit

**Problem:** No middleware-level body size limit. Fiber default is 4MB, but this is implicit, not explicit.

**Solution:**
- Add `BodyLimit` to `fiber.Config` in the HTTP server initialization, driven by `cfg.HTTP.BodyLimit` (default: `4MB`)
- Add the config field to `config.go`

**Note:** `pkg/httpserver/server.go` receives config via the `Option` functional options pattern (same as `ShutdownTimeout`, `ReadTimeout`, etc.). A new `BodyLimit` option must be added to `options.go`, a `bodyLimit` field added to the `Server` struct, and then set in `fiber.Config{BodyLimit: s.bodyLimit}`.

**Files:**
- `config/config.go` — add `BodyLimit` field to `HTTP` config struct
- `config/config.yaml` (or equivalent) — add default
- `pkg/httpserver/options.go` — add `BodyLimit(n int) Option` function
- `pkg/httpserver/server.go` — add `bodyLimit` field to `Server` struct, set in `fiber.Config{BodyLimit: s.bodyLimit}`
- `internal/app/app.go` — pass `httpserver.BodyLimit(cfg.HTTP.BodyLimit)` in `initHTTPServer()`

### Phase 1 Exit Criteria

- No security holes in existing features
- `make check-all` passes
- CI green
- All existing tests pass with modifications
- New tests cover auth on article routes and UserID extraction

---

## Phase 2: Production Hardening

Make the boilerplate safe to deploy and operate. All scaling concerns go behind interfaces with in-memory defaults.

### 2.1 Rate Limiter Behind Interface

**Problem:** Fiber rate limiter uses in-memory storage. Breaks under horizontal scaling.

**Design:**
- Fiber's limiter already accepts a `fiber.Storage` interface. Create a config option `RATE_LIMITER_STORE=memory` (default) that uses Fiber's built-in memory store.
- Document that setting `RATE_LIMITER_STORE=redis` (implemented in Phase 3) swaps to Redis-backed storage.
- The interface is already there (Fiber's `fiber.Storage`); we just need the config wiring and documentation.

**Files:**
- `config/config.go` — add `Store` field to rate limiter config
- `internal/handlers/http/router.go` — select storage backend based on config
- Document the swap mechanism in config comments

### 2.2 Decouple Repo Contracts from DTOs

**Problem:** `repo/contracts.go` imports `dto/article` and `dto/translation` for query parameters. This bleeds presentation concerns into the repo layer.

**Solution:**
- Create `internal/repo/params.go` with `ArticleListParams` and `TranslationHistoryParams` structs
- Update `repo/contracts.go` to use the new param types
- Update repo implementations to use the new types
- Update usecases to map from DTOs to repo params
- Run `go generate ./...` after the interface change to regenerate mocks (the `//go:generate` directive in `contracts.go` generates `internal/usecase/mocks_repo_test.go` — stale mocks will fail to compile)

**Files:**
- `internal/repo/params.go` — new file with query param structs
- `internal/repo/contracts.go` — replace DTO imports
- `internal/repo/persistent/article.go` — update method signatures
- `internal/repo/persistent/translation.go` — update method signatures
- `internal/usecase/article/list.go` — map DTO to repo params
- `internal/usecase/translation/history.go` — map DTO to repo params
- `internal/usecase/mocks_repo_test.go` — regenerated by `go generate`
- Update affected tests

### 2.3 Wire AuditLog

**Problem:** AuditLog entity, migration, and `pkg/audit` exist but are not wired. The `audit_logs` table migration (`000007_create_audit_logs`) already exists, but the entity is missing from `AutoMigrate`'s list and no usecase calls the audit logger.

**Solution:**
- **2.7 must be completed before 2.3.** Once AutoMigrate is replaced by golang-migrate on startup, the existing migration `000007_create_audit_logs` will create the table. If for any reason 2.3 is done before 2.7, `entity.AuditLog` must be temporarily added to the `AutoMigrate` list in `app.go` to avoid a runtime panic when writing to a non-existent table in development. Verify the migration file exists and is correct.
- Wire `audit.Logger` (Postgres implementation) into the DI container
- Add `audit.Logger` interface to usecase constructors that need it (auth, article)
- Add audit logging to auth usecase: login success, login failure, register, password change
- Add audit logging to article usecase: create, update, delete
- Expose an admin endpoint to query audit logs (optional, lower priority)

**Files:**
- `internal/app/app.go` — wire audit logger, pass to usecases
- `internal/usecase/auth/auth.go` — add `auditLogger` field and builder method
- `internal/usecase/auth/login.go`, `register.go` — add audit calls
- `internal/usecase/article/article.go` — add `auditLogger` field
- `internal/usecase/article/create.go`, `update.go`, `delete.go` — add audit calls
- `migrations/000007_create_audit_logs.up.sql` — verify completeness
- Tests for audit logging behavior

### 2.4 RS256/ES256 JWT Option

**Problem:** HS256 is symmetric. Limits multi-service trust delegation.

**Solution:**
- Extend `pkg/jwt` to support algorithm selection via config: `JWT_ALGORITHM=hs256|rs256|es256`
- HS256 remains the default (no breaking change)
- For RS256/ES256: load key files from `JWT_PRIVATE_KEY_PATH` and `JWT_PUBLIC_KEY_PATH`
- The `jwt.Service` interface stays the same; only the internal signing/verification changes
- Replace the current `New(secretKey, accessExpiry, refreshExpiry)` constructor with a factory function that switches on algorithm: `NewFromConfig(cfg config.JWT)` or use functional options. The current `New()` can remain for HS256 backward compatibility, with a new `NewRS256(privateKeyPath, publicKeyPath, accessExpiry, refreshExpiry)` constructor
- Add config validation: if algorithm is asymmetric, key paths must be set

**Files:**
- `config/config.go` — add `Algorithm`, `PrivateKeyPath`, `PublicKeyPath` to JWT config
- `pkg/jwt/jwt.go` — support multiple signing methods, add new constructor(s)
- `pkg/jwt/jwt_test.go` — add tests for RS256/ES256
- `config/config.go` `Validate()` — add key-path validation for asymmetric algorithms
- `internal/app/app.go` — update `initJWT()` to select constructor based on config algorithm

### 2.5 Configurable Shutdown Timeout

**Problem:** `ShutdownWithTimeout(3s)` is hardcoded. Too short for production.

**Solution:**
- Add `ShutdownTimeout` to `HTTP` config (default: 15s)
- Pass it to `ShutdownWithTimeout()` in the app bootstrap

**Note:** `pkg/httpserver/options.go` already defines a `ShutdownTimeout` option function and `server.go` has `_defaultShutdownTimeout = 3 * time.Second`. The option exists but is not passed from config. The fix is simpler than adding new code to the httpserver package.

**Files:**
- `config/config.go` — add `ShutdownTimeout` field to `HTTP` config struct
- `internal/app/app.go` — pass `httpserver.ShutdownTimeout(cfg.HTTP.ShutdownTimeout)` in `initHTTPServer()`

### 2.6 Expand Validation Error Mappings

**Problem:** `ParseValidationErrors()` only maps 4 tags. Others become "Invalid value".

**Solution:**
- Add mappings for: `oneof`, `url`, `uuid`, `gte`, `lte`, `len`, `alphanum`, `numeric`, `boolean`, `datetime`
- Use the validator's `FieldError.Param()` to include allowed values in messages (e.g., "Must be one of: draft, published")

**Files:**
- `internal/handlers/http/v1/helper.go` — expand the switch/map
- Add tests for new validation tag messages

### 2.7 Disable AutoMigrate, Use Migrations Only

**Problem:** Two schema management paths (AutoMigrate in dev, golang-migrate in prod) can drift.

**Solution:**
- Remove GORM `AutoMigrate` call from `app.go`
- In development mode, run `golang-migrate` up automatically on startup (replacing AutoMigrate)
- This means migrations are the single source of truth in all environments
- Keep the seeder, but trigger it after migrations complete (check if roles exist before seeding)

**Files:**
- `internal/app/app.go` — replace AutoMigrate with migrate-up call
- May need to import `golang-migrate` as a library or shell out to the binary
- `internal/app/seeder.go` — adjust trigger mechanism

### Phase 2 Exit Criteria

- All interfaces have in-memory/default implementations
- Config documentation covers every new field
- `make check-all` passes
- Existing tests still green
- New tests cover: RS256 JWT, expanded validation, audit logging, rate limiter config switching

---

## Phase 3: Extensibility & Polish

Features that make the boilerplate stand out and enable multi-service/multi-replica deployments.

### 3.1 Redis-Backed Rate Limiter

**Problem:** Phase 2 created the config switch but only the memory backend exists.

**Solution:**
- Implement a `RedisStore` that satisfies Fiber's `fiber.Storage` interface using `pkg/redis`
- Config switch: `RATE_LIMITER_STORE=redis` activates it
- Plug-and-play: no code changes needed, just config

**Files:**
- `pkg/ratelimiter/redis_store.go` — new file implementing `fiber.Storage` with Redis
- `pkg/ratelimiter/redis_store_test.go` — tests
- `internal/handlers/http/router.go` — add Redis case to the switch
- Update config documentation

### 3.2 Distributed Lock Abstraction

**Problem:** `pkg/redis` has `SetNX` but no lock abstraction for concurrent write protection.

**Solution:**
- Create `pkg/lock/lock.go` with interface: `Lock(ctx, key, ttl) error`, `Unlock(ctx, key) error`, `WithLock(ctx, key, ttl, fn) error`
- `RedisLock` implementation using `SetNX` + TTL + unique value for safe unlock (check-and-delete via Lua script)
- `NoopLock` for single-instance mode (always succeeds)
- Config switch: `LOCK_PROVIDER=noop|redis`

**Files:**
- `pkg/lock/lock.go` — interface + NoopLock
- `pkg/lock/redis.go` — RedisLock implementation
- `pkg/lock/redis_test.go` — tests
- `config/config.go` — add lock provider config

### 3.3 Idempotency Key Middleware

**Problem:** No protection against duplicate POST/PUT requests from client retries.

**Solution:**
- New middleware that reads `Idempotency-Key` header on non-GET requests
- On first request: execute handler, cache response (status + body) in `pkg/cache` with configurable TTL (default: 24h)
- On duplicate key: return cached response without executing handler
- Missing key on POST: return 400 (configurable: can be made optional)
- Uses `pkg/cache` interface, so it works with both `NoopCache` (disabled) and `RedisCache`

**Files:**
- `internal/handlers/http/middleware/idempotency.go` — middleware implementation
- `internal/handlers/http/middleware/idempotency_test.go` — tests
- `config/config.go` — add idempotency config (enabled, TTL, required-for-post)
- `internal/handlers/http/router.go` — register middleware (config-gated)

### 3.4 Structured Request/Response Logging

**Problem:** Logger middleware logs requests but lacks structured fields for production debugging.

**Solution:**
- Extend `middleware/logger.go` to emit structured fields: `method`, `path`, `status`, `latency_ms`, `request_id`, `user_id` (from JWT if present), `ip`, `user_agent`
- Add configurable request body logging with PII redaction: strip fields matching patterns (`password`, `token`, `secret`, `authorization`)
- Add response body logging option (disabled by default, enable for debugging)
- Log level based on status: 2xx→info, 4xx→warn, 5xx→error

**Files:**
- `internal/handlers/http/middleware/logger.go` — rewrite with structured fields
- `config/config.go` — add logging config (log_body, redact_fields)
- Tests for redaction behavior

### 3.5 Cache Invalidation Pattern

**Problem:** `pkg/cache` has get/set/delete but no pattern for invalidating stale data on writes.

**Solution:**
- Add `DeleteByPrefix(ctx, prefix) error` to the `Cache` interface (uses Redis `SCAN` + `DEL`, noop returns nil)
- **Breaking interface change:** Adding a method to the `Cache` interface requires updating ALL implementations simultaneously (currently: `NoopCache` in `noop.go`, `RedisCache` in `redis.go`). Before implementing, grep for `cache.Cache` in test files to find any hand-written or generated mock implementations that also need updating.
- Document the write-through pattern: after a successful write to DB, delete the relevant cache keys
- Implement as a reference in the article usecase: `article:list:*` invalidated on create/update/delete, `article:{id}` invalidated on update/delete
- Add a `CacheKeyBuilder` helper for consistent key construction

**Files:**
- `pkg/cache/cache.go` — add `DeleteByPrefix` to interface
- `pkg/cache/redis.go` — implement with SCAN+DEL
- `pkg/cache/noop.go` — noop implementation
- Any test files with mock `Cache` implementations — update to satisfy new interface
- `pkg/cache/keys.go` — key builder helper
- `internal/usecase/article/create.go`, `update.go`, `delete.go` — add cache invalidation calls
- Tests

### Phase 3 Exit Criteria

- All new features have tests (unit + handler-level)
- Swagger docs updated for new middleware behaviors
- Config docs cover every scaling switch (`RATE_LIMITER_STORE`, `LOCK_PROVIDER`, idempotency)
- README documents the scaling story: what to change when going from single-instance to multi-replica
- `make check-all` passes

---

## Dependency Graph

```
Phase 1 (mostly independent)
  ├── 1.2 Auth on article routes (do FIRST — adds JWT middleware to write routes)
  ├── 1.1 UserID from JWT (depends on 1.2 — GetUserID(c) requires JWT middleware in place)
  ├── 1.3 Dockerfile version (independent)
  ├── 1.4 Postgres version (independent)
  ├── 1.5 Integration tests blocking (independent)
  └── 1.6 Body size limit (independent)

Phase 2 (some dependencies)
  ├── 2.1 Rate limiter interface (independent)
  ├── 2.2 Decouple repo/DTO (independent)
  ├── 2.3 Wire AuditLog (depends on 2.7 — must be done after AutoMigrate removal)
  ├── 2.4 RS256 JWT (independent)
  ├── 2.5 Shutdown timeout (independent)
  ├── 2.6 Validation errors (independent)
  └── 2.7 Disable AutoMigrate (independent, but do before 2.3)

Phase 3 (some dependencies)
  ├── 3.1 Redis rate limiter (depends on 2.1)
  ├── 3.2 Distributed lock (independent)
  ├── 3.3 Idempotency middleware (independent, but benefits from 3.2 for dedup)
  ├── 3.4 Structured logging (independent)
  └── 3.5 Cache invalidation (independent)
```

## Non-Goals

- Full rewrite of any existing layer
- Adding new domain features (articles, media, etc.)
- Kubernetes manifests or Helm charts
- Frontend/admin UI
- Database sharding or read replicas
