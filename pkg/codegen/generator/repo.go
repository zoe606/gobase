package generator

import (
	"fmt"
	"strings"
)

// GenerateRepository generates the repository interface and implementation.
func (g *Generator) GenerateRepository() error {
	// Generate interface addition to contracts.go
	interfaceContent := g.buildRepoInterfaceContent()
	contractsPath := "internal/repo/contracts.go"

	// Try to append to existing file
	err := g.appendToFile(contractsPath, interfaceContent, "")
	if err != nil {
		// If file doesn't exist or can't be modified, print instruction
		if g.config.DryRun {
			fmt.Printf("\n=== Add to %s ===\n", contractsPath)
			fmt.Println(interfaceContent)
		} else {
			fmt.Printf("Please add the following to %s:\n%s\n", contractsPath, interfaceContent)
		}
	}

	// Generate PostgreSQL implementation
	implContent := g.buildRepoImplContent()
	implPath := fmt.Sprintf("internal/repo/persistent/%s_postgres.go", g.packageName())
	if err := g.writeFile(implPath, implContent); err != nil {
		return err
	}

	return nil
}

// buildRepoInterfaceContent builds the repository interface content.
func (g *Generator) buildRepoInterfaceContent() string {
	var sb strings.Builder

	entityName := g.entityName()
	varName := g.varName()

	sb.WriteString(fmt.Sprintf("\n\t// %sRepo defines %s repository operations.\n", entityName, entityName))
	sb.WriteString(fmt.Sprintf("\t%sRepo interface {\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\tCreate(ctx context.Context, %s *entity.%s) error\n", varName, entityName))
	sb.WriteString(fmt.Sprintf("\t\tGetByID(ctx context.Context, id uint) (*entity.%s, error)\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\tList(ctx context.Context, limit, offset int) ([]*entity.%s, int64, error)\n", entityName))
	sb.WriteString(fmt.Sprintf("\t\tUpdate(ctx context.Context, %s *entity.%s) error\n", varName, entityName))
	sb.WriteString("\t\tDelete(ctx context.Context, id uint) error\n")
	sb.WriteString("\t}\n")

	return sb.String()
}

// buildRepoImplContent builds the PostgreSQL repository implementation.
func (g *Generator) buildRepoImplContent() string {
	var sb strings.Builder

	entityName := g.entityName()
	varName := g.varName()
	pkgName := g.packageName()

	// Package declaration
	sb.WriteString("package persistent\n\n")

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"context\"\n")
	sb.WriteString("\t\"errors\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/entity"))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/repo"))
	sb.WriteString("\t\"gorm.io/gorm\"\n")
	sb.WriteString(")\n\n")

	// Struct
	sb.WriteString(fmt.Sprintf("// %sPostgres implements repo.%sRepo using PostgreSQL.\n", entityName, entityName))
	sb.WriteString(fmt.Sprintf("type %sPostgres struct {\n", entityName))
	sb.WriteString("\tdb *gorm.DB\n")
	sb.WriteString("}\n\n")

	// Constructor
	sb.WriteString(fmt.Sprintf("// New%sPostgres creates a new %s repository.\n", entityName, entityName))
	sb.WriteString(fmt.Sprintf("func New%sPostgres(db *gorm.DB) *%sPostgres {\n", entityName, entityName))
	sb.WriteString(fmt.Sprintf("\treturn &%sPostgres{db: db}\n", entityName))
	sb.WriteString("}\n\n")

	// Create method
	sb.WriteString(fmt.Sprintf("// Create creates a new %s.\n", pkgName))
	sb.WriteString(fmt.Sprintf("func (r *%sPostgres) Create(ctx context.Context, %s *entity.%s) error {\n", entityName, varName, entityName))
	sb.WriteString(fmt.Sprintf("\treturn r.db.WithContext(ctx).Create(%s).Error\n", varName))
	sb.WriteString("}\n\n")

	// GetByID method
	sb.WriteString(fmt.Sprintf("// GetByID retrieves a %s by ID.\n", pkgName))
	sb.WriteString(fmt.Sprintf("func (r *%sPostgres) GetByID(ctx context.Context, id uint) (*entity.%s, error) {\n", entityName, entityName))
	sb.WriteString(fmt.Sprintf("\tvar %s entity.%s\n", varName, entityName))
	sb.WriteString(fmt.Sprintf("\terr := r.db.WithContext(ctx).First(&%s, id).Error\n", varName))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\tif errors.Is(err, gorm.ErrRecordNotFound) {\n")
	sb.WriteString("\t\t\treturn nil, repo.ErrNotFound\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t\treturn nil, err\n")
	sb.WriteString("\t}\n")
	sb.WriteString(fmt.Sprintf("\treturn &%s, nil\n", varName))
	sb.WriteString("}\n\n")

	// List method
	sb.WriteString(fmt.Sprintf("// List retrieves a paginated list of %ss.\n", pkgName))
	sb.WriteString(fmt.Sprintf("func (r *%sPostgres) List(ctx context.Context, limit, offset int) ([]*entity.%s, int64, error) {\n", entityName, entityName))
	sb.WriteString(fmt.Sprintf("\tvar %ss []*entity.%s\n", varName, entityName))
	sb.WriteString("\tvar total int64\n\n")
	sb.WriteString("\t// Count total\n")
	sb.WriteString(fmt.Sprintf("\tif err := r.db.WithContext(ctx).Model(&entity.%s{}).Count(&total).Error; err != nil {\n", entityName))
	sb.WriteString("\t\treturn nil, 0, err\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\t// Fetch paginated results\n")
	sb.WriteString("\terr := r.db.WithContext(ctx).\n")
	sb.WriteString("\t\tLimit(limit).\n")
	sb.WriteString("\t\tOffset(offset).\n")
	sb.WriteString("\t\tOrder(\"id DESC\").\n")
	sb.WriteString(fmt.Sprintf("\t\tFind(&%ss).Error\n", varName))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\treturn nil, 0, err\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString(fmt.Sprintf("\treturn %ss, total, nil\n", varName))
	sb.WriteString("}\n\n")

	// Update method
	sb.WriteString(fmt.Sprintf("// Update updates a %s.\n", pkgName))
	sb.WriteString(fmt.Sprintf("func (r *%sPostgres) Update(ctx context.Context, %s *entity.%s) error {\n", entityName, varName, entityName))
	sb.WriteString(fmt.Sprintf("\tresult := r.db.WithContext(ctx).Save(%s)\n", varName))
	sb.WriteString("\tif result.Error != nil {\n")
	sb.WriteString("\t\treturn result.Error\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\tif result.RowsAffected == 0 {\n")
	sb.WriteString("\t\treturn repo.ErrNotFound\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n\n")

	// Delete method
	sb.WriteString(fmt.Sprintf("// Delete deletes a %s by ID.\n", pkgName))
	sb.WriteString(fmt.Sprintf("func (r *%sPostgres) Delete(ctx context.Context, id uint) error {\n", entityName))
	sb.WriteString(fmt.Sprintf("\tresult := r.db.WithContext(ctx).Delete(&entity.%s{}, id)\n", entityName))
	sb.WriteString("\tif result.Error != nil {\n")
	sb.WriteString("\t\treturn result.Error\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\tif result.RowsAffected == 0 {\n")
	sb.WriteString("\t\treturn repo.ErrNotFound\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n")

	return sb.String()
}
