package user

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	userdto "go-boilerplate/internal/dto/user"
	v1 "go-boilerplate/internal/handlers/http/v1"
	userusecase "go-boilerplate/internal/usecase/user"
	"go-boilerplate/pkg/response"
)

// Create godoc
// @Summary     Create user
// @Description Create a new user
// @ID          user-create
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body userdto.CreateRequest true "User data"
// @Success     201 {object} response.Response[userdto.Response]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     409 {object} response.ErrorResponse "Email already exists"
// @Failure     500 {object} response.ErrorResponse
// @Router      /users [post]
func (h *Handler) Create(ctx *fiber.Ctx) error {
	var req userdto.CreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(ctx, v1.ParseValidationErrors(err))
	}

	result, err := h.userUC.Create(ctx.UserContext(), req)
	if err != nil {
		if errors.Is(err, userusecase.ErrEmailExists) {
			return response.Conflict(ctx, "Email already exists")
		}
		if errors.Is(err, userusecase.ErrRoleNotFound) {
			return response.BadRequest(ctx, "ROLE_NOT_FOUND", "Specified role does not exist")
		}
		h.l.Error(err, "handlers - http - v1 - user - Create")
		return response.InternalError(ctx)
	}

	return response.Created(ctx, result)
}
