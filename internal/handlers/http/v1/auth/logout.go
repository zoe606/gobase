package auth

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/dto/auth"
	v1 "go-boilerplate/internal/handlers/http/v1"
	"go-boilerplate/pkg/response"
)

// Logout godoc
// @Summary     Logout
// @Description Invalidate refresh token
// @ID          auth-logout
// @Tags        auth
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body authdto.RefreshRequest true "Refresh token to invalidate"
// @Success     204
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Router      /auth/logout [post]
func (h *Handler) Logout(ctx *fiber.Ctx) error {
	var req authdto.RefreshRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(ctx, v1.ParseValidationErrors(err))
	}

	if err := h.authUC.Logout(ctx.UserContext(), req.RefreshToken); err != nil {
		h.l.Error(err, "handlers - http - v1 - auth - Logout")
		return response.InternalError(ctx)
	}

	return response.NoContent(ctx)
}
