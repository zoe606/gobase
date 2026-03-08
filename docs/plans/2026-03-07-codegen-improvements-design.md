# Code Generator Improvements Design

**Date:** 2026-03-07
**Status:** Approved
**Goal:** Reduce manual work when adding new features — generate near-production-ready code and auto-wire DI/routes.

## Problem

After running `make gen-full`, developers still spend 15-20 minutes per feature:
1. Editing generated code (missing swagger, validation, error handling, tests)
2. Manually wiring contracts in `usecase/contracts.go` and `repo/contracts.go`
3. Manually registering routes in `router.go`
4. Manually wiring DI in `app.go`

## Solution: Approach C — Migration-Driven Generator + Auto-Wiring

Two improvements:
1. **Better templates** — `make gen-full` produces near-final code
2. **New `make wire` command** — Auto-appends DI, routes, and contracts

## Improved Generator Templates

### Entity (no change)
Already generates correct GORM structs.

### DTO
- `package <name>dto` convention (matches recent refactor)
- `ListRequest` embeds `pagination.Params`
- Proper `omitempty` on optional response fields

### Repository
- Interface with standard CRUD signatures
- GORM implementation with:
  - `repo.ErrNotFound` sentinel error handling
  - Pagination support in `List()` using `pagination.Params`
  - Proper error wrapping
- Optional soft delete support via `OPTS=soft-delete` flag

### Usecase
- SOLID file structure (1 file per method)
- `errors.go` with domain-specific errors (ErrNotFound, ErrInvalid)
- Each method file has proper error handling with `errors.Is()`
- Test file per method with table-driven tests and mock setup
- `mocks_test.go` with mock repo implementation

### Handler
- SOLID file structure (1 file per handler method)
- Full swagger annotations (tags, params, success/failure responses)
- Request validation with `v1.ParseValidationErrors()`
- Proper HTTP status mapping (201 for create, 204 for delete)
- Auth middleware placeholder in `RegisterRoutes()`
- `handler_test.go` with test setup
- `mocks_test.go` with mock usecase implementation

### Generator Flags

```bash
make gen-full MIGRATION=000012_create_orders.up.sql              # standard CRUD
make gen-full MIGRATION=000012_create_orders.up.sql OPTS=soft-delete  # with soft delete
```

Supported opts:
- `soft-delete` — Adds `DeletedAt gorm.DeletedAt` handling, uses `Unscoped()` where needed

Future opts (out of scope for now):
- `audit` — Adds audit log hooks
- `cache` — Adds cache-aside pattern in usecase

## The `make wire` Command

### What It Does

Scans generated packages and appends missing entries to 4 files:

| File | What Gets Appended |
|------|--------------------|
| `internal/repo/contracts.go` | Repo interface |
| `internal/usecase/contracts.go` | Usecase interface |
| `internal/handlers/http/router.go` | Import + handler registration in `setupAPIRoutes()` |
| `internal/app/app.go` | Repo, usecase, handler instantiation |

### How It Works

1. Scan `internal/usecase/*/` for packages with `New()` constructors
2. Scan `internal/repo/persistent/` for repo implementations
3. Parse target files using `go/parser` AST
4. Compare: extract existing interface/import names
5. Append only what's missing

### Safety

- **Idempotent** — Running twice won't duplicate entries
- **Append-only** — Never modifies existing code
- **Prints changes** — Shows what was added for review
- **Graceful fallback** — If AST parsing fails, prints warning and skips

### Implementation

- Located at `pkg/codegen/cmd/wire/main.go`
- Uses Go standard library: `go/parser`, `go/ast`, `go/printer`
- No external dependencies

## End-to-End Workflow

```bash
# 1. Create migration
make migrate-create name=create_orders

# 2. Edit the .up.sql and .down.sql files

# 3. Run migration
POSTGRES_URL="..." make migrate-up

# 4. Generate all layers
make gen-full MIGRATION=000012_create_orders.up.sql

# 5. Auto-wire DI, routes, contracts
make wire

# 6. Verify
make check-all

# 7. Regenerate swagger
make swag
```

## What You Still Customize

- Business logic beyond basic CRUD
- Custom validation rules
- Authorization rules (which roles access which endpoints)
- Custom query filters
- Relationship handling

## Out of Scope

- Foreign key / relationship handling in generated code
- Auto-generating migrations from Go structs
- Multi-tenancy scaffolding
- Interactive CLI wizard

## Generated File Structure (per feature)

```
internal/entity/<name>.go
internal/dto/<name>/request.go, response.go
internal/repo/persistent/<name>.go
internal/usecase/<name>/
    <name>.go, errors.go
    create.go, create_test.go
    get_by_id.go, get_by_id_test.go
    list.go, list_test.go
    update.go, update_test.go
    delete.go, delete_test.go
    mocks_test.go
internal/handlers/http/v1/<name>/
    handler.go
    create.go, get_by_id.go, list.go, update.go, delete.go
    handler_test.go, mocks_test.go
```
