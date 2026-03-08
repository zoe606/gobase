package generator

import (
	"strings"
	"testing"

	"go-boilerplate/pkg/codegen/parser"
)

func TestBuildUseCaseInterfaceContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildUseCaseInterfaceContent()

	// Check import hint references the DTO package path
	if !strings.Contains(content, `"go-boilerplate/internal/dto/article"`) {
		t.Error("expected import hint for internal/dto/article")
	}

	// Check interface name
	if !strings.Contains(content, "Article interface") {
		t.Error("expected Article interface")
	}

	// Check method signatures with DTO package prefix (articledto)
	if !strings.Contains(content, "Create(ctx context.Context, req articledto.CreateRequest) (*articledto.Response, error)") {
		t.Error("expected Create method with DTO types")
	}
	if !strings.Contains(content, "GetByID(ctx context.Context, id uint) (*articledto.Response, error)") {
		t.Error("expected GetByID method")
	}
	if !strings.Contains(content, "List(ctx context.Context, req articledto.ListRequest) (*articledto.ListResponse, error)") {
		t.Error("expected List method")
	}
	if !strings.Contains(content, "Update(ctx context.Context, id uint, req articledto.UpdateRequest) (*articledto.Response, error)") {
		t.Error("expected Update method")
	}
	if !strings.Contains(content, "Delete(ctx context.Context, id uint) error") {
		t.Error("expected Delete method")
	}
}

func TestBuildUseCaseMainContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildUseCaseMainContent()

	// Check package declaration
	if !strings.Contains(content, "package article") {
		t.Error("expected package article")
	}

	// Check go:generate directive
	if !strings.Contains(content, "//go:generate mockgen") {
		t.Error("expected go:generate mockgen directive")
	}

	// Check repo import
	if !strings.Contains(content, `"go-boilerplate/internal/repo"`) {
		t.Error("expected repo import")
	}

	// Check struct
	if !strings.Contains(content, "type UseCase struct") {
		t.Error("expected UseCase struct")
	}
	if !strings.Contains(content, "articleRepo repo.ArticleRepo") {
		t.Error("expected articleRepo field")
	}

	// Check constructor
	if !strings.Contains(content, "func New(articleRepo repo.ArticleRepo) *UseCase") {
		t.Error("expected New constructor")
	}
}

func TestBuildUseCaseErrorsContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildUseCaseErrorsContent()

	// Check package
	if !strings.Contains(content, "package article") {
		t.Error("expected package article")
	}

	// Check error definitions
	if !strings.Contains(content, "ErrNotFound = errors.New") {
		t.Error("expected ErrNotFound")
	}
	if !strings.Contains(content, "ErrAlreadyExists = errors.New") {
		t.Error("expected ErrAlreadyExists")
	}
	if !strings.Contains(content, "ErrInvalid = errors.New") {
		t.Error("expected ErrInvalid")
	}

	// Check error messages contain the package name
	if !strings.Contains(content, `"article not found"`) {
		t.Error("expected article not found message")
	}
}

func TestBuildUseCaseCreateContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
		Fields: []parser.GoField{
			{Name: "ID", ColumnName: "id", Type: "uint"},
			{Name: "Title", ColumnName: "title", Type: "string"},
			{Name: "Slug", ColumnName: "slug", Type: "string"},
			{Name: "Content", ColumnName: "content", Type: "string"},
			{Name: "CreatedAt", ColumnName: "created_at", Type: "time.Time"},
			{Name: "UpdatedAt", ColumnName: "updated_at", Type: "time.Time"},
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildUseCaseCreateContent()

	// Check package
	if !strings.Contains(content, "package article") {
		t.Error("expected package article")
	}

	// Check imports (no alias needed since package is articledto)
	if !strings.Contains(content, `"context"`) {
		t.Error("expected context import")
	}
	if !strings.Contains(content, `"go-boilerplate/internal/dto/article"`) {
		t.Error("expected DTO import")
	}
	if !strings.Contains(content, `"go-boilerplate/internal/entity"`) {
		t.Error("expected entity import")
	}

	// Check method signature
	if !strings.Contains(content, "func (uc *UseCase) Create(ctx context.Context, req articledto.CreateRequest) (*articledto.Response, error)") {
		t.Error("expected Create method signature")
	}

	// Check entity field mapping from request
	if !strings.Contains(content, "Title: req.Title,") {
		t.Error("expected Title field mapping from request")
	}
	if !strings.Contains(content, "Slug: req.Slug,") {
		t.Error("expected Slug field mapping from request")
	}
	if !strings.Contains(content, "Content: req.Content,") {
		t.Error("expected Content field mapping from request")
	}

	// Should NOT include auto-generated fields
	if strings.Contains(content, "ID: req.ID,") {
		t.Error("should not map ID field from request")
	}
	if strings.Contains(content, "CreatedAt: req.CreatedAt,") {
		t.Error("should not map CreatedAt field from request")
	}

	// Check repo call
	if !strings.Contains(content, "uc.articleRepo.Create(ctx, article)") {
		t.Error("expected articleRepo.Create call")
	}

	// Check response
	if !strings.Contains(content, "articledto.NewResponse(article)") {
		t.Error("expected NewResponse call")
	}
}

func TestBuildUseCaseGetByIDContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildUseCaseGetByIDContent()

	// Check method signature
	if !strings.Contains(content, "func (uc *UseCase) GetByID(ctx context.Context, id uint) (*articledto.Response, error)") {
		t.Error("expected GetByID method signature")
	}

	// Check ErrNotFound handling
	if !strings.Contains(content, "errors.Is(err, repo.ErrNotFound)") {
		t.Error("expected repo.ErrNotFound check")
	}
	if !strings.Contains(content, "return nil, ErrNotFound") {
		t.Error("expected ErrNotFound return")
	}

	// Check response
	if !strings.Contains(content, "articledto.NewResponse(article)") {
		t.Error("expected NewResponse call")
	}
}

func TestBuildUseCaseListContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildUseCaseListContent()

	// Check method signature
	if !strings.Contains(content, "func (uc *UseCase) List(ctx context.Context, req articledto.ListRequest) (*articledto.ListResponse, error)") {
		t.Error("expected List method signature")
	}

	// Check Normalize call
	if !strings.Contains(content, "req.Params.Normalize()") {
		t.Error("expected Normalize() call")
	}

	// Check repo call with req.Params
	if !strings.Contains(content, "uc.articleRepo.List(ctx, req.Params)") {
		t.Error("expected articleRepo.List call with req.Params")
	}

	// Check response with req.Params
	if !strings.Contains(content, "articledto.NewListResponse(articles, total, req.Params)") {
		t.Error("expected NewListResponse call with req.Params")
	}
}

func TestBuildUseCaseUpdateContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
		Fields: []parser.GoField{
			{Name: "ID", ColumnName: "id", Type: "uint"},
			{Name: "Title", ColumnName: "title", Type: "string"},
			{Name: "Slug", ColumnName: "slug", Type: "string"},
			{Name: "Content", ColumnName: "content", Type: "string"},
			{Name: "CreatedAt", ColumnName: "created_at", Type: "time.Time"},
			{Name: "UpdatedAt", ColumnName: "updated_at", Type: "time.Time"},
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildUseCaseUpdateContent()

	// Check method signature
	if !strings.Contains(content, "func (uc *UseCase) Update(ctx context.Context, id uint, req articledto.UpdateRequest) (*articledto.Response, error)") {
		t.Error("expected Update method signature")
	}

	// Check GetByID call to fetch existing
	if !strings.Contains(content, "uc.articleRepo.GetByID(ctx, id)") {
		t.Error("expected GetByID call to fetch existing record")
	}

	// Check ErrNotFound handling
	if !strings.Contains(content, "errors.Is(err, repo.ErrNotFound)") {
		t.Error("expected repo.ErrNotFound check")
	}
	if !strings.Contains(content, "return nil, ErrNotFound") {
		t.Error("expected ErrNotFound return")
	}

	// Check pointer field update checks
	if !strings.Contains(content, "if req.Title != nil {") {
		t.Error("expected pointer check for Title")
	}
	if !strings.Contains(content, "article.Title = *req.Title") {
		t.Error("expected Title assignment from pointer")
	}
	if !strings.Contains(content, "if req.Slug != nil {") {
		t.Error("expected pointer check for Slug")
	}
	if !strings.Contains(content, "if req.Content != nil {") {
		t.Error("expected pointer check for Content")
	}

	// Should NOT include auto-generated fields
	if strings.Contains(content, "if req.ID != nil {") {
		t.Error("should not have pointer check for ID")
	}
	if strings.Contains(content, "if req.CreatedAt != nil {") {
		t.Error("should not have pointer check for CreatedAt")
	}

	// Check Update call
	if !strings.Contains(content, "uc.articleRepo.Update(ctx, article)") {
		t.Error("expected Update call")
	}

	// Check response
	if !strings.Contains(content, "articledto.NewResponse(article)") {
		t.Error("expected NewResponse call")
	}
}

func TestBuildUseCaseDeleteContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildUseCaseDeleteContent()

	// Check method signature
	if !strings.Contains(content, "func (uc *UseCase) Delete(ctx context.Context, id uint) error") {
		t.Error("expected Delete method signature")
	}

	// Check repo call
	if !strings.Contains(content, "uc.articleRepo.Delete(ctx, id)") {
		t.Error("expected articleRepo.Delete call")
	}

	// Check ErrNotFound handling
	if !strings.Contains(content, "errors.Is(err, repo.ErrNotFound)") {
		t.Error("expected repo.ErrNotFound check")
	}
	if !strings.Contains(content, "return ErrNotFound") {
		t.Error("expected ErrNotFound return")
	}
}

func TestBuildUseCaseTestContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)

	// Test Create test content
	t.Run("Create", func(t *testing.T) {
		content := gen.buildUseCaseTestContent("Create")

		// Check package (should be _test)
		if !strings.Contains(content, "package article_test") {
			t.Error("expected package article_test")
		}

		// Check imports
		if !strings.Contains(content, `"context"`) {
			t.Error("expected context import")
		}
		if !strings.Contains(content, `"testing"`) {
			t.Error("expected testing import")
		}
		if !strings.Contains(content, `"go-boilerplate/internal/usecase/article"`) {
			t.Error("expected usecase import")
		}

		// Check test function
		if !strings.Contains(content, "func TestCreate(t *testing.T)") {
			t.Error("expected TestCreate function")
		}

		// Check table-driven test structure
		if !strings.Contains(content, "t.Parallel()") {
			t.Error("expected t.Parallel()")
		}
		if !strings.Contains(content, "for _, tt := range tests") {
			t.Error("expected table-driven test loop")
		}
		if !strings.Contains(content, "require.NoError") {
			t.Error("expected require.NoError")
		}
		if !strings.Contains(content, "require.Error") {
			t.Error("expected require.Error")
		}
		if !strings.Contains(content, "mockFn") {
			t.Error("expected mockFn in test struct")
		}
		if !strings.Contains(content, "mockArticleRepo") {
			t.Error("expected mockArticleRepo usage")
		}

		// Should NOT contain t.Skip
		if strings.Contains(content, "t.Skip") {
			t.Error("should not contain t.Skip")
		}

		// Check New constructor usage with mock
		if !strings.Contains(content, "article.New(mockRepo)") {
			t.Error("expected article.New(mockRepo) usage")
		}
	})

	// Test GetByID test content
	t.Run("GetByID", func(t *testing.T) {
		content := gen.buildUseCaseTestContent("GetByID")

		if !strings.Contains(content, "func TestGetByID(t *testing.T)") {
			t.Error("expected TestGetByID function")
		}
		if !strings.Contains(content, "repo.ErrNotFound") {
			t.Error("expected repo.ErrNotFound in not found test case")
		}
		if !strings.Contains(content, "for _, tt := range tests") {
			t.Error("expected table-driven test loop")
		}
		if strings.Contains(content, "t.Skip") {
			t.Error("should not contain t.Skip")
		}
	})

	// Test List test content
	t.Run("List", func(t *testing.T) {
		content := gen.buildUseCaseTestContent("List")

		if !strings.Contains(content, "func TestList(t *testing.T)") {
			t.Error("expected TestList function")
		}
		if !strings.Contains(content, "pagination.NewParams()") {
			t.Error("expected pagination.NewParams() in test")
		}
		if !strings.Contains(content, "for _, tt := range tests") {
			t.Error("expected table-driven test loop")
		}
		if strings.Contains(content, "t.Skip") {
			t.Error("should not contain t.Skip")
		}
	})

	// Test Update test content
	t.Run("Update", func(t *testing.T) {
		content := gen.buildUseCaseTestContent("Update")

		if !strings.Contains(content, "func TestUpdate(t *testing.T)") {
			t.Error("expected TestUpdate function")
		}
		if !strings.Contains(content, "repo.ErrNotFound") {
			t.Error("expected repo.ErrNotFound in not found test case")
		}
		if !strings.Contains(content, "for _, tt := range tests") {
			t.Error("expected table-driven test loop")
		}
		if strings.Contains(content, "t.Skip") {
			t.Error("should not contain t.Skip")
		}
	})

	// Test Delete test content
	t.Run("Delete", func(t *testing.T) {
		content := gen.buildUseCaseTestContent("Delete")

		if !strings.Contains(content, "func TestDelete(t *testing.T)") {
			t.Error("expected TestDelete function")
		}
		if !strings.Contains(content, "repo.ErrNotFound") {
			t.Error("expected repo.ErrNotFound in not found test case")
		}
		if !strings.Contains(content, "for _, tt := range tests") {
			t.Error("expected table-driven test loop")
		}
		if strings.Contains(content, "t.Skip") {
			t.Error("should not contain t.Skip")
		}
	})
}

func TestBuildUseCaseMocksTestContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildUseCaseMocksTestContent()

	// Check package
	if !strings.Contains(content, "package article_test") {
		t.Error("expected package article_test")
	}

	// Check mock struct
	if !strings.Contains(content, "mockArticleRepo") {
		t.Error("expected mockArticleRepo struct")
	}

	// Check function fields
	if !strings.Contains(content, "createFn") {
		t.Error("expected createFn field")
	}
	if !strings.Contains(content, "getByIDFn") {
		t.Error("expected getByIDFn field")
	}
	if !strings.Contains(content, "listFn") {
		t.Error("expected listFn field")
	}
	if !strings.Contains(content, "updateFn") {
		t.Error("expected updateFn field")
	}
	if !strings.Contains(content, "deleteFn") {
		t.Error("expected deleteFn field")
	}

	// Check method implementations delegate to function fields
	if !strings.Contains(content, "m.createFn(ctx,") {
		t.Error("expected Create method to delegate to createFn")
	}
	if !strings.Contains(content, "m.getByIDFn(ctx,") {
		t.Error("expected GetByID method to delegate to getByIDFn")
	}
	if !strings.Contains(content, "m.listFn(ctx,") {
		t.Error("expected List method to delegate to listFn")
	}
	if !strings.Contains(content, "m.updateFn(ctx,") {
		t.Error("expected Update method to delegate to updateFn")
	}
	if !strings.Contains(content, "m.deleteFn(ctx,") {
		t.Error("expected Delete method to delegate to deleteFn")
	}

	// Check entity and pagination imports
	if !strings.Contains(content, `"go-boilerplate/internal/entity"`) {
		t.Error("expected entity import")
	}
	if !strings.Contains(content, `"go-boilerplate/pkg/pagination"`) {
		t.Error("expected pagination import")
	}
}

func TestUseCaseWithDifferentTableNames(t *testing.T) {
	tests := []struct {
		name           string
		tableName      string
		expectedPkg    string
		expectedEntity string
		expectedVar    string
	}{
		{
			name:           "users table",
			tableName:      "users",
			expectedPkg:    "package user",
			expectedEntity: "User",
			expectedVar:    "userRepo",
		},
		{
			name:           "user_roles table",
			tableName:      "user_roles",
			expectedPkg:    "package userrole",
			expectedEntity: "UserRole",
			expectedVar:    "userRoleRepo",
		},
		{
			name:           "media table",
			tableName:      "media",
			expectedPkg:    "package media",
			expectedEntity: "Media",
			expectedVar:    "mediaRepo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parseResult := &parser.ParseResult{
				Table: parser.Table{
					Name: tt.tableName,
				},
			}

			gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
			content := gen.buildUseCaseMainContent()

			if !strings.Contains(content, tt.expectedPkg) {
				t.Errorf("expected package %s", tt.expectedPkg)
			}
			if !strings.Contains(content, tt.expectedVar+" repo."+tt.expectedEntity+"Repo") {
				t.Errorf("expected %s repo.%sRepo field", tt.expectedVar, tt.expectedEntity)
			}
		})
	}
}

func TestBuildUseCaseCreateContentWithFields(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "products",
		},
		Fields: []parser.GoField{
			{Name: "ID", ColumnName: "id", Type: "uint"},
			{Name: "Name", ColumnName: "name", Type: "string"},
			{Name: "Price", ColumnName: "price", Type: "float64"},
			{Name: "CategoryID", ColumnName: "category_id", Type: "uint"},
			{Name: "CreatedAt", ColumnName: "created_at", Type: "time.Time"},
			{Name: "UpdatedAt", ColumnName: "updated_at", Type: "time.Time"},
			{Name: "DeletedAt", ColumnName: "deleted_at", Type: "gorm.DeletedAt"},
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildUseCaseCreateContent()

	// Should include user-defined fields
	if !strings.Contains(content, "Name: req.Name,") {
		t.Error("expected Name field mapping")
	}
	if !strings.Contains(content, "Price: req.Price,") {
		t.Error("expected Price field mapping")
	}
	if !strings.Contains(content, "CategoryID: req.CategoryID,") {
		t.Error("expected CategoryID field mapping")
	}

	// Should NOT include auto-generated fields
	if strings.Contains(content, "ID: req.ID") {
		t.Error("should not map ID")
	}
	if strings.Contains(content, "CreatedAt: req.CreatedAt") {
		t.Error("should not map CreatedAt")
	}
	if strings.Contains(content, "UpdatedAt: req.UpdatedAt") {
		t.Error("should not map UpdatedAt")
	}
	if strings.Contains(content, "DeletedAt: req.DeletedAt") {
		t.Error("should not map DeletedAt")
	}
}

func TestBuildUseCaseUpdateContentWithFields(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "products",
		},
		Fields: []parser.GoField{
			{Name: "ID", ColumnName: "id", Type: "uint"},
			{Name: "Name", ColumnName: "name", Type: "string"},
			{Name: "Price", ColumnName: "price", Type: "float64"},
			{Name: "CreatedAt", ColumnName: "created_at", Type: "time.Time"},
			{Name: "UpdatedAt", ColumnName: "updated_at", Type: "time.Time"},
			{Name: "DeletedAt", ColumnName: "deleted_at", Type: "gorm.DeletedAt"},
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildUseCaseUpdateContent()

	// Check pointer field checks for user-defined fields
	if !strings.Contains(content, "if req.Name != nil {") {
		t.Error("expected pointer check for Name")
	}
	if !strings.Contains(content, "product.Name = *req.Name") {
		t.Error("expected Name assignment from pointer")
	}
	if !strings.Contains(content, "if req.Price != nil {") {
		t.Error("expected pointer check for Price")
	}
	if !strings.Contains(content, "product.Price = *req.Price") {
		t.Error("expected Price assignment from pointer")
	}

	// Should NOT include auto-generated fields
	if strings.Contains(content, "if req.ID != nil {") {
		t.Error("should not have pointer check for ID")
	}
	if strings.Contains(content, "if req.CreatedAt != nil {") {
		t.Error("should not have pointer check for CreatedAt")
	}
	if strings.Contains(content, "if req.DeletedAt != nil {") {
		t.Error("should not have pointer check for DeletedAt")
	}
}
