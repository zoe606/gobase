package generator

import (
	"strings"
	"testing"

	"go-boilerplate/pkg/codegen/parser"
)

func TestBuildEntityContent(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name:    "users",
			Comment: "User accounts",
		},
		Fields: []parser.GoField{
			{Name: "ID", ColumnName: "id", Type: "uint", JSONTag: "id", GormTags: "primaryKey"},
			{Name: "Name", ColumnName: "name", Type: "string", JSONTag: "name", GormTags: "not null;size:255"},
			{Name: "Email", ColumnName: "email", Type: "string", JSONTag: "email", GormTags: "uniqueIndex;not null;size:255"},
			{Name: "CreatedAt", ColumnName: "created_at", Type: "time.Time", JSONTag: "created_at", GormTags: "autoCreateTime"},
		},
		Imports: []string{"time"},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildEntityContent()

	// Check package declaration
	if !strings.Contains(content, "package entity") {
		t.Error("expected package entity declaration")
	}

	// Check import
	if !strings.Contains(content, `"time"`) {
		t.Error("expected time import")
	}

	// Check struct name
	if !strings.Contains(content, "type User struct") {
		t.Error("expected User struct declaration")
	}

	// Check struct comment
	if !strings.Contains(content, "// User User accounts") {
		t.Error("expected struct comment")
	}

	// Check fields
	if !strings.Contains(content, "ID") {
		t.Error("expected ID field")
	}
	if !strings.Contains(content, "Name") {
		t.Error("expected Name field")
	}
	if !strings.Contains(content, "Email") {
		t.Error("expected Email field")
	}

	// Check GORM tags
	if !strings.Contains(content, `gorm:"primaryKey"`) {
		t.Error("expected primaryKey gorm tag")
	}
	if !strings.Contains(content, `gorm:"uniqueIndex`) {
		t.Error("expected uniqueIndex gorm tag")
	}

	// Check JSON tags
	if !strings.Contains(content, `json:"id"`) {
		t.Error("expected id json tag")
	}
	if !strings.Contains(content, `json:"email"`) {
		t.Error("expected email json tag")
	}

	// Check TableName method
	if !strings.Contains(content, `func (User) TableName() string`) {
		t.Error("expected TableName method")
	}
	if !strings.Contains(content, `return "users"`) {
		t.Error("expected return users in TableName")
	}
}

func TestBuildEntityImports(t *testing.T) {
	tests := []struct {
		name            string
		imports         []string
		fields          []parser.GoField
		expectedImports []string
	}{
		{
			name:            "time import from result",
			imports:         []string{"time"},
			fields:          []parser.GoField{},
			expectedImports: []string{"time"},
		},
		{
			name:    "gorm import for DeletedAt",
			imports: []string{},
			fields: []parser.GoField{
				{Name: "DeletedAt", Type: "gorm.DeletedAt"},
			},
			expectedImports: []string{"gorm.io/gorm"},
		},
		{
			name:    "both time and gorm",
			imports: []string{"time"},
			fields: []parser.GoField{
				{Name: "DeletedAt", Type: "gorm.DeletedAt"},
			},
			expectedImports: []string{"gorm.io/gorm", "time"},
		},
		{
			name:            "no imports",
			imports:         []string{},
			fields:          []parser.GoField{},
			expectedImports: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := New(Config{}, &parser.ParseResult{
				Imports: tt.imports,
				Fields:  tt.fields,
			})

			imports := gen.buildEntityImports()

			if len(imports) != len(tt.expectedImports) {
				t.Errorf("import count = %d, want %d", len(imports), len(tt.expectedImports))
			}

			for _, expected := range tt.expectedImports {
				found := false
				for _, imp := range imports {
					if imp == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected import %q not found", expected)
				}
			}
		})
	}
}

func TestEntityWithForeignKey(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "articles",
		},
		Fields: []parser.GoField{
			{Name: "ID", ColumnName: "id", Type: "uint", JSONTag: "id", GormTags: "primaryKey"},
			{Name: "UserID", ColumnName: "user_id", Type: "uint", JSONTag: "user_id", GormTags: "not null"},
			{Name: "Title", ColumnName: "title", Type: "string", JSONTag: "title", GormTags: "not null;size:255"},
		},
		Relations: []parser.GoRelation{
			{Name: "User", Type: "*User", ForeignKey: "UserID", JSONTag: "user,omitempty", GormTags: "foreignKey:UserID"},
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildEntityContent()

	// Check FK field
	if !strings.Contains(content, "UserID") {
		t.Error("expected UserID field")
	}

	// Check relation field
	if !strings.Contains(content, "User") && !strings.Contains(content, "*User") {
		t.Error("expected User relation field")
	}

	// Check relation GORM tag
	if !strings.Contains(content, `gorm:"foreignKey:UserID"`) {
		t.Error("expected foreignKey gorm tag")
	}

	// Check relation JSON tag
	if !strings.Contains(content, `json:"user,omitempty"`) {
		t.Error("expected user json tag with omitempty")
	}
}

func TestEntityWithJSONB(t *testing.T) {
	parseResult := &parser.ParseResult{
		Table: parser.Table{
			Name: "settings",
		},
		Fields: []parser.GoField{
			{Name: "ID", ColumnName: "id", Type: "uint", JSONTag: "id", GormTags: "primaryKey"},
			{Name: "Config", ColumnName: "config", Type: "JSONMap", JSONTag: "config,omitempty", GormTags: "type:jsonb"},
		},
	}

	gen := New(Config{ModuleName: "go-boilerplate"}, parseResult)
	content := gen.buildEntityContent()

	// Check JSONMap type
	if !strings.Contains(content, "JSONMap") {
		t.Error("expected JSONMap type")
	}

	// Check GORM jsonb tag
	if !strings.Contains(content, `gorm:"type:jsonb"`) {
		t.Error("expected type:jsonb gorm tag")
	}
}

func TestBuildFieldLine(t *testing.T) {
	gen := New(Config{}, &parser.ParseResult{})

	tests := []struct {
		name     string
		field    string
		goType   string
		jsonTag  string
		gormTags string
		contains []string
	}{
		{
			name:     "basic field",
			field:    "Name",
			goType:   "string",
			jsonTag:  "name",
			gormTags: "",
			contains: []string{"Name", "string", `json:"name"`},
		},
		{
			name:     "field with gorm tag",
			field:    "ID",
			goType:   "uint",
			jsonTag:  "id",
			gormTags: "primaryKey",
			contains: []string{"ID", "uint", `json:"id"`, `gorm:"primaryKey"`},
		},
		{
			name:     "pointer type",
			field:    "Bio",
			goType:   "*string",
			jsonTag:  "bio,omitempty",
			gormTags: "",
			contains: []string{"Bio", "*string", `json:"bio,omitempty"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := gen.buildFieldLine(tt.field, tt.goType, tt.jsonTag, tt.gormTags)

			for _, c := range tt.contains {
				if !strings.Contains(line, c) {
					t.Errorf("expected %q in line: %s", c, line)
				}
			}
		})
	}
}

func TestEntityName(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		expected  string
	}{
		{"users table", "users", "User"},
		{"articles table", "articles", "Article"},
		{"user_roles table", "user_roles", "UserRole"},
		{"media table", "media", "Media"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := New(Config{}, &parser.ParseResult{
				Table: parser.Table{Name: tt.tableName},
			})

			if gen.entityName() != tt.expected {
				t.Errorf("entityName() = %q, want %q", gen.entityName(), tt.expected)
			}
		})
	}
}

func TestPackageName(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		expected  string
	}{
		{"users table", "users", "user"},
		{"articles table", "articles", "article"},
		{"user_roles table", "user_roles", "userrole"},
		{"media table", "media", "media"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := New(Config{}, &parser.ParseResult{
				Table: parser.Table{Name: tt.tableName},
			})

			if gen.packageName() != tt.expected {
				t.Errorf("packageName() = %q, want %q", gen.packageName(), tt.expected)
			}
		})
	}
}
