package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go-boilerplate/pkg/codegen/typemap"
)

// PostgresParser parses PostgreSQL CREATE TABLE statements.
type PostgresParser struct{}

// NewPostgresParser creates a new PostgreSQL parser.
func NewPostgresParser() *PostgresParser {
	return &PostgresParser{}
}

// Parse parses SQL content and extracts table definitions.
func (p *PostgresParser) Parse(sql string) (*ParseResult, error) {
	// Find CREATE TABLE statement
	table, err := p.parseCreateTable(sql)
	if err != nil {
		return nil, err
	}

	// Parse indexes
	table.Indexes = p.parseIndexes(sql, table.Name)

	// Parse table comment
	table.Comment = p.parseTableComment(sql, table.Name)

	// Generate Go fields and relations
	result := p.generateGoTypes(table)

	return result, nil
}

// parseCreateTable extracts table definition from CREATE TABLE statement.
func (p *PostgresParser) parseCreateTable(sql string) (*Table, error) {
	// Match CREATE TABLE statement
	// Pattern: CREATE TABLE [IF NOT EXISTS] table_name (columns...)
	tablePattern := regexp.MustCompile(`(?is)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)\s*\((.*?)\)\s*;`)
	matches := tablePattern.FindStringSubmatch(sql)
	if matches == nil {
		return nil, fmt.Errorf("no CREATE TABLE statement found")
	}

	table := &Table{
		Name:    matches[1],
		Columns: make([]Column, 0),
	}

	// Parse column definitions
	columnsSQL := matches[2]
	columns := p.splitColumns(columnsSQL)

	for _, colSQL := range columns {
		colSQL = strings.TrimSpace(colSQL)
		if colSQL == "" {
			continue
		}

		// Skip table-level constraints
		upperCol := strings.ToUpper(colSQL)
		if strings.HasPrefix(upperCol, "PRIMARY KEY") ||
			strings.HasPrefix(upperCol, "FOREIGN KEY") ||
			strings.HasPrefix(upperCol, "UNIQUE") ||
			strings.HasPrefix(upperCol, "CHECK") ||
			strings.HasPrefix(upperCol, "CONSTRAINT") {
			// Parse constraint and add to table
			constraint := p.parseConstraint(colSQL)
			if constraint != nil {
				table.Constraints = append(table.Constraints, *constraint)
			}
			continue
		}

		// Parse column definition
		col, err := p.parseColumn(colSQL)
		if err != nil {
			// Skip unparseable columns
			continue
		}

		table.Columns = append(table.Columns, *col)
	}

	return table, nil
}

// splitColumns splits the column definitions, handling nested parentheses.
func (p *PostgresParser) splitColumns(columnsSQL string) []string {
	var columns []string
	var current strings.Builder
	depth := 0

	for _, char := range columnsSQL {
		switch char {
		case '(':
			depth++
			current.WriteRune(char)
		case ')':
			depth--
			current.WriteRune(char)
		case ',':
			if depth == 0 {
				columns = append(columns, current.String())
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	// Add the last column
	if current.Len() > 0 {
		columns = append(columns, current.String())
	}

	return columns
}

// parseColumn parses a single column definition.
func (p *PostgresParser) parseColumn(colSQL string) (*Column, error) {
	// Normalize whitespace
	colSQL = strings.Join(strings.Fields(colSQL), " ")

	// Pattern to match column: name TYPE[(size)] [constraints...]
	parts := strings.Fields(colSQL)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid column definition: %s", colSQL)
	}

	col := &Column{
		Name:       parts[0],
		IsNullable: true, // Default to nullable
	}

	// Parse type and size
	typeStr, parts := p.parseColumnType(parts)
	col.SQLType = typeStr
	col.BaseType, col.Size = p.parseTypeSize(typeStr)

	// Parse constraints from remaining parts
	constraintStr := strings.Join(parts[2:], " ")
	p.parseColumnConstraints(col, constraintStr)

	return col, nil
}

// parseColumnType extracts the column type from parts, handling multi-word types.
func (p *PostgresParser) parseColumnType(parts []string) (typeStr string, updatedParts []string) {
	typeStr = parts[1]
	updatedParts = parts

	// Handle CHARACTER VARYING
	if len(updatedParts) > 2 && updatedParts[2] == "VARYING" {
		typeStr = updatedParts[1] + " " + updatedParts[2]
		updatedParts = append(updatedParts[:2], updatedParts[3:]...)
	}

	// Handle TIMESTAMP WITH TIME ZONE
	if len(updatedParts) > 4 && strings.EqualFold(updatedParts[1], "TIMESTAMP") &&
		strings.EqualFold(updatedParts[2], "WITH") && strings.EqualFold(updatedParts[3], "TIME") &&
		strings.EqualFold(updatedParts[4], "ZONE") {
		typeStr = "TIMESTAMP WITH TIME ZONE"
		updatedParts = append(updatedParts[:2], updatedParts[5:]...)
	}

	return typeStr, updatedParts
}

// parseColumnConstraints parses and applies column constraints.
func (p *PostgresParser) parseColumnConstraints(col *Column, constraintStr string) {
	remainder := strings.ToUpper(constraintStr)

	// Check PRIMARY KEY
	if strings.Contains(col.BaseType, "SERIAL") || strings.Contains(remainder, "PRIMARY KEY") {
		col.IsPrimaryKey = true
		col.IsNullable = false
	}

	// Check NOT NULL
	if strings.Contains(remainder, "NOT NULL") {
		col.IsNullable = false
	}

	// Check UNIQUE
	if strings.Contains(remainder, "UNIQUE") {
		col.IsUnique = true
	}

	// Check DEFAULT
	p.parseDefaultValue(col, constraintStr)

	// Check REFERENCES (foreign key)
	p.parseForeignKeyRef(col, constraintStr)
}

// parseDefaultValue extracts the default value from constraint string.
func (p *PostgresParser) parseDefaultValue(col *Column, constraintStr string) {
	defaultPattern := regexp.MustCompile(`(?i)DEFAULT\s+(\S+)`)
	if defaultMatch := defaultPattern.FindStringSubmatch(constraintStr); defaultMatch != nil {
		col.HasDefault = true
		col.DefaultValue = defaultMatch[1]
	}
}

// parseForeignKeyRef extracts foreign key reference from constraint string.
func (p *PostgresParser) parseForeignKeyRef(col *Column, constraintStr string) {
	refPattern := regexp.MustCompile(`(?i)REFERENCES\s+(\w+)\s*\((\w+)\)`)
	refMatch := refPattern.FindStringSubmatch(constraintStr)
	if refMatch == nil {
		return
	}

	col.ForeignKey = &ForeignKey{
		RefTable:  refMatch[1],
		RefColumn: refMatch[2],
	}

	// Parse ON DELETE
	onDeletePattern := regexp.MustCompile(`(?i)ON\s+DELETE\s+(CASCADE|RESTRICT|NO\s+ACTION|SET\s+NULL|SET\s+DEFAULT)`)
	if onDeleteMatch := onDeletePattern.FindStringSubmatch(constraintStr); onDeleteMatch != nil {
		col.ForeignKey.OnDelete = strings.ToUpper(strings.ReplaceAll(onDeleteMatch[1], "  ", " "))
	}

	// Parse ON UPDATE
	onUpdatePattern := regexp.MustCompile(`(?i)ON\s+UPDATE\s+(CASCADE|RESTRICT|NO\s+ACTION|SET\s+NULL|SET\s+DEFAULT)`)
	if onUpdateMatch := onUpdatePattern.FindStringSubmatch(constraintStr); onUpdateMatch != nil {
		col.ForeignKey.OnUpdate = strings.ToUpper(strings.ReplaceAll(onUpdateMatch[1], "  ", " "))
	}
}

// parseTypeSize extracts the base type and size from a type string.
func (p *PostgresParser) parseTypeSize(typeStr string) (baseType string, size int) {
	// Pattern: TYPE(size) or TYPE(precision,scale)
	sizePattern := regexp.MustCompile(`(\w+(?:\s+\w+)*)\s*(?:\((\d+)(?:,\s*(\d+))?\))?`)
	matches := sizePattern.FindStringSubmatch(typeStr)

	if matches == nil {
		return strings.ToUpper(typeStr), 0
	}

	baseType = strings.ToUpper(matches[1])
	if matches[2] != "" {
		size, _ = strconv.Atoi(matches[2])
	}

	return baseType, size
}

// parseConstraint parses a table-level constraint.
func (p *PostgresParser) parseConstraint(constraintSQL string) *Constraint {
	upper := strings.ToUpper(constraintSQL)

	constraint := &Constraint{}

	// Check constraint type
	switch {
	case strings.Contains(upper, "PRIMARY KEY"):
		constraint.Type = "PRIMARY KEY"
	case strings.Contains(upper, "FOREIGN KEY"):
		constraint.Type = "FOREIGN KEY"
	case strings.Contains(upper, "UNIQUE"):
		constraint.Type = "UNIQUE"
	case strings.Contains(upper, "CHECK"):
		constraint.Type = "CHECK"
	}

	// Extract columns
	colPattern := regexp.MustCompile(`\(([^)]+)\)`)
	if matches := colPattern.FindStringSubmatch(constraintSQL); matches != nil {
		cols := strings.Split(matches[1], ",")
		for _, col := range cols {
			constraint.Columns = append(constraint.Columns, strings.TrimSpace(col))
		}
	}

	return constraint
}

// parseIndexes extracts index definitions from SQL.
func (p *PostgresParser) parseIndexes(sql, tableName string) []Index {
	var indexes []Index

	// Pattern: CREATE [UNIQUE] INDEX name ON table(columns)
	indexPattern := regexp.MustCompile(`(?im)CREATE\s+(UNIQUE\s+)?INDEX\s+(\w+)\s+ON\s+` + tableName + `\s*\(([^)]+)\)`)
	matches := indexPattern.FindAllStringSubmatch(sql, -1)

	for _, match := range matches {
		index := Index{
			Name:   match[2],
			Unique: strings.TrimSpace(match[1]) != "",
		}

		// Parse columns
		cols := strings.Split(match[3], ",")
		for _, col := range cols {
			index.Columns = append(index.Columns, strings.TrimSpace(col))
		}

		indexes = append(indexes, index)
	}

	return indexes
}

// parseTableComment extracts the table comment from SQL.
func (p *PostgresParser) parseTableComment(sql, tableName string) string {
	// Pattern: COMMENT ON TABLE table_name IS 'comment'
	commentPattern := regexp.MustCompile(`(?im)COMMENT\s+ON\s+TABLE\s+` + tableName + `\s+IS\s+'([^']*)'`)
	if match := commentPattern.FindStringSubmatch(sql); match != nil {
		return match[1]
	}
	return ""
}

// generateGoTypes generates Go field definitions from parsed table.
func (p *PostgresParser) generateGoTypes(table *Table) *ParseResult {
	result := &ParseResult{
		Table:     *table,
		Fields:    make([]GoField, 0),
		Relations: make([]GoRelation, 0),
		Imports:   make([]string, 0),
	}

	importSet := make(map[string]bool)

	for _, col := range table.Columns {
		// Map column to Go type
		mapResult := typemap.MapColumnWithFK(
			col.SQLType,
			col.Size,
			col.IsNullable,
			col.IsPrimaryKey,
			col.IsUnique,
			col.HasDefault,
			col.DefaultValue,
			col.Name,
			col.ForeignKey != nil,
		)

		// Build GORM tag string
		gormTag := ""
		if len(mapResult.GormTags) > 0 {
			gormTag = strings.Join(mapResult.GormTags, ";")
		}

		// Build JSON tag
		jsonTag := typemap.ToJSONTag(col.Name, col.IsNullable && !col.IsPrimaryKey)

		// Special handling for password fields - hide from JSON
		if col.Name == "password" || strings.HasSuffix(col.Name, "_hash") {
			jsonTag = "-"
		}

		// Special handling for deleted_at - hide from JSON
		if col.Name == "deleted_at" {
			jsonTag = "-"
		}

		field := GoField{
			Name:       typemap.ToGoFieldName(col.Name),
			ColumnName: col.Name,
			Type:       mapResult.GoType,
			JSONTag:    jsonTag,
			GormTags:   gormTag,
			Comment:    col.Comment,
		}

		result.Fields = append(result.Fields, field)

		// Collect imports
		for _, imp := range mapResult.Imports {
			importSet[imp] = true
		}

		// Generate relation field for foreign keys
		if col.ForeignKey != nil {
			fkFieldName := typemap.ToGoFieldName(col.Name)

			// Derive relation name from column name (strip _id suffix) for better naming
			// e.g., avatar_media_id -> Avatar, user_id -> User
			relName := toPascalCase(strings.TrimSuffix(col.Name, "_id"))

			// The type comes from the referenced table
			refEntityName := toPascalCase(singularize(col.ForeignKey.RefTable))

			// Build relation GORM tag
			relGormTag := fmt.Sprintf("foreignKey:%s", fkFieldName)

			relation := GoRelation{
				Name:       relName,
				Type:       "*" + refEntityName,
				ForeignKey: fkFieldName,
				JSONTag:    toCamelCase(relName) + ",omitempty",
				GormTags:   relGormTag,
			}

			result.Relations = append(result.Relations, relation)
		}
	}

	// Convert import set to slice
	for imp := range importSet {
		result.Imports = append(result.Imports, imp)
	}

	return result
}

// ParseFile parses a migration file and returns the parse result.
func ParseFile(content string) (*ParseResult, error) {
	parser := NewPostgresParser()
	return parser.Parse(content)
}
