package installment

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/pkg/response"
)

// GetByID returns an installment with its linked line items.
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, err := parseUint(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "INVALID_ID", "Invalid installment ID")
	}

	result, err := h.installmentUC.GetByID(c.UserContext(), id)
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - installment - GetByID")
		return response.InternalError(c)
	}

	return response.OK(c, result)
}
