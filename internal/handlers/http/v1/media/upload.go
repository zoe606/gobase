package media

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/dto/media"
	"go-boilerplate/pkg/response"
)

// Upload handles multipart file upload.
// @Summary     Upload a file
// @Description Uploads a file and attaches it to an entity
// @Tags        media
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       file            formData file   true  "File to upload"
// @Param       attachable_type formData string true  "Entity type (e.g., users, posts)"
// @Param       attachable_id   formData int    true  "Entity ID"
// @Param       collection      formData string false "Collection name (e.g., avatar, gallery)"
// @Success     201 {object} response.Response[mediadto.MediaResponse]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Router      /media/upload [post]
func (h *Handler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "INVALID_FILE", "No file provided")
	}

	if file.Size > h.maxSize {
		return response.BadRequest(c, "FILE_TOO_LARGE", "File exceeds maximum size")
	}

	// Open file.
	src, err := file.Open()
	if err != nil {
		h.l.Error(err, "Failed to open uploaded file")
		return response.InternalError(c)
	}
	defer src.Close()

	// Parse form values.
	attachableType := c.FormValue("attachable_type")
	if attachableType == "" {
		return response.BadRequest(c, "MISSING_FIELD", "attachable_type is required")
	}

	attachableIDStr := c.FormValue("attachable_id")
	attachableID, err := strconv.ParseUint(attachableIDStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "INVALID_FIELD", "attachable_id must be a valid number")
	}

	collection := c.FormValue("collection", "default")

	result, err := h.mediaUC.Upload(c.UserContext(), mediadto.UploadRequest{
		File:           src,
		Filename:       file.Filename,
		Size:           file.Size,
		MimeType:       file.Header.Get("Content-Type"),
		AttachableType: attachableType,
		AttachableID:   uint(attachableID),
		Collection:     collection,
	})
	if err != nil {
		h.l.Error(err, "Upload failed")
		return response.InternalError(c)
	}

	return response.Created(c, result)
}
