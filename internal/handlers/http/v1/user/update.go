package user

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	userdto "go-boilerplate/internal/dto/user"
	v1 "go-boilerplate/internal/handlers/http/v1"
	userusecase "go-boilerplate/internal/usecase/user"
	"go-boilerplate/pkg/response"
)

// Update godoc
// @Summary     Update user
// @Description Update an existing user
// @ID          user-update
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "User ID"
// @Param       request body userdto.UpdateRequest true "User data"
// @Success     200 {object} response.Response[userdto.Response]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     409 {object} response.ErrorResponse "Email already exists"
// @Failure     500 {object} response.ErrorResponse
// @Router      /users/{id} [put]
func (h *Handler) Update(ctx *fiber.Ctx) error {
	id, err := parseUint(ctx.Params("id"))
	if err != nil {
		return response.BadRequest(ctx, "INVALID_ID", "Invalid user ID")
	}

	var req userdto.UpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(ctx, v1.ParseValidationErrors(err))
	}

	result, err := h.userUC.Update(ctx.UserContext(), id, req)
	if err != nil {
		if errors.Is(err, userusecase.ErrUserNotFound) {
			return response.NotFound(ctx, "User not found")
		}
		if errors.Is(err, userusecase.ErrEmailExists) {
			return response.Conflict(ctx, "Email already exists")
		}
		if errors.Is(err, userusecase.ErrRoleNotFound) {
			return response.BadRequest(ctx, "ROLE_NOT_FOUND", "Specified role does not exist")
		}
		h.l.Error(err, "handlers - http - v1 - user - Update")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
