// Package parser provides SQL parsing functionality for code generation.
package parser

// Table represents a parsed SQL table definition.
type Table struct {
	Name        string       // Table name (e.g., "profiles")
	Columns     []Column     // Column definitions
	Constraints []Constraint // Table-level constraints
	Indexes     []Index      // Index definitions
	Comment     string       // Table comment if any
}

// Column represents a single column in a table.
type Column struct {
	Name         string      // Column name (e.g., "user_id")
	SQLType      string      // Original SQL type (e.g., "VARCHAR(255)")
	BaseType     string      // Base type without size (e.g., "VARCHAR")
	Size         int         // Size/length if applicable (e.g., 255)
	Precision    int         // Precision for numeric types
	Scale        int         // Scale for numeric types
	IsNullable   bool        // Whether NULL is allowed
	IsPrimaryKey bool        // Whether this is a primary key
	IsUnique     bool        // Whether this has a unique constraint
	HasDefault   bool        // Whether a default value is set
	DefaultValue string      // Default value expression
	ForeignKey   *ForeignKey // Foreign key reference if any
	Comment      string      // Column comment if any
}

// ForeignKey represents a foreign key constraint.
type ForeignKey struct {
	RefTable  string // Referenced table name
	RefColumn string // Referenced column name
	OnDelete  string // ON DELETE action (CASCADE, SET NULL, etc.)
	OnUpdate  string // ON UPDATE action
}

// Constraint represents a table-level constraint.
type Constraint struct {
	Name    string   // Constraint name
	Type    string   // PRIMARY KEY, UNIQUE, FOREIGN KEY, CHECK
	Columns []string // Columns involved in the constraint
}

// Index represents an index definition.
type Index struct {
	Name    string   // Index name
	Columns []string // Columns in the index
	Unique  bool     // Whether this is a unique index
}

// GoField represents the Go struct field generated from a column.
type GoField struct {
	Name       string // Go field name (PascalCase)
	ColumnName string // Original SQL column name (snake_case)
	Type       string // Go type (e.g., "string", "*uint", "time.Time")
	JSONTag    string // JSON tag value
	GormTags   string // GORM tag value
	Comment    string // Field comment
}

// GoRelation represents a relation field in the Go struct.
type GoRelation struct {
	Name       string // Relation field name (e.g., "User")
	Type       string // Relation type (e.g., "*User")
	ForeignKey string // Foreign key field name (e.g., "UserID")
	JSONTag    string // JSON tag value
	GormTags   string // GORM tag value
}

// ParseResult contains the complete parsing result for code generation.
type ParseResult struct {
	Table     Table        // Parsed table definition
	Fields    []GoField    // Generated Go fields
	Relations []GoRelation // Generated relation fields
	Imports   []string     // Required imports for the entity
}

// EntityName returns the singular PascalCase entity name from table name.
// E.g., "profiles" -> "Profile", "user_roles" -> "UserRole".
func (t *Table) EntityName() string {
	return toPascalCase(singularize(t.Name))
}

// VarName returns the camelCase variable name from table name.
// E.g., "profiles" -> "profile", "user_roles" -> "userRole".
func (t *Table) VarName() string {
	return toCamelCase(singularize(t.Name))
}

// Helper functions for name conversion.

// toPascalCase converts snake_case to PascalCase.
func toPascalCase(s string) string {
	if s == "" {
		return s
	}

	result := make([]byte, 0, len(s))
	capitalizeNext := true

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' {
			capitalizeNext = true
			continue
		}

		if capitalizeNext && c >= 'a' && c <= 'z' {
			result = append(result, c-32) // Convert to uppercase
		} else {
			result = append(result, c)
		}
		capitalizeNext = false
	}

	return string(result)
}

// toCamelCase converts snake_case to camelCase.
func toCamelCase(s string) string {
	pascal := toPascalCase(s)
	if pascal == "" {
		return pascal
	}

	// Lowercase the first character
	first := pascal[0]
	if first >= 'A' && first <= 'Z' {
		return string(first+32) + pascal[1:]
	}
	return pascal
}

// irregularPlurals maps irregular plural forms to their singular.
var irregularPlurals = map[string]string{
	"people":   "person",
	"children": "child",
	"men":      "man",
	"women":    "woman",
	"media":    "media", // Keep as-is (already singular in this context)
}

// singularize converts a plural table name to singular.
// Simple implementation that handles common cases.
func singularize(s string) string {
	if s == "" {
		return s
	}

	// Handle irregular plurals first
	if singular, ok := irregularPlurals[s]; ok {
		return singular
	}

	// Handle common plural endings using helper functions
	if result, ok := singularizeIES(s); ok {
		return result
	}
	if result, ok := singularizeVES(s); ok {
		return result
	}
	if result, ok := singularizeES(s); ok {
		return result
	}
	if result, ok := singularizeS(s); ok {
		return result
	}

	return s
}

// singularizeIES handles -ies -> -y transformation.
func singularizeIES(s string) (string, bool) {
	if len(s) > 3 && s[len(s)-3:] == "ies" {
		return s[:len(s)-3] + "y", true
	}
	return "", false
}

// singularizeVES handles -ves -> -f transformation.
func singularizeVES(s string) (string, bool) {
	if len(s) > 3 && s[len(s)-3:] == "ves" {
		return s[:len(s)-3] + "f", true
	}
	return "", false
}

// singularizeES handles -es endings (-ses, -xes, -zes, -ches, -shes).
func singularizeES(s string) (string, bool) {
	if len(s) <= 2 || s[len(s)-2:] != "es" {
		return "", false
	}

	// Handle -sses, -shes, -ches
	if len(s) > 3 {
		suffix := s[len(s)-4 : len(s)-2]
		if suffix == "ss" || suffix == "sh" || suffix == "ch" {
			return s[:len(s)-2], true
		}
	}

	// Handle -xes, -zes
	if len(s) > 2 && (s[len(s)-3] == 'x' || s[len(s)-3] == 'z') {
		return s[:len(s)-2], true
	}

	// For other -es endings, just remove 's'
	return s[:len(s)-1], true
}

// singularizeS handles simple -s removal.
func singularizeS(s string) (string, bool) {
	if len(s) > 1 && s[len(s)-1] == 's' && s[len(s)-2] != 's' {
		return s[:len(s)-1], true
	}
	return "", false
}
