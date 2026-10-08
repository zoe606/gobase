package scaffold_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/project"
	"go-boilerplate/pkg/scaffold"
)

func TestGenerateMinimal(t *testing.T) {
	for _, engine := range []project.Engine{project.Gin, project.Stdlib, project.Fiber} {
		t.Run(string(engine), func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "service")
			require.NoError(t, scaffold.Generate(scaffold.Config{Source: templateRoot(t), Output: output, Module: "example.com/service", Name: "service", Engine: engine, Profile: project.Minimal}))
			settings, err := project.Read(output)
			require.NoError(t, err)
			require.Equal(t, project.Settings{Engine: engine, Profile: project.Minimal}, settings)
			for _, path := range []string{"cmd/worker", "migrations", "internal/entity", "internal/repo", "internal/usecase", "internal/dto", "pkg/postgres", "pkg/redis", "pkg/asynq", "pkg/telemetry", "docs/roadmap.md", "pkg/scaffold"} {
				_, err := os.Stat(filepath.Join(output, path))
				require.True(t, os.IsNotExist(err), path)
			}
			for _, path := range []string{"cmd/app/main.go", "internal/app/app.go", "config/config.go", "config/config.example.yaml", "internal/handlers/http/router_test.go", "pkg/codegen/cmd/codegen/main.go", "pkg/codegen/cmd/wire/main.go", "deployment/docker/Dockerfile", ".env.example", ".github/CONTRIBUTING.md", ".github/workflows/ci.yml"} {
				require.FileExists(t, filepath.Join(output, path))
			}
			mod, err := os.ReadFile(filepath.Join(output, "go.mod"))
			require.NoError(t, err)
			for _, forbidden := range []string{"gorm.io", "github.com/redis", "github.com/hibiken", "github.com/minio", "github.com/resend", "github.com/golang-jwt", "swaggo", "opentelemetry", "air-verse"} {
				require.NotContains(t, string(mod), forbidden)
			}
			config, err := os.ReadFile(filepath.Join(output, "config", "config.example.yaml"))
			require.NoError(t, err)
			require.NotContains(t, string(config), "postgres")
			require.NotContains(t, string(config), "redis")
			require.Contains(t, string(config), "name: service")
			readme, err := os.ReadFile(filepath.Join(output, "README.md"))
			require.NoError(t, err)
			require.Contains(t, string(readme), "Profile: `minimal`")
			require.Contains(t, string(readme), "HTTP engine: `"+string(engine)+"`")
			require.NotContains(t, string(readme), "{{")
		})
	}
}

func TestInvalidProfileCreatesNoOutput(t *testing.T) {
	output := filepath.Join(t.TempDir(), "service")
	err := scaffold.Generate(scaffold.Config{Source: templateRoot(t), Output: output, Module: "example.com/service", Name: "service", Profile: "unknown"})
	require.ErrorContains(t, err, "choose full or minimal")
	require.NoDirExists(t, output)
}
