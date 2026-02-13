package installment

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/pkg/response"
)

// Delete removes an installment.
func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := parseUint(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "INVALID_ID", "Invalid installment ID")
	}

	if err := h.installmentUC.Delete(c.UserContext(), id); err != nil {
		h.l.Error(err, "handlers - http - v1 - installment - Delete")
		return response.InternalError(c)
	}

	return response.NoContent(c)
}
