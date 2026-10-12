package media

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/pkg/response"
)

// GetURL returns a URL to access the media.
// @Summary     Get media URL
// @Description Returns a signed URL to access the media file
// @Tags        media
// @Security    BearerAuth
// @Produce     json
// @Param       id      path  int    true  "Media ID"
// @Param       variant query string false "Variant name (e.g., thumb, medium, large)"
// @Success     200 {object} response.Response[map[string]string]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Router      /media/{id}/url [get]
func (h *Handler) GetURL(c *fiber.Ctx) error {
	id, err := parseUint(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "INVALID_ID", "Invalid media ID")
	}

	variant := c.Query("variant", "")

	media, err := h.mediaUC.GetByID(c.UserContext(), id)
	if err != nil {
		return response.NotFound(c, "Media not found")
	}

	url, err := h.mediaUC.GetURL(c.UserContext(), media, variant)
	if err != nil {
		h.l.Error(err, "Failed to get media URL")
		return response.InternalError(c)
	}

	return response.OK(c, fiber.Map{"url": url})
}
