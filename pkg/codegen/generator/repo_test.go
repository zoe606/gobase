package generator

import (
	"strings"
	"testing"

	"go-boilerplate/pkg/codegen/parser"
)

func TestBuildRepoInterfaceContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildRepoInterfaceContent()

	// Check interface name
	if !strings.Contains(content, "ArticleRepo interface") {
		t.Error("expected ArticleRepo interface")
	}

	// Check CRUD method signatures
	methods := []string{
		"Create(ctx context.Context, article *entity.Article) error",
		"GetByID(ctx context.Context, id uint) (*entity.Article, error)",
		"List(ctx context.Context, limit, offset int) ([]*entity.Article, int64, error)",
		"Update(ctx context.Context, article *entity.Article) error",
		"Delete(ctx context.Context, id uint) error",
	}

	for _, method := range methods {
		if !strings.Contains(content, method) {
			t.Errorf("expected method signature: %s", method)
		}
	}

	// Check interface documentation
	if !strings.Contains(content, "// ArticleRepo defines Article repository operations.") {
		t.Error("expected interface documentation")
	}
}

func TestBuildRepoImplContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildRepoImplContent()

	// Check package declaration
	if !strings.Contains(content, "package persistent") {
		t.Error("expected package persistent")
	}

	// Check imports
	if !strings.Contains(content, `"context"`) {
		t.Error("expected context import")
	}
	if !strings.Contains(content, `"errors"`) {
		t.Error("expected errors import")
	}
	if !strings.Contains(content, `"go-boilerplate/internal/entity"`) {
		t.Error("expected entity import")
	}
	if !strings.Contains(content, `"go-boilerplate/internal/repo"`) {
		t.Error("expected repo import")
	}
	if !strings.Contains(content, `"gorm.io/gorm"`) {
		t.Error("expected gorm import")
	}

	// Check struct
	if !strings.Contains(content, "type ArticleRepo struct") {
		t.Error("expected ArticleRepo struct")
	}
	if !strings.Contains(content, "db *gorm.DB") {
		t.Error("expected db field")
	}

	// Check constructor
	if !strings.Contains(content, "func NewArticleRepo(db *gorm.DB) *ArticleRepo") {
		t.Error("expected NewArticleRepo constructor")
	}

	// Check CRUD methods exist
	methods := []string{
		"func (r *ArticleRepo) Create(ctx context.Context",
		"func (r *ArticleRepo) GetByID(ctx context.Context",
		"func (r *ArticleRepo) List(ctx context.Context",
		"func (r *ArticleRepo) Update(ctx context.Context",
		"func (r *ArticleRepo) Delete(ctx context.Context",
	}

	for _, method := range methods {
		if !strings.Contains(content, method) {
			t.Errorf("expected method: %s", method)
		}
	}
}

func TestRepoCreateMethod(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildRepoImplContent()

	// Check Create implementation
	if !strings.Contains(content, "r.db.WithContext(ctx).Create(article)") {
		t.Error("expected Create implementation with WithContext")
	}
}

func TestRepoGetByIDMethod(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildRepoImplContent()

	// Check GetByID implementation
	if !strings.Contains(content, "r.db.WithContext(ctx).First(&article, id)") {
		t.Error("expected First query in GetByID")
	}

	// Check ErrNotFound handling
	if !strings.Contains(content, "errors.Is(err, gorm.ErrRecordNotFound)") {
		t.Error("expected ErrRecordNotFound check")
	}
	if !strings.Contains(content, "return nil, repo.ErrNotFound") {
		t.Error("expected repo.ErrNotFound return")
	}
}

func TestRepoListMethod(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildRepoImplContent()

	// Check List implementation has count
	if !strings.Contains(content, ".Count(&total)") {
		t.Error("expected Count query")
	}

	// Check pagination
	if !strings.Contains(content, "Limit(limit)") {
		t.Error("expected Limit clause")
	}
	if !strings.Contains(content, "Offset(offset)") {
		t.Error("expected Offset clause")
	}

	// Check ordering
	if !strings.Contains(content, `Order("id DESC")`) {
		t.Error("expected Order clause")
	}
}

func TestRepoUpdateMethod(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildRepoImplContent()

	// Check Update implementation
	if !strings.Contains(content, "r.db.WithContext(ctx).Save(article)") {
		t.Error("expected Save in Update")
	}

	// Check RowsAffected check
	if !strings.Contains(content, "result.RowsAffected == 0") {
		t.Error("expected RowsAffected check")
	}
}

func TestRepoDeleteMethod(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildRepoImplContent()

	// Check Delete implementation
	if !strings.Contains(content, "r.db.WithContext(ctx).Delete(&entity.Article{}, id)") {
		t.Error("expected Delete implementation")
	}

	// Check RowsAffected check
	if !strings.Contains(content, "result.RowsAffected == 0") {
		t.Error("expected RowsAffected check in Delete")
	}
}

func TestRepoWithDifferentTableNames(t *testing.T) {
	tests := []struct {
		name           string
		tableName      string
		expectedStruct string
		expectedVar    string
	}{
		{
			name:           "users table",
			tableName:      "users",
			expectedStruct: "UserRepo",
			expectedVar:    "user",
		},
		{
			name:           "user_roles table",
			tableName:      "user_roles",
			expectedStruct: "UserRoleRepo",
			expectedVar:    "userRole",
		},
		{
			name:           "media table",
			tableName:      "media",
			expectedStruct: "MediaRepo",
			expectedVar:    "media",
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
			content := gen.buildRepoImplContent()

			if !strings.Contains(content, "type "+tt.expectedStruct+" struct") {
				t.Errorf("expected struct %s", tt.expectedStruct)
			}

			// Check variable name is used in Create method
			expectedCreate := "Create(ctx context.Context, " + tt.expectedVar + " *entity."
			if !strings.Contains(content, expectedCreate) {
				t.Errorf("expected Create with variable %s", tt.expectedVar)
			}
		})
	}
}
