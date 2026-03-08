package wire

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// wireRouter adds handler registration to internal/handlers/http/router.go.
func wireRouter(cfg Config, f Feature) error {
	relPath := "internal/handlers/http/router.go"
	fullPath := filepath.Join(cfg.OutputDir, relPath)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", relPath, err)
	}

	contentStr := string(content)

	// Check if already wired by looking for handler import
	handlerImport := fmt.Sprintf(`/v1/%s"`, f.PackageName)
	if strings.Contains(contentStr, handlerImport) {
		fmt.Printf("  [skip] router: %s handler already registered\n", f.EntityName)
		return nil
	}

	// 1. Add handler import
	importAlias := f.VarName + "handler"
	importLine := fmt.Sprintf(`%s "%s/internal/handlers/http/v1/%s"`, importAlias, cfg.ModuleName, f.PackageName)
	contentStr = ensureImport(contentStr, importLine)

	// 2. Add parameter to SetupRoutes function signature
	newParam := fmt.Sprintf(", %sUC usecase.%s", f.VarName, f.EntityName)
	contentStr = addRouterParam(contentStr, "func SetupRoutes(", newParam)

	// 3. Add parameter to setupAPIRoutes call in SetupRoutes body
	contentStr = addSetupAPIRoutesCallArg(contentStr, f.VarName+"UC")

	// 4. Add parameter to setupAPIRoutes function signature
	contentStr = addRouterParam(contentStr, "func setupAPIRoutes(", newParam)

	// 5. Add handler creation at end of setupAPIRoutes body
	handlerCode := fmt.Sprintf("\n\t%sHandler := %s.New(%sUC, l)\n\t%sHandler.RegisterRoutes(apiV1Group)\n",
		f.VarName, importAlias, f.VarName, f.VarName)
	contentStr = appendToSetupAPIRoutes(contentStr, handlerCode)

	if cfg.DryRun {
		fmt.Printf("  [dry-run] would update %s: add %s handler\n", relPath, f.EntityName)
		return nil
	}

	if err := os.WriteFile(fullPath, []byte(contentStr), 0o600); err != nil { //nolint:gosec // trusted local path from codegen
		return fmt.Errorf("writing %s: %w", relPath, err)
	}

	fmt.Printf("  [updated] %s: added %s handler registration\n", relPath, f.EntityName)
	return nil
}

// addRouterParam adds a parameter before the last parameter group in a function signature.
// It inserts before ", jwtService jwt.Service" or before ", l logger.Interface)" in the
// setupAPIRoutes signature.
func addRouterParam(content, funcPrefix, newParam string) string {
	funcIdx := strings.Index(content, funcPrefix)
	if funcIdx == -1 {
		return content
	}

	// Find the end of the function signature (the closing paren followed by opening brace)
	sigStart := funcIdx + len(funcPrefix)
	// Find the closing ) { pattern for this function
	remaining := content[sigStart:]
	parenDepth := 1 // We're after the opening (
	endIdx := -1
	for i := 0; i < len(remaining); i++ {
		switch remaining[i] {
		case '(':
			parenDepth++
		case ')':
			parenDepth--
			if parenDepth == 0 {
				endIdx = i
			}
		}
		if endIdx != -1 {
			break
		}
	}

	if endIdx == -1 {
		return content
	}

	// Check if this parameter is already present
	sigContent := remaining[:endIdx]
	if strings.Contains(sigContent, newParam) {
		return content
	}

	// Insert before ", jwtService" if it exists, otherwise before ", l logger.Interface"
	// In SetupRoutes, insert before "jwtService"; in setupAPIRoutes, insert before "jwtService"
	// insertPoint is absolute position in content
	insertPoint := sigStart + endIdx
	insertMarkers := []string{", jwtService", ", healthChecker"}

	for _, marker := range insertMarkers {
		markerIdx := strings.Index(remaining[:endIdx], marker)
		if markerIdx != -1 {
			insertPoint = sigStart + markerIdx
			break
		}
	}

	return content[:insertPoint] + newParam + content[insertPoint:]
}

// addSetupAPIRoutesCallArg adds an argument to the setupAPIRoutes() call inside SetupRoutes.
func addSetupAPIRoutesCallArg(content, newArg string) string {
	// Find "setupAPIRoutes(" call
	callIdx := strings.Index(content, "setupAPIRoutes(")
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

	// Insert before ", jwtService" in the call arguments
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

// appendToSetupAPIRoutes appends handler registration code to the end of the
// setupAPIRoutes function body (before its closing brace).
func appendToSetupAPIRoutes(content, code string) string {
	// Find the setupAPIRoutes function
	funcIdx := strings.Index(content, "func setupAPIRoutes(")
	if funcIdx == -1 {
		return content
	}

	// Find the opening brace
	remaining := content[funcIdx:]
	braceIdx := strings.Index(remaining, "{")
	if braceIdx == -1 {
		return content
	}

	// Find the matching closing brace
	braceStart := funcIdx + braceIdx
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

	// Insert before the closing brace
	return content[:closingBrace] + code + content[closingBrace:]
}
