package generator

import (
	"fmt"
	"strings"
)

// GenerateUseCase generates the usecase package files.
func (g *Generator) GenerateUseCase() error {
	pkgName := g.packageName()
	basePath := fmt.Sprintf("internal/usecase/%s", pkgName)

	// Generate interface addition to contracts.go
	interfaceContent := g.buildUseCaseInterfaceContent()
	contractsPath := "internal/usecase/contracts.go"

	// Try to append to existing file
	err := g.appendToFile(contractsPath, interfaceContent, "")
	if err != nil {
		if g.config.DryRun {
			fmt.Printf("\n=== Add to %s ===\n", contractsPath)
			fmt.Println(interfaceContent)
		} else {
			fmt.Printf("Please add the following to %s:\n%s\n", contractsPath, interfaceContent)
		}
	}

	// Generate main usecase file
	mainContent := g.buildUseCaseMainContent()
	if err := g.writeFile(basePath+"/"+pkgName+".go", mainContent); err != nil {
		return err
	}

	// Generate errors file
	errorsContent := g.buildUseCaseErrorsContent()
	if err := g.writeFile(basePath+"/errors.go", errorsContent); err != nil {
		return err
	}

	// Generate CRUD method files
	methods := []struct {
		name     string
		fileName string
		content  func() string
	}{
		{"Create", "create.go", g.buildUseCaseCreateContent},
		{"GetByID", "get_by_id.go", g.buildUseCaseGetByIDContent},
		{"List", "list.go", g.buildUseCaseListContent},
		{"Update", "update.go", g.buildUseCaseUpdateContent},
		{"Delete", "delete.go", g.buildUseCaseDeleteContent},
	}

	for _, method := range methods {
		content := method.content()
		if err := g.writeFile(basePath+"/"+method.fileName, content); err != nil {
			return err
		}

		// Generate test file
		testContent := g.buildUseCaseTestContent(method.name)
		if err := g.writeFile(basePath+"/"+strings.TrimSuffix(method.fileName, ".go")+"_test.go", testContent); err != nil {
			return err
		}
	}

	// Generate mocks_test.go
	mocksContent := g.buildUseCaseMocksTestContent()
	if err := g.writeFile(basePath+"/mocks_test.go", mocksContent); err != nil {
		return err
	}

	return nil
}

// buildUseCaseInterfaceContent builds the usecase interface for contracts.go.
func (g *Generator) buildUseCaseInterfaceContent() string {
	var sb strings.Builder

	entityName := g.entityName()
	pkgName := g.packageName()
	dtoAlias := pkgName + "dto"
	dtoImport := g.config.ModuleName + "/internal/dto/" + pkgName

	// Include import hint so the caller knows which import to add
	sb.WriteString(fmt.Sprintf("\n\t// import %q\n", dtoImport))

	sb.WriteString(fmt.Sprintf("\n\t// %s defines %s use case operations.\n", entityName, entityName))
	sb.WriteString(fmt.Sprintf("\t%s interface {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\tCreate(ctx context.Context, req %s.CreateRequest) (*%s.Response, error)\n", dtoAlias, dtoAlias))
	sb.WriteString(fmt.Sprintf("\t\tGetByID(ctx context.Context, id uint) (*%s.Response, error)\n", dtoAlias))
	sb.WriteString(fmt.Sprintf("\t\tList(ctx context.Context, req %s.ListRequest) (*%s.ListResponse, error)\n", dtoAlias, dtoAlias))
	sb.WriteString(fmt.Sprintf("\t\tUpdate(ctx context.Context, id uint, req %s.UpdateRequest) (*%s.Response, error)\n", dtoAlias, dtoAlias))
	sb.WriteString("\t\tDelete(ctx context.Context, id uint) error\n")
	sb.WriteString("\t}\n")

	return sb.String()
}

// buildUseCaseMainContent builds the main usecase file content.
func (g *Generator) buildUseCaseMainContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()

	// Package declaration with go:generate
	sb.WriteString(fmt.Sprintf("// Package %s provides %s management use cases.\n", pkgName, pkgName))
	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	sb.WriteString("//go:generate mockgen -source=../../repo/contracts.go -destination=mocks_repo_test.go -package=" + pkgName + "_test\n\n")

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/repo"))
	sb.WriteString(")\n\n")

	// Struct
	sb.WriteString(fmt.Sprintf("// UseCase implements %s business logic.\n", pkgName))
	sb.WriteString("type UseCase struct {\n")
	sb.WriteString(fmt.Sprintf("\t%sRepo repo.%sRepo\n", g.varName(), entityName))
	sb.WriteString("}\n\n")

	// Constructor
	sb.WriteString(fmt.Sprintf("// New creates a new %s use case.\n", pkgName))
	sb.WriteString(fmt.Sprintf("func New(%sRepo repo.%sRepo) *UseCase {\n", g.varName(), entityName))
	sb.WriteString("\treturn &UseCase{\n")
	sb.WriteString(fmt.Sprintf("\t\t%sRepo: %sRepo,\n", g.varName(), g.varName()))
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// buildUseCaseErrorsContent builds the errors.go file content.
func (g *Generator) buildUseCaseErrorsContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()

	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))
	sb.WriteString("import \"errors\"\n\n")

	sb.WriteString("// Error definitions.\n")
	sb.WriteString("var (\n")
	sb.WriteString(fmt.Sprintf("\t// ErrNotFound indicates that the %s was not found.\n", pkgName))
	sb.WriteString(fmt.Sprintf("\tErrNotFound = errors.New(\"%s not found\")\n\n", pkgName))
	sb.WriteString(fmt.Sprintf("\t// ErrAlreadyExists indicates that the %s already exists.\n", pkgName))
	sb.WriteString(fmt.Sprintf("\tErrAlreadyExists = errors.New(\"%s already exists\")\n\n", pkgName))
	sb.WriteString(fmt.Sprintf("\t// ErrInvalid indicates invalid %s data.\n", entityName))
	sb.WriteString(fmt.Sprintf("\tErrInvalid = errors.New(\"invalid %s data\")\n", pkgName))
	sb.WriteString(")\n")

	return sb.String()
}

// buildUseCaseCreateContent builds the create.go file content.
func (g *Generator) buildUseCaseCreateContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()
	varName := g.varName()
	dtoAlias := pkgName + "dto"

	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"fmt\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/dto/"+pkgName))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/entity"))
	sb.WriteString(")\n\n")

	// Method
	sb.WriteString(fmt.Sprintf("// Create creates a new %s.\n", pkgName))
	sb.WriteString(fmt.Sprintf("func (uc *UseCase) Create(ctx context.Context, req %s.CreateRequest) (*%s.Response, error) {\n", dtoAlias, dtoAlias))

	// Build entity from request fields
	sb.WriteString(fmt.Sprintf("\t%s := &entity.%s{\n", varName, entityName))
	for _, field := range g.result.Fields {
		if g.isCreateRequestField(field.ColumnName) {
			sb.WriteString(fmt.Sprintf("\t\t%s: req.%s,\n", field.Name, field.Name))
		}
	}
	sb.WriteString("\t}\n\n")

	sb.WriteString(fmt.Sprintf("\tif err := uc.%sRepo.Create(ctx, %s); err != nil {\n", varName, varName))
	sb.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"%s - Create: %%w\", err)\n", pkgName))
	sb.WriteString("\t}\n\n")
	sb.WriteString(fmt.Sprintf("\treturn %s.NewResponse(%s), nil\n", dtoAlias, varName))
	sb.WriteString("}\n")

	return sb.String()
}

// buildUseCaseGetByIDContent builds the get_by_id.go file content.
func (g *Generator) buildUseCaseGetByIDContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	varName := g.varName()
	dtoAlias := pkgName + "dto"

	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"errors\"\n")
	sb.WriteString("\t\"fmt\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/dto/"+pkgName))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/repo"))
	sb.WriteString(")\n\n")

	// Method
	sb.WriteString(fmt.Sprintf("// GetByID retrieves a %s by ID.\n", pkgName))
	sb.WriteString(fmt.Sprintf("func (uc *UseCase) GetByID(ctx context.Context, id uint) (*%s.Response, error) {\n", dtoAlias))
	sb.WriteString(fmt.Sprintf("\t%s, err := uc.%sRepo.GetByID(ctx, id)\n", varName, varName))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\tif errors.Is(err, repo.ErrNotFound) {\n")
	sb.WriteString("\t\t\treturn nil, ErrNotFound\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"%s - GetByID: %%w\", err)\n", pkgName))
	sb.WriteString("\t}\n\n")
	sb.WriteString(fmt.Sprintf("\treturn %s.NewResponse(%s), nil\n", dtoAlias, varName))
	sb.WriteString("}\n")

	return sb.String()
}

// buildUseCaseListContent builds the list.go file content.
func (g *Generator) buildUseCaseListContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	varName := g.varName()
	dtoAlias := pkgName + "dto"

	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"fmt\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/dto/"+pkgName))
	sb.WriteString(")\n\n")

	// Method
	sb.WriteString(fmt.Sprintf("// List retrieves a paginated list of %ss with filters.\n", pkgName))
	sb.WriteString(fmt.Sprintf("func (uc *UseCase) List(ctx context.Context, req %s.ListRequest) (*%s.ListResponse, error) {\n", dtoAlias, dtoAlias))
	sb.WriteString("\treq.Params.Normalize()\n\n")
	sb.WriteString(fmt.Sprintf("\t%ss, total, err := uc.%sRepo.List(ctx, req.Params)\n", varName, varName))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"%s - List: %%w\", err)\n", pkgName))
	sb.WriteString("\t}\n\n")
	sb.WriteString(fmt.Sprintf("\treturn %s.NewListResponse(%ss, total, req.Params), nil\n", dtoAlias, varName))
	sb.WriteString("}\n")

	return sb.String()
}

// buildUseCaseUpdateContent builds the update.go file content.
func (g *Generator) buildUseCaseUpdateContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	varName := g.varName()
	dtoAlias := pkgName + "dto"

	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"errors\"\n")
	sb.WriteString("\t\"fmt\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/dto/"+pkgName))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/repo"))
	sb.WriteString(")\n\n")

	// Method
	sb.WriteString(fmt.Sprintf("// Update updates a %s.\n", pkgName))
	sb.WriteString(fmt.Sprintf("func (uc *UseCase) Update(ctx context.Context, id uint, req %s.UpdateRequest) (*%s.Response, error) {\n", dtoAlias, dtoAlias))
	sb.WriteString(fmt.Sprintf("\t%s, err := uc.%sRepo.GetByID(ctx, id)\n", varName, varName))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\tif errors.Is(err, repo.ErrNotFound) {\n")
	sb.WriteString("\t\t\treturn nil, ErrNotFound\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"%s - Update: %%w\", err)\n", pkgName))
	sb.WriteString("\t}\n\n")

	// Generate partial update field checks
	sb.WriteString("\t// Apply partial updates\n")
	for _, field := range g.result.Fields {
		if g.isUpdateRequestField(field.ColumnName) {
			sb.WriteString(fmt.Sprintf("\tif req.%s != nil {\n", field.Name))
			sb.WriteString(fmt.Sprintf("\t\t%s.%s = *req.%s\n", varName, field.Name, field.Name))
			sb.WriteString("\t}\n")
		}
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("\tif err := uc.%sRepo.Update(ctx, %s); err != nil {\n", varName, varName))
	sb.WriteString(fmt.Sprintf("\t\treturn nil, fmt.Errorf(\"%s - Update: %%w\", err)\n", pkgName))
	sb.WriteString("\t}\n\n")
	sb.WriteString(fmt.Sprintf("\treturn %s.NewResponse(%s), nil\n", dtoAlias, varName))
	sb.WriteString("}\n")

	return sb.String()
}

// buildUseCaseDeleteContent builds the delete.go file content.
func (g *Generator) buildUseCaseDeleteContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	varName := g.varName()

	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"errors\"\n")
	sb.WriteString("\t\"fmt\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/repo"))
	sb.WriteString(")\n\n")

	// Method
	sb.WriteString(fmt.Sprintf("// Delete deletes a %s by ID.\n", pkgName))
	sb.WriteString("func (uc *UseCase) Delete(ctx context.Context, id uint) error {\n")
	sb.WriteString(fmt.Sprintf("\tif err := uc.%sRepo.Delete(ctx, id); err != nil {\n", varName))
	sb.WriteString("\t\tif errors.Is(err, repo.ErrNotFound) {\n")
	sb.WriteString("\t\t\treturn ErrNotFound\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"%s - Delete: %%w\", err)\n", pkgName))
	sb.WriteString("\t}\n\n")
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n")

	return sb.String()
}

// buildUseCaseTestContent builds a test file scaffold.
func (g *Generator) buildUseCaseTestContent(methodName string) string {
	var sb strings.Builder

	pkgName := g.packageName()

	sb.WriteString(fmt.Sprintf("package %s_test\n\n", pkgName))

	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"testing\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/usecase/"+pkgName))
	sb.WriteString(")\n\n")

	sb.WriteString(fmt.Sprintf("func Test%s(t *testing.T) {\n", methodName))
	sb.WriteString("\t// TODO: Implement test\n")
	sb.WriteString("\t_ = context.Background()\n")
	sb.WriteString(fmt.Sprintf("\t_ = %s.New(nil)\n", pkgName))
	sb.WriteString("\tt.Skip(\"Test not implemented\")\n")
	sb.WriteString("}\n")

	return sb.String()
}

// buildUseCaseMocksTestContent builds the mocks_test.go file content.
func (g *Generator) buildUseCaseMocksTestContent() string {
	pkgName := g.packageName()

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("package %s_test\n", pkgName))

	return sb.String()
}
