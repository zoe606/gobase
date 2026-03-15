# Phase 4a: Critical Fixes — Design Spec

**Date:** 2026-03-15
**Status:** Approved
**Scope:** 3 critical bugs in the article module and idempotency middleware

---

## Problem Statement

The article CRUD module has three critical bugs that make it non-functional and insecure:

1. **Create/Update field mapping is broken** — articles are saved with empty fields
2. **No ownership check on Update/Delete** — any authenticated user can modify/delete any article (IDOR vulnerability)
3. **Idempotency cache key is not scoped** — different users sharing the same Idempotency-Key get each other's cached responses

---

## Fix 1: Article Create/Update Field Mapping

### Current State

`internal/usecase/article/create.go` builds an `entity.Article` with only `UserID` set — all other fields from `CreateRequest` are ignored (TODO comment left in place). The created article has empty title, slug, content, etc.

`internal/usecase/article/update.go` fetches the existing article but never applies the `UpdateRequest` fields. It saves the unchanged entity back to the database (TODO comment + `_ = article` placeholder).

### DTO Cleanup

`internal/dto/article/request.go` changes:

1. **Remove `ViewCount`** from both `CreateRequest` and `UpdateRequest` — users should never set view counts (managed by the system).
2. **Make `Status` optional** in `CreateRequest` — remove `validate:"required"`, add `validate:"omitempty,oneof=draft published"`. Default to `"draft"` in the usecase.
3. **Change `PublishedAt`** in `CreateRequest` from `time.Time` (value, required) to `*time.Time` (pointer, optional) — articles start as drafts without a publish date.

**CreateRequest after fix:**
```go
type CreateRequest struct {
    Title        string     `json:"title" validate:"required"`
    Slug         string     `json:"slug" validate:"required"`
    Content      string     `json:"content" validate:"required"`
    Excerpt      string     `json:"excerpt" validate:"required"`
    CoverMediaID uint       `json:"cover_media_id" validate:"required"`
    Status       string     `json:"status" validate:"omitempty,oneof=draft published"`
    PublishedAt  *time.Time `json:"published_at"`
}
```

**UpdateRequest after fix** (only `ViewCount` removed):
```go
type UpdateRequest struct {
    Title        *string    `json:"title,omitempty"`
    Slug         *string    `json:"slug,omitempty"`
    Content      *string    `json:"content,omitempty"`
    Excerpt      *string    `json:"excerpt,omitempty"`
    CoverMediaID *uint      `json:"cover_media_id,omitempty"`
    Status       *string    `json:"status,omitempty"`
    PublishedAt  *time.Time `json:"published_at,omitempty"`
}
```

### Create Fix

Map all `CreateRequest` fields to `entity.Article`, handling the entity's pointer types:

```go
article := &entity.Article{
    UserID:       userID,
    Title:        req.Title,
    Slug:         req.Slug,
    Content:      &req.Content,
    Excerpt:      &req.Excerpt,
    CoverMediaID: &req.CoverMediaID,
    Status:       ptrOrDefault(req.Status, "draft"),
    PublishedAt:  req.PublishedAt,
}
```

A small `ptrOrDefault` helper returns a `*string` — if the input is empty, it returns a pointer to the default value. Place this helper in `internal/usecase/article/article.go` alongside the struct and constructor (per SOLID convention: shared helpers live in the main struct file).

Note: the `"draft"` default intentionally duplicates the entity's GORM `default:'draft'` tag. This ensures the correct value is set at the application level before the DB call, rather than relying solely on the database default.

### Update Fix

Apply non-nil pointer fields from `UpdateRequest` to the fetched entity:

```go
if req.Title != nil       { article.Title = *req.Title }
if req.Slug != nil        { article.Slug = *req.Slug }
if req.Content != nil     { article.Content = req.Content }
if req.Excerpt != nil     { article.Excerpt = req.Excerpt }
if req.CoverMediaID != nil { article.CoverMediaID = req.CoverMediaID }
if req.Status != nil      { article.Status = req.Status }
if req.PublishedAt != nil { article.PublishedAt = req.PublishedAt }
```

### Files Changed

| File | Action |
|------|--------|
| `internal/dto/article/request.go` | Remove `ViewCount` from Create/Update; change `PublishedAt` to `*time.Time`; make `Status` optional |
| `internal/usecase/article/article.go` | Add `ptrOrDefault` helper |
| `internal/usecase/article/create.go` | Map all request fields to entity |
| `internal/usecase/article/create_test.go` | Test field mapping, default status |
| `internal/usecase/article/update.go` | Apply non-nil fields from request |
| `internal/usecase/article/update_test.go` | Test partial update, no-op update |

---

## Fix 2: IDOR Ownership Check on Update/Delete

### Current State

`Update` and `Delete` accept only an article `id` — no `userID` parameter. Any authenticated user can modify or delete any article. The routes are behind `JWTAuth` middleware, so the user is authenticated, but authorization (ownership) is not checked.

### Interface Change

Add `userID uint` parameter to both methods in `internal/usecase/contracts.go`:

```go
Update(ctx context.Context, userID uint, id uint, req articledto.UpdateRequest) (*articledto.Response, error)
Delete(ctx context.Context, userID uint, id uint) error
```

**Ripple effects of this interface change:**
- All existing **usecase test cases** for Update and Delete must add a `userID` field to their test struct and pass it in the call (e.g., `uc.Update(ctx, userID, id, req)` instead of `uc.Update(ctx, id, req)`).
- The **handler mock** (`internal/handlers/http/v1/article/mocks_test.go`) must be updated to match the new signatures.
- All existing **handler test cases** for Update and Delete must update their mock expectations to include the `userID` parameter.

### Usecase Implementation — Update

`update.go` already fetches the article via `GetByID`. Add ownership check after the fetch:

```go
func (uc *UseCase) Update(ctx context.Context, userID uint, id uint, req articledto.UpdateRequest) (*articledto.Response, error) {
    article, err := uc.articleRepo.GetByID(ctx, id)
    // ... error handling (existing)

    if article.UserID != userID {
        return nil, ErrForbidden
    }

    // ... apply fields, save, audit, cache (existing + Fix 1 changes)
}
```

Pass `&userID` to `uc.auditLogger.LogUpdate` (currently `nil`). Also capture old values before mutation for the audit trail:

```go
oldValues := map[string]any{"title": article.Title, "slug": article.Slug}
// ... apply fields ...
_ = uc.auditLogger.LogUpdate(ctx, "article", article.ID, &userID, oldValues, map[string]any{"title": article.Title})
```

### Usecase Implementation — Delete

`delete.go` currently calls `articleRepo.Delete(ctx, id)` directly without fetching first. Must change to fetch-then-check-then-delete:

```go
func (uc *UseCase) Delete(ctx context.Context, userID uint, id uint) error {
    // Fetch article to verify ownership
    article, err := uc.articleRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, repo.ErrNotFound) {
            return ErrNotFound
        }
        return fmt.Errorf("article - Delete - articleRepo.GetByID: %w", err)
    }

    if article.UserID != userID {
        return ErrForbidden
    }

    if err := uc.articleRepo.Delete(ctx, id); err != nil {
        return fmt.Errorf("article - Delete - articleRepo.Delete: %w", err)
    }

    // Audit log (best-effort)
    _ = uc.auditLogger.LogDelete(ctx, "article", id, &userID, nil)

    // Invalidate caches
    _ = uc.cache.Delete(ctx, uc.cacheKeys.ID(id))
    _ = uc.cache.DeleteByPrefix(ctx, uc.cacheKeys.ListPrefix())

    return nil
}
```

Note: Delete tests must now mock both `GetByID` and `Delete` repo calls (currently only `Delete` is mocked).

### Error Definition

Add `ErrForbidden` to `internal/usecase/article/errors.go`:

```go
ErrForbidden = errors.New("not authorized to modify this article")
```

### Handler Changes

Both `update.go` and `delete.go` handlers extract `userID` from the JWT context via `middleware.GetUserID(ctx)` and pass it to the usecase. Map `ErrForbidden` to HTTP 403.

Also add missing Swagger annotations:
- Both handlers: add `@Failure 403 {object} response.ErrorResponse`
- `delete.go`: add missing `@Failure 401 {object} response.ErrorResponse` and `@Security BearerAuth` (pre-existing omission, fix alongside)

Run `make swag` after all handler annotation changes.

### Files Changed

| File | Action |
|------|--------|
| `internal/usecase/contracts.go` | Add `userID` param to `Update` and `Delete` |
| `internal/usecase/article/errors.go` | Add `ErrForbidden` |
| `internal/usecase/article/update.go` | Add `userID` param, ownership check, audit old values |
| `internal/usecase/article/update_test.go` | Update all existing tests with `userID` param; add ownership check tests (pass + forbidden) |
| `internal/usecase/article/delete.go` | Add `userID` param, fetch-then-check-then-delete pattern |
| `internal/usecase/article/delete_test.go` | Update all existing tests with `userID` param; mock `GetByID` + `Delete`; add ownership check tests |
| `internal/handlers/http/v1/article/update.go` | Extract userID, pass to usecase, handle 403, add Swagger `@Failure 403` |
| `internal/handlers/http/v1/article/delete.go` | Extract userID, pass to usecase, handle 403, add Swagger `@Failure 401/403` + `@Security BearerAuth` |
| `internal/handlers/http/v1/article/mocks_test.go` | Update mock `Update`/`Delete` signatures to include `userID` |
| `internal/handlers/http/v1/article/handler_test.go` | Update all existing Update/Delete mock expectations; add 403 test cases |

---

## Fix 3: Scope Idempotency Cache Key by Method+Path

### Current State

`internal/handlers/http/middleware/idempotency.go` line 48:

```go
cacheKey := "idempotency:" + key
```

The raw client-provided `Idempotency-Key` header becomes the cache key. Two problems:
- **Cross-user collision**: User A and User B sending the same key get each other's responses
- **Cross-endpoint collision**: Same key on `POST /articles` and `POST /auth/login` share a cache entry

### Middleware Ordering Constraint

The idempotency middleware is registered at the **global app level** in `router.go:76-78`, before any route-specific middleware. JWTAuth is applied per-route inside `RegisterRoutes()`. This means `ctx.Locals(middleware.UserIDKey)` is **not set** when idempotency runs — the user hasn't been authenticated yet.

### Fix

Scope the key by HTTP method + path. This eliminates cross-endpoint collision entirely. For cross-user collision: idempotency keys are client-generated UUIDs, so collision between different users is astronomically unlikely. The method+path scoping is the practical fix that works with the current middleware ordering.

```go
cacheKey := "idempotency:" + ctx.Method() + ":" + ctx.Path() + ":" + key
```

Use the `middleware.UserIDKey` constant (not a string literal) for the Locals lookup, keeping the door open for future per-user scoping if the middleware is ever moved to run after JWTAuth:

```go
// Best-effort user scoping (only works if JWTAuth ran before this middleware)
userPart := "shared"
if id, ok := ctx.Locals(middleware.UserIDKey).(uint); ok {
    userPart = strconv.FormatUint(uint64(id), 10)
}
cacheKey := "idempotency:" + userPart + ":" + ctx.Method() + ":" + ctx.Path() + ":" + key
```

This handles both the current global-level registration (falls back to "shared") and any future route-level registration (uses real user ID).

### Files Changed

| File | Action |
|------|--------|
| `internal/handlers/http/middleware/idempotency.go` | Scope cache key by user+method+path; use `UserIDKey` constant |
| `internal/handlers/http/middleware/idempotency_test.go` | Test cross-endpoint isolation; test with/without user ID in Locals |

---

## Testing Strategy

- **Unit tests** for all usecase changes (TDD — write failing tests first)
- **Handler tests** for 403 responses, updated mock signatures, and field validation
- **Middleware tests** for idempotency key scoping (cross-endpoint, with/without auth)
- All existing tests must continue to pass (signature changes propagate to all existing test cases)
- Run `make check-all` before every commit
- Run `make swag` after handler annotation changes

## Out of Scope

- OTel MeterProvider setup (Phase 4b)
- Article list validation missing call (Phase 4b)
- Email/password reset handler stubs (Phase 4b)
- Redis client consolidation (Phase 4b)
- Dockerfile non-root user (Phase 4b)
