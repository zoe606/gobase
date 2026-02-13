package installment

import (
	"github.com/gofiber/fiber/v2"

	installmentdto "go-boilerplate/internal/dto/installment"
	"go-boilerplate/pkg/response"
)

// Update updates an installment.
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := parseUint(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "INVALID_ID", "Invalid installment ID")
	}

	var req installmentdto.UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_JSON", "Invalid request body")
	}

	result, err := h.installmentUC.Update(c.UserContext(), id, req)
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - installment - Update")
		return response.InternalError(c)
	}

	return response.OK(c, result)
}
