package user

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/handlers/http/middleware"
	userusecase "go-boilerplate/internal/usecase/user"
	"go-boilerplate/pkg/response"
)

// Delete godoc
// @Summary     Delete user
// @Description Delete a user by ID
// @ID          user-delete
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "User ID"
// @Success     204 "No Content"
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /users/{id} [delete]
func (h *Handler) Delete(ctx *fiber.Ctx) error {
	id, err := parseUint(ctx.Params("id"))
	if err != nil {
		return response.BadRequest(ctx, "INVALID_ID", "Invalid user ID")
	}

	currentUserID := middleware.GetUserID(ctx)

	err = h.userUC.Delete(ctx.UserContext(), id, currentUserID)
	if err != nil {
		if errors.Is(err, userusecase.ErrUserNotFound) {
			return response.NotFound(ctx, "User not found")
		}
		if errors.Is(err, userusecase.ErrCannotDeleteSelf) {
			return response.BadRequest(ctx, "CANNOT_DELETE_SELF", "Cannot delete your own account")
		}
		h.l.Error(err, "handlers - http - v1 - user - Delete")
		return response.InternalError(ctx)
	}

	return response.NoContent(ctx)
}
