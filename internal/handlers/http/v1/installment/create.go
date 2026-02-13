package installment

import (
	"github.com/gofiber/fiber/v2"

	installmentdto "go-boilerplate/internal/dto/installment"
	"go-boilerplate/internal/handlers/http/middleware"
	v1 "go-boilerplate/internal/handlers/http/v1"
	"go-boilerplate/pkg/response"
)

// Create creates a new installment.
func (h *Handler) Create(c *fiber.Ctx) error {
	var req installmentdto.CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_JSON", "Invalid request body")
	}

	req.UserID = middleware.GetUserID(c)

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(c, v1.ParseValidationErrors(err))
	}

	result, err := h.installmentUC.Create(c.UserContext(), req)
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - installment - Create")
		return response.InternalError(c)
	}

	return response.Created(c, result)
}
