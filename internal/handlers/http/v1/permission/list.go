package permission

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/pkg/response"
)

// List godoc
// @Summary     List permissions
// @Description Get all available permissions
// @ID          permission-list
// @Tags        permissions
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.Response[[]entity.Permission]
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /permissions [get]
func (h *Handler) List(ctx *fiber.Ctx) error {
	permissions, err := h.permissionUC.List(ctx.UserContext())
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - permission - List")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, permissions)
}
