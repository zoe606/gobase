package role

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	roledto "go-boilerplate/internal/dto/role"
	v1 "go-boilerplate/internal/handlers/http/v1"
	roleusecase "go-boilerplate/internal/usecase/role"
	"go-boilerplate/pkg/response"
)

// AssignPermissions godoc
// @Summary     Assign permissions to role
// @Description Assign permissions to a role, replacing existing permissions
// @ID          role-assign-permissions
// @Tags        roles
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Role ID"
// @Param       request body roledto.AssignPermissionsRequest true "Permission IDs"
// @Success     200 {object} response.Response[roledto.Response]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /roles/{id}/permissions [put]
func (h *Handler) AssignPermissions(ctx *fiber.Ctx) error {
	id, err := parseUint(ctx.Params("id"))
	if err != nil {
		return response.BadRequest(ctx, "INVALID_ID", "Invalid role ID")
	}

	var req roledto.AssignPermissionsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(ctx, v1.ParseValidationErrors(err))
	}

	result, err := h.roleUC.AssignPermissions(ctx.UserContext(), id, req.PermissionIDs)
	if err != nil {
		if errors.Is(err, roleusecase.ErrRoleNotFound) {
			return response.NotFound(ctx, "Role not found")
		}
		h.l.Error(err, "handlers - http - v1 - role - AssignPermissions")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
