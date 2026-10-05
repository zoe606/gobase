package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	authdto "go-boilerplate/internal/dto/auth"
	v1 "go-boilerplate/internal/handlers/http/v1"
	autherrors "go-boilerplate/internal/usecase/auth"
	"go-boilerplate/pkg/response"
)

// Login godoc
// @Summary     Login
// @Description Authenticate user and get tokens
// @ID          auth-login
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body authdto.LoginRequest true "Login credentials"
// @Success     200 {object} response.Response[authdto.LoginResponse]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Router      /auth/login [post]
func (h *Handler) Login(ctx *fiber.Ctx) error {
	var req authdto.LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(ctx, v1.ParseValidationErrors(err))
	}

	result, err := h.authUC.Login(ctx.UserContext(), req)
	if err != nil {
		if errors.Is(err, autherrors.ErrInvalidCredentials) {
			return response.Unauthorized(ctx, "Invalid email or password")
		}
		if errors.Is(err, autherrors.ErrUserNotActive) {
			return response.Forbidden(ctx, "Account is not active")
		}
		h.l.Error(err, "handlers - http - v1 - auth - Login")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
