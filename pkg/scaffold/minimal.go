package scaffold

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/imports"

	"go-boilerplate/pkg/project"
)

func (c *projectCopier) copyMinimal() error {
	paths := []string{"pkg/logger", "pkg/json", "pkg/response", "pkg/apperror", "pkg/project", "pkg/httpserver/options.go", "pkg/httpserver/server_test.go", "internal/handlers/http/v1/helper.go", "internal/handlers/http/v1/helper_test.go", "integration-test/runtimecheck", "config/env.go", "go.sum", "LICENSE", ".gitignore", ".dockerignore", ".golangci.yml"}
	for _, path := range paths {
		if err := c.output.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return err
		}
		if err := c.copyPath(path); err != nil {
			return err
		}
	}
	if err := c.writeMinimalModule(); err != nil {
		return err
	}
	if err := c.applyMinimalTemplates("minimal/common"); err != nil {
		return err
	}
	if c.cfg.Engine != project.Fiber {
		if err := c.applyMinimalTemplates("minimal/nethttp"); err != nil {
			return err
		}
	}
	if err := c.applyMinimalTemplates("minimal/" + string(c.cfg.Engine)); err != nil {
		return err
	}
	if c.cfg.Engine == project.Fiber {
		return c.copyMinimalHTTPHelpers("fiber", []string{"pkg/httpserver/server.go", "pkg/response/helpers.go"})
	}
	return c.copyMinimalHTTPHelpers("nethttp", []string{"pkg/request/request.go", "pkg/response/helpers.go"})
}

func (c *projectCopier) writeMinimalModule() error {
	source, err := c.source.ReadFile("go.mod")
	if err != nil {
		return err
	}
	original, err := modfile.Parse("go.mod", source, nil)
	if err != nil {
		return err
	}
	if original.Go == nil {
		return fmt.Errorf("template go.mod must specify a Go version")
	}
	mod, err := modfile.Parse("go.mod", []byte("module "+c.cfg.Module+"\n\ngo "+original.Go.Version+"\n"), nil)
	if err != nil {
		return err
	}
	dependencies := []string{"github.com/spf13/viper", "go.uber.org/zap", "github.com/goccy/go-json", "github.com/go-playground/validator/v10", "github.com/stretchr/testify", "golang.org/x/vuln"}
	if c.cfg.Engine == project.Fiber {
		dependencies = append(dependencies, "github.com/gofiber/fiber/v2", "golang.org/x/sync")
	}
	versions := make(map[string]string)
	for _, req := range original.Require {
		versions[req.Mod.Path] = req.Mod.Version
	}
	for _, dependency := range dependencies {
		version := versions[dependency]
		if version == "" {
			return fmt.Errorf("template has no pinned version for %s", dependency)
		}
		if err := mod.AddRequire(dependency, version); err != nil {
			return err
		}
	}
	if err := mod.AddTool("golang.org/x/vuln/cmd/govulncheck"); err != nil {
		return err
	}
	data, err := mod.Format()
	if err != nil {
		return err
	}
	return c.output.WriteFile("go.mod", data, 0o600)
}

func (c *projectCopier) applyMinimalTemplates(directory string) error {
	prefix := "templates/" + directory + "/"
	return fs.WalkDir(templates, strings.TrimSuffix(prefix, "/"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		return c.writeMinimalTemplate(path, strings.TrimSuffix(strings.TrimPrefix(path, prefix), ".tmpl"))
	})
}

func (c *projectCopier) copyMinimalHTTPHelpers(engine string, paths []string) error {
	for _, path := range paths {
		if err := c.writeMinimalTemplate("templates/"+engine+"/"+path+".tmpl", path); err != nil {
			return err
		}
	}
	return nil
}

func (c *projectCopier) writeMinimalTemplate(source, target string) error {
	data, err := templates.ReadFile(source)
	if err != nil {
		return err
	}
	routerType, routerNew, routerImport := "*http.ServeMux", "http.NewServeMux()", ""
	if c.cfg.Engine == project.Gin {
		routerType, routerNew, routerImport = "*gin.Engine", "gin.New()", `"github.com/gin-gonic/gin"`
	}
	data = []byte(strings.NewReplacer("go-boilerplate/", c.cfg.Module+"/", "{{APP_NAME}}", c.cfg.Name, "{{MODULE}}", c.cfg.Module, "{{ENGINE}}", string(c.cfg.Engine), "{{NATIVE_ROUTER_TYPE}}", routerType, "{{NATIVE_ROUTER_NEW}}", routerNew, "{{NATIVE_ROUTER_IMPORT}}", routerImport).Replace(string(data)))
	if strings.HasSuffix(target, ".go") {
		data, err = imports.Process(target, data, &imports.Options{Comments: true, FormatOnly: true})
		if err != nil {
			return fmt.Errorf("format %s: %w", target, err)
		}
	}
	if err := c.output.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		return err
	}
	return c.output.WriteFile(target, data, 0o600)
}
