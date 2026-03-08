// Package wire provides auto-wiring of generated features into the DI container,
// router, and contract files.
package wire

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Config holds configuration for the wiring process.
type Config struct {
	ModuleName string    // Go module name (e.g., "go-boilerplate")
	OutputDir  string    // Project root directory
	DryRun     bool      // If true, print changes without writing
	Output     io.Writer // Output writer for log messages (defaults to os.Stdout)
}

// output returns the configured writer or os.Stdout.
func (c Config) output() io.Writer {
	if c.Output != nil {
		return c.Output
	}
	return os.Stdout
}

// Feature represents a detected feature that may need wiring.
type Feature struct {
	Name        string // Directory name (e.g., "article")
	EntityName  string // PascalCase name (e.g., "Article")
	PackageName string // Package name (e.g., "article")
	VarName     string // camelCase variable name (e.g., "article")
}

// Wirer orchestrates the auto-wiring process.
type Wirer struct {
	config Config
}

// New creates a new Wirer with the given configuration.
func New(config Config) *Wirer {
	return &Wirer{config: config}
}

// Run scans for features and wires any that are missing from contracts, router, or app.go.
func (w *Wirer) Run() error {
	features, err := scanFeatures(w.config.OutputDir)
	if err != nil {
		return fmt.Errorf("scanning features: %w", err)
	}

	out := w.config.output()

	if len(features) == 0 {
		fmt.Fprintln(out, "No features found in internal/usecase/")
		return nil
	}

	fmt.Fprintf(out, "Found %d feature(s): %s\n", len(features), featureNames(features))

	unwired, err := findUnwired(w.config.OutputDir, features)
	if err != nil {
		return fmt.Errorf("checking wired status: %w", err)
	}

	if len(unwired) == 0 {
		fmt.Fprintln(out, "All features are already wired.")
		return nil
	}

	fmt.Fprintf(out, "Unwired feature(s): %s\n\n", featureNames(unwired))

	for _, f := range unwired {
		if err := w.wireFeature(f); err != nil {
			return fmt.Errorf("wiring feature %s: %w", f.Name, err)
		}
	}

	return nil
}

// wireFeature wires a single feature into all target files.
func (w *Wirer) wireFeature(f Feature) error {
	fmt.Fprintf(w.config.output(), "Wiring feature: %s\n", f.EntityName)

	// Wire repo contracts
	if err := wireRepoContract(w.config, f); err != nil {
		return fmt.Errorf("repo contracts: %w", err)
	}

	// Wire usecase contracts
	if err := wireUsecaseContract(w.config, f); err != nil {
		return fmt.Errorf("usecase contracts: %w", err)
	}

	// Wire router
	if err := wireRouter(w.config, f); err != nil {
		return fmt.Errorf("router: %w", err)
	}

	// Wire app.go (DI container)
	if err := wireApp(w.config, f); err != nil {
		return fmt.Errorf("app.go: %w", err)
	}

	fmt.Fprintln(w.config.output())
	return nil
}

// featureNames returns a comma-separated list of feature names.
func featureNames(features []Feature) string {
	names := make([]string, len(features))
	for i, f := range features {
		names[i] = f.Name
	}
	return strings.Join(names, ", ")
}

// toPascalCase converts a snake_case string to PascalCase.
func toPascalCase(s string) string {
	if s == "" {
		return s
	}

	result := make([]byte, 0, len(s))
	capitalizeNext := true

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' {
			capitalizeNext = true
			continue
		}

		if capitalizeNext && c >= 'a' && c <= 'z' {
			result = append(result, c-32)
		} else {
			result = append(result, c)
		}
		capitalizeNext = false
	}

	return string(result)
}

// toCamelCase converts a snake_case string to camelCase.
func toCamelCase(s string) string {
	pascal := toPascalCase(s)
	if pascal == "" {
		return pascal
	}

	first := pascal[0]
	if first >= 'A' && first <= 'Z' {
		return string(first+32) + pascal[1:]
	}
	return pascal
}
