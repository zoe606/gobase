package generator

import (
	"fmt"
	"strings"
)

// GenerateDTO generates the DTO request and response files.
func (g *Generator) GenerateDTO() error {
	// Generate request.go
	requestContent := g.buildRequestDTOContent()
	requestPath := fmt.Sprintf("internal/dto/%s/request.go", g.packageName())
	if err := g.writeFile(requestPath, requestContent); err != nil {
		return err
	}

	// Generate response.go
	responseContent := g.buildResponseDTOContent()
	responsePath := fmt.Sprintf("internal/dto/%s/response.go", g.packageName())
	if err := g.writeFile(responsePath, responseContent); err != nil {
		return err
	}

	return nil
}

// buildRequestDTOContent builds the request DTO file content.
func (g *Generator) buildRequestDTOContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()

	// Package declaration
	sb.WriteString(fmt.Sprintf("package %sdto\n\n", pkgName))

	// Imports
	sb.WriteString(fmt.Sprintf("import (\n\t%q\n)\n\n", g.config.ModuleName+"/pkg/pagination"))

	// CreateRequest
	sb.WriteString(fmt.Sprintf("// CreateRequest represents the request to create a %s.\n", entityName))
	sb.WriteString("type CreateRequest struct {\n")
	for _, field := range g.result.Fields {
		if g.isCreateRequestField(field.ColumnName) {
			line := g.buildDTOFieldLine(field.ColumnName, g.dtoType(field.Type), true)
			sb.WriteString(line)
		}
	}
	sb.WriteString("}\n\n")

	// UpdateRequest
	sb.WriteString(fmt.Sprintf("// UpdateRequest represents the request to update a %s.\n", entityName))
	sb.WriteString("type UpdateRequest struct {\n")
	for _, field := range g.result.Fields {
		if g.isUpdateRequestField(field.ColumnName) {
			// Make all fields pointers for partial updates
			fieldType := g.dtoType(field.Type)
			if !strings.HasPrefix(fieldType, "*") {
				fieldType = "*" + fieldType
			}
			line := g.buildDTOFieldLine(field.ColumnName, fieldType, false)
			sb.WriteString(line)
		}
	}
	sb.WriteString("}\n\n")

	// ListRequest (pagination)
	sb.WriteString(fmt.Sprintf("// ListRequest represents the request to list %ss with filters.\n", strings.ToLower(entityName)))
	sb.WriteString("type ListRequest struct {\n")
	sb.WriteString("\tpagination.Params\n")
	sb.WriteString("}\n")

	return sb.String()
}

// buildResponseDTOContent builds the response DTO file content.
func (g *Generator) buildResponseDTOContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()
	varName := g.varName()

	// Package declaration
	sb.WriteString(fmt.Sprintf("package %sdto\n\n", pkgName))

	// Imports - check if we need time package
	needsTime := false
	for _, field := range g.result.Fields {
		if strings.Contains(field.Type, "time.Time") {
			needsTime = true
			break
		}
	}

	sb.WriteString("import (\n")
	if needsTime {
		sb.WriteString("\t\"time\"\n\n")
	}
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/entity"))
	sb.WriteString(")\n\n")

	// Response
	sb.WriteString(fmt.Sprintf("// Response represents a %s response.\n", entityName))
	sb.WriteString("type Response struct {\n")
	for _, field := range g.result.Fields {
		if g.isResponseField(field.ColumnName) {
			line := g.buildResponseFieldLine(field.Name, g.responseType(field.Type), field.ColumnName)
			sb.WriteString(line)
		}
	}
	sb.WriteString("}\n\n")

	// NewResponse constructor
	sb.WriteString(fmt.Sprintf("// NewResponse creates a Response from an entity.%s.\n", entityName))
	sb.WriteString(fmt.Sprintf("func NewResponse(%s *entity.%s) *Response {\n", varName, entityName))
	sb.WriteString(fmt.Sprintf("\tif %s == nil {\n", varName))
	sb.WriteString("\t\treturn nil\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn &Response{\n")
	for _, field := range g.result.Fields {
		if g.isResponseField(field.ColumnName) {
			sb.WriteString(fmt.Sprintf("\t\t%s: %s.%s,\n", field.Name, varName, field.Name))
		}
	}
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	// ListResponse
	sb.WriteString(fmt.Sprintf("// ListResponse represents a list of %s responses.\n", entityName))
	sb.WriteString("type ListResponse struct {\n")
	sb.WriteString("\tData       []*Response `json:\"data\"`\n")
	sb.WriteString("\tTotal      int64       `json:\"total\"`\n")
	sb.WriteString("\tPage       int         `json:\"page\"`\n")
	sb.WriteString("\tPageSize   int         `json:\"page_size\"`\n")
	sb.WriteString("\tTotalPages int         `json:\"total_pages\"`\n")
	sb.WriteString("}\n\n")

	// NewListResponse constructor
	sb.WriteString(fmt.Sprintf("// NewListResponse creates a ListResponse from a slice of %ss.\n", entityName))
	sb.WriteString(fmt.Sprintf("func NewListResponse(%ss []*entity.%s, total int64, page, pageSize int) *ListResponse {\n", varName, entityName))
	sb.WriteString(fmt.Sprintf("\tdata := make([]*Response, len(%ss))\n", varName))
	sb.WriteString(fmt.Sprintf("\tfor i, %s := range %ss {\n", varName, varName))
	sb.WriteString(fmt.Sprintf("\t\tdata[i] = NewResponse(%s)\n", varName))
	sb.WriteString("\t}\n\n")
	sb.WriteString("\ttotalPages := int(total) / pageSize\n")
	sb.WriteString("\tif int(total)%pageSize > 0 {\n")
	sb.WriteString("\t\ttotalPages++\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\treturn &ListResponse{\n")
	sb.WriteString("\t\tData:       data,\n")
	sb.WriteString("\t\tTotal:      total,\n")
	sb.WriteString("\t\tPage:       page,\n")
	sb.WriteString("\t\tPageSize:   pageSize,\n")
	sb.WriteString("\t\tTotalPages: totalPages,\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// isCreateRequestField returns true if the field should be in CreateRequest.
func (g *Generator) isCreateRequestField(fieldName string) bool {
	// Exclude auto-generated fields
	excluded := map[string]bool{
		"id":         true,
		"created_at": true,
		"updated_at": true,
		"deleted_at": true,
	}
	return !excluded[fieldName]
}

// isUpdateRequestField returns true if the field should be in UpdateRequest.
func (g *Generator) isUpdateRequestField(fieldName string) bool {
	// Exclude auto-generated fields and usually immutable fields
	excluded := map[string]bool{
		"id":         true,
		"created_at": true,
		"updated_at": true,
		"deleted_at": true,
	}
	return !excluded[fieldName]
}

// isResponseField returns true if the field should be in Response.
func (g *Generator) isResponseField(fieldName string) bool {
	// Exclude sensitive fields
	excluded := map[string]bool{
		"password":   true,
		"deleted_at": true,
	}
	return !excluded[fieldName]
}

// dtoType converts entity type to DTO type.
func (g *Generator) dtoType(entityType string) string {
	// Remove pointer for request fields
	t := strings.TrimPrefix(entityType, "*")

	// Handle special types
	switch t {
	case "gorm.DeletedAt":
		return "time.Time"
	case "JSONMap":
		return "map[string]interface{}"
	}

	return t
}

// responseType converts entity type to response type.
func (g *Generator) responseType(entityType string) string {
	// Handle special types
	switch entityType {
	case "gorm.DeletedAt":
		return "*time.Time"
	case "JSONMap":
		return "map[string]interface{}"
	}
	return entityType
}

// goFieldName converts a column name to Go field name.
func (g *Generator) goFieldName(columnName string) string {
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

// buildDTOFieldLine builds a DTO struct field line.
func (g *Generator) buildDTOFieldLine(columnName, goType string, required bool) string {
	goName := g.goFieldName(columnName)

	// Build JSON tag (use snake_case column name)
	jsonTag := columnName
	if !required {
		jsonTag += ",omitempty"
	}

	// Build validation tag
	validateTag := ""
	if required {
		validateTag = " validate:\"required\""
	}

	return fmt.Sprintf("\t%s %s `json:%q%s`\n", goName, goType, jsonTag, validateTag)
}

// buildResponseFieldLine builds a response struct field line.
// goFieldName is the PascalCase Go field name, columnName is the snake_case column name for JSON tag.
func (g *Generator) buildResponseFieldLine(goFieldName, goType, columnName string) string {
	// Build JSON tag from column name
	jsonTag := columnName
	if columnName != "id" {
		jsonTag += ",omitempty"
	}

	return fmt.Sprintf("\t%s %s `json:%q`\n", goFieldName, goType, jsonTag)
}
