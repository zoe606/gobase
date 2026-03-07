package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/dto/auth"
	v1 "go-boilerplate/internal/handlers/http/v1"
	autherrors "go-boilerplate/internal/usecase/auth"
	"go-boilerplate/pkg/response"
)

// Register godoc
// @Summary     Register a new user
// @Description Register a new user account
// @ID          auth-register
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body authdto.RegisterRequest true "Registration details"
// @Success     201 {object} response.Response[authdto.LoginResponse]
// @Failure     400 {object} response.ErrorResponse
// @Failure     409 {object} response.ErrorResponse
// @Router      /auth/register [post]
func (h *Handler) Register(ctx *fiber.Ctx) error {
	var req authdto.RegisterRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(ctx, v1.ParseValidationErrors(err))
	}

	result, err := h.authUC.Register(ctx.UserContext(), req)
	if err != nil {
		if errors.Is(err, autherrors.ErrEmailExists) {
			return response.Conflict(ctx, "Email already exists")
		}
		h.l.Error(err, "handlers - http - v1 - auth - Register")
		return response.InternalError(ctx)
	}

	return response.Created(ctx, result)
}
