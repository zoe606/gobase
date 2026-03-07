package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/dto/auth"
	v1 "go-boilerplate/internal/handlers/http/v1"
	autherrors "go-boilerplate/internal/usecase/auth"
	"go-boilerplate/pkg/response"
)

// Refresh godoc
// @Summary     Refresh tokens
// @Description Refresh access token using refresh token
// @ID          auth-refresh
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body authdto.RefreshRequest true "Refresh token"
// @Success     200 {object} response.Response[authdto.TokenResponse]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Router      /auth/refresh [post]
func (h *Handler) Refresh(ctx *fiber.Ctx) error {
	var req authdto.RefreshRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(ctx, v1.ParseValidationErrors(err))
	}

	result, err := h.authUC.Refresh(ctx.UserContext(), req)
	if err != nil {
		if errors.Is(err, autherrors.ErrInvalidToken) {
			return response.Unauthorized(ctx, "Invalid or expired refresh token")
		}
		h.l.Error(err, "handlers - http - v1 - auth - Refresh")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
