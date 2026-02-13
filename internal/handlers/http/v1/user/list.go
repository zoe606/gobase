package user

import (
	"github.com/gofiber/fiber/v2"

	userdto "go-boilerplate/internal/dto/user"
	"go-boilerplate/pkg/response"
)

// List godoc
// @Summary     List users
// @Description Get a paginated list of users
// @ID          user-list
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       page query int false "Page number" default(1)
// @Param       page_size query int false "Page size" default(20)
// @Param       search query string false "Search by name or email"
// @Param       role_id query int false "Filter by role ID"
// @Param       active query bool false "Filter by active status"
// @Success     200 {object} response.Response[userdto.ListResponse]
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /users [get]
func (h *Handler) List(ctx *fiber.Ctx) error {
	var req userdto.ListRequest
	if err := ctx.QueryParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_QUERY", "Invalid query parameters")
	}

	result, err := h.userUC.List(ctx.UserContext(), req)
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - user - List")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
