package wire

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// wireApp adds feature wiring to internal/app/app.go:
// - Import for the usecase package.
// - Field in repositories struct.
// - Field in usecases struct.
// - Initialization in initRepositories().
// - Initialization in initUseCases().
// - Parameter in SetupRoutes call.
// - Entity in AutoMigrate call.
func wireApp(cfg Config, f Feature) error {
	relPath := "internal/app/app.go"
	fullPath := filepath.Join(cfg.OutputDir, relPath)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", relPath, err)
	}

	contentStr := string(content)

	// Check if already wired by looking for the usecase import
	ucImport := fmt.Sprintf(`"%s/internal/usecase/%s"`, cfg.ModuleName, f.PackageName)
	if strings.Contains(contentStr, ucImport) {
		fmt.Printf("  [skip] app.go: %s already wired\n", f.EntityName)
		return nil
	}

	// 1. Add usecase package import
	contentStr = ensureImport(contentStr, ucImport)

	// 2. Add field to repositories struct
	repoField := fmt.Sprintf("\t%s repo.%sRepo\n", f.VarName, f.EntityName)
	contentStr = addStructField(contentStr, "type repositories struct", repoField)

	// 3. Add field to usecases struct
	ucField := fmt.Sprintf("\t%s usecase.%s\n", f.VarName, f.EntityName)
	contentStr = addStructField(contentStr, "type usecases struct", ucField)

	// 4. Add line to initRepositories() return
	repoInit := fmt.Sprintf("\t\t%s: persistent.New%sRepo(db),\n", f.VarName, f.EntityName)
	contentStr = addToReturnStruct(contentStr, "func initRepositories(", repoInit)

	// 5. Add to initUseCases()
	contentStr = addUseCaseInit(contentStr, f)

	// 6. Add to SetupRoutes call in initHTTPServer
	contentStr = addSetupRoutesArg(contentStr, "uc."+f.VarName)

	// 7. Add entity to AutoMigrate call
	contentStr = addAutoMigrateEntity(contentStr, f.EntityName)

	if cfg.DryRun {
		fmt.Printf("  [dry-run] would update %s: wire %s\n", relPath, f.EntityName)
		return nil
	}

	if err := os.WriteFile(fullPath, []byte(contentStr), 0o600); err != nil { //nolint:gosec // dev tool operating on trusted local paths
		return fmt.Errorf("writing %s: %w", relPath, err)
	}

	fmt.Printf("  [updated] %s: wired %s\n", relPath, f.EntityName)
	return nil
}

// addStructField adds a field before the closing brace of a struct definition.
func addStructField(content, structMarker, field string) string {
	idx := strings.Index(content, structMarker)
	if idx == -1 {
		return content
	}

	// Check if field already exists
	if strings.Contains(content, strings.TrimSpace(field)) {
		return content
	}

	// Find the opening brace of the struct
	remaining := content[idx:]
	braceIdx := strings.Index(remaining, "{")
	if braceIdx == -1 {
		return content
	}

	// Find the closing brace
	braceStart := idx + braceIdx
	depth := 0
	closingBrace := -1
	for i := braceStart; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				closingBrace = i
				break
			}
		}
		if closingBrace != -1 {
			break
		}
	}

	if closingBrace == -1 {
		return content
	}

	return content[:closingBrace] + field + content[closingBrace:]
}

// addToReturnStruct adds a field initialization to a return &struct{...} block in a function.
func addToReturnStruct(content, funcMarker, initLine string) string {
	funcIdx := strings.Index(content, funcMarker)
	if funcIdx == -1 {
		return content
	}

	// Check if already present
	if strings.Contains(content[funcIdx:], strings.TrimSpace(initLine)) {
		return content
	}

	// Find "return &repositories{" or "return &usecases{" after the function
	remaining := content[funcIdx:]
	returnIdx := strings.Index(remaining, "return &")
	if returnIdx == -1 {
		return content
	}

	// Find the opening brace of the return struct literal
	afterReturn := remaining[returnIdx:]
	braceIdx := strings.Index(afterReturn, "{")
	if braceIdx == -1 {
		return content
	}

	braceStart := funcIdx + returnIdx + braceIdx
	depth := 0
	closingBrace := -1
	for i := braceStart; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				closingBrace = i
				break
			}
		}
		if closingBrace != -1 {
			break
		}
	}

	if closingBrace == -1 {
		return content
	}

	return content[:closingBrace] + initLine + content[closingBrace:]
}

// addUseCaseInit adds usecase initialization to initUseCases() function.
// It adds the UC variable creation and adds the field to the return struct.
func addUseCaseInit(content string, f Feature) string {
	funcMarker := "func initUseCases("
	funcIdx := strings.Index(content, funcMarker)
	if funcIdx == -1 {
		return content
	}

	// Check if already present
	ucVarLine := fmt.Sprintf("%sUC := %s.New(", f.VarName, f.PackageName)
	if strings.Contains(content[funcIdx:], ucVarLine) {
		return content
	}

	// Find "return &usecases{" in the function
	remaining := content[funcIdx:]
	returnIdx := strings.Index(remaining, "return &usecases{")
	if returnIdx == -1 {
		return content
	}

	// Insert UC creation before the return statement
	insertPos := funcIdx + returnIdx
	ucCreation := fmt.Sprintf("\t%sUC := %s.New(repos.%s)\n\n\t", f.VarName, f.PackageName, f.VarName)

	content = content[:insertPos] + ucCreation + content[insertPos:]

	// Now add to the return struct
	ucField := fmt.Sprintf("\t\t%s: %sUC,\n", f.VarName, f.VarName)
	content = addToReturnStruct(content, funcMarker, ucField)

	return content
}

// addSetupRoutesArg adds an argument to the httphandler.SetupRoutes() call
// in initHTTPServer.
func addSetupRoutesArg(content, newArg string) string {
	// Find "httphandler.SetupRoutes("
	callMarker := "httphandler.SetupRoutes("
	callIdx := strings.Index(content, callMarker)
	if callIdx == -1 {
		return content
	}

	// Find the closing ) of this call
	remaining := content[callIdx:]
	parenDepth := 0
	endIdx := -1
	for i := 0; i < len(remaining); i++ {
		switch remaining[i] {
		case '(':
			parenDepth++
		case ')':
			parenDepth--
			if parenDepth == 0 {
				endIdx = i
				break
			}
		}
		if endIdx != -1 {
			break
		}
	}

	if endIdx == -1 {
		return content
	}

	// Check if already present
	callContent := remaining[:endIdx]
	if strings.Contains(callContent, newArg) {
		return content
	}

	// Insert before ", jwtService" in the call
	insertPoint := callIdx + endIdx
	insertMarkers := []string{", jwtService"}

	for _, marker := range insertMarkers {
		markerIdx := strings.Index(callContent, marker)
		if markerIdx != -1 {
			insertPoint = callIdx + markerIdx
			break
		}
	}

	return content[:insertPoint] + ", " + newArg + content[insertPoint:]
}

// addAutoMigrateEntity adds an entity to the db.AutoMigrate() call.
func addAutoMigrateEntity(content, entityName string) string {
	marker := "&entity." + entityName + "{}"
	if strings.Contains(content, marker) {
		return content
	}

	// Find db.AutoMigrate(
	migrateIdx := strings.Index(content, "db.AutoMigrate(")
	if migrateIdx == -1 {
		return content
	}

	// Find the closing ) of AutoMigrate call
	remaining := content[migrateIdx:]
	parenDepth := 0
	endIdx := -1
	for i := 0; i < len(remaining); i++ {
		switch remaining[i] {
		case '(':
			parenDepth++
		case ')':
			parenDepth--
			if parenDepth == 0 {
				endIdx = i
				break
			}
		}
		if endIdx != -1 {
			break
		}
	}

	if endIdx == -1 {
		return content
	}

	// Insert before the closing paren, after the last entity
	insertPos := migrateIdx + endIdx
	newEntity := fmt.Sprintf("\t\t%s,\n\t", marker)

	return content[:insertPos] + newEntity + content[insertPos:]
}
