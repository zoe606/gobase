package media

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/dto/media"
	"go-boilerplate/pkg/response"
)

// GetByID returns a media item by ID.
// @Summary     Get media by ID
// @Description Returns a media item by its ID
// @Tags        media
// @Security    BearerAuth
// @Produce     json
// @Param       id path int true "Media ID"
// @Success     200 {object} response.Response[mediadto.MediaResponse]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Router      /media/{id} [get]
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, err := parseUint(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "INVALID_ID", "Invalid media ID")
	}

	media, err := h.mediaUC.GetByID(c.UserContext(), id)
	if err != nil {
		return response.NotFound(c, "Media not found")
	}

	return response.OK(c, mediadto.FromEntity(media))
}
