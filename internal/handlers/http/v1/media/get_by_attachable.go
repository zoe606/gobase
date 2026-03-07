package media

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/dto/media"
	"go-boilerplate/pkg/response"
)

// GetByAttachable returns media for an attachable entity.
// @Summary     Get media by attachable
// @Description Returns all media items for an attachable entity
// @Tags        media
// @Security    BearerAuth
// @Produce     json
// @Param       attachable_type query string true  "Entity type (e.g., users, posts)"
// @Param       attachable_id   query int    true  "Entity ID"
// @Param       collection      query string false "Collection name"
// @Success     200 {object} response.Response[mediadto.MediaListResponse]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Router      /media [get]
func (h *Handler) GetByAttachable(c *fiber.Ctx) error {
	attachableType := c.Query("attachable_type")
	if attachableType == "" {
		return response.BadRequest(c, "MISSING_PARAM", "attachable_type is required")
	}

	attachableIDStr := c.Query("attachable_id")
	attachableID, err := strconv.ParseUint(attachableIDStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "INVALID_PARAM", "attachable_id must be a valid number")
	}

	collection := c.Query("collection", "")

	result, err := h.mediaUC.GetByAttachable(c.UserContext(), mediadto.GetMediaRequest{
		AttachableType: attachableType,
		AttachableID:   uint(attachableID),
		Collection:     collection,
	})
	if err != nil {
		h.l.Error(err, "Failed to get media by attachable")
		return response.InternalError(c)
	}

	return response.OK(c, result)
}
