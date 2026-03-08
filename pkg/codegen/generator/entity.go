package generator

import (
	"fmt"
	"sort"
	"strings"
)

// GenerateEntity generates the entity file.
func (g *Generator) GenerateEntity() error {
	content := g.buildEntityContent()
	relPath := fmt.Sprintf("internal/entity/%s.go", g.packageName())
	return g.writeFile(relPath, content)
}

// buildEntityContent builds the entity file content.
func (g *Generator) buildEntityContent() string {
	var sb strings.Builder

	entityName := g.entityName()
	tableName := g.tableName()
	comment := g.result.Table.Comment
	if comment == "" {
		comment = fmt.Sprintf("%s represents a %s record.", entityName, tableName)
	}

	// Package declaration
	sb.WriteString("package entity\n\n")

	// Imports
	imports := g.buildEntityImports()
	if len(imports) > 0 {
		sb.WriteString("import (\n")
		for _, imp := range imports {
			fmt.Fprintf(&sb, "\t%q\n", imp)
		}
		sb.WriteString(")\n\n")
	}

	// Struct comment
	fmt.Fprintf(&sb, "// %s %s\n", entityName, comment)

	// Struct definition
	fmt.Fprintf(&sb, "type %s struct {\n", entityName)

	// Fields
	for _, field := range g.result.Fields {
		line := g.buildFieldLine(field.Name, field.Type, field.JSONTag, field.GormTags)
		sb.WriteString(line)
	}

	// Relation fields
	if len(g.result.Relations) > 0 {
		for _, rel := range g.result.Relations {
			line := g.buildFieldLine(rel.Name, rel.Type, rel.JSONTag, rel.GormTags)
			sb.WriteString(line)
		}
	}

	sb.WriteString("}\n\n")

	// TableName method
	sb.WriteString("// TableName returns the table name.\n")
	fmt.Fprintf(&sb, "func (%s) TableName() string {\n", entityName)
	fmt.Fprintf(&sb, "\treturn %q\n", tableName)
	sb.WriteString("}\n")

	return sb.String()
}

// buildEntityImports builds the import list for the entity.
func (g *Generator) buildEntityImports() []string {
	importSet := make(map[string]bool)

	// Add imports from parsed result
	for _, imp := range g.result.Imports {
		importSet[imp] = true
	}

	// Check fields for additional imports
	for _, field := range g.result.Fields {
		if strings.Contains(field.Type, "gorm.DeletedAt") {
			importSet["gorm.io/gorm"] = true
		}
	}

	// Convert to sorted slice
	imports := make([]string, 0, len(importSet))
	for imp := range importSet {
		imports = append(imports, imp)
	}
	sort.Strings(imports)

	return imports
}

// buildFieldLine builds a struct field line.
func (g *Generator) buildFieldLine(name, goType, jsonTag, gormTags string) string {
	// Build the tag string
	var tags []string

	if jsonTag != "" {
		tags = append(tags, fmt.Sprintf("json:%q", jsonTag))
	}

	if gormTags != "" {
		tags = append(tags, fmt.Sprintf("gorm:%q", gormTags))
	}

	tagStr := ""
	if len(tags) > 0 {
		tagStr = " `" + strings.Join(tags, " ") + "`"
	}

	// Calculate spacing for alignment
	// Name is padded to 20 chars, Type to 20 chars
	namePadded := name
	if len(name) < 20 {
		namePadded = name + strings.Repeat(" ", 20-len(name))
	}

	typePadded := goType
	if len(goType) < 15 {
		typePadded = goType + strings.Repeat(" ", 15-len(goType))
	}

	return fmt.Sprintf("\t%s %s%s\n", namePadded, typePadded, tagStr)
}
