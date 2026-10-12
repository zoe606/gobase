package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

func (g *Generator) addContractImport(source, path string) (string, error) {
	importPath := g.config.ModuleName + "/pkg/pagination"
	alias := ""
	if path == "internal/usecase/contracts.go" {
		importPath = g.config.ModuleName + "/internal/dto/" + g.packageName()
		alias = g.packageName() + "dto"
	} else if path != "internal/repo/contracts.go" {
		return source, nil
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, source, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("parse contracts: %w", err)
	}
	for _, imp := range file.Imports {
		if imp.Path.Value == strconv.Quote(importPath) {
			return source, nil
		}
	}
	astutil.AddNamedImport(fset, file, alias, importPath)
	var result bytes.Buffer
	if err := format.Node(&result, fset, file); err != nil {
		return "", err
	}
	return result.String(), nil
}
