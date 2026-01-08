package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/response"
)

// Me godoc
// @Summary     Get current user
// @Description Get currently authenticated user info
// @ID          auth-me
// @Tags        auth
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.Response[authdto.UserResponse]
// @Failure     401 {object} response.ErrorResponse
// @Router      /auth/me [get]
func (h *Handler) Me(ctx *fiber.Ctx) error {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		return response.Unauthorized(ctx, "User not found in context")
	}

	result, err := h.authUC.GetCurrentUser(ctx.UserContext(), userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return response.NotFound(ctx, "User not found")
		}
		h.l.Error(err, "handlers - http - v1 - auth - Me")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
