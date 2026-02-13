# Master Bank CRUD - Backend Implementation Plan

## Context

The bank-statement feature currently treats banks as read-only reference data (seeded BCA/BRI). There's no API to manage banks, add new ones, or set default PDF passwords. We need CRUD endpoints for banks and move the existing `GET /v1/banks` from the bankstatement handler to a new dedicated bank handler.

## Current State

- `BankRepo` is read-only: `List`, `GetByID`, `GetByCode` (in `internal/repo/contracts.go` lines 113-118)
- `ListBanks` is embedded in the BankStatement usecase (in `internal/usecase/contracts.go` line 106)
- `GET /v1/banks/` route lives in bankstatement handler (in `internal/handlers/http/v1/bankstatement/handler.go` lines 45-49)
- Entity `Bank` has fields: ID, Name, Code, DefaultPassword (*string), timestamps, soft delete

---

## Step 1: Expand BankRepo Interface

**File: `internal/repo/contracts.go`** — Add `Create`, `Update`, `Delete` to BankRepo:

```go
BankRepo interface {
    Create(ctx context.Context, bank *entity.Bank) error
    GetByID(ctx context.Context, id uint) (*entity.Bank, error)
    GetByCode(ctx context.Context, code string) (*entity.Bank, error)
    List(ctx context.Context) ([]*entity.Bank, error)
    Update(ctx context.Context, bank *entity.Bank) error
    Delete(ctx context.Context, id uint) error
}
```

**File: `internal/repo/persistent/bank.go`** — Implement Create, Update, Delete:

```go
func (r *BankRepo) Create(ctx context.Context, bank *entity.Bank) error {
    db := tx.DBFromContext(ctx, r.db)
    return db.Create(bank).Error
}

func (r *BankRepo) Update(ctx context.Context, bank *entity.Bank) error {
    db := tx.DBFromContext(ctx, r.db)
    return db.Save(bank).Error
}

func (r *BankRepo) Delete(ctx context.Context, id uint) error {
    db := tx.DBFromContext(ctx, r.db)
    return db.Delete(&entity.Bank{}, id).Error
}
```

---

## Step 2: Create Bank DTOs

**Create `internal/dto/bank/request.go`:**

```go
package bank

type CreateRequest struct {
    Name            string  `json:"name" validate:"required,min=2,max=100"`
    Code            string  `json:"code" validate:"required,min=2,max=20,uppercase"`
    DefaultPassword *string `json:"default_password,omitempty" validate:"omitempty,min=1,max=255"`
}

type UpdateRequest struct {
    Name            *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
    Code            *string `json:"code,omitempty" validate:"omitempty,min=2,max=20,uppercase"`
    DefaultPassword *string `json:"default_password,omitempty" validate:"omitempty,min=1,max=255"`
}
```

**Create `internal/dto/bank/response.go`:**

```go
package bank

type Response struct {
    ID          uint      `json:"id"`
    Name        string    `json:"name"`
    Code        string    `json:"code"`
    HasPassword bool      `json:"has_password"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type DetailResponse struct {
    ID              uint      `json:"id"`
    Name            string    `json:"name"`
    Code            string    `json:"code"`
    DefaultPassword *string   `json:"default_password,omitempty"`
    HasPassword     bool      `json:"has_password"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

type ListResponse struct {
    Banks []*Response `json:"banks"`
}

// Factory methods: NewResponse, NewDetailResponse, NewListResponse
// NewResponse hides actual password, just shows has_password bool
// NewDetailResponse includes password (for edit forms)
```

---

## Step 3: Create Bank Usecase

**Add to `internal/usecase/contracts.go`:**

```go
import bankdto "go-boilerplate/internal/dto/bank"

Bank interface {
    Create(ctx context.Context, req bankdto.CreateRequest) (*bankdto.Response, error)
    GetByID(ctx context.Context, id uint) (*bankdto.DetailResponse, error)
    List(ctx context.Context) (*bankdto.ListResponse, error)
    Update(ctx context.Context, id uint, req bankdto.UpdateRequest) (*bankdto.Response, error)
    Delete(ctx context.Context, id uint) error
}
```

**Remove `ListBanks` from BankStatement interface** (line 106).

**Create files in `internal/usecase/bank/`:**

- `bank.go` — UseCase struct with bankRepo dependency + constructor
- `errors.go` — `ErrBankNotFound`, `ErrDuplicateBankCode`
- `create.go` — Check duplicate code via `bankRepo.GetByCode`, then `bankRepo.Create`
- `list.go` — `bankRepo.List()` → `NewListResponse`
- `get_by_id.go` — `bankRepo.GetByID()` → `NewDetailResponse` (includes password)
- `update.go` — Partial update with pointer fields, duplicate code check when changing code
- `delete.go` — `bankRepo.Delete()`

**Delete `internal/usecase/bankstatement/list_banks.go`** — logic moved to bank usecase.

---

## Step 4: Create Bank Handler

**Create files in `internal/handlers/http/v1/bank/`:**

- `handler.go` — Handler struct + RegisterRoutes:
  ```
  GET    /v1/banks/     RequireAnyPermission("banks:read", "bank-statement:read")  ← backward compat
  POST   /v1/banks/     RequirePermission("banks:write")
  GET    /v1/banks/:id  RequirePermission("banks:read")
  PUT    /v1/banks/:id  RequirePermission("banks:write")
  DELETE /v1/banks/:id  RequirePermission("banks:delete")
  ```
- `create.go` — POST handler with validation
- `list.go` — GET list handler
- `get_by_id.go` — GET by ID handler
- `update.go` — PUT handler with validation
- `delete.go` — DELETE handler

**Modify `internal/handlers/http/v1/bankstatement/handler.go`:**
- Remove `/banks` route group (lines 45-49)

**Delete `internal/handlers/http/v1/bankstatement/list_banks.go`** — endpoint moved.

---

## Step 5: Wire Up DI

**Modify `internal/app/app.go`:**
- Add `bank usecase.Bank` to `usecases` struct
- In `initUseCases`: `bankUC := bankuc.New(repos.bank)`
- In `initHTTPServer`: pass `uc.bank` to SetupRoutes

**Modify `internal/handlers/http/router.go`:**
- Add import: `bankhandler "go-boilerplate/internal/handlers/http/v1/bank"`
- Add `bankUC usecase.Bank` param to `SetupRoutes` and `setupAPIRoutes`
- Register: `bankhandler.New(bankUC, jwtService, l).RegisterRoutes(apiV1Group)`

---

## Step 6: Permissions

**Modify `internal/app/seeder.go`:**
- Add to `defaultPermissions`:
  ```go
  {Name: "banks:read", Resource: "banks", Action: "read"},
  {Name: "banks:write", Resource: "banks", Action: "write"},
  {Name: "banks:delete", Resource: "banks", Action: "delete"},
  ```
- Add to admin role permissions list: `"banks:read", "banks:write", "banks:delete"`

**Create `migrations/000018_seed_bank_permissions.up.sql`:**
```sql
INSERT INTO permissions (name, resource, action, created_at, updated_at)
VALUES
  ('banks:read', 'banks', 'read', NOW(), NOW()),
  ('banks:write', 'banks', 'write', NOW(), NOW()),
  ('banks:delete', 'banks', 'delete', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;
```

**Create `migrations/000018_seed_bank_permissions.down.sql`:**
```sql
DELETE FROM permissions WHERE name IN ('banks:read', 'banks:write', 'banks:delete');
```

---

## Step 7: Regenerate Mocks & Fix Tests

- Run `go generate ./...` (BankRepo and BankStatement interfaces changed)
- Fix bankstatement handler test — remove `TestHandler_ListBanks` test
- Fix bankstatement usecase — remove `list_banks` references from mock expectations

---

## Files Summary

### CREATE (17 files)
| File | Purpose |
|------|---------|
| `internal/dto/bank/request.go` | Create/Update request DTOs |
| `internal/dto/bank/response.go` | Response/DetailResponse/ListResponse DTOs |
| `internal/usecase/bank/bank.go` | UseCase struct + constructor |
| `internal/usecase/bank/errors.go` | Error definitions |
| `internal/usecase/bank/create.go` | Create with duplicate check |
| `internal/usecase/bank/list.go` | List all banks |
| `internal/usecase/bank/get_by_id.go` | Get bank detail (with password) |
| `internal/usecase/bank/update.go` | Partial update |
| `internal/usecase/bank/delete.go` | Soft delete |
| `internal/handlers/http/v1/bank/handler.go` | Handler + routes |
| `internal/handlers/http/v1/bank/create.go` | POST handler |
| `internal/handlers/http/v1/bank/list.go` | GET list handler |
| `internal/handlers/http/v1/bank/get_by_id.go` | GET by ID handler |
| `internal/handlers/http/v1/bank/update.go` | PUT handler |
| `internal/handlers/http/v1/bank/delete.go` | DELETE handler |
| `migrations/000018_seed_bank_permissions.up.sql` | Permission seed |
| `migrations/000018_seed_bank_permissions.down.sql` | Rollback |

### MODIFY (7 files)
| File | Change |
|------|--------|
| `internal/repo/contracts.go` | Add Create/Update/Delete to BankRepo |
| `internal/repo/persistent/bank.go` | Implement Create/Update/Delete |
| `internal/usecase/contracts.go` | Add Bank interface; remove ListBanks from BankStatement |
| `internal/app/app.go` | Add bank usecase to struct + wiring |
| `internal/app/seeder.go` | Add banks:read/write/delete permissions |
| `internal/handlers/http/router.go` | Register bank handler; add bankUC param |
| `internal/handlers/http/v1/bankstatement/handler.go` | Remove /banks route group |

### DELETE (2 files)
| File | Reason |
|------|--------|
| `internal/usecase/bankstatement/list_banks.go` | Moved to bank usecase |
| `internal/handlers/http/v1/bankstatement/list_banks.go` | Moved to bank handler |
