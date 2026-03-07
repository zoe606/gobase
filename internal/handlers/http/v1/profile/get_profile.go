package profile

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/dto/profile"
	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/pkg/response"
)

// Ensure profiledto is used for swagger.
var _ = profiledto.ProfileResponse{}

// GetProfile godoc
// @Summary     Get current user's profile
// @Description Returns the profile of the currently authenticated user
// @ID          profile-get
// @Tags        profile
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.Response[profiledto.ProfileResponse]
// @Failure     401 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /profile [get]
func (h *Handler) GetProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		return response.Unauthorized(c, "User not found in context")
	}

	result, err := h.profileUC.GetProfile(c.UserContext(), userID)
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - profile - GetProfile")
		return response.InternalError(c)
	}

	return response.OK(c, result)
}
