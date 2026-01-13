// Package main provides the CLI for the code generator.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-boilerplate/pkg/codegen/generator"
	"go-boilerplate/pkg/codegen/parser"
)

func main() {
	// Define flags
	migration := flag.String("m", "", "Migration file path or number (e.g., 000010 or migrations/000010_create_profiles.up.sql)")
	layers := flag.String("l", "entity,dto,repo", "Comma-separated layers to generate: entity,dto,repo,usecase,handler")
	outputDir := flag.String("o", ".", "Output directory (project root)")
	moduleName := flag.String("mod", "", "Go module name (reads from go.mod if not specified)")
	dryRun := flag.Bool("n", false, "Dry run - print output without writing files")
	force := flag.Bool("f", false, "Force overwrite existing files")

	// Custom usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: codegen [options]\n\n")
		fmt.Fprintf(os.Stderr, "Generate Go code from SQL migration files.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  codegen -m 000010                          # Generate entity,dto,repo from migration 000010\n")
		fmt.Fprintf(os.Stderr, "  codegen -m 000010 -l entity                 # Generate entity only\n")
		fmt.Fprintf(os.Stderr, "  codegen -m 000010 -l entity,dto,repo,usecase,handler  # Generate all layers\n")
		fmt.Fprintf(os.Stderr, "  codegen -m migrations/000010_create_profiles.up.sql   # Use full path\n")
		fmt.Fprintf(os.Stderr, "  codegen -m 000010 -n                        # Dry run (print without writing)\n")
	}

	flag.Parse()

	// Validate migration flag
	if *migration == "" {
		fmt.Fprintln(os.Stderr, "Error: -m (migration) flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Resolve migration file path
	migrationPath, err := resolveMigrationPath(*migration, *outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Read migration file
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading migration file: %v\n", err)
		os.Exit(1)
	}

	// Parse migration
	result, err := parser.ParseFile(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing migration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Parsed table: %s (%d columns)\n", result.Table.Name, len(result.Table.Columns))

	// Resolve module name
	modName := *moduleName
	if modName == "" {
		modName, err = readModuleName(*outputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading module name: %v\n", err)
			os.Exit(1)
		}
	}

	// Parse layers
	layerList := strings.Split(*layers, ",")
	for i, l := range layerList {
		layerList[i] = strings.TrimSpace(l)
	}

	// Create generator config
	config := generator.Config{
		ModuleName: modName,
		OutputDir:  *outputDir,
		Layers:     layerList,
		DryRun:     *dryRun,
		Force:      *force,
	}

	// Generate code
	gen := generator.New(config, result)
	if err := gen.Generate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}

	if *dryRun {
		fmt.Println("\n(Dry run - no files were written)")
	} else {
		fmt.Println("\nCode generation complete!")
	}
}

// resolveMigrationPath resolves the migration file path from a number or full path.
func resolveMigrationPath(migration, outputDir string) (string, error) {
	// Check if it's already a full path
	if strings.HasSuffix(migration, ".sql") {
		fullPath := migration
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(outputDir, migration)
		}
		if _, err := os.Stat(fullPath); err != nil {
			return "", fmt.Errorf("migration file not found: %s", fullPath)
		}
		return fullPath, nil
	}

	// Treat as migration number - search for matching file
	migrationsDir := filepath.Join(outputDir, "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return "", fmt.Errorf("could not read migrations directory: %w", err)
	}

	// Look for a file starting with the migration number and ending in .up.sql
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, migration) && strings.HasSuffix(name, ".up.sql") {
			return filepath.Join(migrationsDir, name), nil
		}
	}

	return "", fmt.Errorf("no migration file found matching: %s", migration)
}

// readModuleName reads the module name from go.mod.
func readModuleName(outputDir string) (string, error) {
	goModPath := filepath.Join(outputDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("could not read go.mod: %w", err)
	}

	// Parse module line
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module "), nil
		}
	}

	return "", fmt.Errorf("module name not found in go.mod")
}
