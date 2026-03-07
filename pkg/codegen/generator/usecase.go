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

// buildUseCaseTestContent builds a table-driven test file for a usecase method.
func (g *Generator) buildUseCaseTestContent(methodName string) string {
	switch methodName {
	case "Create":
		return g.buildUseCaseCreateTestContent()
	case "GetByID":
		return g.buildUseCaseGetByIDTestContent()
	case "List":
		return g.buildUseCaseListTestContent()
	case "Update":
		return g.buildUseCaseUpdateTestContent()
	case "Delete":
		return g.buildUseCaseDeleteTestContent()
	default:
		return g.buildUseCaseGenericTestContent(methodName)
	}
}

// buildUseCaseCreateTestContent builds the create_test.go file content.
func (g *Generator) buildUseCaseCreateTestContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()
	dtoAlias := pkgName + "dto"

	sb.WriteString(fmt.Sprintf("package %s_test\n\n", pkgName))

	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"errors\"\n")
	sb.WriteString("\t\"testing\"\n\n")
	sb.WriteString("\t\"github.com/stretchr/testify/require\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/dto/"+pkgName))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/entity"))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/usecase/"+pkgName))
	sb.WriteString(")\n\n")

	sb.WriteString("func TestCreate(t *testing.T) {\n")
	sb.WriteString("\tt.Parallel()\n\n")

	sb.WriteString("\ttests := []struct {\n")
	sb.WriteString("\t\tname    string\n")
	sb.WriteString(fmt.Sprintf("\t\treq     %s.CreateRequest\n", dtoAlias))
	sb.WriteString(fmt.Sprintf("\t\tmockFn  func(*mock%sRepo)\n", entityName))
	sb.WriteString("\t\twantErr bool\n")
	sb.WriteString("\t}{\n")

	// Success case
	sb.WriteString("\t\t{\n")
	sb.WriteString("\t\t\tname: \"success\",\n")
	sb.WriteString(fmt.Sprintf("\t\t\treq:  %s.CreateRequest{},\n", dtoAlias))
	sb.WriteString(fmt.Sprintf("\t\t\tmockFn: func(m *mock%sRepo) {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\t\t\tm.createFn = func(ctx context.Context, e *entity.%s) error {\n", entityName))
	sb.WriteString("\t\t\t\t\treturn nil\n")
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t},\n")

	// Error case
	sb.WriteString("\t\t{\n")
	sb.WriteString("\t\t\tname: \"repo error\",\n")
	sb.WriteString(fmt.Sprintf("\t\t\treq:  %s.CreateRequest{},\n", dtoAlias))
	sb.WriteString(fmt.Sprintf("\t\t\tmockFn: func(m *mock%sRepo) {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\t\t\tm.createFn = func(ctx context.Context, e *entity.%s) error {\n", entityName))
	sb.WriteString("\t\t\t\t\treturn errors.New(\"db error\")\n")
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t\twantErr: true,\n")
	sb.WriteString("\t\t},\n")

	sb.WriteString("\t}\n\n")

	// Test runner
	sb.WriteString("\tfor _, tt := range tests {\n")
	sb.WriteString("\t\tt.Run(tt.name, func(t *testing.T) {\n")
	sb.WriteString("\t\t\tt.Parallel()\n\n")
	sb.WriteString(fmt.Sprintf("\t\t\tmockRepo := &mock%sRepo{}\n", entityName))
	sb.WriteString("\t\t\ttt.mockFn(mockRepo)\n")
	sb.WriteString(fmt.Sprintf("\t\t\tuc := %s.New(mockRepo)\n\n", pkgName))
	sb.WriteString("\t\t\tresult, err := uc.Create(context.Background(), tt.req)\n")
	sb.WriteString("\t\t\tif tt.wantErr {\n")
	sb.WriteString("\t\t\t\trequire.Error(t, err)\n")
	sb.WriteString("\t\t\t\treturn\n")
	sb.WriteString("\t\t\t}\n")
	sb.WriteString("\t\t\trequire.NoError(t, err)\n")
	sb.WriteString("\t\t\trequire.NotNil(t, result)\n")
	sb.WriteString("\t\t})\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// buildUseCaseGetByIDTestContent builds the get_by_id_test.go file content.
func (g *Generator) buildUseCaseGetByIDTestContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()

	sb.WriteString(fmt.Sprintf("package %s_test\n\n", pkgName))

	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"errors\"\n")
	sb.WriteString("\t\"testing\"\n\n")
	sb.WriteString("\t\"github.com/stretchr/testify/require\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/entity"))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/repo"))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/usecase/"+pkgName))
	sb.WriteString(")\n\n")

	sb.WriteString("func TestGetByID(t *testing.T) {\n")
	sb.WriteString("\tt.Parallel()\n\n")

	sb.WriteString("\ttests := []struct {\n")
	sb.WriteString("\t\tname    string\n")
	sb.WriteString("\t\tid      uint\n")
	sb.WriteString(fmt.Sprintf("\t\tmockFn  func(*mock%sRepo)\n", entityName))
	sb.WriteString("\t\twantErr bool\n")
	sb.WriteString("\t}{\n")

	// Success case
	sb.WriteString("\t\t{\n")
	sb.WriteString("\t\t\tname: \"success\",\n")
	sb.WriteString("\t\t\tid:   1,\n")
	sb.WriteString(fmt.Sprintf("\t\t\tmockFn: func(m *mock%sRepo) {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\t\t\tm.getByIDFn = func(ctx context.Context, id uint) (*entity.%s, error) {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\t\t\t\treturn &entity.%s{}, nil\n", entityName))
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t},\n")

	// Not found case
	sb.WriteString("\t\t{\n")
	sb.WriteString("\t\t\tname: \"not found\",\n")
	sb.WriteString("\t\t\tid:   999,\n")
	sb.WriteString(fmt.Sprintf("\t\t\tmockFn: func(m *mock%sRepo) {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\t\t\tm.getByIDFn = func(ctx context.Context, id uint) (*entity.%s, error) {\n", entityName))
	sb.WriteString("\t\t\t\t\treturn nil, repo.ErrNotFound\n")
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t\twantErr: true,\n")
	sb.WriteString("\t\t},\n")

	// Repo error case
	sb.WriteString("\t\t{\n")
	sb.WriteString("\t\t\tname: \"repo error\",\n")
	sb.WriteString("\t\t\tid:   1,\n")
	sb.WriteString(fmt.Sprintf("\t\t\tmockFn: func(m *mock%sRepo) {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\t\t\tm.getByIDFn = func(ctx context.Context, id uint) (*entity.%s, error) {\n", entityName))
	sb.WriteString("\t\t\t\t\treturn nil, errors.New(\"db error\")\n")
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t\twantErr: true,\n")
	sb.WriteString("\t\t},\n")

	sb.WriteString("\t}\n\n")

	// Test runner
	sb.WriteString("\tfor _, tt := range tests {\n")
	sb.WriteString("\t\tt.Run(tt.name, func(t *testing.T) {\n")
	sb.WriteString("\t\t\tt.Parallel()\n\n")
	sb.WriteString(fmt.Sprintf("\t\t\tmockRepo := &mock%sRepo{}\n", entityName))
	sb.WriteString("\t\t\ttt.mockFn(mockRepo)\n")
	sb.WriteString(fmt.Sprintf("\t\t\tuc := %s.New(mockRepo)\n\n", pkgName))
	sb.WriteString("\t\t\tresult, err := uc.GetByID(context.Background(), tt.id)\n")
	sb.WriteString("\t\t\tif tt.wantErr {\n")
	sb.WriteString("\t\t\t\trequire.Error(t, err)\n")
	sb.WriteString("\t\t\t\treturn\n")
	sb.WriteString("\t\t\t}\n")
	sb.WriteString("\t\t\trequire.NoError(t, err)\n")
	sb.WriteString("\t\t\trequire.NotNil(t, result)\n")
	sb.WriteString("\t\t})\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// buildUseCaseListTestContent builds the list_test.go file content.
func (g *Generator) buildUseCaseListTestContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()
	dtoAlias := pkgName + "dto"

	sb.WriteString(fmt.Sprintf("package %s_test\n\n", pkgName))

	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"errors\"\n")
	sb.WriteString("\t\"testing\"\n\n")
	sb.WriteString("\t\"github.com/stretchr/testify/require\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/dto/"+pkgName))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/entity"))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/usecase/"+pkgName))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/pkg/pagination"))
	sb.WriteString(")\n\n")

	sb.WriteString("func TestList(t *testing.T) {\n")
	sb.WriteString("\tt.Parallel()\n\n")

	sb.WriteString("\ttests := []struct {\n")
	sb.WriteString("\t\tname    string\n")
	sb.WriteString(fmt.Sprintf("\t\treq     %s.ListRequest\n", dtoAlias))
	sb.WriteString(fmt.Sprintf("\t\tmockFn  func(*mock%sRepo)\n", entityName))
	sb.WriteString("\t\twantErr bool\n")
	sb.WriteString("\t}{\n")

	// Success case
	sb.WriteString("\t\t{\n")
	sb.WriteString("\t\t\tname: \"success\",\n")
	sb.WriteString(fmt.Sprintf("\t\t\treq:  %s.ListRequest{Params: pagination.NewParams()},\n", dtoAlias))
	sb.WriteString(fmt.Sprintf("\t\t\tmockFn: func(m *mock%sRepo) {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\t\t\tm.listFn = func(ctx context.Context, params pagination.Params) ([]*entity.%s, int64, error) {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\t\t\t\treturn []*entity.%s{{}}, 1, nil\n", entityName))
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t},\n")

	// Repo error case
	sb.WriteString("\t\t{\n")
	sb.WriteString("\t\t\tname: \"repo error\",\n")
	sb.WriteString(fmt.Sprintf("\t\t\treq:  %s.ListRequest{Params: pagination.NewParams()},\n", dtoAlias))
	sb.WriteString(fmt.Sprintf("\t\t\tmockFn: func(m *mock%sRepo) {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\t\t\tm.listFn = func(ctx context.Context, params pagination.Params) ([]*entity.%s, int64, error) {\n", entityName))
	sb.WriteString("\t\t\t\t\treturn nil, 0, errors.New(\"db error\")\n")
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t\twantErr: true,\n")
	sb.WriteString("\t\t},\n")

	sb.WriteString("\t}\n\n")

	// Test runner
	sb.WriteString("\tfor _, tt := range tests {\n")
	sb.WriteString("\t\tt.Run(tt.name, func(t *testing.T) {\n")
	sb.WriteString("\t\t\tt.Parallel()\n\n")
	sb.WriteString(fmt.Sprintf("\t\t\tmockRepo := &mock%sRepo{}\n", entityName))
	sb.WriteString("\t\t\ttt.mockFn(mockRepo)\n")
	sb.WriteString(fmt.Sprintf("\t\t\tuc := %s.New(mockRepo)\n\n", pkgName))
	sb.WriteString("\t\t\tresult, err := uc.List(context.Background(), tt.req)\n")
	sb.WriteString("\t\t\tif tt.wantErr {\n")
	sb.WriteString("\t\t\t\trequire.Error(t, err)\n")
	sb.WriteString("\t\t\t\treturn\n")
	sb.WriteString("\t\t\t}\n")
	sb.WriteString("\t\t\trequire.NoError(t, err)\n")
	sb.WriteString("\t\t\trequire.NotNil(t, result)\n")
	sb.WriteString("\t\t})\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// buildUseCaseUpdateTestContent builds the update_test.go file content.
func (g *Generator) buildUseCaseUpdateTestContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()
	dtoAlias := pkgName + "dto"

	sb.WriteString(fmt.Sprintf("package %s_test\n\n", pkgName))

	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"errors\"\n")
	sb.WriteString("\t\"testing\"\n\n")
	sb.WriteString("\t\"github.com/stretchr/testify/require\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/dto/"+pkgName))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/entity"))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/repo"))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/usecase/"+pkgName))
	sb.WriteString(")\n\n")

	sb.WriteString("func TestUpdate(t *testing.T) {\n")
	sb.WriteString("\tt.Parallel()\n\n")

	sb.WriteString("\ttests := []struct {\n")
	sb.WriteString("\t\tname    string\n")
	sb.WriteString("\t\tid      uint\n")
	sb.WriteString(fmt.Sprintf("\t\treq     %s.UpdateRequest\n", dtoAlias))
	sb.WriteString(fmt.Sprintf("\t\tmockFn  func(*mock%sRepo)\n", entityName))
	sb.WriteString("\t\twantErr bool\n")
	sb.WriteString("\t}{\n")

	// Success case
	sb.WriteString("\t\t{\n")
	sb.WriteString("\t\t\tname: \"success\",\n")
	sb.WriteString("\t\t\tid:   1,\n")
	sb.WriteString(fmt.Sprintf("\t\t\treq:  %s.UpdateRequest{},\n", dtoAlias))
	sb.WriteString(fmt.Sprintf("\t\t\tmockFn: func(m *mock%sRepo) {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\t\t\tm.getByIDFn = func(ctx context.Context, id uint) (*entity.%s, error) {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\t\t\t\treturn &entity.%s{}, nil\n", entityName))
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString(fmt.Sprintf("\t\t\t\tm.updateFn = func(ctx context.Context, e *entity.%s) error {\n", entityName))
	sb.WriteString("\t\t\t\t\treturn nil\n")
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t},\n")

	// Not found case
	sb.WriteString("\t\t{\n")
	sb.WriteString("\t\t\tname: \"not found\",\n")
	sb.WriteString("\t\t\tid:   999,\n")
	sb.WriteString(fmt.Sprintf("\t\t\treq:  %s.UpdateRequest{},\n", dtoAlias))
	sb.WriteString(fmt.Sprintf("\t\t\tmockFn: func(m *mock%sRepo) {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\t\t\tm.getByIDFn = func(ctx context.Context, id uint) (*entity.%s, error) {\n", entityName))
	sb.WriteString("\t\t\t\t\treturn nil, repo.ErrNotFound\n")
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t\twantErr: true,\n")
	sb.WriteString("\t\t},\n")

	// Repo error case
	sb.WriteString("\t\t{\n")
	sb.WriteString("\t\t\tname: \"repo error\",\n")
	sb.WriteString("\t\t\tid:   1,\n")
	sb.WriteString(fmt.Sprintf("\t\t\treq:  %s.UpdateRequest{},\n", dtoAlias))
	sb.WriteString(fmt.Sprintf("\t\t\tmockFn: func(m *mock%sRepo) {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\t\t\tm.getByIDFn = func(ctx context.Context, id uint) (*entity.%s, error) {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\t\t\t\treturn &entity.%s{}, nil\n", entityName))
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString(fmt.Sprintf("\t\t\t\tm.updateFn = func(ctx context.Context, e *entity.%s) error {\n", entityName))
	sb.WriteString("\t\t\t\t\treturn errors.New(\"db error\")\n")
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t\twantErr: true,\n")
	sb.WriteString("\t\t},\n")

	sb.WriteString("\t}\n\n")

	// Test runner
	sb.WriteString("\tfor _, tt := range tests {\n")
	sb.WriteString("\t\tt.Run(tt.name, func(t *testing.T) {\n")
	sb.WriteString("\t\t\tt.Parallel()\n\n")
	sb.WriteString(fmt.Sprintf("\t\t\tmockRepo := &mock%sRepo{}\n", entityName))
	sb.WriteString("\t\t\ttt.mockFn(mockRepo)\n")
	sb.WriteString(fmt.Sprintf("\t\t\tuc := %s.New(mockRepo)\n\n", pkgName))
	sb.WriteString("\t\t\tresult, err := uc.Update(context.Background(), tt.id, tt.req)\n")
	sb.WriteString("\t\t\tif tt.wantErr {\n")
	sb.WriteString("\t\t\t\trequire.Error(t, err)\n")
	sb.WriteString("\t\t\t\treturn\n")
	sb.WriteString("\t\t\t}\n")
	sb.WriteString("\t\t\trequire.NoError(t, err)\n")
	sb.WriteString("\t\t\trequire.NotNil(t, result)\n")
	sb.WriteString("\t\t})\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// buildUseCaseDeleteTestContent builds the delete_test.go file content.
func (g *Generator) buildUseCaseDeleteTestContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()

	sb.WriteString(fmt.Sprintf("package %s_test\n\n", pkgName))

	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"errors\"\n")
	sb.WriteString("\t\"testing\"\n\n")
	sb.WriteString("\t\"github.com/stretchr/testify/require\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/repo"))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/usecase/"+pkgName))
	sb.WriteString(")\n\n")

	sb.WriteString("func TestDelete(t *testing.T) {\n")
	sb.WriteString("\tt.Parallel()\n\n")

	sb.WriteString("\ttests := []struct {\n")
	sb.WriteString("\t\tname    string\n")
	sb.WriteString("\t\tid      uint\n")
	sb.WriteString(fmt.Sprintf("\t\tmockFn  func(*mock%sRepo)\n", entityName))
	sb.WriteString("\t\twantErr bool\n")
	sb.WriteString("\t}{\n")

	// Success case
	sb.WriteString("\t\t{\n")
	sb.WriteString("\t\t\tname: \"success\",\n")
	sb.WriteString("\t\t\tid:   1,\n")
	sb.WriteString(fmt.Sprintf("\t\t\tmockFn: func(m *mock%sRepo) {\n", entityName))
	sb.WriteString("\t\t\t\tm.deleteFn = func(ctx context.Context, id uint) error {\n")
	sb.WriteString("\t\t\t\t\treturn nil\n")
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t},\n")

	// Not found case
	sb.WriteString("\t\t{\n")
	sb.WriteString("\t\t\tname: \"not found\",\n")
	sb.WriteString("\t\t\tid:   999,\n")
	sb.WriteString(fmt.Sprintf("\t\t\tmockFn: func(m *mock%sRepo) {\n", entityName))
	sb.WriteString("\t\t\t\tm.deleteFn = func(ctx context.Context, id uint) error {\n")
	sb.WriteString("\t\t\t\t\treturn repo.ErrNotFound\n")
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t\twantErr: true,\n")
	sb.WriteString("\t\t},\n")

	// Repo error case
	sb.WriteString("\t\t{\n")
	sb.WriteString("\t\t\tname: \"repo error\",\n")
	sb.WriteString("\t\t\tid:   1,\n")
	sb.WriteString(fmt.Sprintf("\t\t\tmockFn: func(m *mock%sRepo) {\n", entityName))
	sb.WriteString("\t\t\t\tm.deleteFn = func(ctx context.Context, id uint) error {\n")
	sb.WriteString("\t\t\t\t\treturn errors.New(\"db error\")\n")
	sb.WriteString("\t\t\t\t}\n")
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t\twantErr: true,\n")
	sb.WriteString("\t\t},\n")

	sb.WriteString("\t}\n\n")

	// Test runner
	sb.WriteString("\tfor _, tt := range tests {\n")
	sb.WriteString("\t\tt.Run(tt.name, func(t *testing.T) {\n")
	sb.WriteString("\t\t\tt.Parallel()\n\n")
	sb.WriteString(fmt.Sprintf("\t\t\tmockRepo := &mock%sRepo{}\n", entityName))
	sb.WriteString("\t\t\ttt.mockFn(mockRepo)\n")
	sb.WriteString(fmt.Sprintf("\t\t\tuc := %s.New(mockRepo)\n\n", pkgName))
	sb.WriteString("\t\t\terr := uc.Delete(context.Background(), tt.id)\n")
	sb.WriteString("\t\t\tif tt.wantErr {\n")
	sb.WriteString("\t\t\t\trequire.Error(t, err)\n")
	sb.WriteString("\t\t\t\treturn\n")
	sb.WriteString("\t\t\t}\n")
	sb.WriteString("\t\t\trequire.NoError(t, err)\n")
	sb.WriteString("\t\t})\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// buildUseCaseGenericTestContent builds a generic test stub for unknown methods.
func (g *Generator) buildUseCaseGenericTestContent(methodName string) string {
	var sb strings.Builder

	pkgName := g.packageName()

	sb.WriteString(fmt.Sprintf("package %s_test\n\n", pkgName))

	sb.WriteString("import (\n")
	sb.WriteString("\t\"testing\"\n")
	sb.WriteString(")\n\n")

	sb.WriteString(fmt.Sprintf("func Test%s(t *testing.T) {\n", methodName))
	sb.WriteString("\tt.Parallel()\n")
	sb.WriteString("\t// TODO: Implement test\n")
	sb.WriteString("}\n")

	return sb.String()
}

// buildUseCaseMocksTestContent builds the mocks_test.go file with function-based mocks.
func (g *Generator) buildUseCaseMocksTestContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()
	varName := g.varName()

	sb.WriteString(fmt.Sprintf("package %s_test\n\n", pkgName))

	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/entity"))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/pkg/pagination"))
	sb.WriteString(")\n\n")

	// Mock struct
	sb.WriteString(fmt.Sprintf("type mock%sRepo struct {\n", entityName))
	sb.WriteString(fmt.Sprintf("\tcreateFn  func(ctx context.Context, %s *entity.%s) error\n", varName, entityName))
	sb.WriteString(fmt.Sprintf("\tgetByIDFn func(ctx context.Context, id uint) (*entity.%s, error)\n", entityName))
	sb.WriteString(fmt.Sprintf("\tlistFn    func(ctx context.Context, params pagination.Params) ([]*entity.%s, int64, error)\n", entityName))
	sb.WriteString(fmt.Sprintf("\tupdateFn  func(ctx context.Context, %s *entity.%s) error\n", varName, entityName))
	sb.WriteString("\tdeleteFn  func(ctx context.Context, id uint) error\n")
	sb.WriteString("}\n\n")

	// Create method
	sb.WriteString(fmt.Sprintf("func (m *mock%sRepo) Create(ctx context.Context, %s *entity.%s) error {\n", entityName, varName, entityName))
	sb.WriteString(fmt.Sprintf("\treturn m.createFn(ctx, %s)\n", varName))
	sb.WriteString("}\n\n")

	// GetByID method
	sb.WriteString(fmt.Sprintf("func (m *mock%sRepo) GetByID(ctx context.Context, id uint) (*entity.%s, error) {\n", entityName, entityName))
	sb.WriteString("\treturn m.getByIDFn(ctx, id)\n")
	sb.WriteString("}\n\n")

	// List method
	sb.WriteString(fmt.Sprintf("func (m *mock%sRepo) List(ctx context.Context, params pagination.Params) ([]*entity.%s, int64, error) {\n", entityName, entityName))
	sb.WriteString("\treturn m.listFn(ctx, params)\n")
	sb.WriteString("}\n\n")

	// Update method
	sb.WriteString(fmt.Sprintf("func (m *mock%sRepo) Update(ctx context.Context, %s *entity.%s) error {\n", entityName, varName, entityName))
	sb.WriteString(fmt.Sprintf("\treturn m.updateFn(ctx, %s)\n", varName))
	sb.WriteString("}\n\n")

	// Delete method
	sb.WriteString(fmt.Sprintf("func (m *mock%sRepo) Delete(ctx context.Context, id uint) error {\n", entityName))
	sb.WriteString("\treturn m.deleteFn(ctx, id)\n")
	sb.WriteString("}\n")

	return sb.String()
}
