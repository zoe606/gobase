package media

import (
	"github.com/gofiber/fiber/v2"

	mediadto "go-boilerplate/internal/dto/media"
	v1 "go-boilerplate/internal/handlers/http/v1"
	"go-boilerplate/pkg/response"
)

// GetPresignedURL returns a presigned URL for direct upload.
// @Summary     Get presigned upload URL
// @Description Returns a presigned URL for direct client-to-storage upload
// @Tags        media
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       request body mediadto.PresignedURLRequest true "Request body"
// @Success     200 {object} response.Response[mediadto.PresignedURLResponse]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Router      /media/presigned-url [post]
func (h *Handler) GetPresignedURL(c *fiber.Ctx) error {
	var req mediadto.PresignedURLRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_JSON", "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(c, v1.ParseValidationErrors(err))
	}

	result, err := h.mediaUC.GetPresignedUploadURL(c.UserContext(), req.Filename)
	if err != nil {
		h.l.Error(err, "Failed to get presigned URL")
		return response.InternalError(c)
	}

	return response.OK(c, result)
}
