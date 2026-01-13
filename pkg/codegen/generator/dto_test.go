package generator

import (
	"strings"
	"testing"

	"go-boilerplate/pkg/codegen/parser"
)

func TestBuildRequestDTOContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
		Fields: []parser.GoField{
			{Name: "ID", ColumnName: "id", Type: "uint", JSONTag: "id", GormTags: "primaryKey"},
			{Name: "UserID", ColumnName: "user_id", Type: "uint", JSONTag: "user_id", GormTags: "not null"},
			{Name: "Title", ColumnName: "title", Type: "string", JSONTag: "title", GormTags: "not null;size:255"},
			{Name: "Content", ColumnName: "content", Type: "*string", JSONTag: "content,omitempty"},
			{Name: "CreatedAt", ColumnName: "created_at", Type: "time.Time", JSONTag: "created_at", GormTags: "autoCreateTime"},
			{Name: "UpdatedAt", ColumnName: "updated_at", Type: "time.Time", JSONTag: "updated_at", GormTags: "autoUpdateTime"},
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildRequestDTOContent()

	// Check package declaration
	if !strings.Contains(content, "package article") {
		t.Error("expected package article declaration")
	}

	// Check CreateRequest struct
	if !strings.Contains(content, "type CreateRequest struct") {
		t.Error("expected CreateRequest struct")
	}

	// CreateRequest should have user_id and title (not id, created_at, updated_at)
	if strings.Contains(content, `json:"id"`) {
		t.Error("CreateRequest should not have id field")
	}
	if !strings.Contains(content, `json:"user_id"`) {
		t.Error("CreateRequest should have user_id field")
	}
	if !strings.Contains(content, `json:"title"`) {
		t.Error("CreateRequest should have title field")
	}

	// Check UpdateRequest struct
	if !strings.Contains(content, "type UpdateRequest struct") {
		t.Error("expected UpdateRequest struct")
	}

	// UpdateRequest should have pointer types for partial updates
	// Note: Fields in UpdateRequest should be pointers

	// Check ListRequest struct
	if !strings.Contains(content, "type ListRequest struct") {
		t.Error("expected ListRequest struct")
	}
	if !strings.Contains(content, `query:"page"`) {
		t.Error("expected page query param")
	}
	if !strings.Contains(content, `query:"page_size"`) {
		t.Error("expected page_size query param")
	}

	// Check helper methods
	if !strings.Contains(content, "func (r *ListRequest) GetPageSize()") {
		t.Error("expected GetPageSize method")
	}
	if !strings.Contains(content, "func (r *ListRequest) GetOffset()") {
		t.Error("expected GetOffset method")
	}
}

func TestBuildResponseDTOContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
		Fields: []parser.GoField{
			{Name: "ID", ColumnName: "id", Type: "uint", JSONTag: "id"},
			{Name: "UserID", ColumnName: "user_id", Type: "uint", JSONTag: "user_id"},
			{Name: "Title", ColumnName: "title", Type: "string", JSONTag: "title"},
			{Name: "CreatedAt", ColumnName: "created_at", Type: "time.Time", JSONTag: "created_at"},
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildResponseDTOContent()

	// Check package declaration
	if !strings.Contains(content, "package article") {
		t.Error("expected package article declaration")
	}

	// Check imports
	if !strings.Contains(content, `"go-boilerplate/internal/entity"`) {
		t.Error("expected entity import")
	}

	// Check Response struct
	if !strings.Contains(content, "type Response struct") {
		t.Error("expected Response struct")
	}

	// Check NewResponse constructor
	if !strings.Contains(content, "func NewResponse(article *entity.Article) *Response") {
		t.Error("expected NewResponse constructor")
	}

	// Check nil check in NewResponse
	if !strings.Contains(content, "if article == nil") {
		t.Error("expected nil check in NewResponse")
	}

	// Check ListResponse struct
	if !strings.Contains(content, "type ListResponse struct") {
		t.Error("expected ListResponse struct")
	}
	if !strings.Contains(content, `json:"data"`) {
		t.Error("expected data field in ListResponse")
	}
	if !strings.Contains(content, `json:"total"`) {
		t.Error("expected total field in ListResponse")
	}
	if !strings.Contains(content, `json:"page"`) {
		t.Error("expected page field in ListResponse")
	}
	if !strings.Contains(content, `json:"total_pages"`) {
		t.Error("expected total_pages field in ListResponse")
	}

	// Check NewListResponse constructor
	if !strings.Contains(content, "func NewListResponse(articles []*entity.Article") {
		t.Error("expected NewListResponse constructor")
	}
}

func TestResponseDTOWithTimeImport(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
		Fields: []parser.GoField{
			{Name: "ID", ColumnName: "id", Type: "uint", JSONTag: "id"},
			{Name: "CreatedAt", ColumnName: "created_at", Type: "time.Time", JSONTag: "created_at"},
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildResponseDTOContent()

	// Should have time import when fields contain time.Time
	if !strings.Contains(content, `"time"`) {
		t.Error("expected time import")
	}
}

func TestIsCreateRequestField(t *testing.T) {
	gen := New(Config{}, &parser.ParseResult{})

	tests := []struct {
		name     string
		field    string
		expected bool
	}{
		{"id excluded", "id", false},
		{"created_at excluded", "created_at", false},
		{"updated_at excluded", "updated_at", false},
		{"deleted_at excluded", "deleted_at", false},
		{"user_id included", "user_id", true},
		{"title included", "title", true},
		{"content included", "content", true},
		{"name included", "name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.isCreateRequestField(tt.field)
			if result != tt.expected {
				t.Errorf("isCreateRequestField(%q) = %v, want %v", tt.field, result, tt.expected)
			}
		})
	}
}

func TestIsUpdateRequestField(t *testing.T) {
	gen := New(Config{}, &parser.ParseResult{})

	tests := []struct {
		name     string
		field    string
		expected bool
	}{
		{"id excluded", "id", false},
		{"created_at excluded", "created_at", false},
		{"updated_at excluded", "updated_at", false},
		{"deleted_at excluded", "deleted_at", false},
		{"user_id included", "user_id", true},
		{"title included", "title", true},
		{"content included", "content", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.isUpdateRequestField(tt.field)
			if result != tt.expected {
				t.Errorf("isUpdateRequestField(%q) = %v, want %v", tt.field, result, tt.expected)
			}
		})
	}
}

func TestIsResponseField(t *testing.T) {
	gen := New(Config{}, &parser.ParseResult{})

	tests := []struct {
		name     string
		field    string
		expected bool
	}{
		{"password excluded", "password", false},
		{"deleted_at excluded", "deleted_at", false},
		{"id included", "id", true},
		{"user_id included", "user_id", true},
		{"title included", "title", true},
		{"created_at included", "created_at", true},
		{"updated_at included", "updated_at", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.isResponseField(tt.field)
			if result != tt.expected {
				t.Errorf("isResponseField(%q) = %v, want %v", tt.field, result, tt.expected)
			}
		})
	}
}

func TestDtoType(t *testing.T) {
	gen := New(Config{}, &parser.ParseResult{})

	tests := []struct {
		name       string
		entityType string
		expected   string
	}{
		{"string stays string", "string", "string"},
		{"pointer string removes pointer", "*string", "string"},
		{"gorm.DeletedAt becomes time.Time", "gorm.DeletedAt", "time.Time"},
		{"JSONMap becomes map", "JSONMap", "map[string]interface{}"},
		{"uint stays uint", "uint", "uint"},
		{"int64 stays int64", "int64", "int64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.dtoType(tt.entityType)
			if result != tt.expected {
				t.Errorf("dtoType(%q) = %q, want %q", tt.entityType, result, tt.expected)
			}
		})
	}
}

func TestResponseType(t *testing.T) {
	gen := New(Config{}, &parser.ParseResult{})

	tests := []struct {
		name       string
		entityType string
		expected   string
	}{
		{"string stays string", "string", "string"},
		{"pointer stays pointer", "*string", "*string"},
		{"gorm.DeletedAt becomes *time.Time", "gorm.DeletedAt", "*time.Time"},
		{"JSONMap becomes map", "JSONMap", "map[string]interface{}"},
		{"time.Time stays time.Time", "time.Time", "time.Time"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.responseType(tt.entityType)
			if result != tt.expected {
				t.Errorf("responseType(%q) = %q, want %q", tt.entityType, result, tt.expected)
			}
		})
	}
}

func TestGoFieldName(t *testing.T) {
	gen := New(Config{}, &parser.ParseResult{})

	tests := []struct {
		name       string
		columnName string
		expected   string
	}{
		{"id becomes ID", "id", "ID"},
		{"user_id becomes UserID", "user_id", "UserID"},
		{"api_key becomes APIKey", "api_key", "APIKey"},
		{"url becomes URL", "url", "URL"},
		{"created_at becomes CreatedAt", "created_at", "CreatedAt"},
		{"title becomes Title", "title", "Title"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.goFieldName(tt.columnName)
			if result != tt.expected {
				t.Errorf("goFieldName(%q) = %q, want %q", tt.columnName, result, tt.expected)
			}
		})
	}
}
