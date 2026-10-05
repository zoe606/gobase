package scaffold_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/project"
	"go-boilerplate/pkg/scaffold"
)

func templateRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
func TestGenerateEngines(t *testing.T) {
	for _, engine := range []project.Engine{project.Gin, project.Stdlib, project.Fiber, ""} {
		t.Run(string(engine), func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "app")
			require.NoError(t, scaffold.Generate(scaffold.Config{Source: templateRoot(t), Output: output, Module: "example.com/app", Name: "app", Engine: engine}))
			settings, err := project.Read(output)
			require.NoError(t, err)
			if engine == "" {
				engine = project.Gin
			}
			require.Equal(t, engine, settings.Engine)
			mod, err := os.ReadFile(filepath.Join(output, "go.mod"))
			require.NoError(t, err)
			require.Contains(t, string(mod), "module example.com/app")
			handler, err := os.ReadFile(filepath.Join(output, "internal", "handlers", "http", "v1", "auth", "login.go"))
			require.NoError(t, err)
			switch engine {
			case project.Fiber:
				require.Contains(t, string(handler), "*fiber.Ctx")
			case project.Gin:
				require.Contains(t, string(handler), "*gin.Context")
				require.NotContains(t, string(handler), "gofiber")
			case project.Stdlib:
				require.Contains(t, string(handler), "http.ResponseWriter")
				require.NotContains(t, string(handler), "gofiber")
			}
			original, err := os.ReadFile(filepath.Join(templateRoot(t), "internal", "usecase", "auth", "login.go"))
			require.NoError(t, err)
			copied, err := os.ReadFile(filepath.Join(output, "internal", "usecase", "auth", "login.go"))
			require.NoError(t, err)
			require.NotEmpty(t, original)
			require.Contains(t, string(copied), "example.com/app/")
			require.NoFileExists(t, filepath.Join(output, ".env"))
			require.NoFileExists(t, filepath.Join(output, "config", "config.yaml"))
			require.NoDirExists(t, filepath.Join(output, "pkg", "scaffold"))
			require.FileExists(t, filepath.Join(output, ".dockerignore"))
			require.NoFileExists(t, filepath.Join(output, "docs", "http-engines.md"))
			require.NoFileExists(t, filepath.Join(output, "docs", "roadmap.md"))
			require.NoDirExists(t, filepath.Join(output, "docs", "superpowers"))
			readme, err := os.ReadFile(filepath.Join(output, "README.md"))
			require.NoError(t, err)
			require.Contains(t, string(readme), "HTTP engine: `"+string(engine)+"`")
			require.NotContains(t, string(readme), "make init")
			require.NotContains(t, string(readme), "test-engines")
			guide, err := os.ReadFile(filepath.Join(output, ".github", "CONTRIBUTING.md"))
			require.NoError(t, err)
			require.Contains(t, string(guide), "`"+string(engine)+"` HTTP engine")
			require.Error(t, scaffold.Generate(scaffold.Config{Source: templateRoot(t), Output: output, Module: "example.com/app", Name: "app", Engine: engine}))
		})
	}
}

func TestEngineSourcesDoNotDependOnCheckoutHandlers(t *testing.T) {
	source := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(source, "go.mod"), []byte("module go-boilerplate\n\ngo 1.27.1\n"), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(source, "pkg", "scaffold", "templates"), 0o750))
	handlers := filepath.Join(source, "internal", "handlers", "http", "v1", "auth")
	require.NoError(t, os.MkdirAll(handlers, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(handlers, "login.go"), []byte("invalid checkout handler source"), 0o600))
	for _, engine := range []project.Engine{project.Gin, project.Stdlib, project.Fiber} {
		output := filepath.Join(t.TempDir(), "app")
		require.NoError(t, scaffold.Generate(scaffold.Config{Source: source, Output: output, Module: "example.com/api", Name: "api", Engine: engine}))
		data, err := os.ReadFile(filepath.Join(output, "internal", "handlers", "http", "v1", "auth", "login.go"))
		require.NoError(t, err)
		require.Contains(t, string(data), "func (h *Handler) Login")
		require.NotContains(t, string(data), "invalid checkout handler source")
	}
}
func TestInvalidGeneration(t *testing.T) {
	base := scaffold.Config{Source: templateRoot(t), Output: filepath.Join(t.TempDir(), "app"), Module: "example.com/app", Name: "app", Engine: project.Gin}
	for _, change := range []func(*scaffold.Config){func(c *scaffold.Config) { c.Engine = "bad" }, func(c *scaffold.Config) { c.Module = "../bad" }, func(c *scaffold.Config) { c.Name = "../bad" }, func(c *scaffold.Config) { c.Output = c.Source }, func(c *scaffold.Config) { c.Output = filepath.Join(c.Source, "generated") }, func(c *scaffold.Config) { c.Source = filepath.Join(t.TempDir(), "missing") }} {
		cfg := base
		change(&cfg)
		require.Error(t, scaffold.Generate(cfg))
		require.NoDirExists(t, base.Output)
	}
}

func TestOutputSymlinkInsideSource(t *testing.T) {
	source := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(source, "go.mod"), []byte("module example.com/source\n"), 0o600))
	link := filepath.Join(t.TempDir(), "linked-source")
	require.NoError(t, os.Symlink(source, link))
	err := scaffold.Generate(scaffold.Config{Source: source, Output: filepath.Join(link, "app"), Module: "example.com/app", Name: "app", Engine: project.Fiber})
	require.ErrorContains(t, err, "output must be outside")
	require.NoDirExists(t, filepath.Join(source, "app"))
}
