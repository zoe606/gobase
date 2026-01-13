// Package main provides a CLI tool to rename the boilerplate project.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	oldModuleName = "go-boilerplate"
	oldAppName    = "go-boilerplate"
)

func main() {
	moduleName := flag.String("module", "", "New Go module name (e.g., github.com/mycompany/myproject)")
	appName := flag.String("app-name", "", "New application name (e.g., myapp)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: rename [options]\n\n")
		fmt.Fprintf(os.Stderr, "Rename the boilerplate project to a new module name and app name.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  rename -module=github.com/mycompany/myproject -app-name=myapp\n")
	}

	flag.Parse()

	if *moduleName == "" || *appName == "" {
		fmt.Fprintln(os.Stderr, "Error: -module and -app-name flags are required")
		flag.Usage()
		os.Exit(1)
	}

	// Validate inputs
	if err := validateModuleName(*moduleName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := validateAppName(*appName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Renaming project:\n")
	fmt.Printf("  Module: %s -> %s\n", oldModuleName, *moduleName)
	fmt.Printf("  App:    %s -> %s\n", oldAppName, *appName)
	fmt.Println()

	// Update go.mod
	if err := updateGoMod(*moduleName); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating go.mod: %v\n", err)
		os.Exit(1)
	}

	// Update Go imports in all .go files
	if err := updateGoImports(*moduleName); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating Go imports: %v\n", err)
		os.Exit(1)
	}

	// Update config files
	if err := updateConfigFiles(*appName); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating config files: %v\n", err)
		os.Exit(1)
	}

	// Update environment files
	if err := updateEnvFiles(*appName); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating env files: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nRename complete!")
	fmt.Println("Run 'go mod tidy' and 'make build' to verify the changes.")
}

// validateModuleName validates the Go module name.
func validateModuleName(name string) error {
	// Basic validation - must not be empty and should look like a module path
	if name == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	// Check for invalid characters
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
	if !validPattern.MatchString(name) {
		return fmt.Errorf("module name contains invalid characters")
	}

	return nil
}

// validateAppName validates the application name.
func validateAppName(name string) error {
	if name == "" {
		return fmt.Errorf("app name cannot be empty")
	}

	// Check for invalid characters (allow alphanumeric, hyphens, underscores)
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(name) {
		return fmt.Errorf("app name contains invalid characters (only alphanumeric, hyphens, underscores allowed)")
	}

	return nil
}

// updateGoMod updates the go.mod file with the new module name.
func updateGoMod(newModule string) error {
	fmt.Println("Updating go.mod...")

	// Use go mod edit to update the module name
	cmd := exec.CommandContext(context.Background(), "go", "mod", "edit", "-module", newModule)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go mod edit failed: %w\n%s", err, output)
	}

	fmt.Println("  Updated go.mod")
	return nil
}

// updateGoImports updates all Go import statements in .go files.
func updateGoImports(newModule string) error {
	fmt.Println("Updating Go imports...")

	count := 0
	oldImport := `"` + oldModuleName + `/`
	newImport := `"` + newModule + `/`

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() {
			// Skip vendor, .git, and other non-source directories
			if info.Name() == "vendor" || info.Name() == ".git" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		// Check if file contains old import
		if !strings.Contains(string(content), oldImport) {
			return nil
		}

		// Replace imports
		newContent := strings.ReplaceAll(string(content), oldImport, newImport)

		// Write file
		if err := os.WriteFile(path, []byte(newContent), info.Mode()); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}

		count++
		return nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("  Updated %d Go files\n", count)
	return nil
}

// updateConfigFiles updates the app name in config files.
func updateConfigFiles(newAppName string) error {
	fmt.Println("Updating config files...")

	configFiles := []string{
		"config/config.yaml",
		"config/config.example.yaml",
		"config/config.go",
	}

	for _, file := range configFiles {
		if err := updateConfigFile(file, newAppName); err != nil {
			// File might not exist, that's okay
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("updating %s: %w", file, err)
		}
		fmt.Printf("  Updated %s\n", file)
	}

	return nil
}

// updateConfigFile updates a single config file.
func updateConfigFile(path, newAppName string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	newContent := string(content)

	// Update YAML config files (name: go-boilerplate)
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		newContent = regexp.MustCompile(`name:\s*`+oldAppName).ReplaceAllString(newContent, "name: "+newAppName)
	}

	// Update config.go (viper.SetDefault("app.name", "go-boilerplate"))
	if strings.HasSuffix(path, ".go") {
		newContent = strings.ReplaceAll(newContent,
			`viper.SetDefault("app.name", "`+oldAppName+`")`,
			`viper.SetDefault("app.name", "`+newAppName+`")`)
	}

	if newContent == string(content) {
		return nil // No changes needed
	}

	return os.WriteFile(path, []byte(newContent), 0o600)
}

// updateEnvFiles updates the app name in .env files.
func updateEnvFiles(newAppName string) error {
	fmt.Println("Updating environment files...")

	envFiles := []string{
		".env",
		".env.example",
	}

	for _, file := range envFiles {
		if err := updateEnvFile(file, newAppName); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("updating %s: %w", file, err)
		}
		fmt.Printf("  Updated %s\n", file)
	}

	return nil
}

// updateEnvFile updates a single .env file.
func updateEnvFile(path, newAppName string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	modified := false

	for scanner.Scan() {
		line := scanner.Text()

		// Update APP_NAME=go-boilerplate
		if strings.HasPrefix(line, "APP_NAME=") {
			line = "APP_NAME=" + newAppName
			modified = true
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if !modified {
		return nil
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600)
}
