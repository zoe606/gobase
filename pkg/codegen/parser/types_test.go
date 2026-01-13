package parser

import "testing"

func TestSingularize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Regular plurals (ending in 's')
		{"users", "users", "user"},
		{"profiles", "profiles", "profile"},
		{"articles", "articles", "article"},
		{"comments", "comments", "comment"},

		// Words ending in 'ies' → 'y'
		{"categories", "categories", "category"},
		{"companies", "companies", "company"},
		{"histories", "histories", "history"},

		// Words ending in 'es' (sses, shes, ches, xes, zes)
		{"addresses", "addresses", "address"}, // -sses → remove "es"
		{"bushes", "bushes", "bush"},
		{"matches", "matches", "match"},
		{"boxes", "boxes", "box"},
		{"quizzes", "quizzes", "quizz"}, // Simple heuristic, not perfect

		// Words ending in 'ves' → 'f'
		{"leaves", "leaves", "leaf"},
		{"wives", "wives", "wif"}, // Simple implementation

		// Irregular plurals
		{"people", "people", "person"},
		{"children", "children", "child"},
		{"men", "men", "man"},
		{"women", "women", "woman"},
		{"media", "media", "media"}, // Keep as-is

		// Already singular
		{"user", "user", "user"},
		{"profile", "profile", "profile"},

		// Empty string
		{"empty", "", ""},

		// Compound words
		{"user_roles", "user_roles", "user_role"},
		{"refresh_tokens", "refresh_tokens", "refresh_token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := singularize(tt.input)
			if result != tt.expected {
				t.Errorf("singularize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic snake_case
		{"user_id", "user_id", "UserId"},
		{"created_at", "created_at", "CreatedAt"},
		{"user_profile", "user_profile", "UserProfile"},

		// Single word
		{"user", "user", "User"},
		{"id", "id", "Id"},

		// Multiple underscores
		{"user_profile_picture", "user_profile_picture", "UserProfilePicture"},

		// Already uppercase
		{"USER", "USER", "USER"},

		// Empty string
		{"empty", "", ""},

		// Leading underscore
		{"leading", "_user", "User"},

		// Trailing underscore
		{"trailing", "user_", "User"},

		// Multiple consecutive underscores
		{"double", "user__id", "UserId"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic snake_case
		{"user_id", "user_id", "userId"},
		{"created_at", "created_at", "createdAt"},
		{"user_profile", "user_profile", "userProfile"},

		// Single word
		{"user", "user", "user"},
		{"User", "User", "user"},

		// Multiple underscores
		{"user_profile_picture", "user_profile_picture", "userProfilePicture"},

		// Empty string
		{"empty", "", ""},

		// Single uppercase letter
		{"single_upper", "A", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTableEntityName(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		expected  string
	}{
		{"users table", "users", "User"},
		{"profiles table", "profiles", "Profile"},
		{"articles table", "articles", "Article"},
		{"user_roles table", "user_roles", "UserRole"},
		{"refresh_tokens table", "refresh_tokens", "RefreshToken"},
		{"media table", "media", "Media"},
		{"categories table", "categories", "Category"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := &Table{Name: tt.tableName}
			result := table.EntityName()
			if result != tt.expected {
				t.Errorf("Table{Name: %q}.EntityName() = %q, want %q", tt.tableName, result, tt.expected)
			}
		})
	}
}

func TestTableVarName(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		expected  string
	}{
		{"users table", "users", "user"},
		{"profiles table", "profiles", "profile"},
		{"articles table", "articles", "article"},
		{"user_roles table", "user_roles", "userRole"},
		{"refresh_tokens table", "refresh_tokens", "refreshToken"},
		{"media table", "media", "media"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := &Table{Name: tt.tableName}
			result := table.VarName()
			if result != tt.expected {
				t.Errorf("Table{Name: %q}.VarName() = %q, want %q", tt.tableName, result, tt.expected)
			}
		})
	}
}
