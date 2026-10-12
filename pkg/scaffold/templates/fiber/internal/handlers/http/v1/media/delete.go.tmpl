package media

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/pkg/response"
)

// Delete removes a media item.
// @Summary     Delete media
// @Description Deletes a media item and its associated files
// @Tags        media
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Media ID"
// @Success     200 {object} response.Response[map[string]string]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Router      /media/{id} [delete]
func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := parseUint(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "INVALID_ID", "Invalid media ID")
	}

	if err := h.mediaUC.Delete(c.UserContext(), id); err != nil {
		return response.NotFound(c, "Media not found")
	}

	return response.OK(c, fiber.Map{"message": "Media deleted successfully"})
}
