// Command verifyengines checks generated projects and their code generators.
package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go-boilerplate/pkg/project"
	"go-boilerplate/pkg/scaffold"
)

//go:embed testdata
var testSources embed.FS

func main() {
	engine := flag.String("engine", "", "one engine to check; empty checks all engines")
	output := flag.String("output", "", "directory for generated projects; empty uses a temporary directory")
	lint := flag.Bool("lint", false, "run golangci-lint in generated projects")
	flag.Parse()
	if err := verify(*engine, *output, *lint); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func verify(selected, output string, lint bool) error {
	engines := []project.Engine{project.Gin, project.Stdlib, project.Fiber}
	if selected != "" {
		engines = []project.Engine{project.Engine(selected)}
	}
	if output == "" {
		var err error
		output, err = os.MkdirTemp("", "gobase-engines-")
		if err != nil {
			return err
		}
		defer func() { _ = os.RemoveAll(output) }()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	for _, engine := range engines {
		if err := engine.Validate(); err != nil {
			return err
		}
		path := filepath.Join(output, string(engine))
		fmt.Printf("Checking %s in %s\n", engine, path)
		if err := scaffold.Generate(scaffold.Config{Source: ".", Output: path, Module: "example.com/" + string(engine) + "app", Name: string(engine) + "app", Engine: engine}); err != nil {
			return err
		}
		if err := checkProject(ctx, path, engine, lint); err != nil {
			return fmt.Errorf("%s: %w", engine, err)
		}
	}
	return nil
}

func run(ctx context.Context, path, command string, args ...string) error {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func checkProject(ctx context.Context, path string, engine project.Engine, lint bool) error {
	if err := run(ctx, path, "go", "mod", "tidy"); err != nil {
		return err
	}
	if err := run(ctx, path, "go", "mod", "verify"); err != nil {
		return err
	}
	if err := run(ctx, path, "go", "build", "./..."); err != nil {
		return err
	}
	if err := checkDependencies(ctx, path, engine); err != nil {
		return err
	}
	if err := run(ctx, path, "go", "test", "-race", "./internal/...", "./pkg/..."); err != nil {
		return err
	}
	if err := checkCodegen(ctx, path); err != nil {
		return err
	}
	if lint {
		return run(ctx, path, "golangci-lint", "run")
	}
	return nil
}

func checkDependencies(ctx context.Context, path string, engine project.Engine) error {
	cmd := exec.CommandContext(ctx, "go", "list", "-deps", "./cmd/app", "./cmd/worker")
	cmd.Dir = path
	data, err := cmd.Output()
	if err != nil {
		return err
	}
	forbidden := []string{"github.com/gin-gonic/gin"}
	if engine != project.Fiber {
		forbidden = []string{"github.com/gofiber/", "github.com/valyala/fasthttp"}
	}
	if engine == project.Stdlib {
		forbidden = append(forbidden, "github.com/gin-gonic/gin")
	}
	for _, dependency := range forbidden {
		if strings.Contains(string(data), dependency) {
			return fmt.Errorf("unexpected runtime dependency %s", dependency)
		}
	}
	return nil
}

func checkCodegen(ctx context.Context, path string) error {
	root, err := os.OpenRoot(path)
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	migration := "CREATE TABLE widgets (id BIGSERIAL PRIMARY KEY, name VARCHAR(255) NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT NOW(), updated_at TIMESTAMP NOT NULL DEFAULT NOW());\n"
	name := "migrations/000099_create_widgets.up.sql"
	if err := root.WriteFile(name, []byte(migration), 0o600); err != nil {
		return err
	}
	if err := root.WriteFile("migrations/000099_create_widgets.down.sql", []byte("DROP TABLE IF EXISTS widgets;\n"), 0o600); err != nil {
		return err
	}
	if err := run(ctx, path, "go", "run", "./pkg/codegen/cmd/codegen", "-m", name, "-l", "entity,dto,repo,usecase,handler"); err != nil {
		return err
	}
	if err := run(ctx, path, "go", "run", "./pkg/codegen/cmd/wire"); err != nil {
		return err
	}
	integration, err := testSources.ReadFile("testdata/widget_integration_test.go.txt")
	if err != nil {
		return err
	}
	if err := root.WriteFile("integration-test/codegen_test.go", integration, 0o600); err != nil {
		return err
	}
	if err := run(ctx, path, "go", "fmt", "./..."); err != nil {
		return err
	}
	if err := run(ctx, path, "go", "tool", "swag", "init", "-g", "internal/handlers/http/router.go", "--parseDependency", "--parseInternal"); err != nil {
		return err
	}
	if err := checkSwagger(root); err != nil {
		return err
	}
	if err := run(ctx, path, "go", "build", "./..."); err != nil {
		return err
	}
	return run(ctx, path, "go", "test", "./internal/usecase/widget/...", "./internal/handlers/http/...")
}

type swaggerOperation struct {
	Responses  map[string]json.RawMessage `json:"responses"`
	Parameters []struct {
		In     string          `json:"in"`
		Schema json.RawMessage `json:"schema"`
	} `json:"parameters"`
}

func (operation swaggerOperation) hasRequestBody() bool {
	for _, parameter := range operation.Parameters {
		if parameter.In == "body" && len(parameter.Schema) > 0 {
			return true
		}
	}
	return false
}

func checkSwagger(root *os.Root) error {
	data, err := root.ReadFile("docs/swagger.json")
	if err != nil {
		return err
	}
	var spec struct {
		Paths map[string]map[string]swaggerOperation `json:"paths"`
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		return err
	}
	for _, route := range []struct{ path, method, status string }{
		{"/widgets", "post", "201"}, {"/widgets", "get", "200"},
		{"/widgets/{id}", "get", "200"}, {"/widgets/{id}", "put", "200"}, {"/widgets/{id}", "delete", "204"},
	} {
		operation := spec.Paths[route.path][route.method]
		for _, status := range []string{route.status, "400"} {
			if len(operation.Responses[status]) == 0 {
				return fmt.Errorf("swagger missing %s %s response %s", route.method, route.path, status)
			}
		}
		if (route.method == "post" || route.method == "put") && !operation.hasRequestBody() {
			return fmt.Errorf("swagger missing %s %s request schema", route.method, route.path)
		}
	}
	return nil
}
