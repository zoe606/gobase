// Package typemap provides SQL to Go type mapping for code generation.
package typemap

import (
	"fmt"
	"strings"
)

// GoTypeInfo contains information about a Go type mapped from SQL.
type GoTypeInfo struct {
	Type       string // Go type (e.g., "string", "int", "time.Time")
	GormTag    string // Default GORM tag (e.g., "primaryKey", "type:jsonb")
	Import     string // Required import if any (e.g., "time")
	NeedsSize  bool   // Whether the type needs size in GORM tag
	IsNullable bool   // Whether the type should be a pointer when nullable
}

// SQLToGoType maps SQL types to Go types.
var SQLToGoType = map[string]GoTypeInfo{
	// Integer types
	"SERIAL":      {Type: "uint", GormTag: "primaryKey", IsNullable: false},
	"BIGSERIAL":   {Type: "uint", GormTag: "primaryKey", IsNullable: false},
	"SMALLSERIAL": {Type: "uint", GormTag: "primaryKey", IsNullable: false},
	"INTEGER":     {Type: "int", IsNullable: true},
	"INT":         {Type: "int", IsNullable: true},
	"INT4":        {Type: "int", IsNullable: true},
	"BIGINT":      {Type: "int64", IsNullable: true},
	"INT8":        {Type: "int64", IsNullable: true},
	"SMALLINT":    {Type: "int16", IsNullable: true},
	"INT2":        {Type: "int16", IsNullable: true},

	// Boolean
	"BOOLEAN": {Type: "bool", IsNullable: true},
	"BOOL":    {Type: "bool", IsNullable: true},

	// String types
	"VARCHAR":           {Type: "string", NeedsSize: true, IsNullable: true},
	"CHARACTER VARYING": {Type: "string", NeedsSize: true, IsNullable: true},
	"CHAR":              {Type: "string", NeedsSize: true, IsNullable: true},
	"CHARACTER":         {Type: "string", NeedsSize: true, IsNullable: true},
	"TEXT":              {Type: "string", IsNullable: true},

	// Timestamp types
	"TIMESTAMP WITH TIME ZONE":    {Type: "time.Time", Import: "time", IsNullable: true},
	"TIMESTAMP WITHOUT TIME ZONE": {Type: "time.Time", Import: "time", IsNullable: true},
	"TIMESTAMPTZ":                 {Type: "time.Time", Import: "time", IsNullable: true},
	"TIMESTAMP":                   {Type: "time.Time", Import: "time", IsNullable: true},
	"DATE":                        {Type: "time.Time", Import: "time", IsNullable: true},
	"TIME WITH TIME ZONE":         {Type: "time.Time", Import: "time", IsNullable: true},
	"TIME WITHOUT TIME ZONE":      {Type: "time.Time", Import: "time", IsNullable: true},
	"TIME":                        {Type: "time.Time", Import: "time", IsNullable: true},

	// JSON types
	"JSONB": {Type: "JSONMap", GormTag: "type:jsonb", IsNullable: true},
	"JSON":  {Type: "JSONMap", GormTag: "type:json", IsNullable: true},

	// Numeric types
	"REAL":             {Type: "float32", IsNullable: true},
	"FLOAT4":           {Type: "float32", IsNullable: true},
	"DOUBLE PRECISION": {Type: "float64", IsNullable: true},
	"FLOAT8":           {Type: "float64", IsNullable: true},
	"NUMERIC":          {Type: "float64", IsNullable: true},
	"DECIMAL":          {Type: "float64", IsNullable: true},

	// Binary types
	"BYTEA": {Type: "[]byte", IsNullable: true},

	// UUID
	"UUID": {Type: "string", GormTag: "type:uuid", IsNullable: true},
}

// MapResult contains the result of mapping an SQL column to Go.
type MapResult struct {
	GoType   string   // The Go type to use
	GormTags []string // GORM struct tags
	Imports  []string // Required imports
}

// MapColumn maps an SQL column to a Go type with appropriate tags.
func MapColumn(
	sqlType string,
	size int,
	isNullable bool,
	isPrimaryKey bool,
	isUnique bool,
	hasDefault bool,
	defaultValue string,
	columnName string,
) MapResult {
	return MapColumnWithFK(sqlType, size, isNullable, isPrimaryKey, isUnique, hasDefault, defaultValue, columnName, false)
}

// columnContext holds context for column mapping.
type columnContext struct {
	sqlType      string
	size         int
	isNullable   bool
	isPrimaryKey bool
	isUnique     bool
	hasDefault   bool
	defaultValue string
	columnName   string
	isForeignKey bool
	baseType     string
	typeInfo     GoTypeInfo
}

// MapColumnWithFK maps an SQL column to a Go type, considering foreign key status.
func MapColumnWithFK(
	sqlType string,
	size int,
	isNullable bool,
	isPrimaryKey bool,
	isUnique bool,
	hasDefault bool,
	defaultValue string,
	columnName string,
	isForeignKey bool,
) MapResult {
	ctx := &columnContext{
		sqlType:      sqlType,
		size:         size,
		isNullable:   isNullable,
		isPrimaryKey: isPrimaryKey,
		isUnique:     isUnique,
		hasDefault:   hasDefault,
		defaultValue: defaultValue,
		columnName:   columnName,
		isForeignKey: isForeignKey,
		baseType:     normalizeType(sqlType),
	}

	// Look up the type info
	typeInfo, found := SQLToGoType[ctx.baseType]
	if !found {
		typeInfo = GoTypeInfo{Type: "string", IsNullable: true}
	}
	ctx.typeInfo = typeInfo

	// Handle special deleted_at case first
	if result, ok := handleDeletedAt(ctx); ok {
		return result
	}

	result := MapResult{
		GormTags: make([]string, 0),
		Imports:  make([]string, 0),
	}

	// Determine Go type
	result.GoType = determineGoType(ctx)

	// Add import if needed
	if ctx.typeInfo.Import != "" {
		result.Imports = append(result.Imports, ctx.typeInfo.Import)
	}

	// Build GORM tags
	result.GormTags = buildGormTags(ctx)

	return result
}

// handleDeletedAt handles the special deleted_at column case.
func handleDeletedAt(ctx *columnContext) (MapResult, bool) {
	if ctx.columnName != "deleted_at" || !strings.Contains(ctx.baseType, "TIMESTAMP") {
		return MapResult{}, false
	}

	return MapResult{
		GoType:   "gorm.DeletedAt",
		Imports:  []string{"gorm.io/gorm"},
		GormTags: []string{"index"},
	}, true
}

// determineGoType determines the Go type for a column.
func determineGoType(ctx *columnContext) string {
	goType := ctx.typeInfo.Type

	// Foreign key columns should use uint to match entity IDs
	if ctx.isForeignKey && (goType == "int" || goType == "int64" || goType == "int16") {
		goType = "uint"
	}

	// Handle nullable types (make them pointers)
	if shouldBePointer(ctx, goType) {
		goType = "*" + goType
	}

	return goType
}

// shouldBePointer determines if a type should be a pointer for nullable columns.
func shouldBePointer(ctx *columnContext, goType string) bool {
	if !ctx.isNullable || !ctx.typeInfo.IsNullable || ctx.isPrimaryKey {
		return false
	}

	// Exception: created_at/updated_at with defaults should not be pointers
	isTimestampWithDefault := (ctx.columnName == "created_at" || ctx.columnName == "updated_at") && ctx.hasDefault
	if isTimestampWithDefault {
		return false
	}

	// Slice types don't need pointers
	return goType != "[]byte" && goType != "JSONMap"
}

// buildGormTags builds the GORM tags for a column.
func buildGormTags(ctx *columnContext) []string {
	tags := make([]string, 0)

	if ctx.isPrimaryKey {
		tags = append(tags, "primaryKey")
	}

	if ctx.isUnique && !ctx.isPrimaryKey {
		tags = append(tags, "uniqueIndex")
	}

	if !ctx.isNullable && !ctx.isPrimaryKey {
		tags = append(tags, "not null")
	}

	// Add size for varchar types
	if ctx.typeInfo.NeedsSize && ctx.size > 0 {
		tags = append(tags, fmt.Sprintf("size:%d", ctx.size))
	}

	// Add type-specific GORM tags
	if ctx.typeInfo.GormTag != "" && ctx.typeInfo.GormTag != "primaryKey" {
		tags = append(tags, ctx.typeInfo.GormTag)
	}

	// Add default value
	if ctx.hasDefault && ctx.defaultValue != "" && !isTimestampDefault(ctx.defaultValue) {
		tags = append(tags, fmt.Sprintf("default:%s", ctx.defaultValue))
	}

	// Add autoCreateTime/autoUpdateTime for timestamp columns
	tags = appendTimestampTags(tags, ctx)

	return tags
}

// appendTimestampTags appends timestamp-specific GORM tags.
func appendTimestampTags(tags []string, ctx *columnContext) []string {
	if !strings.Contains(ctx.baseType, "TIMESTAMP") {
		return tags
	}

	switch ctx.columnName {
	case "created_at":
		tags = append(tags, "autoCreateTime")
	case "updated_at":
		tags = append(tags, "autoUpdateTime")
	}

	return tags
}

// normalizeType normalizes an SQL type to a standard form.
func normalizeType(sqlType string) string {
	// Uppercase and trim
	normalized := strings.ToUpper(strings.TrimSpace(sqlType))

	// Remove size/precision from type for lookup
	// e.g., "VARCHAR(255)" -> "VARCHAR"
	if idx := strings.Index(normalized, "("); idx != -1 {
		normalized = strings.TrimSpace(normalized[:idx])
	}

	return normalized
}

// isTimestampDefault checks if a default value is a timestamp default.
func isTimestampDefault(defaultValue string) bool {
	upper := strings.ToUpper(defaultValue)
	return strings.Contains(upper, "CURRENT_TIMESTAMP") ||
		strings.Contains(upper, "NOW()") ||
		strings.Contains(upper, "CURRENT_DATE") ||
		strings.Contains(upper, "CURRENT_TIME")
}

// ToGoFieldName converts a column name to a Go field name (PascalCase).
func ToGoFieldName(columnName string) string {
	// Handle common abbreviations
	acronyms := map[string]string{
		"id":   "ID",
		"url":  "URL",
		"uri":  "URI",
		"api":  "API",
		"http": "HTTP",
		"json": "JSON",
		"xml":  "XML",
		"sql":  "SQL",
		"ip":   "IP",
		"uuid": "UUID",
	}

	parts := strings.Split(columnName, "_")
	result := make([]string, len(parts))

	for i, part := range parts {
		if acronym, ok := acronyms[strings.ToLower(part)]; ok {
			result[i] = acronym
		} else if part != "" {
			result[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}

	return strings.Join(result, "")
}

// ToJSONTag converts a column name to a JSON tag.
func ToJSONTag(columnName string, isNullable bool) string {
	tag := columnName
	if isNullable {
		tag += ",omitempty"
	}
	return tag
}
