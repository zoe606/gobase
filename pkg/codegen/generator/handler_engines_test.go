package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	codegenparser "go-boilerplate/pkg/codegen/parser"
	"go-boilerplate/pkg/project"
)

func TestGenerateNativeHandlers(t *testing.T) {
	for _, engine := range []project.Engine{project.Gin, project.Stdlib, project.Fiber} {
		for _, table := range []string{"articles", "user_profiles"} {
			t.Run(string(engine)+"/"+table, func(t *testing.T) {
				output := t.TempDir()
				gen := New(Config{Engine: engine, ModuleName: "example.com/api", OutputDir: output}, &codegenparser.ParseResult{Table: codegenparser.Table{Name: table}})
				require.NoError(t, gen.GenerateHandler())
				for _, name := range []string{"handler", "create", "get_by_id", "list", "update", "delete"} {
					path := filepath.Join(output, "internal", "handlers", "http", "v1", gen.packageName(), name+".go")
					data, err := os.ReadFile(path)
					require.NoError(t, err)
					file, err := parser.ParseFile(token.NewFileSet(), path, data, parser.ParseComments)
					require.NoError(t, err)
					require.Contains(t, string(data), "example.com/api/")
					if engine != project.Fiber {
						require.NotContains(t, string(data), "fiber")
					}
					if engine != project.Gin {
						require.NotContains(t, string(data), "gin-gonic")
					}
					for _, declaration := range file.Decls {
						fn, ok := declaration.(*ast.FuncDecl)
						if !ok || fn.Recv == nil || fn.Name.Name == "RegisterRoutes" {
							continue
						}
						require.NotNil(t, fn.Doc)
						require.Contains(t, fn.Doc.Text(), "@Router")
						switch engine {
						case project.Gin:
							require.Contains(t, string(data), "c *gin.Context")
							require.Len(t, fn.Type.Params.List, 1)
						case project.Stdlib:
							require.Contains(t, string(data), "w http.ResponseWriter, r *http.Request")
							require.Len(t, fn.Type.Params.List, 2)
						case project.Fiber:
							require.Contains(t, string(data), "ctx *fiber.Ctx")
							require.Len(t, fn.Type.Params.List, 1)
						}
					}
				}
			})
		}
	}
}
