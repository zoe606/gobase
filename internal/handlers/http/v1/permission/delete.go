package permission

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	permissionusecase "go-boilerplate/internal/usecase/permission"
	"go-boilerplate/pkg/response"
)

// Delete godoc
// @Summary     Delete permission
// @Description Delete a permission by ID
// @ID          permission-delete
// @Tags        permissions
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Permission ID"
// @Success     204 "No Content"
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /permissions/{id} [delete]
func (h *Handler) Delete(ctx *fiber.Ctx) error {
	id, err := parseUint(ctx.Params("id"))
	if err != nil {
		return response.BadRequest(ctx, "INVALID_ID", "Invalid permission ID")
	}

	err = h.permissionUC.Delete(ctx.UserContext(), id)
	if err != nil {
		if errors.Is(err, permissionusecase.ErrPermissionNotFound) {
			return response.NotFound(ctx, "Permission not found")
		}
		if errors.Is(err, permissionusecase.ErrPermissionInUse) {
			return response.BadRequest(ctx, "PERMISSION_IN_USE", "Permission is assigned to one or more roles")
		}
		h.l.Error(err, "handlers - http - v1 - permission - Delete")
		return response.InternalError(ctx)
	}

	return response.NoContent(ctx)
}
