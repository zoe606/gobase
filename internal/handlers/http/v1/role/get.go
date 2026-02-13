package role

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	roleusecase "go-boilerplate/internal/usecase/role"
	"go-boilerplate/pkg/response"
)

// GetByID godoc
// @Summary     Get role
// @Description Get a role by ID
// @ID          role-get
// @Tags        roles
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Role ID"
// @Success     200 {object} response.Response[roledto.Response]
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /roles/{id} [get]
func (h *Handler) GetByID(ctx *fiber.Ctx) error {
	id, err := parseUint(ctx.Params("id"))
	if err != nil {
		return response.BadRequest(ctx, "INVALID_ID", "Invalid role ID")
	}

	result, err := h.roleUC.GetByID(ctx.UserContext(), id)
	if err != nil {
		if errors.Is(err, roleusecase.ErrRoleNotFound) {
			return response.NotFound(ctx, "Role not found")
		}
		h.l.Error(err, "handlers - http - v1 - role - GetByID")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
