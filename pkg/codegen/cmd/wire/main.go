// Package main provides the CLI for the auto-wiring tool.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-boilerplate/pkg/codegen/wire"
)

func main() {
	dryRun := flag.Bool("n", false, "Dry run - show what would be changed without writing")
	outputDir := flag.String("o", ".", "Output directory (project root)")
	moduleName := flag.String("mod", "", "Go module name (reads from go.mod if not specified)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: wire [options]\n\n")
		fmt.Fprintf(os.Stderr, "Auto-wire generated features into DI container, routes, and contracts.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  wire            # Wire all unwired features\n")
		fmt.Fprintf(os.Stderr, "  wire -n         # Dry run (show what would be changed)\n")
	}

	flag.Parse()

	modName := *moduleName
	if modName == "" {
		var err error
		modName, err = readModuleName(*outputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading module name: %v\n", err)
			os.Exit(1)
		}
	}

	cfg := wire.Config{
		ModuleName: modName,
		OutputDir:  *outputDir,
		DryRun:     *dryRun,
	}

	w := wire.New(cfg)
	if err := w.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// readModuleName reads the module name from go.mod.
func readModuleName(outputDir string) (string, error) {
	goModPath := filepath.Join(outputDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("could not read go.mod: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module "), nil
		}
	}

	return "", fmt.Errorf("module name not found in go.mod")
}
