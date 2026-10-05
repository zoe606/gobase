// Package scaffold creates projects with one selected HTTP engine.
package scaffold

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/imports"

	"go-boilerplate/pkg/project"
)

//go:embed templates application
var templates embed.FS

// Config identifies the template checkout and the new project.
type Config struct {
	Source string
	Output string
	Module string
	Name   string
	Engine project.Engine
}

// Generate writes a new project. Existing output directories are rejected.
func Generate(cfg Config) error {
	var err error
	cfg, err = validateConfig(cfg)
	if err != nil {
		return err
	}
	source, err := os.OpenRoot(cfg.Source)
	if err != nil {
		return err
	}
	defer func() { _ = source.Close() }()
	module, err := source.ReadFile("go.mod")
	if err != nil {
		return err
	}
	fields := strings.Fields(string(module))
	if len(fields) < 2 || fields[0] != "module" {
		return fmt.Errorf("source has no Go module")
	}
	output, err := createOutput(cfg.Output)
	if err != nil {
		return err
	}
	defer func() { _ = output.Close() }()
	copier := projectCopier{cfg: cfg, source: source, output: output, oldModule: fields[1]}
	if err := copier.copyProject(); err != nil {
		return err
	}
	return copier.writeSettings()
}

func validateConfig(cfg Config) (Config, error) {
	if cfg.Engine == "" {
		cfg.Engine = project.Gin
	}
	if err := cfg.Engine.Validate(); err != nil {
		return cfg, err
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]*$`).MatchString(cfg.Module) || strings.Contains(cfg.Module, "..") {
		return cfg, fmt.Errorf("invalid Go module name %q", cfg.Module)
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`).MatchString(cfg.Name) {
		return cfg, fmt.Errorf("invalid application name %q", cfg.Name)
	}
	var err error
	cfg.Source, err = resolvePath(cfg.Source)
	if err != nil {
		return cfg, err
	}
	cfg.Output, err = resolvePath(cfg.Output)
	if err != nil {
		return cfg, err
	}
	relative, err := filepath.Rel(cfg.Source, cfg.Output)
	if err != nil || relative == "." || !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return cfg, fmt.Errorf("output must be outside the template checkout")
	}
	if _, err := os.Stat(filepath.Join(cfg.Source, "pkg", "scaffold", "templates")); err != nil {
		return cfg, fmt.Errorf("create projects from the original gobase template checkout")
	}
	return cfg, nil
}

func resolvePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err == nil {
		return resolved, nil
	}
	if !os.IsNotExist(err) || filepath.Dir(abs) == abs {
		return "", err
	}
	parent, err := resolvePath(filepath.Dir(abs))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(abs)), nil
}

func createOutput(path string) (*os.Root, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, err
	}
	if err := os.Mkdir(path, 0o750); err != nil {
		return nil, fmt.Errorf("create output: %w", err)
	}
	return os.OpenRoot(path)
}

type projectCopier struct {
	cfg            Config
	source, output *os.Root
	oldModule      string
}

func (c *projectCopier) copyProject() error {
	paths := []string{"cmd", "config", "deployment", "docs", "integration-test", "internal", "migrations", "pkg", ".github", ".githooks", "Makefile", "go.mod", "go.sum", "README.md", "LICENSE", ".gitignore", ".dockerignore", ".golangci.yml", ".air.toml", ".env.example", "AGENTS.md"}
	for _, path := range paths {
		if _, err := c.source.Lstat(path); os.IsNotExist(err) {
			continue
		}
		if err := c.copyPath(path); err != nil {
			return fmt.Errorf("generate %s: %w", path, err)
		}
	}
	if c.cfg.Engine != project.Fiber {
		if err := c.applyTemplates("nethttp"); err != nil {
			return err
		}
	}
	return c.applyTemplates(string(c.cfg.Engine))
}

func (c *projectCopier) copyPath(root string) error {
	return fs.WalkDir(c.source.FS(), root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "pkg/scaffold" || path == "pkg/tools/init" || path == "pkg/tools/verifyengines" || path == "docs/superpowers" || path == "docs/plans" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("template contains symlink %s", path)
		}
		if entry.IsDir() {
			return c.output.MkdirAll(path, 0o750)
		}
		if skipPrivateFile(path) {
			return nil
		}
		return c.copyFile(path, entry)
	})
}

func skipPrivateFile(path string) bool {
	name := filepath.Base(path)
	return path == ".github/workflows/engines.yml" || path == "docs/http-engines.md" || path == "docs/http-engines-verification.md" || path == "docs/roadmap.md" || path == "docs/release-readiness.md" || path == "config/config.yaml" || strings.HasSuffix(name, ".pem") || strings.HasSuffix(name, ".key") || (strings.HasPrefix(name, ".env") && name != ".env.example")
}

func (c *projectCopier) copyFile(path string, entry fs.DirEntry) error {
	data, err := c.source.ReadFile(path)
	if err != nil {
		return err
	}
	if engineSourceFile(path) || (c.cfg.Engine != project.Fiber && strings.HasSuffix(path, "_test.go") && importsFiber(data)) {
		return nil
	}
	data, err = c.renderFile(path, data)
	if err != nil {
		return err
	}
	info, err := entry.Info()
	if err != nil {
		return err
	}
	mode := fs.FileMode(0o600)
	if info.Mode()&0o111 != 0 {
		mode = 0o750
	}
	return c.output.WriteFile(path, data, mode)
}

func (c *projectCopier) renderFile(path string, data []byte) ([]byte, error) {
	if path == "Makefile" {
		start := bytes.Index(data, []byte("ENGINE ?= gin"))
		end := bytes.Index(data, []byte(".PHONY: rename"))
		if start >= 0 && end > start {
			data = append(data[:start], data[end:]...)
		}
	}
	var docTemplate string
	switch path {
	case "README.md":
		docTemplate = "application/README.md.tmpl"
	case ".github/CONTRIBUTING.md":
		docTemplate = "application/CONTRIBUTING.md.tmpl"
	}
	if docTemplate != "" {
		var err error
		data, err = templates.ReadFile(docTemplate)
		if err != nil {
			return nil, err
		}
		data = []byte(strings.NewReplacer("{{APP_NAME}}", c.cfg.Name, "{{MODULE}}", c.cfg.Module, "{{ENGINE}}", string(c.cfg.Engine)).Replace(string(data)))
	}
	if path == "go.mod" {
		data = []byte(strings.Replace(string(data), "module "+c.oldModule, "module "+c.cfg.Module, 1))
	} else if !strings.HasSuffix(path, ".sum") && !strings.Contains(path, "/img/") {
		data = []byte(strings.ReplaceAll(string(data), c.oldModule+"/", c.cfg.Module+"/"))
		if strings.HasPrefix(path, "pkg/codegen/") {
			data = []byte(strings.ReplaceAll(string(data), strconv.Quote(c.oldModule), strconv.Quote(c.cfg.Module)))
		}
		data = []byte(strings.ReplaceAll(string(data), "go-boilerplate", c.cfg.Name))
	}
	if strings.HasSuffix(path, ".go") {
		formatted, err := imports.Process(path, data, &imports.Options{Comments: true, FormatOnly: true})
		if err != nil {
			return nil, fmt.Errorf("format %s: %w", path, err)
		}
		return formatted, nil
	}
	return data, nil
}

func engineSourceFile(path string) bool {
	if strings.HasSuffix(path, "_test.go") || path == "internal/handlers/http/middleware/idempotency_scope.go" {
		return false
	}
	if strings.HasPrefix(path, "internal/handlers/http/middleware/") || path == "internal/handlers/http/router.go" || path == "pkg/response/helpers.go" || path == "pkg/httpserver/server.go" {
		return true
	}
	return strings.HasPrefix(path, "internal/handlers/http/v1/") && filepath.Dir(path) != "internal/handlers/http/v1"
}

func importsFiber(source []byte) bool {
	file, err := parser.ParseFile(token.NewFileSet(), "template.go", source, parser.ImportsOnly)
	if err != nil {
		return false
	}
	for _, imp := range file.Imports {
		if strings.HasPrefix(imp.Path.Value, "\"github.com/gofiber/") {
			return true
		}
	}
	return false
}

func (c *projectCopier) applyTemplates(engine string) error {
	return fs.WalkDir(templates, "templates/"+engine, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative := strings.TrimSuffix(strings.TrimPrefix(path, "templates/"+engine+"/"), ".tmpl")
		data, err := templates.ReadFile(path)
		if err != nil {
			return err
		}
		data = []byte(strings.ReplaceAll(string(data), "go-boilerplate/", c.cfg.Module+"/"))
		if strings.HasSuffix(relative, ".go") {
			data, err = imports.Process(relative, data, &imports.Options{Comments: true, FormatOnly: true})
			if err != nil {
				return fmt.Errorf("format %s: %w", relative, err)
			}
		}
		if err := c.output.MkdirAll(filepath.Dir(relative), 0o750); err != nil {
			return err
		}
		return c.output.WriteFile(relative, data, 0o600)
	})
}

func (c *projectCopier) writeSettings() error {
	data, err := json.MarshalIndent(project.Settings{Engine: c.cfg.Engine}, "", "  ")
	if err != nil {
		return err
	}
	if err := c.output.WriteFile(project.ConfigFile, append(data, '\n'), 0o600); err != nil {
		return err
	}
	if c.cfg.Engine != project.Gin {
		return nil
	}
	data, err = c.output.ReadFile("go.mod")
	if err != nil {
		return err
	}
	data = append(data, []byte("\nrequire (\n\tgithub.com/gin-gonic/gin v1.12.0\n\tgithub.com/quic-go/quic-go v0.59.1 // indirect\n)\n")...)
	return c.output.WriteFile("go.mod", data, 0o600)
}
