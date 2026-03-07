package generator

import (
	"fmt"
	"strings"
)

// GenerateHandler generates the HTTP handler package files.
func (g *Generator) GenerateHandler() error {
	pkgName := g.packageName()
	basePath := fmt.Sprintf("internal/handlers/http/v1/%s", pkgName)

	// Generate main handler file
	mainContent := g.buildHandlerMainContent()
	if err := g.writeFile(basePath+"/handler.go", mainContent); err != nil {
		return err
	}

	// Generate CRUD handler method files
	methods := []struct {
		name     string
		fileName string
		content  func() string
	}{
		{"Create", "create.go", g.buildHandlerCreateContent},
		{"GetByID", "get_by_id.go", g.buildHandlerGetByIDContent},
		{"List", "list.go", g.buildHandlerListContent},
		{"Update", "update.go", g.buildHandlerUpdateContent},
		{"Delete", "delete.go", g.buildHandlerDeleteContent},
	}

	for _, method := range methods {
		content := method.content()
		if err := g.writeFile(basePath+"/"+method.fileName, content); err != nil {
			return err
		}
	}

	return nil
}

// buildHandlerMainContent builds the main handler.go file content.
func (g *Generator) buildHandlerMainContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()

	// Package declaration
	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"github.com/go-playground/validator/v10\"\n")
	sb.WriteString("\t\"github.com/gofiber/fiber/v2\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/usecase"))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/pkg/logger"))
	sb.WriteString(")\n\n")

	// Handler struct
	sb.WriteString(fmt.Sprintf("// Handler handles %s endpoints.\n", pkgName))
	sb.WriteString("type Handler struct {\n")
	sb.WriteString(fmt.Sprintf("\t%sUC usecase.%s\n", g.varName(), entityName))
	sb.WriteString("\tl   logger.Interface\n")
	sb.WriteString("\tv   *validator.Validate\n")
	sb.WriteString("}\n\n")

	// Constructor
	sb.WriteString(fmt.Sprintf("// New creates a new %s handler.\n", pkgName))
	sb.WriteString(fmt.Sprintf("func New(%sUC usecase.%s, l logger.Interface) *Handler {\n", g.varName(), entityName))
	sb.WriteString("\treturn &Handler{\n")
	sb.WriteString(fmt.Sprintf("\t\t%sUC: %sUC,\n", g.varName(), g.varName()))
	sb.WriteString("\t\tl:   l,\n")
	sb.WriteString("\t\tv:   validator.New(validator.WithRequiredStructEnabled()),\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	// RegisterRoutes
	sb.WriteString("// RegisterRoutes sets up routes.\n")
	sb.WriteString("func (h *Handler) RegisterRoutes(router fiber.Router) {\n")
	sb.WriteString(fmt.Sprintf("\t%ss := router.Group(\"/%ss\")\n", g.varName(), g.varName()))
	sb.WriteString(fmt.Sprintf("\t// TODO: Add auth middleware: %ss.Use(middleware.JWT(jwtService))\n", g.varName()))
	sb.WriteString("\t{\n")
	sb.WriteString(fmt.Sprintf("\t\t%ss.Post(\"/\", h.Create)\n", g.varName()))
	sb.WriteString(fmt.Sprintf("\t\t%ss.Get(\"/\", h.List)\n", g.varName()))
	sb.WriteString(fmt.Sprintf("\t\t%ss.Get(\"/:id\", h.GetByID)\n", g.varName()))
	sb.WriteString(fmt.Sprintf("\t\t%ss.Put(\"/:id\", h.Update)\n", g.varName()))
	sb.WriteString(fmt.Sprintf("\t\t%ss.Delete(\"/:id\", h.Delete)\n", g.varName()))
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// buildHandlerCreateContent builds the create.go handler file content.
func (g *Generator) buildHandlerCreateContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()
	dtoAlias := pkgName + "dto"
	ucAlias := pkgName + "uc"

	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"github.com/gofiber/fiber/v2\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/dto/"+pkgName))
	sb.WriteString(fmt.Sprintf("\tv1 %q\n", g.config.ModuleName+"/internal/handlers/http/v1"))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/pkg/response"))
	sb.WriteString(")\n\n")

	// Swagger annotation
	sb.WriteString("// Create godoc\n")
	sb.WriteString(fmt.Sprintf("// @Summary     Create %s\n", pkgName))
	sb.WriteString(fmt.Sprintf("// @Description Create a new %s\n", pkgName))
	sb.WriteString(fmt.Sprintf("// @ID          %s-create\n", pkgName))
	sb.WriteString(fmt.Sprintf("// @Tags        %ss\n", pkgName))
	sb.WriteString("// @Accept      json\n")
	sb.WriteString("// @Produce     json\n")
	sb.WriteString(fmt.Sprintf("// @Param       request body %s.CreateRequest true \"Create %s request\"\n", dtoAlias, entityName))
	sb.WriteString(fmt.Sprintf("// @Success     201 {object} response.Response[%s.Response]\n", dtoAlias))
	sb.WriteString("// @Failure     400 {object} response.ErrorResponse\n")
	sb.WriteString("// @Failure     500 {object} response.ErrorResponse\n")
	sb.WriteString(fmt.Sprintf("// @Router      /%ss [post]\n", g.varName()))

	// Method
	sb.WriteString("func (h *Handler) Create(ctx *fiber.Ctx) error {\n")
	sb.WriteString(fmt.Sprintf("\tvar req %s.CreateRequest\n", dtoAlias))
	sb.WriteString("\tif err := ctx.BodyParser(&req); err != nil {\n")
	sb.WriteString("\t\treturn response.BadRequest(ctx, \"INVALID_JSON\", \"Invalid request body\")\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\tif err := h.v.Struct(req); err != nil {\n")
	sb.WriteString("\t\treturn response.ValidationError(ctx, v1.ParseValidationErrors(err))\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString(fmt.Sprintf("\tresult, err := h.%sUC.Create(ctx.UserContext(), req)\n", g.varName()))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString(fmt.Sprintf("\t\th.l.Error(err, \"handlers - http - v1 - %s - Create\")\n", pkgName))
	sb.WriteString("\t\treturn response.InternalError(ctx)\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\treturn response.Created(ctx, result)\n")
	sb.WriteString("}\n")

	// Mark ucAlias as used (for future error handling)
	_ = ucAlias

	return sb.String()
}

// buildHandlerGetByIDContent builds the get_by_id.go handler file content.
func (g *Generator) buildHandlerGetByIDContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()
	dtoAlias := pkgName + "dto"
	ucAlias := pkgName + "uc"

	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"errors\"\n")
	sb.WriteString("\t\"strconv\"\n\n")
	sb.WriteString("\t\"github.com/gofiber/fiber/v2\"\n\n")
	sb.WriteString(fmt.Sprintf("\t_ %q // swagger type resolution\n", g.config.ModuleName+"/internal/dto/"+pkgName))
	sb.WriteString(fmt.Sprintf("\t%s %q\n", ucAlias, g.config.ModuleName+"/internal/usecase/"+pkgName))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/pkg/response"))
	sb.WriteString(")\n\n")

	// Swagger annotation
	sb.WriteString("// GetByID godoc\n")
	sb.WriteString(fmt.Sprintf("// @Summary     Get %s by ID\n", pkgName))
	sb.WriteString(fmt.Sprintf("// @Description Get a %s by its ID\n", pkgName))
	sb.WriteString(fmt.Sprintf("// @ID          %s-get-by-id\n", pkgName))
	sb.WriteString(fmt.Sprintf("// @Tags        %ss\n", pkgName))
	sb.WriteString("// @Accept      json\n")
	sb.WriteString("// @Produce     json\n")
	sb.WriteString(fmt.Sprintf("// @Param       id path int true \"%s ID\"\n", entityName))
	sb.WriteString(fmt.Sprintf("// @Success     200 {object} response.Response[%s.Response]\n", dtoAlias))
	sb.WriteString("// @Failure     404 {object} response.ErrorResponse\n")
	sb.WriteString("// @Failure     500 {object} response.ErrorResponse\n")
	sb.WriteString(fmt.Sprintf("// @Router      /%ss/{id} [get]\n", g.varName()))

	// Method
	sb.WriteString("func (h *Handler) GetByID(ctx *fiber.Ctx) error {\n")
	sb.WriteString("\tid, err := strconv.ParseUint(ctx.Params(\"id\"), 10, 32)\n")
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString(fmt.Sprintf("\t\treturn response.BadRequest(ctx, \"INVALID_ID\", \"Invalid %s ID\")\n", pkgName))
	sb.WriteString("\t}\n\n")
	sb.WriteString(fmt.Sprintf("\tresult, err := h.%sUC.GetByID(ctx.UserContext(), uint(id))\n", g.varName()))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString(fmt.Sprintf("\t\tif errors.Is(err, %s.ErrNotFound) {\n", ucAlias))
	sb.WriteString(fmt.Sprintf("\t\t\treturn response.NotFound(ctx, \"%s not found\")\n", entityName))
	sb.WriteString("\t\t}\n")
	sb.WriteString(fmt.Sprintf("\t\th.l.Error(err, \"handlers - http - v1 - %s - GetByID\")\n", pkgName))
	sb.WriteString("\t\treturn response.InternalError(ctx)\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\treturn response.OK(ctx, result)\n")
	sb.WriteString("}\n")

	return sb.String()
}

// buildHandlerListContent builds the list.go handler file content.
func (g *Generator) buildHandlerListContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	dtoAlias := pkgName + "dto"

	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"github.com/gofiber/fiber/v2\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/dto/"+pkgName))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/pkg/response"))
	sb.WriteString(")\n\n")

	// Swagger annotation
	sb.WriteString("// List godoc\n")
	sb.WriteString(fmt.Sprintf("// @Summary     List %ss\n", pkgName))
	sb.WriteString(fmt.Sprintf("// @Description Get a paginated list of %ss\n", pkgName))
	sb.WriteString(fmt.Sprintf("// @ID          %s-list\n", pkgName))
	sb.WriteString(fmt.Sprintf("// @Tags        %ss\n", pkgName))
	sb.WriteString("// @Accept      json\n")
	sb.WriteString("// @Produce     json\n")
	sb.WriteString("// @Param       page query int false \"Page number\" default(1)\n")
	sb.WriteString("// @Param       page_size query int false \"Page size\" default(20)\n")
	sb.WriteString(fmt.Sprintf("// @Success     200 {object} response.Response[%s.ListResponse]\n", dtoAlias))
	sb.WriteString("// @Failure     500 {object} response.ErrorResponse\n")
	sb.WriteString(fmt.Sprintf("// @Router      /%ss [get]\n", g.varName()))

	// Method
	sb.WriteString("func (h *Handler) List(ctx *fiber.Ctx) error {\n")
	sb.WriteString(fmt.Sprintf("\tvar req %s.ListRequest\n", dtoAlias))
	sb.WriteString("\tif err := ctx.QueryParser(&req); err != nil {\n")
	sb.WriteString("\t\treturn response.BadRequest(ctx, \"INVALID_QUERY\", \"Invalid query parameters\")\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString(fmt.Sprintf("\tresult, err := h.%sUC.List(ctx.UserContext(), req)\n", g.varName()))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString(fmt.Sprintf("\t\th.l.Error(err, \"handlers - http - v1 - %s - List\")\n", pkgName))
	sb.WriteString("\t\treturn response.InternalError(ctx)\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\treturn response.OK(ctx, result)\n")
	sb.WriteString("}\n")

	return sb.String()
}

// buildHandlerUpdateContent builds the update.go handler file content.
func (g *Generator) buildHandlerUpdateContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()
	dtoAlias := pkgName + "dto"
	ucAlias := pkgName + "uc"

	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"errors\"\n")
	sb.WriteString("\t\"strconv\"\n\n")
	sb.WriteString("\t\"github.com/gofiber/fiber/v2\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/internal/dto/"+pkgName))
	sb.WriteString(fmt.Sprintf("\tv1 %q\n", g.config.ModuleName+"/internal/handlers/http/v1"))
	sb.WriteString(fmt.Sprintf("\t%s %q\n", ucAlias, g.config.ModuleName+"/internal/usecase/"+pkgName))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/pkg/response"))
	sb.WriteString(")\n\n")

	// Swagger annotation
	sb.WriteString("// Update godoc\n")
	sb.WriteString(fmt.Sprintf("// @Summary     Update %s\n", pkgName))
	sb.WriteString(fmt.Sprintf("// @Description Update an existing %s\n", pkgName))
	sb.WriteString(fmt.Sprintf("// @ID          %s-update\n", pkgName))
	sb.WriteString(fmt.Sprintf("// @Tags        %ss\n", pkgName))
	sb.WriteString("// @Accept      json\n")
	sb.WriteString("// @Produce     json\n")
	sb.WriteString(fmt.Sprintf("// @Param       id path int true \"%s ID\"\n", entityName))
	sb.WriteString(fmt.Sprintf("// @Param       request body %s.UpdateRequest true \"Update %s request\"\n", dtoAlias, entityName))
	sb.WriteString(fmt.Sprintf("// @Success     200 {object} response.Response[%s.Response]\n", dtoAlias))
	sb.WriteString("// @Failure     400 {object} response.ErrorResponse\n")
	sb.WriteString("// @Failure     404 {object} response.ErrorResponse\n")
	sb.WriteString("// @Failure     500 {object} response.ErrorResponse\n")
	sb.WriteString(fmt.Sprintf("// @Router      /%ss/{id} [put]\n", g.varName()))

	// Method
	sb.WriteString("func (h *Handler) Update(ctx *fiber.Ctx) error {\n")
	sb.WriteString("\tid, err := strconv.ParseUint(ctx.Params(\"id\"), 10, 32)\n")
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString(fmt.Sprintf("\t\treturn response.BadRequest(ctx, \"INVALID_ID\", \"Invalid %s ID\")\n", pkgName))
	sb.WriteString("\t}\n\n")
	sb.WriteString(fmt.Sprintf("\tvar req %s.UpdateRequest\n", dtoAlias))
	sb.WriteString("\tif err := ctx.BodyParser(&req); err != nil {\n")
	sb.WriteString("\t\treturn response.BadRequest(ctx, \"INVALID_JSON\", \"Invalid request body\")\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\tif err := h.v.Struct(req); err != nil {\n")
	sb.WriteString("\t\treturn response.ValidationError(ctx, v1.ParseValidationErrors(err))\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString(fmt.Sprintf("\tresult, err := h.%sUC.Update(ctx.UserContext(), uint(id), req)\n", g.varName()))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString(fmt.Sprintf("\t\tif errors.Is(err, %s.ErrNotFound) {\n", ucAlias))
	sb.WriteString(fmt.Sprintf("\t\t\treturn response.NotFound(ctx, \"%s not found\")\n", entityName))
	sb.WriteString("\t\t}\n")
	sb.WriteString(fmt.Sprintf("\t\th.l.Error(err, \"handlers - http - v1 - %s - Update\")\n", pkgName))
	sb.WriteString("\t\treturn response.InternalError(ctx)\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\treturn response.OK(ctx, result)\n")
	sb.WriteString("}\n")

	return sb.String()
}

// buildHandlerDeleteContent builds the delete.go handler file content.
func (g *Generator) buildHandlerDeleteContent() string {
	var sb strings.Builder

	pkgName := g.packageName()
	entityName := g.entityName()
	ucAlias := pkgName + "uc"

	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Imports
	sb.WriteString("import (\n")
	sb.WriteString("\t\"errors\"\n")
	sb.WriteString("\t\"strconv\"\n\n")
	sb.WriteString("\t\"github.com/gofiber/fiber/v2\"\n\n")
	sb.WriteString(fmt.Sprintf("\t%s %q\n", ucAlias, g.config.ModuleName+"/internal/usecase/"+pkgName))
	sb.WriteString(fmt.Sprintf("\t%q\n", g.config.ModuleName+"/pkg/response"))
	sb.WriteString(")\n\n")

	// Swagger annotation
	sb.WriteString("// Delete godoc\n")
	sb.WriteString(fmt.Sprintf("// @Summary     Delete %s\n", pkgName))
	sb.WriteString(fmt.Sprintf("// @Description Delete a %s by ID\n", pkgName))
	sb.WriteString(fmt.Sprintf("// @ID          %s-delete\n", pkgName))
	sb.WriteString(fmt.Sprintf("// @Tags        %ss\n", pkgName))
	sb.WriteString("// @Accept      json\n")
	sb.WriteString("// @Produce     json\n")
	sb.WriteString(fmt.Sprintf("// @Param       id path int true \"%s ID\"\n", entityName))
	sb.WriteString("// @Success     204 \"No Content\"\n")
	sb.WriteString("// @Failure     404 {object} response.ErrorResponse\n")
	sb.WriteString("// @Failure     500 {object} response.ErrorResponse\n")
	sb.WriteString(fmt.Sprintf("// @Router      /%ss/{id} [delete]\n", g.varName()))

	// Method
	sb.WriteString("func (h *Handler) Delete(ctx *fiber.Ctx) error {\n")
	sb.WriteString("\tid, err := strconv.ParseUint(ctx.Params(\"id\"), 10, 32)\n")
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString(fmt.Sprintf("\t\treturn response.BadRequest(ctx, \"INVALID_ID\", \"Invalid %s ID\")\n", pkgName))
	sb.WriteString("\t}\n\n")
	sb.WriteString(fmt.Sprintf("\tif err := h.%sUC.Delete(ctx.UserContext(), uint(id)); err != nil {\n", g.varName()))
	sb.WriteString(fmt.Sprintf("\t\tif errors.Is(err, %s.ErrNotFound) {\n", ucAlias))
	sb.WriteString(fmt.Sprintf("\t\t\treturn response.NotFound(ctx, \"%s not found\")\n", entityName))
	sb.WriteString("\t\t}\n")
	sb.WriteString(fmt.Sprintf("\t\th.l.Error(err, \"handlers - http - v1 - %s - Delete\")\n", pkgName))
	sb.WriteString("\t\treturn response.InternalError(ctx)\n")
	sb.WriteString("\t}\n\n")
	sb.WriteString("\treturn response.NoContent(ctx)\n")
	sb.WriteString("}\n")

	return sb.String()
}
