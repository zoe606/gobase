package wire

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// wireRepoContract adds a repository interface to internal/repo/contracts.go.
func wireRepoContract(cfg Config, f Feature) error {
	relPath := "internal/repo/contracts.go"
	fullPath := filepath.Join(cfg.OutputDir, relPath)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", relPath, err)
	}

	contentStr := string(content)

	// Check if already wired
	if strings.Contains(contentStr, f.EntityName+"Repo interface") {
		fmt.Printf("  [skip] repo contract: %sRepo already exists\n", f.EntityName)
		return nil
	}

	// Build the interface block
	iface := fmt.Sprintf(`
	// %sRepo defines the %s repository interface.
	%sRepo interface {
		Create(ctx context.Context, %s *entity.%s) error
		GetByID(ctx context.Context, id uint) (*entity.%s, error)
		List(ctx context.Context, req %sdto.ListRequest) ([]*entity.%s, int64, error)
		Update(ctx context.Context, %s *entity.%s) error
		Delete(ctx context.Context, id uint) error
	}
`,
		f.EntityName, f.VarName,
		f.EntityName,
		f.VarName, f.EntityName,
		f.EntityName,
		f.PackageName, f.EntityName,
		f.VarName, f.EntityName,
	)

	// Add DTO import if not present
	dtoImport := fmt.Sprintf(`"%s/internal/dto/%s"`, cfg.ModuleName, f.PackageName)
	contentStr = ensureImport(contentStr, dtoImport)

	// Find the last closing paren of the type block and insert before it
	lastParen := strings.LastIndex(contentStr, ")")
	if lastParen == -1 {
		return fmt.Errorf("could not find closing ) in %s", relPath)
	}

	newContent := contentStr[:lastParen] + iface + contentStr[lastParen:]

	if cfg.DryRun {
		fmt.Printf("  [dry-run] would update %s: add %sRepo interface\n", relPath, f.EntityName)
		return nil
	}

	if err := os.WriteFile(fullPath, []byte(newContent), 0o600); err != nil { //nolint:gosec // dev tool operating on trusted local paths
		return fmt.Errorf("writing %s: %w", relPath, err)
	}

	fmt.Printf("  [updated] %s: added %sRepo interface\n", relPath, f.EntityName)
	return nil
}

// wireUsecaseContract adds a usecase interface to internal/usecase/contracts.go.
func wireUsecaseContract(cfg Config, f Feature) error {
	relPath := "internal/usecase/contracts.go"
	fullPath := filepath.Join(cfg.OutputDir, relPath)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", relPath, err)
	}

	contentStr := string(content)

	// Check if already wired
	if strings.Contains(contentStr, f.EntityName+" interface") {
		fmt.Printf("  [skip] usecase contract: %s already exists\n", f.EntityName)
		return nil
	}

	// Build the interface block
	iface := fmt.Sprintf(`
	// %s defines the %s use case interface.
	%s interface {
		Create(ctx context.Context, req %sdto.CreateRequest) (*%sdto.Response, error)
		GetByID(ctx context.Context, id uint) (*%sdto.Response, error)
		List(ctx context.Context, req %sdto.ListRequest) (*%sdto.ListResponse, error)
		Update(ctx context.Context, id uint, req %sdto.UpdateRequest) (*%sdto.Response, error)
		Delete(ctx context.Context, id uint) error
	}
`,
		f.EntityName, f.VarName,
		f.EntityName,
		f.PackageName, f.PackageName,
		f.PackageName,
		f.PackageName, f.PackageName,
		f.PackageName, f.PackageName,
	)

	// Add DTO import if not present
	dtoImport := fmt.Sprintf(`"%s/internal/dto/%s"`, cfg.ModuleName, f.PackageName)
	contentStr = ensureImport(contentStr, dtoImport)

	// Find the last closing paren and insert before it
	lastParen := strings.LastIndex(contentStr, ")")
	if lastParen == -1 {
		return fmt.Errorf("could not find closing ) in %s", relPath)
	}

	newContent := contentStr[:lastParen] + iface + contentStr[lastParen:]

	if cfg.DryRun {
		fmt.Printf("  [dry-run] would update %s: add %s interface\n", relPath, f.EntityName)
		return nil
	}

	if err := os.WriteFile(fullPath, []byte(newContent), 0o600); err != nil { //nolint:gosec // dev tool operating on trusted local paths
		return fmt.Errorf("writing %s: %w", relPath, err)
	}

	fmt.Printf("  [updated] %s: added %s interface\n", relPath, f.EntityName)
	return nil
}

// ensureImport adds an import line to a Go file if it's not already present.
// It handles both grouped imports (parenthesized) and single imports.
func ensureImport(content, importPath string) string {
	// Already imported
	if strings.Contains(content, importPath) {
		return content
	}

	// Find the import block
	importStart := strings.Index(content, "import (")
	if importStart == -1 {
		// No grouped import found; this shouldn't happen in our target files
		return content
	}

	importEnd := strings.Index(content[importStart:], ")")
	if importEnd == -1 {
		return content
	}

	importEnd += importStart

	// Insert the new import before the closing paren of the import block
	// Add a blank line before if the last line isn't empty
	insertPos := importEnd
	newImport := "\t" + importPath + "\n"

	newContent := content[:insertPos] + newImport + content[insertPos:]

	return newContent
}
