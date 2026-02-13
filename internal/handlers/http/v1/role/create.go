package role

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	roledto "go-boilerplate/internal/dto/role"
	v1 "go-boilerplate/internal/handlers/http/v1"
	roleusecase "go-boilerplate/internal/usecase/role"
	"go-boilerplate/pkg/response"
)

// Create godoc
// @Summary     Create role
// @Description Create a new role
// @ID          role-create
// @Tags        roles
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body roledto.CreateRequest true "Role data"
// @Success     201 {object} response.Response[roledto.Response]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     409 {object} response.ErrorResponse "Role name already exists"
// @Failure     500 {object} response.ErrorResponse
// @Router      /roles [post]
func (h *Handler) Create(ctx *fiber.Ctx) error {
	var req roledto.CreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(ctx, v1.ParseValidationErrors(err))
	}

	result, err := h.roleUC.Create(ctx.UserContext(), req)
	if err != nil {
		if errors.Is(err, roleusecase.ErrRoleNameExists) {
			return response.Conflict(ctx, "Role name already exists")
		}
		h.l.Error(err, "handlers - http - v1 - role - Create")
		return response.InternalError(ctx)
	}

	return response.Created(ctx, result)
}
