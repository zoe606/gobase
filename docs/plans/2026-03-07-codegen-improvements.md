# Code Generator Improvements Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Improve code generator templates to produce near-production-ready code and add a `make wire` command to auto-wire DI, routes, and contracts.

**Architecture:** Enhance existing `pkg/codegen/generator/*.go` templates to match the quality of hand-written reference code (article feature). Add a new `pkg/codegen/cmd/wire/` CLI that uses Go AST parsing to append missing entries to contracts, router, and DI files.

**Tech Stack:** Go standard library (`go/parser`, `go/ast`, `go/format`, `text/template`), existing codegen infrastructure.

---

## Phase 1: Improve Generator Templates

### Task 1: Improve DTO Generator — ListRequest with pagination.Params

The current DTO generator creates a custom ListRequest with manual Page/PageSize fields. The actual codebase uses `pagination.Params` from `pkg/pagination`. Update the DTO generator to match.

**Files:**
- Modify: `pkg/codegen/generator/dto.go`
- Modify: `pkg/codegen/generator/dto_test.go`

**Step 1: Update the test to expect pagination.Params**

In `pkg/codegen/generator/dto_test.go`, find `TestBuildRequestDTOContent`. Update assertions:

Replace the ListRequest checks:
```go
// Check ListRequest uses pagination.Params
if !strings.Contains(content, "pagination.Params") {
    t.Error("expected ListRequest to embed pagination.Params")
}
```

Remove the old assertions for `query:"page"`, `query:"page_size"`, `GetPageSize`, `GetOffset` — these are provided by `pagination.Params`.

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/codegen/generator/ -run TestBuildRequestDTOContent -v`
Expected: FAIL

**Step 3: Update dto.go buildRequestDTOContent()**

Replace the `ListRequest` generation block (the ListRequest struct, GetPageSize, and GetOffset methods) with:

```go
// ListRequest (pagination)
sb.WriteString(fmt.Sprintf("// ListRequest represents the request to list %ss with filters.\n", strings.ToLower(entityName)))
sb.WriteString("type ListRequest struct {\n")
sb.WriteString("\tpagination.Params\n")
sb.WriteString("}\n")
```

Add `"go-boilerplate/pkg/pagination"` to the imports block. The import needs to be written at the top of the request.go output. Update the function to add an imports section:

```go
sb.WriteString(fmt.Sprintf("import (\n\t%q\n)\n\n", g.config.ModuleName+"/pkg/pagination"))
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/codegen/generator/ -run TestBuildRequestDTOContent -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/codegen/generator/dto.go pkg/codegen/generator/dto_test.go
git commit -m "feat(codegen): Use pagination.Params in generated ListRequest DTO"
```

---

### Task 2: Improve Repository Generator — Pagination and tx support

The current repo generator creates a basic `List(ctx, limit, offset)`. The actual codebase uses `pagination.Params` and `tx.DBFromContext()`. Update the repo generator to match.

**Files:**
- Modify: `pkg/codegen/generator/repo.go`
- Modify: `pkg/codegen/generator/repo_test.go`

**Step 1: Update repo tests**

In `repo_test.go`, update `TestBuildRepoImplContent` to check for:
- `tx.DBFromContext` usage
- `pagination.Params` in List method signature
- `params.Limit` and `params.Offset()` in List implementation

Update `TestBuildRepoInterfaceContent` to check:
- `List(ctx context.Context, params pagination.Params)` signature

**Step 2: Run tests to verify they fail**

Run: `go test ./pkg/codegen/generator/ -run TestRepo -v`
Expected: FAIL

**Step 3: Update repo.go**

In `buildRepoInterfaceContent()`, update the `List` signature:
```go
List(ctx context.Context, params pagination.Params) ([]*entity.%s, int64, error)
```

Add `pagination` import to the interface imports.

In `buildRepoImplContent()`, update all methods to use `tx.DBFromContext`:
```go
db := tx.DBFromContext(ctx, r.db)
```

Update the `List` method to use `params.Limit` and `params.Offset()`:
```go
func (r *{Entity}Repo) List(ctx context.Context, params pagination.Params) ([]*entity.{Entity}, int64, error) {
    db := tx.DBFromContext(ctx, r.db)
    var items []*entity.{Entity}
    var total int64

    if err := db.Model(&entity.{Entity}{}).Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("{package}repo - List - Count: %w", err)
    }

    if err := db.Order("id DESC").Limit(params.Limit).Offset(params.Offset()).Find(&items).Error; err != nil {
        return nil, 0, fmt.Errorf("{package}repo - List - Find: %w", err)
    }

    return items, total, nil
}
```

Add imports for `tx` and `pagination` packages in the implementation.

**Step 4: Run tests to verify they pass**

Run: `go test ./pkg/codegen/generator/ -run TestRepo -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/codegen/generator/repo.go pkg/codegen/generator/repo_test.go
git commit -m "feat(codegen): Add pagination.Params and tx support to generated repos"
```

---

### Task 3: Improve UseCase Generator — Real method implementations

The current usecase generator creates method files with TODO comments. Generate actual implementations that match the article pattern.

**Files:**
- Modify: `pkg/codegen/generator/usecase.go`
- Modify: `pkg/codegen/generator/usecase_test.go`

**Step 1: Update usecase tests**

Update test assertions for method content:
- `create.go` should contain: entity construction from request fields, `repo.Create()` call, `NewResponse()` return
- `get_by_id.go` should contain: `errors.Is(err, repo.ErrNotFound)` check, `ErrNotFound` return
- `list.go` should contain: `pagination.Params` usage, `NewListResponse()` return
- `update.go` should contain: fetch-then-update pattern, partial update logic
- `delete.go` should contain: `repo.Delete()` call, ErrNotFound check

**Step 2: Run tests to verify they fail**

Run: `go test ./pkg/codegen/generator/ -run TestUseCase -v`
Expected: FAIL

**Step 3: Update usecase.go method builders**

Update `buildUseCaseCreateContent()`:
```go
func (uc *UseCase) Create(ctx context.Context, req {pkgdto}.CreateRequest) (*{pkgdto}.Response, error) {
    {var} := &entity.{Entity}{
        // Map request fields to entity fields
    }

    if err := uc.{var}Repo.Create(ctx, {var}); err != nil {
        return nil, fmt.Errorf("{pkg} - Create: %w", err)
    }

    return {pkgdto}.NewResponse({var}), nil
}
```

The field mapping in Create should iterate `g.result.Fields` and map create-eligible fields from request to entity.

Update `buildUseCaseGetByIDContent()`:
```go
func (uc *UseCase) GetByID(ctx context.Context, id uint) (*{pkgdto}.Response, error) {
    {var}, err := uc.{var}Repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, repo.ErrNotFound) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("{pkg} - GetByID: %w", err)
    }

    return {pkgdto}.NewResponse({var}), nil
}
```

Update `buildUseCaseListContent()`:
```go
func (uc *UseCase) List(ctx context.Context, req {pkgdto}.ListRequest) (*{pkgdto}.ListResponse, error) {
    req.Params.Normalize()

    items, total, err := uc.{var}Repo.List(ctx, req.Params)
    if err != nil {
        return nil, fmt.Errorf("{pkg} - List: %w", err)
    }

    return {pkgdto}.NewListResponse(items, total, req.Params), nil
}
```

Update `buildUseCaseUpdateContent()`:
```go
func (uc *UseCase) Update(ctx context.Context, id uint, req {pkgdto}.UpdateRequest) (*{pkgdto}.Response, error) {
    {var}, err := uc.{var}Repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, repo.ErrNotFound) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("{pkg} - Update - GetByID: %w", err)
    }

    // Apply partial updates (only non-nil fields)
    // ... iterate pointer fields from UpdateRequest ...

    if err := uc.{var}Repo.Update(ctx, {var}); err != nil {
        return nil, fmt.Errorf("{pkg} - Update: %w", err)
    }

    return {pkgdto}.NewResponse({var}), nil
}
```

Update `buildUseCaseDeleteContent()`:
```go
func (uc *UseCase) Delete(ctx context.Context, id uint) error {
    if err := uc.{var}Repo.Delete(ctx, id); err != nil {
        if errors.Is(err, repo.ErrNotFound) {
            return ErrNotFound
        }
        return fmt.Errorf("{pkg} - Delete: %w", err)
    }

    return nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./pkg/codegen/generator/ -run TestUseCase -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/codegen/generator/usecase.go pkg/codegen/generator/usecase_test.go
git commit -m "feat(codegen): Generate real usecase method implementations"
```

---

### Task 4: Improve UseCase Generator — Real test files

The current generator creates `t.Skip("Test not implemented")` stubs. Generate actual table-driven tests with mock setup.

**Files:**
- Modify: `pkg/codegen/generator/usecase.go`
- Modify: `pkg/codegen/generator/usecase_test.go`

**Step 1: Update test assertions**

In `TestBuildUseCaseTestContent`, check that generated test content includes:
- `func TestCreate(t *testing.T)` with subtests (success, repo error)
- Mock repo setup: `mockRepo := &mockArticleRepo{}`
- UseCase construction: `uc := New(mockRepo)`
- `require.NoError` / `require.Error` assertions
- Table-driven test structure with `for _, tt := range tests`

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/codegen/generator/ -run TestBuildUseCaseTestContent -v`
Expected: FAIL

**Step 3: Update usecase.go test builders**

Generate a `mocks_test.go` with a simple mock repo struct:
```go
type mock{Entity}Repo struct {
    createFn  func(ctx context.Context, {var} *entity.{Entity}) error
    getByIDFn func(ctx context.Context, id uint) (*entity.{Entity}, error)
    listFn    func(ctx context.Context, params pagination.Params) ([]*entity.{Entity}, int64, error)
    updateFn  func(ctx context.Context, {var} *entity.{Entity}) error
    deleteFn  func(ctx context.Context, id uint) error
}
```

Each mock method delegates to the function field (e.g., `r.createFn(ctx, entity)`).

Generate `create_test.go` with:
```go
func TestCreate(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        req     {pkgdto}.CreateRequest
        mockFn  func(repo *mock{Entity}Repo)
        wantErr bool
    }{
        {
            name: "success",
            req:  {pkgdto}.CreateRequest{...sample fields...},
            mockFn: func(repo *mock{Entity}Repo) {
                repo.createFn = func(ctx context.Context, e *entity.{Entity}) error {
                    return nil
                }
            },
        },
        {
            name: "repo error",
            req:  {pkgdto}.CreateRequest{...sample fields...},
            mockFn: func(repo *mock{Entity}Repo) {
                repo.createFn = func(ctx context.Context, e *entity.{Entity}) error {
                    return errors.New("db error")
                }
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            mockRepo := &mock{Entity}Repo{}
            tt.mockFn(mockRepo)
            uc := New(mockRepo)

            result, err := uc.Create(context.Background(), tt.req)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            require.NotNil(t, result)
        })
    }
}
```

Similar pattern for GetByID (success, not found, repo error), List, Update, Delete tests.

**Step 4: Run tests to verify they pass**

Run: `go test ./pkg/codegen/generator/ -run TestBuildUseCaseTestContent -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/codegen/generator/usecase.go pkg/codegen/generator/usecase_test.go
git commit -m "feat(codegen): Generate real table-driven usecase tests with mocks"
```

---

### Task 5: Improve Handler Generator — Swagger annotations and error handling

The current handler generator already has swagger annotations but they need to match the `articledto` naming convention and use correct response types.

**Files:**
- Modify: `pkg/codegen/generator/handler.go`
- Modify: `pkg/codegen/generator/handler_test.go`

**Step 1: Update handler tests**

Check that generated handlers include:
- `_ "go-boilerplate/internal/dto/{pkg}"` blank import for swagger resolution (in files that only reference DTO in annotations, like get_by_id.go, list.go, delete.go)
- Correct DTO package references: `{pkg}dto.CreateRequest`, `{pkg}dto.Response`
- Auth middleware comment placeholder in RegisterRoutes
- `v1.ParseValidationErrors(err)` for validation errors

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/codegen/generator/ -run TestHandler -v`
Expected: FAIL

**Step 3: Update handler.go**

Key changes to handler method templates:
- Import `{pkgdto}` correctly using the package's actual name (no alias needed since package is named `{pkg}dto`)
- Add blank import `_ "{module}/internal/dto/{pkg}"` in handler files that don't use DTO in code (get_by_id, list, delete)
- Use `{pkg}dto.CreateRequest` in swagger annotations
- Use `response.Response[{pkg}dto.Response]` in success annotations
- Add `// TODO: Add auth middleware` comment in RegisterRoutes

Update `buildHandlerMainContent()` to include auth middleware placeholder:
```go
func (h *Handler) RegisterRoutes(router fiber.Router) {
    {var}s := router.Group("/{varName}s")
    // TODO: Add auth middleware: {var}s.Use(middleware.JWT(jwtService))
    {var}s.Post("/", h.Create)
    // ...
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./pkg/codegen/generator/ -run TestHandler -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/codegen/generator/handler.go pkg/codegen/generator/handler_test.go
git commit -m "feat(codegen): Improve handler generation with swagger and error handling"
```

---

### Task 6: Update usecase contracts generator — DTO naming

The usecase interface in contracts.go must use `{pkg}dto.` prefix to match the renamed DTO packages.

**Files:**
- Modify: `pkg/codegen/generator/usecase.go` (the `buildUseCaseInterfaceContent` function)
- Modify: `pkg/codegen/generator/usecase_test.go`

**Step 1: Update test for interface content**

Check that the generated interface references `{pkg}dto.CreateRequest`, `{pkg}dto.Response`, etc. and includes the correct import for `{module}/internal/dto/{pkg}`.

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/codegen/generator/ -run TestBuildUseCaseInterfaceContent -v`
Expected: FAIL

**Step 3: Update the interface builder**

Ensure imports include `{module}/internal/dto/{pkg}` and method signatures use `{pkg}dto.` prefix:
```go
Create(ctx context.Context, req {pkg}dto.CreateRequest) (*{pkg}dto.Response, error)
GetByID(ctx context.Context, id uint) (*{pkg}dto.Response, error)
List(ctx context.Context, req {pkg}dto.ListRequest) (*{pkg}dto.ListResponse, error)
Update(ctx context.Context, id uint, req {pkg}dto.UpdateRequest) (*{pkg}dto.Response, error)
Delete(ctx context.Context, id uint) error
```

**Step 4: Run tests to verify they pass**

Run: `go test ./pkg/codegen/generator/ -run TestBuildUseCaseInterfaceContent -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/codegen/generator/usecase.go pkg/codegen/generator/usecase_test.go
git commit -m "feat(codegen): Use correct DTO package naming in usecase contracts"
```

---

## Phase 2: The `make wire` Command

### Task 7: Create wire CLI entry point

**Files:**
- Create: `pkg/codegen/cmd/wire/main.go`

**Step 1: Create the wire command skeleton**

```go
package main

import (
    "flag"
    "fmt"
    "os"

    "go-boilerplate/pkg/codegen/wire"
)

func main() {
    outputDir := flag.String("o", ".", "Project root directory")
    dryRun := flag.Bool("n", false, "Dry run (print changes without writing)")
    flag.Parse()

    moduleName, err := readModuleName(*outputDir)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading module name: %v\n", err)
        os.Exit(1)
    }

    w := wire.New(wire.Config{
        ModuleName: moduleName,
        OutputDir:  *outputDir,
        DryRun:     *dryRun,
    })

    if err := w.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func readModuleName(dir string) (string, error) {
    data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
    if err != nil {
        return "", err
    }
    for _, line := range strings.Split(string(data), "\n") {
        if strings.HasPrefix(line, "module ") {
            return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
        }
    }
    return "", fmt.Errorf("module name not found in go.mod")
}
```

**Step 2: Add Makefile target**

In `Makefile`, under `##@ Code Generation`, add:
```makefile
.PHONY: wire
wire: ## Auto-wire DI, routes, and contracts for generated features
	go run ./pkg/codegen/cmd/wire
```

**Step 3: Commit**

```bash
git add pkg/codegen/cmd/wire/main.go Makefile
git commit -m "feat(codegen): Add wire CLI entry point and Makefile target"
```

---

### Task 8: Wire scanner — Detect generated features

Create the wire package that scans the project to find features that need wiring.

**Files:**
- Create: `pkg/codegen/wire/wire.go`
- Create: `pkg/codegen/wire/scanner.go`
- Create: `pkg/codegen/wire/scanner_test.go`

**Step 1: Write scanner test**

```go
func TestScanFeatures(t *testing.T) {
    // Create a temp directory mimicking the project structure
    // with internal/usecase/order/order.go, internal/repo/persistent/order.go, etc.
    // Verify scanner detects "order" as an unwired feature
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/codegen/wire/ -run TestScanFeatures -v`
Expected: FAIL

**Step 3: Implement scanner**

`scanner.go`:
- `ScanFeatures(outputDir string) ([]Feature, error)`
- A `Feature` struct: `Name`, `EntityName`, `PackageName`, `HasRepo`, `HasUseCase`, `HasHandler`
- Scan `internal/usecase/*/` for directories containing `*.go` with a `New()` function
- Scan `internal/repo/persistent/` for matching repo files
- Scan `internal/handlers/http/v1/*/` for matching handler directories
- Return features not yet in contracts/router/app

`wire.go`:
- `Config` struct: ModuleName, OutputDir, DryRun
- `New(config Config) *Wirer`
- `Run() error`: Calls scanner, then calls each wiring function

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/codegen/wire/ -run TestScanFeatures -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/codegen/wire/
git commit -m "feat(codegen): Add wire scanner to detect unwired features"
```

---

### Task 9: Wire contracts — Append to repo/contracts.go and usecase/contracts.go

**Files:**
- Create: `pkg/codegen/wire/contracts.go`
- Create: `pkg/codegen/wire/contracts_test.go`

**Step 1: Write contract wiring test**

```go
func TestWireRepoContract(t *testing.T) {
    // Create a temp contracts.go with existing interfaces
    // Call wireRepoContract() for "order" feature
    // Verify OrderRepo interface was appended inside the type block
    // Verify running again is idempotent (no duplicate)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/codegen/wire/ -run TestWireRepoContract -v`
Expected: FAIL

**Step 3: Implement contract wiring**

`contracts.go`:
- `wireRepoContract(feature Feature, contractsPath string) error`
  - Read file, check if `{Entity}Repo` already exists, skip if so
  - Find the last `)` in the type block
  - Insert the interface definition before it
  - Write back with `go/format`

- `wireUsecaseContract(feature Feature, contractsPath string) error`
  - Same pattern, insert usecase interface

Use string-based insertion (matching the existing `appendToFile` pattern in generator.go) rather than full AST rewriting — simpler and proven to work.

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/codegen/wire/ -run TestWireRepoContract -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/codegen/wire/contracts.go pkg/codegen/wire/contracts_test.go
git commit -m "feat(codegen): Add contract wiring for repo and usecase interfaces"
```

---

### Task 10: Wire router — Append to router.go

**Files:**
- Create: `pkg/codegen/wire/router.go`
- Create: `pkg/codegen/wire/router_test.go`

**Step 1: Write router wiring test**

```go
func TestWireRouter(t *testing.T) {
    // Create a temp router.go with existing setupAPIRoutes function
    // Call wireRouter() for "order" feature
    // Verify: import added, parameter added to SetupRoutes and setupAPIRoutes
    // Verify: handler creation + RegisterRoutes call appended inside setupAPIRoutes body
    // Verify idempotent
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/codegen/wire/ -run TestWireRouter -v`
Expected: FAIL

**Step 3: Implement router wiring**

`router.go`:
- `wireRouter(feature Feature, routerPath, moduleName string) error`
  - Read router.go content
  - Check if handler import already exists, skip if so
  - Add import: `{pkg}handler "{module}/internal/handlers/http/v1/{pkg}"`
  - Add `{pkg}UC usecase.{Entity}` parameter to `SetupRoutes()` and `setupAPIRoutes()`
  - Append handler creation before the closing `}` of `setupAPIRoutes()`:
    ```go
    {var}Handler := {pkg}handler.New({var}UC, l)
    {var}Handler.RegisterRoutes(apiV1Group)
    ```

Use string search for insertion points:
- Import: Find the import block, insert before closing `)`
- Parameters: Find the `setupAPIRoutes(` signature, insert before closing `)`
- Handler: Find the last `RegisterRoutes(apiV1Group)` line, insert after it

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/codegen/wire/ -run TestWireRouter -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/codegen/wire/router.go pkg/codegen/wire/router_test.go
git commit -m "feat(codegen): Add router wiring for handler registration"
```

---

### Task 11: Wire DI — Append to app.go

**Files:**
- Create: `pkg/codegen/wire/app.go`
- Create: `pkg/codegen/wire/app_test.go`

**Step 1: Write DI wiring test**

```go
func TestWireApp(t *testing.T) {
    // Create a temp app.go with existing struct and init functions
    // Call wireApp() for "order" feature
    // Verify: import added for usecase/order and persistent packages
    // Verify: field added to repositories struct
    // Verify: field added to usecases struct
    // Verify: line added to initRepositories()
    // Verify: line added to initUseCases()
    // Verify: parameter added to SetupRoutes() call in initHTTPServer()
    // Verify idempotent
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/codegen/wire/ -run TestWireApp -v`
Expected: FAIL

**Step 3: Implement DI wiring**

`app.go`:
- `wireApp(feature Feature, appPath, moduleName string) error`

This is the most complex wiring. It modifies 5 locations in app.go:

1. **Import block**: Add `"{module}/internal/usecase/{pkg}"`
2. **repositories struct**: Add `{var} repo.{Entity}Repo`
3. **usecases struct**: Add `{var} usecase.{Entity}`
4. **initRepositories()**: Add `{var}: persistent.New{Entity}Repo(db),`
5. **initUseCases()**: Add `{var}UC := {pkg}.New(repos.{var})` and add to return struct
6. **initHTTPServer()** SetupRoutes call: Add `uc.{var}` parameter

Each insertion uses string search:
- For structs: find `type repositories struct {` or `type usecases struct {`, insert before closing `}`
- For functions: find the return statement or closing `}`, insert before it
- Check if `{Entity}Repo` or `{Entity}` already exists to ensure idempotency

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/codegen/wire/ -run TestWireApp -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/codegen/wire/app.go pkg/codegen/wire/app_test.go
git commit -m "feat(codegen): Add DI wiring for app.go"
```

---

### Task 12: Wire AutoMigrate — Append entity to migration list

**Files:**
- Modify: `pkg/codegen/wire/app.go`
- Modify: `pkg/codegen/wire/app_test.go`

**Step 1: Update test**

Add assertion that wiring also adds `&entity.{Entity}{}` to the `db.AutoMigrate()` call in `runAutoMigrate()`.

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/codegen/wire/ -run TestWireApp -v`
Expected: FAIL

**Step 3: Implement**

In `wireApp()`, add a step:
- Find `db.AutoMigrate(` block
- Check if `entity.{Entity}` already exists
- Insert `&entity.{Entity}{},` before the closing `)`

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/codegen/wire/ -run TestWireApp -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/codegen/wire/app.go pkg/codegen/wire/app_test.go
git commit -m "feat(codegen): Wire entity into AutoMigrate list"
```

---

## Phase 3: Integration and Verification

### Task 13: End-to-end test with a real migration

Verify the complete workflow works by generating a new feature from an existing migration.

**Files:**
- No new files — testing existing infrastructure

**Step 1: Generate a test feature**

```bash
# Use the existing articles migration as a test (already has entity, so use --force)
# Or create a temporary test migration
make gen-full MIGRATION=000011_create_articles.up.sql LAYERS=entity,dto,repo,usecase,handler
```

This should fail because files exist. That's expected — confirms the generator works.

**Step 2: Test with dry run on a hypothetical feature**

```bash
# Test dry run to see what would be generated
go run ./pkg/codegen/cmd/codegen -m 000010_create_profiles.up.sql -l entity,dto,repo,usecase,handler -n
```

Review the output matches the improved templates.

**Step 3: Test wire command dry run**

```bash
make wire -- -n
```

Should report that all existing features are already wired (idempotent check).

**Step 4: Run full test suite**

```bash
make check-all
```

Expected: All tests pass, no lint errors.

**Step 5: Commit any fixes**

```bash
git add -A
git commit -m "test: Verify end-to-end codegen and wire workflow"
```

---

### Task 14: Update documentation

**Files:**
- Modify: `CLAUDE.md`
- Modify: `README.md` (if codegen section exists)

**Step 1: Update CLAUDE.md**

Add the new workflow to the "Adding New Features" section:
```
## Adding New Features (Quick Path)

1. Create migration: `make migrate-create name=create_orders`
2. Edit migration SQL files
3. Run migration: `POSTGRES_URL="..." make migrate-up`
4. Generate all layers: `make gen-full MIGRATION=000012_create_orders.up.sql`
5. Auto-wire: `make wire`
6. Verify: `make check-all`
7. Regenerate swagger: `make swag`
```

Document `make wire` in the Commands section.

**Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: Add make wire and improved codegen workflow to CLAUDE.md"
```
