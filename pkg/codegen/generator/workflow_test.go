package generator

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	codegenparser "go-boilerplate/pkg/codegen/parser"
	"go-boilerplate/pkg/project"
)

func TestGenerateWorkflowWritesContracts(t *testing.T) {
	for _, engine := range []project.Engine{project.Gin, project.Stdlib, project.Fiber} {
		t.Run(string(engine), func(t *testing.T) {
			output := t.TempDir()
			for _, layer := range []string{"repo", "usecase"} {
				directory := filepath.Join(output, "internal", layer)
				require.NoError(t, os.MkdirAll(directory, 0o750))
				require.NoError(t, os.WriteFile(filepath.Join(directory, "contracts.go"), []byte("package "+layer+"\n\nimport \"context\"\n\ntype (\nExisting interface { Ping(context.Context) error }\n)\n"), 0o600))
			}
			result := &codegenparser.ParseResult{
				Table: codegenparser.Table{Name: "widgets"},
				Fields: []codegenparser.GoField{
					{Name: "ID", ColumnName: "id", Type: "uint", JSONTag: "id", GormTags: "primaryKey"},
					{Name: "Name", ColumnName: "name", Type: "string", JSONTag: "name", GormTags: "not null"},
				},
			}
			gen := New(Config{Engine: engine, ModuleName: "example.com/api", OutputDir: output, Layers: []string{"entity", "dto", "repo", "usecase", "handler"}}, result)
			require.NoError(t, gen.Generate())
			require.NoError(t, filepath.WalkDir(output, func(path string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if entry.IsDir() || filepath.Ext(path) != ".go" {
					return nil
				}
				_, err = parser.ParseFile(token.NewFileSet(), path, nil, parser.AllErrors)
				return err
			}))
			for _, item := range []struct{ layer, declaration, dependency string }{
				{"repo", "WidgetRepo interface", "example.com/api/pkg/pagination"},
				{"usecase", "Widget interface", "example.com/api/internal/dto/widget"},
			} {
				data, err := os.ReadFile(filepath.Join(output, "internal", item.layer, "contracts.go"))
				require.NoError(t, err)
				require.Contains(t, string(data), "Existing interface")
				require.Contains(t, string(data), item.declaration)
				require.Contains(t, string(data), item.dependency)
			}
			require.FileExists(t, filepath.Join(output, "internal", "usecase", "widget", "create_test.go"))
			path := filepath.Join(output, "internal", "entity", "widget.go")
			require.NoError(t, os.WriteFile(path, []byte("user-owned content\n"), 0o600))
			require.ErrorContains(t, gen.Generate(), "file exists")
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, "user-owned content\n", string(data))
			gen.config.Force, gen.config.Layers = true, []string{"entity"}
			require.NoError(t, gen.Generate())
			data, err = os.ReadFile(path)
			require.NoError(t, err)
			require.Contains(t, string(data), "type Widget struct")
		})
	}
}
