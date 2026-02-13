package user

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	userusecase "go-boilerplate/internal/usecase/user"
	"go-boilerplate/pkg/response"
)

// GetByID godoc
// @Summary     Get user
// @Description Get a user by ID
// @ID          user-get
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "User ID"
// @Success     200 {object} response.Response[userdto.Response]
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /users/{id} [get]
func (h *Handler) GetByID(ctx *fiber.Ctx) error {
	id, err := parseUint(ctx.Params("id"))
	if err != nil {
		return response.BadRequest(ctx, "INVALID_ID", "Invalid user ID")
	}

	result, err := h.userUC.GetByID(ctx.UserContext(), id)
	if err != nil {
		if errors.Is(err, userusecase.ErrUserNotFound) {
			return response.NotFound(ctx, "User not found")
		}
		h.l.Error(err, "handlers - http - v1 - user - GetByID")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
