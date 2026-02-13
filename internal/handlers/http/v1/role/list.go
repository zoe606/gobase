package role

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/pkg/response"
)

// List godoc
// @Summary     List roles
// @Description Get all roles with their permissions
// @ID          role-list
// @Tags        roles
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.Response[roledto.ListResponse]
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /roles [get]
func (h *Handler) List(ctx *fiber.Ctx) error {
	result, err := h.roleUC.List(ctx.UserContext())
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - role - List")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
