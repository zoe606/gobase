# Phase 4a: Critical Fixes Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix 3 critical bugs: article field mapping, IDOR vulnerability, and idempotency cache key scoping.

**Architecture:** Fixes propagate through DTO → usecase → handler layers. Interface changes to `Update`/`Delete` ripple through contracts, implementations, mocks, and tests. Idempotency fix is isolated to one middleware file.

**Tech Stack:** Go 1.26, Fiber v2, GORM, gomock, testify

**Spec:** `docs/superpowers/specs/2026-03-15-phase4a-critical-fixes-design.md`

---

## File Structure

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/dto/article/request.go` | Modify | Remove ViewCount, change PublishedAt type, make Status optional |
| `internal/usecase/contracts.go` | Modify | Add `userID` param to `Update` and `Delete` |
| `internal/usecase/article/errors.go` | Modify | Add `ErrForbidden` |
| `internal/usecase/article/article.go` | Modify | Add `ptrOrDefault` helper |
| `internal/usecase/article/create.go` | Modify | Map request fields to entity |
| `internal/usecase/article/create_test.go` | Modify | Test field mapping, default status |
| `internal/usecase/article/update.go` | Modify | Add userID param, ownership check, apply fields |
| `internal/usecase/article/update_test.go` | Modify | Add userID param to all tests, add forbidden test |
| `internal/usecase/article/delete.go` | Modify | Add userID param, fetch-then-check pattern |
| `internal/usecase/article/delete_test.go` | Modify | Add userID param, mock GetByID+Delete, add forbidden test |
| `internal/handlers/http/v1/article/update.go` | Modify | Extract userID, pass to usecase, handle 403, Swagger |
| `internal/handlers/http/v1/article/delete.go` | Modify | Extract userID, pass to usecase, handle 403, Swagger |
| `internal/handlers/http/v1/article/mocks_test.go` | Modify | Update Update/Delete mock signatures |
| `internal/handlers/http/v1/article/handler_test.go` | Modify | Update all mock expectations, add 403 tests |
| `internal/handlers/http/middleware/idempotency.go` | Modify | Scope cache key by user+method+path |
| `internal/handlers/http/middleware/idempotency_test.go` | Modify | Add cross-endpoint and user-scoping tests |

---

## Chunk 1: DTO Cleanup and Create Field Mapping

### Task 1: Clean up CreateRequest and UpdateRequest DTOs

**Files:**
- Modify: `internal/dto/article/request.go`

- [ ] **Step 1: Update CreateRequest — remove ViewCount, change PublishedAt, make Status optional**

```go
// CreateRequest represents the request to create a Article.
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

Changes from current:
- `Status`: removed `validate:"required"`, added `validate:"omitempty,oneof=draft published"`
- `PublishedAt`: changed from `time.Time` to `*time.Time`, removed `validate:"required"`
- `ViewCount`: removed entirely

- [ ] **Step 2: Update UpdateRequest — remove ViewCount**

```go
// UpdateRequest represents the request to update a Article.
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

Only change: removed `ViewCount *int` line.

- [ ] **Step 3: Verify compilation**

Run: `go build ./internal/dto/...`
Expected: PASS (no other code references `ViewCount` on these structs except handler tests — those will be fixed in Task 5)

- [ ] **Step 4: Commit**

```bash
git add internal/dto/article/request.go
git commit -m "fix(dto): Remove ViewCount from article DTOs, make Status/PublishedAt optional"
```

---

### Task 2: Add ptrOrDefault helper and fix Create field mapping

**Files:**
- Modify: `internal/usecase/article/article.go`
- Modify: `internal/usecase/article/create.go`
- Modify: `internal/usecase/article/create_test.go`

- [ ] **Step 1: Write failing test for Create field mapping**

In `internal/usecase/article/create_test.go`, replace the existing "success" test case to verify fields are actually mapped. Add a new "success with default status" test case:

```go
func TestCreate(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)

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
		wantTitle string
	}{
		{
			name: "success - all fields mapped",
			args: args{
				ctx:    context.Background(),
				userID: 1,
				req: articledto.CreateRequest{
					Title:        "Test Article",
					Slug:         "test-article",
					Content:      "Some content",
					Excerpt:      "Some excerpt",
					CoverMediaID: 42,
					Status:       "published",
					PublishedAt:  &now,
				},
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, a *entity.Article) error {
						// Verify fields are actually mapped
						require.Equal(t, uint(1), a.UserID)
						require.Equal(t, "Test Article", a.Title)
						require.Equal(t, "test-article", a.Slug)
						require.NotNil(t, a.Content)
						require.Equal(t, "Some content", *a.Content)
						require.NotNil(t, a.Excerpt)
						require.Equal(t, "Some excerpt", *a.Excerpt)
						require.NotNil(t, a.CoverMediaID)
						require.Equal(t, uint(42), *a.CoverMediaID)
						require.NotNil(t, a.Status)
						require.Equal(t, "published", *a.Status)
						require.NotNil(t, a.PublishedAt)
						a.ID = 1
						return nil
					})
			},
			wantErr:    nil,
			wantTitle: "Test Article",
		},
		{
			name: "success - empty status defaults to draft",
			args: args{
				ctx:    context.Background(),
				userID: 2,
				req: articledto.CreateRequest{
					Title:        "Draft Article",
					Slug:         "draft-article",
					Content:      "Draft content",
					Excerpt:      "Draft excerpt",
					CoverMediaID: 1,
				},
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, a *entity.Article) error {
						require.NotNil(t, a.Status)
						require.Equal(t, "draft", *a.Status)
						require.Nil(t, a.PublishedAt)
						a.ID = 2
						return nil
					})
			},
			wantErr:   nil,
			wantTitle: "Draft Article",
		},
		{
			name: "repo error",
			args: args{
				ctx:    context.Background(),
				userID: 1,
				req: articledto.CreateRequest{
					Title:        "Test Article",
					Slug:         "test-article",
					Content:      "Some content",
					Excerpt:      "Some excerpt",
					CoverMediaID: 1,
					Status:       "draft",
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

			uc := article.New(mockArticleRepo, audit.NewNoop(), cache.NewNoop())
			got, err := uc.Create(tt.args.ctx, tt.args.userID, tt.args.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			if tt.wantTitle != "" {
				require.Equal(t, tt.wantTitle, got.Title)
			}
		})
	}
}
```

Add these imports to the import block (if not already present): `"time"`, `"go-boilerplate/internal/entity"`, `"github.com/stretchr/testify/require"`.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -run TestCreate ./internal/usecase/article/`
Expected: FAIL — the "success - all fields mapped" test fails because `DoAndReturn` verifications fail (Title is empty string, Content is nil, etc.)

- [ ] **Step 3: Add ptrOrDefault helper to article.go**

In `internal/usecase/article/article.go`, add after the `New` function:

```go
// ptrOrDefault returns a pointer to s if non-empty, otherwise a pointer to def.
func ptrOrDefault(s, def string) *string {
	if s == "" {
		return &def
	}
	return &s
}
```

- [ ] **Step 4: Implement Create field mapping**

Replace `internal/usecase/article/create.go`:

```go
package article

import (
	"context"
	"fmt"

	articledto "go-boilerplate/internal/dto/article"
	"go-boilerplate/internal/entity"
)

// Create creates a new article.
func (uc *UseCase) Create(ctx context.Context, userID uint, req articledto.CreateRequest) (*articledto.Response, error) {
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

	if err := uc.articleRepo.Create(ctx, article); err != nil {
		return nil, fmt.Errorf("article - Create - articleRepo.Create: %w", err)
	}

	// Audit log (best-effort — don't fail the operation)
	_ = uc.auditLogger.LogCreate(ctx, "article", article.ID, &userID, map[string]any{
		"title": req.Title,
		"slug":  req.Slug,
	})

	// Invalidate list cache (new article changes any list)
	_ = uc.cache.DeleteByPrefix(ctx, uc.cacheKeys.ListPrefix())

	return articledto.NewResponse(article), nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test -v -run TestCreate ./internal/usecase/article/`
Expected: PASS — all 3 test cases pass

- [ ] **Step 6: Commit**

```bash
git add internal/usecase/article/article.go internal/usecase/article/create.go internal/usecase/article/create_test.go
git commit -m "fix(article): Map all CreateRequest fields to entity, add ptrOrDefault helper"
```

---

## Chunk 2: Interface Change, Update Fix, and Ownership Checks

### Task 3: Change Article interface and add ErrForbidden

**Files:**
- Modify: `internal/usecase/contracts.go`
- Modify: `internal/usecase/article/errors.go`

- [ ] **Step 1: Update Article interface in contracts.go**

In `internal/usecase/contracts.go`, change the `Article` interface:

```go
	// Article defines Article use case operations.
	Article interface {
		Create(ctx context.Context, userID uint, req articledto.CreateRequest) (*articledto.Response, error)
		GetByID(ctx context.Context, id uint) (*articledto.Response, error)
		List(ctx context.Context, req articledto.ListRequest) (*articledto.ListResponse, error)
		Update(ctx context.Context, userID uint, id uint, req articledto.UpdateRequest) (*articledto.Response, error)
		Delete(ctx context.Context, userID uint, id uint) error
	}
```

Changes: `Update` gains `userID uint` before `id`; `Delete` gains `userID uint` before `id`.

- [ ] **Step 2: Add ErrForbidden to errors.go**

In `internal/usecase/article/errors.go`, add to the `var` block:

```go
	// ErrForbidden indicates the user is not authorized to modify this article.
	ErrForbidden = errors.New("not authorized to modify this article")
```

**Note:** Do NOT commit yet. The interface change breaks compilation in `update.go`, `delete.go`, handler mocks, and handler tests. Tasks 4 and 5 implement the new signatures. The commit will be made at the end of Task 5 when all usecase code compiles again.

---

### Task 4: Fix Update usecase — ownership check and field mapping

**Files:**
- Modify: `internal/usecase/article/update.go`
- Modify: `internal/usecase/article/update_test.go`

- [ ] **Step 1: Write failing tests for ownership check and field mapping**

Replace the entire `internal/usecase/article/update_test.go`:

```go
package article_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	articledto "go-boilerplate/internal/dto/article"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/usecase/article"
	"go-boilerplate/pkg/audit"
	"go-boilerplate/pkg/cache"
)

func TestUpdate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	content := "Original content"
	status := "draft"
	newTitle := "Updated Title"

	tests := []struct {
		name      string
		userID    uint
		id        uint
		req       articledto.UpdateRequest
		setupMock func(articleRepo *MockArticleRepo)
		wantErr   error
		wantTitle string
	}{
		{
			name:   "success - title updated",
			userID: 1,
			id:     1,
			req: articledto.UpdateRequest{
				Title: &newTitle,
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Article{
						ID:        1,
						UserID:    1,
						Title:     "Original Title",
						Slug:      "original-title",
						Content:   &content,
						Status:    &status,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				articleRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, a *entity.Article) error {
						require.Equal(t, "Updated Title", a.Title)
						return nil
					})
			},
			wantErr:   nil,
			wantTitle: "Updated Title",
		},
		{
			name:   "forbidden - different user",
			userID: 99,
			id:     1,
			req: articledto.UpdateRequest{
				Title: &newTitle,
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Article{
						ID:        1,
						UserID:    1,
						Title:     "Original Title",
						Slug:      "original-title",
						Content:   &content,
						Status:    &status,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			wantErr: article.ErrForbidden,
		},
		{
			name:   "not found",
			userID: 1,
			id:     999,
			req: articledto.UpdateRequest{
				Title: &newTitle,
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(999)).
					Return(nil, repo.ErrNotFound)
			},
			wantErr: article.ErrNotFound,
		},
		{
			name:   "get by id repo error",
			userID: 1,
			id:     1,
			req: articledto.UpdateRequest{
				Title: &newTitle,
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(nil, errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name:   "update repo error",
			userID: 1,
			id:     1,
			req: articledto.UpdateRequest{
				Title: &newTitle,
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Article{
						ID:        1,
						UserID:    1,
						Title:     "Original Title",
						Slug:      "original-title",
						Content:   &content,
						Status:    &status,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				articleRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(errors.New("update failed"))
			},
			wantErr: errors.New("update failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockArticleRepo := NewMockArticleRepo(ctrl)

			tt.setupMock(mockArticleRepo)

			uc := article.New(mockArticleRepo, audit.NewNoop(), cache.NewNoop())
			got, err := uc.Update(context.Background(), tt.userID, tt.id, tt.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, article.ErrNotFound) || errors.Is(tt.wantErr, article.ErrForbidden) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.Contains(t, err.Error(), tt.wantErr.Error())
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, tt.id, got.ID)
			if tt.wantTitle != "" {
				require.Equal(t, tt.wantTitle, got.Title)
			}
		})
	}
}
```

- [ ] **Step 2: Implement Update with ownership check and field mapping**

Replace `internal/usecase/article/update.go`:

```go
package article

import (
	"context"
	"errors"
	"fmt"

	articledto "go-boilerplate/internal/dto/article"
	"go-boilerplate/internal/repo"
)

// Update updates a article.
func (uc *UseCase) Update(ctx context.Context, userID uint, id uint, req articledto.UpdateRequest) (*articledto.Response, error) {
	// Get existing article
	article, err := uc.articleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("article - Update - articleRepo.GetByID: %w", err)
	}

	// Ownership check
	if article.UserID != userID {
		return nil, ErrForbidden
	}

	// Capture old values for audit
	oldValues := map[string]any{"title": article.Title, "slug": article.Slug}

	// Apply non-nil fields from request
	if req.Title != nil {
		article.Title = *req.Title
	}
	if req.Slug != nil {
		article.Slug = *req.Slug
	}
	if req.Content != nil {
		article.Content = req.Content
	}
	if req.Excerpt != nil {
		article.Excerpt = req.Excerpt
	}
	if req.CoverMediaID != nil {
		article.CoverMediaID = req.CoverMediaID
	}
	if req.Status != nil {
		article.Status = req.Status
	}
	if req.PublishedAt != nil {
		article.PublishedAt = req.PublishedAt
	}

	if err := uc.articleRepo.Update(ctx, article); err != nil {
		return nil, fmt.Errorf("article - Update - articleRepo.Update: %w", err)
	}

	// Audit log (best-effort)
	_ = uc.auditLogger.LogUpdate(ctx, "article", article.ID, &userID, oldValues, map[string]any{
		"title": article.Title,
	})

	// Invalidate caches
	_ = uc.cache.Delete(ctx, uc.cacheKeys.ID(id))
	_ = uc.cache.DeleteByPrefix(ctx, uc.cacheKeys.ListPrefix())

	return articledto.NewResponse(article), nil
}
```

- [ ] **Step 3: Run tests to verify they pass**

Run: `go test -v -run TestUpdate ./internal/usecase/article/`
Expected: PASS — all 5 test cases pass (success, forbidden, not found, get error, update error)

**Note:** Do NOT commit yet — Task 5 will commit Update and Delete together with the interface change (atomic commit).

---

### Task 5: Fix Delete usecase — ownership check with fetch-then-delete

**Files:**
- Modify: `internal/usecase/article/delete.go`
- Modify: `internal/usecase/article/delete_test.go`

- [ ] **Step 1: Write failing tests for Delete with ownership check**

Replace `internal/usecase/article/delete_test.go`:

```go
package article_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/usecase/article"
	"go-boilerplate/pkg/audit"
	"go-boilerplate/pkg/cache"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	now := time.Now()
	status := "draft"

	tests := []struct {
		name      string
		userID    uint
		id        uint
		setupMock func(articleRepo *MockArticleRepo)
		wantErr   error
	}{
		{
			name:   "success",
			userID: 1,
			id:     1,
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Article{
						ID:        1,
						UserID:    1,
						Title:     "Test Article",
						Status:    &status,
						CreatedAt: now,
					}, nil)
				articleRepo.EXPECT().
					Delete(gomock.Any(), uint(1)).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:   "forbidden - different user",
			userID: 99,
			id:     1,
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Article{
						ID:        1,
						UserID:    1,
						Title:     "Test Article",
						Status:    &status,
						CreatedAt: now,
					}, nil)
			},
			wantErr: article.ErrForbidden,
		},
		{
			name:   "not found",
			userID: 1,
			id:     999,
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(999)).
					Return(nil, repo.ErrNotFound)
			},
			wantErr: article.ErrNotFound,
		},
		{
			name:   "get by id repo error",
			userID: 1,
			id:     1,
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(nil, errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name:   "delete repo error",
			userID: 1,
			id:     1,
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Article{
						ID:        1,
						UserID:    1,
						Title:     "Test Article",
						Status:    &status,
						CreatedAt: now,
					}, nil)
				articleRepo.EXPECT().
					Delete(gomock.Any(), uint(1)).
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

			uc := article.New(mockArticleRepo, audit.NewNoop(), cache.NewNoop())
			err := uc.Delete(context.Background(), tt.userID, tt.id)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, article.ErrNotFound) || errors.Is(tt.wantErr, article.ErrForbidden) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.Contains(t, err.Error(), tt.wantErr.Error())
				}
				return
			}

			require.NoError(t, err)
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v -run TestDelete ./internal/usecase/article/`
Expected: FAIL — compilation error because `Delete` signature doesn't match yet

- [ ] **Step 3: Implement Delete with fetch-then-check-then-delete**

Replace `internal/usecase/article/delete.go`:

```go
package article

import (
	"context"
	"errors"
	"fmt"

	"go-boilerplate/internal/repo"
)

// Delete deletes a article by ID after verifying ownership.
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

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -v -run TestDelete ./internal/usecase/article/`
Expected: PASS — all 5 test cases pass

- [ ] **Step 5: Commit (atomic: interface change + both implementations)**

This commit includes the interface change from Task 3 alongside the Update (Task 4) and Delete implementations, ensuring the codebase compiles at every commit.

```bash
git add internal/usecase/contracts.go internal/usecase/article/errors.go \
  internal/usecase/article/update.go internal/usecase/article/update_test.go \
  internal/usecase/article/delete.go internal/usecase/article/delete_test.go
git commit -m "fix(article): Add userID ownership check to Update/Delete, fix field mapping"
```

---

## Chunk 3: Handler Updates and Mock Fixes

### Task 6: Update handler mocks to match new interface

**Files:**
- Modify: `internal/handlers/http/v1/article/mocks_test.go`

- [ ] **Step 1: Update MockArticle Update method**

In `internal/handlers/http/v1/article/mocks_test.go`, replace the `Update` mock method and recorder:

```go
// Update mocks base method.
func (m *MockArticle) Update(ctx context.Context, userID uint, id uint, req articledto.UpdateRequest) (*articledto.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, userID, id, req)
	ret0, _ := ret[0].(*articledto.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr *MockArticleMockRecorder) Update(ctx, userID, id, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockArticle)(nil).Update), ctx, userID, id, req)
}
```

- [ ] **Step 2: Update MockArticle Delete method**

In the same file, replace the `Delete` mock method and recorder:

```go
// Delete mocks base method.
func (m *MockArticle) Delete(ctx context.Context, userID uint, id uint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, userID, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockArticleMockRecorder) Delete(ctx, userID, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockArticle)(nil).Delete), ctx, userID, id)
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/handlers/http/v1/article/mocks_test.go
git commit -m "fix(article): Update handler mocks to match new Update/Delete signatures"
```

---

### Task 7: Update Update handler — extract userID, handle 403, Swagger

**Files:**
- Modify: `internal/handlers/http/v1/article/update.go`

- [ ] **Step 1: Update the handler**

Replace `internal/handlers/http/v1/article/update.go`:

```go
package article

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	articledto "go-boilerplate/internal/dto/article"
	"go-boilerplate/internal/handlers/http/middleware"
	v1 "go-boilerplate/internal/handlers/http/v1"
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
// @Param       id path int true "Article ID"
// @Param       request body articledto.UpdateRequest true "Update Article request"
// @Success     200 {object} response.Response[articledto.Response]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
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

	userID := middleware.GetUserID(ctx)

	result, err := h.articleUC.Update(ctx.UserContext(), userID, uint(id), req)
	if err != nil {
		if errors.Is(err, articleuc.ErrNotFound) {
			return response.NotFound(ctx, "Article not found")
		}
		if errors.Is(err, articleuc.ErrForbidden) {
			return response.Forbidden(ctx, "Not authorized to update this article")
		}
		h.l.Error(err, "handlers - http - v1 - article - Update")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
```

Changes from current:
- Added `"go-boilerplate/internal/handlers/http/middleware"` import
- Added `userID := middleware.GetUserID(ctx)` line
- Changed `h.articleUC.Update(ctx.UserContext(), uint(id), req)` to `h.articleUC.Update(ctx.UserContext(), userID, uint(id), req)`
- Added `ErrForbidden` check mapping to 403
- Added `@Failure 403` Swagger annotation

- [ ] **Step 2: Commit**

```bash
git add internal/handlers/http/v1/article/update.go
git commit -m "fix(article): Extract userID in Update handler, add 403 handling"
```

---

### Task 8: Update Delete handler — extract userID, handle 403, Swagger

**Files:**
- Modify: `internal/handlers/http/v1/article/delete.go`

- [ ] **Step 1: Update the handler**

Replace `internal/handlers/http/v1/article/delete.go`:

```go
package article

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/handlers/http/middleware"
	articleuc "go-boilerplate/internal/usecase/article"
	"go-boilerplate/pkg/response"
)

// Delete godoc
// @Summary     Delete article
// @Description Delete a article by ID
// @ID          article-delete
// @Tags        articles
// @Accept      json
// @Produce     json
// @Param       id path int true "Article ID"
// @Success     204 "No Content"
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /articles/{id} [delete]
func (h *Handler) Delete(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		return response.BadRequest(ctx, "INVALID_ID", "Invalid article ID")
	}

	userID := middleware.GetUserID(ctx)

	if err := h.articleUC.Delete(ctx.UserContext(), userID, uint(id)); err != nil {
		if errors.Is(err, articleuc.ErrNotFound) {
			return response.NotFound(ctx, "Article not found")
		}
		if errors.Is(err, articleuc.ErrForbidden) {
			return response.Forbidden(ctx, "Not authorized to delete this article")
		}
		h.l.Error(err, "handlers - http - v1 - article - Delete")
		return response.InternalError(ctx)
	}

	return response.NoContent(ctx)
}
```

Changes from current:
- Added `"go-boilerplate/internal/handlers/http/middleware"` import
- Added `userID := middleware.GetUserID(ctx)` line
- Changed `h.articleUC.Delete(ctx.UserContext(), uint(id))` to `h.articleUC.Delete(ctx.UserContext(), userID, uint(id))`
- Added `ErrForbidden` check mapping to 403
- Added Swagger: `@Failure 400`, `@Failure 401`, `@Failure 403`, `@Security BearerAuth`

- [ ] **Step 2: Commit**

```bash
git add internal/handlers/http/v1/article/delete.go
git commit -m "fix(article): Extract userID in Delete handler, add 403 handling and Swagger"
```

---

### Task 9: Update handler tests — fix mock expectations, add 403 tests

**Files:**
- Modify: `internal/handlers/http/v1/article/handler_test.go`

- [ ] **Step 1: Update TestHandler_Create validBody**

In `TestHandler_Create`, the `validBody` JSON currently includes `"view_count": 1`. Remove it since the field no longer exists:

```go
	validBody := `{
		"title": "Test Article",
		"slug": "test-article",
		"content": "Article content here",
		"excerpt": "Short excerpt",
		"cover_media_id": 1,
		"status": "draft",
		"published_at": "` + now.Format(time.RFC3339) + `"
	}`
```

- [ ] **Step 2: Update TestHandler_Update mock expectations**

All `Update` mock calls need the `userID` parameter added. The test JWT token is for `userID: 1` (see `setupTestApp` line 35). Update every `mockArticleUC.EXPECT().Update(...)` call:

**"success" case** (line 312-315 currently):
```go
mockArticleUC.EXPECT().
	Update(gomock.Any(), uint(1), uint(1), articledto.UpdateRequest{
		Title: &title,
	}).
	Return(&articledto.Response{
		ID:    1,
		Title: "Updated Title",
	}, nil)
```

**"not found" case** (line 345-349 currently):
```go
mockArticleUC.EXPECT().
	Update(gomock.Any(), uint(1), uint(999), articledto.UpdateRequest{
		Title: &title,
	}).
	Return(nil, articleuc.ErrNotFound)
```

**"internal error" case** (line 359-363 currently):
```go
mockArticleUC.EXPECT().
	Update(gomock.Any(), uint(1), uint(1), articledto.UpdateRequest{
		Title: &title,
	}).
	Return(nil, errors.New("database error"))
```

Add a new **"forbidden" test case** to the `tests` slice:
```go
{
	name:    "forbidden - not article owner",
	id:      "1",
	body:    validBody,
	addAuth: true,
	setupMock: func() {
		mockArticleUC.EXPECT().
			Update(gomock.Any(), uint(1), uint(1), articledto.UpdateRequest{
				Title: &title,
			}).
			Return(nil, articleuc.ErrForbidden)
	},
	wantStatus: fiber.StatusForbidden,
},
```

- [ ] **Step 3: Update TestHandler_Delete mock expectations**

All `Delete` mock calls need the `userID` parameter added.

**"success" case** (line 417-419 currently):
```go
mockArticleUC.EXPECT().
	Delete(gomock.Any(), uint(1), uint(1)).
	Return(nil)
```

**"not found" case** (line 435-437 currently):
```go
mockArticleUC.EXPECT().
	Delete(gomock.Any(), uint(1), uint(999)).
	Return(articleuc.ErrNotFound)
```

**"internal error" case** (line 446-448 currently):
```go
mockArticleUC.EXPECT().
	Delete(gomock.Any(), uint(1), uint(1)).
	Return(errors.New("database error"))
```

Add a new **"forbidden" test case** to the `tests` slice:
```go
{
	name:    "forbidden - not article owner",
	id:      "1",
	addAuth: true,
	setupMock: func() {
		mockArticleUC.EXPECT().
			Delete(gomock.Any(), uint(1), uint(1)).
			Return(articleuc.ErrForbidden)
	},
	wantStatus: fiber.StatusForbidden,
},
```

- [ ] **Step 4: Run all handler tests**

Run: `go test -v ./internal/handlers/http/v1/article/`
Expected: PASS — all tests pass including new forbidden cases

- [ ] **Step 5: Commit**

```bash
git add internal/handlers/http/v1/article/handler_test.go
git commit -m "test(article): Update handler test expectations for userID param, add 403 tests"
```

---

## Chunk 4: Idempotency Fix and Final Verification

### Task 10: Scope idempotency cache key by method+path+user

**Files:**
- Modify: `internal/handlers/http/middleware/idempotency.go`
- Modify: `internal/handlers/http/middleware/idempotency_test.go`

- [ ] **Step 1: Write failing test for cross-endpoint isolation**

Add to `internal/handlers/http/middleware/idempotency_test.go`:

```go
func TestIdempotency_ScopedByEndpoint(t *testing.T) {
	t.Parallel()

	mc := newMockIdempotencyCache()
	cfg := config.Idempotency{
		Enabled:         true,
		TTL:             24 * time.Hour,
		RequiredForPost: false,
	}

	createCount := 0
	updateCount := 0

	app := fiber.New()
	app.Use(middleware.Idempotency(mc, cfg))
	app.Post("/create", func(c *fiber.Ctx) error {
		createCount++
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"action": "create"})
	})
	app.Post("/update", func(c *fiber.Ctx) error {
		updateCount++
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"action": "update"})
	})

	sameKey := "shared-key-123"

	// POST /create with key
	req1 := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/create", http.NoBody)
	req1.Header.Set("Idempotency-Key", sameKey)

	resp1, err := app.Test(req1)
	require.NoError(t, err)
	defer resp1.Body.Close() //nolint:errcheck // test

	assert.Equal(t, http.StatusCreated, resp1.StatusCode)
	assert.Equal(t, 1, createCount)

	// POST /update with SAME key — should NOT replay /create response
	req2 := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/update", http.NoBody)
	req2.Header.Set("Idempotency-Key", sameKey)

	resp2, err := app.Test(req2)
	require.NoError(t, err)
	defer resp2.Body.Close() //nolint:errcheck // test

	assert.Equal(t, http.StatusOK, resp2.StatusCode)
	assert.Equal(t, 1, updateCount, "update handler should be called (not replayed from /create)")
	assert.Empty(t, resp2.Header.Get("X-Idempotent-Replay"))
}

func TestIdempotency_ScopedByUser(t *testing.T) {
	t.Parallel()

	mc := newMockIdempotencyCache()
	cfg := config.Idempotency{
		Enabled:         true,
		TTL:             24 * time.Hour,
		RequiredForPost: false,
	}

	callCount := 0

	app := fiber.New()
	// Simulate JWTAuth by setting user_id in Locals via a preceding middleware
	app.Use(func(c *fiber.Ctx) error {
		if uid := c.Get("X-Test-UserID"); uid != "" {
			// Parse and set user_id (mimics what JWTAuth does)
			id, _ := strconv.ParseUint(uid, 10, 32)
			if id > 0 {
				c.Locals(middleware.UserIDKey, uint(id))
			}
		}
		return c.Next()
	})
	app.Use(middleware.Idempotency(mc, cfg))
	app.Post("/create", func(c *fiber.Ctx) error {
		callCount++
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": callCount})
	})

	sameKey := "same-key"

	// User 1 sends POST
	req1 := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/create", http.NoBody)
	req1.Header.Set("Idempotency-Key", sameKey)
	req1.Header.Set("X-Test-UserID", "1")

	resp1, err := app.Test(req1)
	require.NoError(t, err)
	defer resp1.Body.Close() //nolint:errcheck // test

	assert.Equal(t, http.StatusCreated, resp1.StatusCode)
	assert.Equal(t, 1, callCount)

	// User 2 sends same key on same endpoint — should NOT replay (different user)
	req2 := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/create", http.NoBody)
	req2.Header.Set("Idempotency-Key", sameKey)
	req2.Header.Set("X-Test-UserID", "2")

	resp2, err := app.Test(req2)
	require.NoError(t, err)
	defer resp2.Body.Close() //nolint:errcheck // test

	assert.Equal(t, http.StatusCreated, resp2.StatusCode)
	assert.Equal(t, 2, callCount, "different user should not get cached response")
	assert.Empty(t, resp2.Header.Get("X-Idempotent-Replay"))

	// User 1 sends same key again — SHOULD replay
	req3 := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/create", http.NoBody)
	req3.Header.Set("Idempotency-Key", sameKey)
	req3.Header.Set("X-Test-UserID", "1")

	resp3, err := app.Test(req3)
	require.NoError(t, err)
	defer resp3.Body.Close() //nolint:errcheck // test

	assert.Equal(t, http.StatusCreated, resp3.StatusCode)
	assert.Equal(t, 2, callCount, "user 1 should get replay from first request")
	assert.Equal(t, "true", resp3.Header.Get("X-Idempotent-Replay"))
}
```

Add `"strconv"` to the test file's import block for the user ID parsing in `TestIdempotency_ScopedByUser`.

- [ ] **Step 2: Run tests to verify the cross-endpoint test fails**

Run: `go test -v -run TestIdempotency_ScopedByEndpoint ./internal/handlers/http/middleware/`
Expected: FAIL — the `/update` request replays the `/create` response because the cache key doesn't include the path

- [ ] **Step 3: Update idempotency middleware to scope cache key**

In `internal/handlers/http/middleware/idempotency.go`, add `"strconv"` to imports and replace the cache key construction (line 48):

Replace:
```go
		cacheKey := "idempotency:" + key
```

With:
```go
		// Scope cache key by user (if authenticated) + method + path
		userPart := "shared"
		if id, ok := ctx.Locals(UserIDKey).(uint); ok {
			userPart = strconv.FormatUint(uint64(id), 10)
		}
		cacheKey := "idempotency:" + userPart + ":" + ctx.Method() + ":" + ctx.Path() + ":" + key
```

Add `"strconv"` to the import block.

- [ ] **Step 4: Run all idempotency tests**

Run: `go test -v -run TestIdempotency ./internal/handlers/http/middleware/`
Expected: PASS — all tests pass including new scoping tests

- [ ] **Step 5: Commit**

```bash
git add internal/handlers/http/middleware/idempotency.go internal/handlers/http/middleware/idempotency_test.go
git commit -m "fix(idempotency): Scope cache key by user+method+path to prevent collisions"
```

---

### Task 11: Run full check-all and regenerate Swagger

- [ ] **Step 1: Run make check-all**

Run: `make check-all`
Expected: PASS — all lints, tests, vulnerability checks pass

- [ ] **Step 2: Regenerate Swagger docs**

Run: `make swag`
Expected: Updated `docs/swagger.json` and `docs/docs.go` with new 403 failure annotations

- [ ] **Step 3: Run make check-all again after Swagger regeneration**

Run: `make check-all`
Expected: PASS

- [ ] **Step 4: Final commit**

```bash
git add docs/
git commit -m "docs: Regenerate Swagger docs with 403 error responses"
```
