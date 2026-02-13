package bankstatement

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	bankstatementdto "go-boilerplate/internal/dto/bankstatement"
	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/pkg/response"
)

// Upload handles multipart bank statement PDF upload and parsing.
func (h *Handler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "INVALID_FILE", "No file provided")
	}

	bankIDStr := c.FormValue("bank_id")
	bankID, err := strconv.ParseUint(bankIDStr, 10, 64)
	if err != nil {
		return response.BadRequest(c, "INVALID_FIELD", "bank_id must be a valid number")
	}

	src, err := file.Open()
	if err != nil {
		h.l.Error(err, "Failed to open uploaded file")
		return response.InternalError(c)
	}
	defer src.Close()

	password := c.FormValue("password")

	result, err := h.bankStatementUC.Upload(c.UserContext(), bankstatementdto.UploadRequest{
		File:     src,
		Filename: file.Filename,
		Size:     file.Size,
		BankID:   uint(bankID),
		Password: password,
		UserID:   middleware.GetUserID(c),
	})
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - bankstatement - Upload")
		return response.InternalError(c)
	}

	return response.Created(c, result)
}
