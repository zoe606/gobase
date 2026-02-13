package role

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	roleusecase "go-boilerplate/internal/usecase/role"
	"go-boilerplate/pkg/response"
)

// Delete godoc
// @Summary     Delete role
// @Description Delete a role by ID
// @ID          role-delete
// @Tags        roles
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Role ID"
// @Success     204 "No Content"
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /roles/{id} [delete]
func (h *Handler) Delete(ctx *fiber.Ctx) error {
	id, err := parseUint(ctx.Params("id"))
	if err != nil {
		return response.BadRequest(ctx, "INVALID_ID", "Invalid role ID")
	}

	err = h.roleUC.Delete(ctx.UserContext(), id)
	if err != nil {
		if errors.Is(err, roleusecase.ErrRoleNotFound) {
			return response.NotFound(ctx, "Role not found")
		}
		h.l.Error(err, "handlers - http - v1 - role - Delete")
		return response.InternalError(ctx)
	}

	return response.NoContent(ctx)
}
