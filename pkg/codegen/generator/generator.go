// Package generator provides code generation from parsed SQL schemas.
package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-boilerplate/pkg/codegen/parser"
)

// Config holds configuration for code generation.
type Config struct {
	ModuleName string   // Go module name (e.g., "go-boilerplate")
	OutputDir  string   // Output directory (project root)
	Layers     []string // Layers to generate: entity, dto, repo, usecase, handler
	DryRun     bool     // If true, print output instead of writing files
	Force      bool     // If true, overwrite existing files
}

// Generator generates Go code from parsed SQL schemas.
type Generator struct {
	config Config
	result *parser.ParseResult
}

// New creates a new Generator.
func New(config Config, result *parser.ParseResult) *Generator {
	return &Generator{
		config: config,
		result: result,
	}
}

// Generate generates code for all configured layers.
func (g *Generator) Generate() error {
	for _, layer := range g.config.Layers {
		var err error
		switch strings.ToLower(layer) {
		case "entity":
			err = g.GenerateEntity()
		case "dto":
			err = g.GenerateDTO()
		case "repo":
			err = g.GenerateRepository()
		case "usecase":
			err = g.GenerateUseCase()
		case "handler":
			err = g.GenerateHandler()
		default:
			return fmt.Errorf("unknown layer: %s", layer)
		}
		if err != nil {
			return fmt.Errorf("generating %s: %w", layer, err)
		}
	}

	return nil
}

// writeFile writes content to a file, creating directories as needed.
func (g *Generator) writeFile(relPath, content string) error {
	fullPath := filepath.Join(g.config.OutputDir, relPath)

	if g.config.DryRun {
		fmt.Printf("\n=== %s ===\n", relPath)
		fmt.Println(content)
		return nil
	}

	// Check if file exists
	if !g.config.Force {
		if _, err := os.Stat(fullPath); err == nil {
			return fmt.Errorf("file exists: %s (use --force to overwrite)", relPath)
		}
	}

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("writing file %s: %w", fullPath, err)
	}

	fmt.Printf("Generated: %s\n", relPath)
	return nil
}

// appendToFile appends content to an existing file.
func (g *Generator) appendToFile(relPath, content, marker string) error {
	fullPath := filepath.Join(g.config.OutputDir, relPath)

	if g.config.DryRun {
		fmt.Printf("\n=== Append to %s ===\n", relPath)
		fmt.Println(content)
		return nil
	}

	// Read existing file
	existing, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", fullPath, err)
	}

	// Check if content already exists (by looking for the entity name)
	entityName := g.result.Table.EntityName()
	if strings.Contains(string(existing), entityName+"Repo") {
		fmt.Printf("Skipped: %s (interface already exists)\n", relPath)
		return nil
	}

	// Find position to insert (before closing bracket or at marker)
	existingStr := string(existing)
	insertPos := -1

	if marker != "" {
		insertPos = strings.Index(existingStr, marker)
	}

	if insertPos == -1 {
		// Insert before last closing bracket
		insertPos = strings.LastIndex(existingStr, ")")
		if insertPos == -1 {
			return fmt.Errorf("could not find insertion point in %s", relPath)
		}
	}

	// Insert content
	newContent := existingStr[:insertPos] + "\n" + content + existingStr[insertPos:]

	if err := os.WriteFile(fullPath, []byte(newContent), 0o600); err != nil {
		return fmt.Errorf("writing file %s: %w", fullPath, err)
	}

	fmt.Printf("Updated: %s\n", relPath)
	return nil
}

// entityName returns the singular PascalCase entity name.
func (g *Generator) entityName() string {
	return g.result.Table.EntityName()
}

// tableName returns the plural snake_case table name.
func (g *Generator) tableName() string {
	return g.result.Table.Name
}

// varName returns the singular camelCase variable name.
func (g *Generator) varName() string {
	return g.result.Table.VarName()
}

// packageName returns the lowercase package name.
func (g *Generator) packageName() string {
	return strings.ToLower(g.result.Table.VarName())
}
