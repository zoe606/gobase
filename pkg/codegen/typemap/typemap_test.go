package typemap

import (
	"testing"
)

func TestMapColumn(t *testing.T) {
	tests := []struct {
		name         string
		sqlType      string
		size         int
		isNullable   bool
		isPrimaryKey bool
		isUnique     bool
		hasDefault   bool
		defaultValue string
		columnName   string
		expectedType string
		expectTags   []string
	}{
		// Integer types
		{
			name:         "SERIAL primary key",
			sqlType:      "SERIAL",
			isPrimaryKey: true,
			columnName:   "id",
			expectedType: "uint",
			expectTags:   []string{"primaryKey"},
		},
		{
			name:         "BIGSERIAL primary key",
			sqlType:      "BIGSERIAL",
			isPrimaryKey: true,
			columnName:   "id",
			expectedType: "uint",
			expectTags:   []string{"primaryKey"},
		},
		{
			name:         "INTEGER not null",
			sqlType:      "INTEGER",
			isNullable:   false,
			columnName:   "count",
			expectedType: "int",
			expectTags:   []string{"not null"},
		},
		{
			name:         "INTEGER nullable",
			sqlType:      "INTEGER",
			isNullable:   true,
			columnName:   "count",
			expectedType: "*int",
		},
		{
			name:         "BIGINT nullable",
			sqlType:      "BIGINT",
			isNullable:   true,
			columnName:   "big_count",
			expectedType: "*int64",
		},

		// String types
		{
			name:         "VARCHAR with size",
			sqlType:      "VARCHAR(255)",
			size:         255,
			isNullable:   false,
			columnName:   "name",
			expectedType: "string",
			expectTags:   []string{"not null", "size:255"},
		},
		{
			name:         "TEXT nullable",
			sqlType:      "TEXT",
			isNullable:   true,
			columnName:   "bio",
			expectedType: "*string",
		},

		// Boolean
		{
			name:         "BOOLEAN not null",
			sqlType:      "BOOLEAN",
			isNullable:   false,
			columnName:   "is_active",
			expectedType: "bool",
			expectTags:   []string{"not null"},
		},
		{
			name:         "BOOL nullable",
			sqlType:      "BOOL",
			isNullable:   true,
			columnName:   "is_verified",
			expectedType: "*bool",
		},

		// Timestamp types
		{
			name:         "TIMESTAMP WITH TIME ZONE nullable",
			sqlType:      "TIMESTAMP WITH TIME ZONE",
			isNullable:   true,
			columnName:   "published_at",
			expectedType: "*time.Time",
		},
		{
			name:         "TIMESTAMPTZ nullable",
			sqlType:      "TIMESTAMPTZ",
			isNullable:   true,
			columnName:   "expires_at",
			expectedType: "*time.Time",
		},

		// JSON types
		{
			name:         "JSONB",
			sqlType:      "JSONB",
			isNullable:   true,
			columnName:   "data",
			expectedType: "JSONMap",
			expectTags:   []string{"type:jsonb"},
		},
		{
			name:         "JSON",
			sqlType:      "JSON",
			isNullable:   true,
			columnName:   "config",
			expectedType: "JSONMap",
			expectTags:   []string{"type:json"},
		},

		// Numeric types
		{
			name:         "REAL nullable",
			sqlType:      "REAL",
			isNullable:   true,
			columnName:   "rate",
			expectedType: "*float32",
		},
		{
			name:         "DOUBLE PRECISION nullable",
			sqlType:      "DOUBLE PRECISION",
			isNullable:   true,
			columnName:   "amount",
			expectedType: "*float64",
		},
		{
			name:         "NUMERIC nullable",
			sqlType:      "NUMERIC",
			isNullable:   true,
			columnName:   "price",
			expectedType: "*float64",
		},

		// Binary
		{
			name:         "BYTEA",
			sqlType:      "BYTEA",
			isNullable:   true,
			columnName:   "data",
			expectedType: "[]byte",
		},

		// UUID
		{
			name:         "UUID",
			sqlType:      "UUID",
			isNullable:   true,
			columnName:   "external_id",
			expectedType: "*string",
			expectTags:   []string{"type:uuid"},
		},

		// Unique constraint
		{
			name:         "VARCHAR unique",
			sqlType:      "VARCHAR(255)",
			size:         255,
			isNullable:   false,
			isUnique:     true,
			columnName:   "email",
			expectedType: "string",
			expectTags:   []string{"uniqueIndex", "not null", "size:255"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapColumn(
				tt.sqlType,
				tt.size,
				tt.isNullable,
				tt.isPrimaryKey,
				tt.isUnique,
				tt.hasDefault,
				tt.defaultValue,
				tt.columnName,
			)

			if result.GoType != tt.expectedType {
				t.Errorf("GoType = %q, want %q", result.GoType, tt.expectedType)
			}

			// Check expected tags are present
			for _, tag := range tt.expectTags {
				found := false
				for _, resultTag := range result.GormTags {
					if resultTag == tag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected tag %q not found in %v", tag, result.GormTags)
				}
			}
		})
	}
}

func TestMapColumnWithFK(t *testing.T) {
	tests := []struct {
		name         string
		sqlType      string
		isNullable   bool
		isForeignKey bool
		expectedType string
	}{
		{
			name:         "INTEGER FK NOT NULL should be uint",
			sqlType:      "INTEGER",
			isNullable:   false,
			isForeignKey: true,
			expectedType: "uint",
		},
		{
			name:         "BIGINT FK NOT NULL should be uint",
			sqlType:      "BIGINT",
			isNullable:   false,
			isForeignKey: true,
			expectedType: "uint",
		},
		{
			name:         "INTEGER FK nullable should be *uint",
			sqlType:      "INTEGER",
			isNullable:   true,
			isForeignKey: true,
			expectedType: "*uint",
		},
		{
			name:         "INTEGER non-FK nullable stays *int",
			sqlType:      "INTEGER",
			isNullable:   true,
			isForeignKey: false,
			expectedType: "*int",
		},
		{
			name:         "INTEGER non-FK not null stays int",
			sqlType:      "INTEGER",
			isNullable:   false,
			isForeignKey: false,
			expectedType: "int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapColumnWithFK(
				tt.sqlType,
				0,
				tt.isNullable,
				false,
				false,
				false,
				"",
				"some_column",
				tt.isForeignKey,
			)

			if result.GoType != tt.expectedType {
				t.Errorf("GoType = %q, want %q", result.GoType, tt.expectedType)
			}
		})
	}
}

func TestMapColumnSpecialCases(t *testing.T) {
	t.Run("deleted_at should be gorm.DeletedAt", func(t *testing.T) {
		result := MapColumn(
			"TIMESTAMP WITH TIME ZONE",
			0,
			true,
			false,
			false,
			false,
			"",
			"deleted_at",
		)

		if result.GoType != "gorm.DeletedAt" {
			t.Errorf("GoType = %q, want %q", result.GoType, "gorm.DeletedAt")
		}

		// Should have gorm import
		hasGormImport := false
		for _, imp := range result.Imports {
			if imp == "gorm.io/gorm" {
				hasGormImport = true
				break
			}
		}
		if !hasGormImport {
			t.Error("expected gorm.io/gorm import for deleted_at")
		}
	})

	t.Run("created_at with default should not be pointer", func(t *testing.T) {
		result := MapColumn(
			"TIMESTAMP WITH TIME ZONE",
			0,
			true,
			false,
			false,
			true,
			"CURRENT_TIMESTAMP",
			"created_at",
		)

		if result.GoType != "time.Time" {
			t.Errorf("GoType = %q, want %q", result.GoType, "time.Time")
		}

		// Should have autoCreateTime tag
		hasAutoCreate := false
		for _, tag := range result.GormTags {
			if tag == "autoCreateTime" {
				hasAutoCreate = true
				break
			}
		}
		if !hasAutoCreate {
			t.Error("expected autoCreateTime tag for created_at")
		}
	})

	t.Run("updated_at with default should not be pointer", func(t *testing.T) {
		result := MapColumn(
			"TIMESTAMP WITH TIME ZONE",
			0,
			true,
			false,
			false,
			true,
			"CURRENT_TIMESTAMP",
			"updated_at",
		)

		if result.GoType != "time.Time" {
			t.Errorf("GoType = %q, want %q", result.GoType, "time.Time")
		}

		// Should have autoUpdateTime tag
		hasAutoUpdate := false
		for _, tag := range result.GormTags {
			if tag == "autoUpdateTime" {
				hasAutoUpdate = true
				break
			}
		}
		if !hasAutoUpdate {
			t.Error("expected autoUpdateTime tag for updated_at")
		}
	})
}

func TestToGoFieldName(t *testing.T) {
	tests := []struct {
		name       string
		columnName string
		expected   string
	}{
		// Acronyms
		{"id column", "id", "ID"},
		{"user_id column", "user_id", "UserID"},
		{"api_key column", "api_key", "APIKey"},
		{"http_url column", "http_url", "HTTPURL"},
		{"json_data column", "json_data", "JSONData"},
		{"xml_config column", "xml_config", "XMLConfig"},
		{"sql_query column", "sql_query", "SQLQuery"},
		{"ip_address column", "ip_address", "IPAddress"},
		{"uuid column", "uuid", "UUID"},
		{"uri column", "uri", "URI"},

		// Regular columns
		{"created_at", "created_at", "CreatedAt"},
		{"user_profile", "user_profile", "UserProfile"},
		{"is_active", "is_active", "IsActive"},

		// Single words
		{"name", "name", "Name"},
		{"email", "email", "Email"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToGoFieldName(tt.columnName)
			if result != tt.expected {
				t.Errorf("ToGoFieldName(%q) = %q, want %q", tt.columnName, result, tt.expected)
			}
		})
	}
}

func TestToJSONTag(t *testing.T) {
	tests := []struct {
		name       string
		columnName string
		isNullable bool
		expected   string
	}{
		{
			name:       "non-nullable",
			columnName: "user_id",
			isNullable: false,
			expected:   "user_id",
		},
		{
			name:       "nullable",
			columnName: "bio",
			isNullable: true,
			expected:   "bio,omitempty",
		},
		{
			name:       "id non-nullable",
			columnName: "id",
			isNullable: false,
			expected:   "id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToJSONTag(tt.columnName, tt.isNullable)
			if result != tt.expected {
				t.Errorf("ToJSONTag(%q, %v) = %q, want %q", tt.columnName, tt.isNullable, result, tt.expected)
			}
		})
	}
}

func TestNormalizeType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"uppercase", "varchar(255)", "VARCHAR"},
		{"with spaces", "  integer  ", "INTEGER"},
		{"with size", "VARCHAR(100)", "VARCHAR"},
		{"compound type", "timestamp with time zone", "TIMESTAMP WITH TIME ZONE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeType(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsTimestampDefault(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP", true},
		{"NOW()", "NOW()", true},
		{"CURRENT_DATE", "CURRENT_DATE", true},
		{"CURRENT_TIME", "CURRENT_TIME", true},
		{"lowercase now", "now()", true},
		{"regular value", "default_value", false},
		{"number", "0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTimestampDefault(tt.value)
			if result != tt.expected {
				t.Errorf("isTimestampDefault(%q) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}
