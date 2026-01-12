package profile

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	profiledto "go-boilerplate/internal/dto/profile"
	"go-boilerplate/internal/handlers/http/middleware"
	v1 "go-boilerplate/internal/handlers/http/v1"
	profileuc "go-boilerplate/internal/usecase/profile"
	"go-boilerplate/pkg/response"
)

// UpdateProfile godoc
// @Summary     Update current user's profile
// @Description Updates the profile of the currently authenticated user
// @ID          profile-update
// @Tags        profile
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body profiledto.UpdateProfileRequest true "Profile update data"
// @Success     200 {object} response.Response[profiledto.ProfileResponse]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /profile [patch]
func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return response.Unauthorized(c, "User not found in context")
	}

	var req profiledto.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	if err := h.v.Struct(&req); err != nil {
		return response.ValidationError(c, v1.ParseValidationErrors(err))
	}

	result, err := h.profileUC.UpdateProfile(c.UserContext(), userID, req)
	if err != nil {
		if errors.Is(err, profileuc.ErrInvalidMedia) {
			return response.BadRequest(c, "INVALID_MEDIA", "Media does not exist or does not belong to user")
		}
		h.l.Error(err, "handlers - http - v1 - profile - UpdateProfile")
		return response.InternalError(c)
	}

	return response.OK(c, result)
}
