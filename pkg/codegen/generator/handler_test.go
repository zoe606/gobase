package generator

import (
	"strings"
	"testing"

	"go-boilerplate/pkg/codegen/parser"
)

func TestBuildHandlerMainContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildHandlerMainContent()

	// Check package declaration
	if !strings.Contains(content, "package article") {
		t.Error("expected package article")
	}

	// Check imports
	if !strings.Contains(content, `"github.com/go-playground/validator/v10"`) {
		t.Error("expected validator import")
	}
	if !strings.Contains(content, `"github.com/gofiber/fiber/v2"`) {
		t.Error("expected fiber import")
	}
	if !strings.Contains(content, `"go-boilerplate/internal/usecase"`) {
		t.Error("expected usecase import")
	}
	if !strings.Contains(content, `"go-boilerplate/pkg/logger"`) {
		t.Error("expected logger import")
	}

	// Check Handler struct
	if !strings.Contains(content, "type Handler struct") {
		t.Error("expected Handler struct")
	}
	if !strings.Contains(content, "articleUC usecase.Article") {
		t.Error("expected articleUC field")
	}
	if !strings.Contains(content, "l   logger.Interface") {
		t.Error("expected logger field")
	}
	if !strings.Contains(content, "v   *validator.Validate") {
		t.Error("expected validator field")
	}

	// Check constructor
	if !strings.Contains(content, "func New(articleUC usecase.Article, l logger.Interface) *Handler") {
		t.Error("expected New constructor")
	}

	// Check RegisterRoutes
	if !strings.Contains(content, "func (h *Handler) RegisterRoutes(router fiber.Router)") {
		t.Error("expected RegisterRoutes method")
	}
}

func TestHandlerRegisterRoutes(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildHandlerMainContent()

	// Check route group
	if !strings.Contains(content, `router.Group("/articles")`) {
		t.Error("expected articles route group")
	}

	// Check CRUD routes
	routes := []string{
		`articles.Post("/", h.Create)`,
		`articles.Get("/", h.List)`,
		`articles.Get("/:id", h.GetByID)`,
		`articles.Put("/:id", h.Update)`,
		`articles.Delete("/:id", h.Delete)`,
	}

	for _, route := range routes {
		if !strings.Contains(content, route) {
			t.Errorf("expected route: %s", route)
		}
	}
}

func TestBuildHandlerCreateContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildHandlerCreateContent()

	// Check package
	if !strings.Contains(content, "package article") {
		t.Error("expected package article")
	}

	// Check imports
	if !strings.Contains(content, `articledto "go-boilerplate/internal/dto/article"`) {
		t.Error("expected DTO import with alias")
	}
	if !strings.Contains(content, `"go-boilerplate/pkg/response"`) {
		t.Error("expected response import")
	}

	// Check Swagger annotations
	if !strings.Contains(content, "// @Summary     Create article") {
		t.Error("expected Swagger summary")
	}
	if !strings.Contains(content, "// @Tags        articles") {
		t.Error("expected Swagger tags")
	}
	if !strings.Contains(content, "// @Router      /articles [post]") {
		t.Error("expected Swagger router")
	}

	// Check method
	if !strings.Contains(content, "func (h *Handler) Create(ctx *fiber.Ctx) error") {
		t.Error("expected Create method")
	}

	// Check body parser
	if !strings.Contains(content, "ctx.BodyParser(&req)") {
		t.Error("expected BodyParser call")
	}

	// Check validation
	if !strings.Contains(content, "h.v.Struct(req)") {
		t.Error("expected validation call")
	}

	// Check UC call
	if !strings.Contains(content, "h.articleUC.Create(ctx.UserContext(), req)") {
		t.Error("expected articleUC.Create call")
	}

	// Check response
	if !strings.Contains(content, "response.Created(ctx, result)") {
		t.Error("expected response.Created")
	}
}

func TestBuildHandlerGetByIDContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildHandlerGetByIDContent()

	// Check Swagger
	if !strings.Contains(content, "// @Summary     Get article by ID") {
		t.Error("expected Swagger summary")
	}
	if !strings.Contains(content, "// @Param       id path int true") {
		t.Error("expected Swagger id param")
	}
	if !strings.Contains(content, "// @Router      /articles/{id} [get]") {
		t.Error("expected Swagger router")
	}

	// Check ID parsing
	if !strings.Contains(content, `strconv.ParseUint(ctx.Params("id")`) {
		t.Error("expected ID parsing")
	}

	// Check NotFound handling
	if !strings.Contains(content, "errors.Is(err, articleuc.ErrNotFound)") {
		t.Error("expected ErrNotFound check")
	}
	if !strings.Contains(content, `response.NotFound(ctx, "Article not found")`) {
		t.Error("expected NotFound response")
	}
}

func TestBuildHandlerListContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildHandlerListContent()

	// Check Swagger
	if !strings.Contains(content, "// @Summary     List articles") {
		t.Error("expected Swagger summary")
	}
	if !strings.Contains(content, `// @Param       page query int false "Page number"`) {
		t.Error("expected Swagger page param")
	}
	if !strings.Contains(content, `// @Param       page_size query int false "Page size"`) {
		t.Error("expected Swagger page_size param")
	}

	// Check query parser
	if !strings.Contains(content, "ctx.QueryParser(&req)") {
		t.Error("expected QueryParser call")
	}

	// Check UC call
	if !strings.Contains(content, "h.articleUC.List(ctx.UserContext(), req)") {
		t.Error("expected articleUC.List call")
	}
}

func TestBuildHandlerUpdateContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildHandlerUpdateContent()

	// Check Swagger
	if !strings.Contains(content, "// @Summary     Update article") {
		t.Error("expected Swagger summary")
	}
	if !strings.Contains(content, "// @Router      /articles/{id} [put]") {
		t.Error("expected Swagger router")
	}

	// Check ID parsing
	if !strings.Contains(content, `strconv.ParseUint(ctx.Params("id")`) {
		t.Error("expected ID parsing")
	}

	// Check body parser
	if !strings.Contains(content, "ctx.BodyParser(&req)") {
		t.Error("expected BodyParser call")
	}

	// Check validation
	if !strings.Contains(content, "h.v.Struct(req)") {
		t.Error("expected validation call")
	}

	// Check UC call
	if !strings.Contains(content, "h.articleUC.Update(ctx.UserContext(), uint(id), req)") {
		t.Error("expected articleUC.Update call")
	}
}

func TestBuildHandlerDeleteContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildHandlerDeleteContent()

	// Check Swagger
	if !strings.Contains(content, "// @Summary     Delete article") {
		t.Error("expected Swagger summary")
	}
	if !strings.Contains(content, `// @Success     204 "No Content"`) {
		t.Error("expected Swagger 204 success")
	}
	if !strings.Contains(content, "// @Router      /articles/{id} [delete]") {
		t.Error("expected Swagger router")
	}

	// Check UC call
	if !strings.Contains(content, "h.articleUC.Delete(ctx.UserContext(), uint(id))") {
		t.Error("expected articleUC.Delete call")
	}

	// Check NoContent response
	if !strings.Contains(content, "response.NoContent(ctx)") {
		t.Error("expected NoContent response")
	}
}

func TestSwaggerAnnotations(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)

	methods := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:    "Create",
			content: gen.buildHandlerCreateContent(),
			expected: []string{
				"// @Summary",
				"// @Description",
				"// @ID",
				"// @Tags",
				"// @Accept",
				"// @Produce",
				"// @Param",
				"// @Success",
				"// @Failure",
				"// @Router",
			},
		},
		{
			name:    "GetByID",
			content: gen.buildHandlerGetByIDContent(),
			expected: []string{
				"// @Summary",
				"// @Description",
				"// @ID",
				"// @Tags",
				"// @Param       id path",
				"// @Router",
			},
		},
		{
			name:    "List",
			content: gen.buildHandlerListContent(),
			expected: []string{
				"// @Summary",
				"// @Param       page query",
				"// @Param       page_size query",
				"// @Router",
			},
		},
		{
			name:    "Update",
			content: gen.buildHandlerUpdateContent(),
			expected: []string{
				"// @Summary",
				"// @Param       id path",
				"// @Param       request body",
				"// @Router",
			},
		},
		{
			name:    "Delete",
			content: gen.buildHandlerDeleteContent(),
			expected: []string{
				"// @Summary",
				"// @Param       id path",
				"// @Success     204",
				"// @Router",
			},
		},
	}

	for _, method := range methods {
		t.Run(method.name, func(t *testing.T) {
			for _, expected := range method.expected {
				if !strings.Contains(method.content, expected) {
					t.Errorf("%s: expected Swagger annotation %q", method.name, expected)
				}
			}
		})
	}
}

func TestHandlerWithDifferentTableNames(t *testing.T) {
	tests := []struct {
		name          string
		tableName     string
		expectedPkg   string
		expectedRoute string
		expectedUC    string
	}{
		{
			name:          "users table",
			tableName:     "users",
			expectedPkg:   "package user",
			expectedRoute: `router.Group("/users")`,
			expectedUC:    "userUC usecase.User",
		},
		{
			name:          "user_roles table",
			tableName:     "user_roles",
			expectedPkg:   "package userrole",
			expectedRoute: `router.Group("/userRoles")`,
			expectedUC:    "userRoleUC usecase.UserRole",
		},
		{
			name:          "media table",
			tableName:     "media",
			expectedPkg:   "package media",
			expectedRoute: `router.Group("/medias")`,
			expectedUC:    "mediaUC usecase.Media",
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
			content := gen.buildHandlerMainContent()

			if !strings.Contains(content, tt.expectedPkg) {
				t.Errorf("expected package %s", tt.expectedPkg)
			}
			if !strings.Contains(content, tt.expectedRoute) {
				t.Errorf("expected route %s", tt.expectedRoute)
			}
			if !strings.Contains(content, tt.expectedUC) {
				t.Errorf("expected UC field %s", tt.expectedUC)
			}
		})
	}
}
