package generator

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"golang.org/x/tools/imports"
)

//go:embed templates/handler
var handlerTemplates embed.FS

type handlerTemplateData struct {
	Module   string
	Package  string
	Entity   string
	Variable string
}

// GenerateHandler writes handlers from the selected engine's templates.
func (g *Generator) GenerateHandler() error {
	if err := g.config.Engine.Validate(); err != nil {
		return err
	}
	basePath := fmt.Sprintf("internal/handlers/http/v1/%s", g.packageName())
	for _, name := range []string{"handler", "create", "get_by_id", "list", "update", "delete"} {
		content, err := g.handlerContent(name)
		if err != nil {
			return err
		}
		path := basePath + "/" + name + ".go"
		formatted, err := imports.Process(path, []byte(content), &imports.Options{Comments: true, FormatOnly: true})
		if err != nil {
			return fmt.Errorf("format handler %s: %w", path, err)
		}
		if err := g.writeFile(path, string(formatted)); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) handlerContent(name string) (string, error) {
	path := "templates/handler/" + string(g.config.Engine) + "/" + name + ".go.tmpl"
	source, err := handlerTemplates.ReadFile(path)
	if err != nil {
		return "", err
	}
	renderer, err := template.New(name).Option("missingkey=error").Parse(string(source))
	if err != nil {
		return "", err
	}
	data := handlerTemplateData{Module: g.config.ModuleName, Package: g.packageName(), Entity: g.entityName(), Variable: g.varName()}
	var result bytes.Buffer
	if err := renderer.Execute(&result, data); err != nil {
		return "", err
	}
	return result.String(), nil
}
