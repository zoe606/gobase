package bankstatement

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/pkg/response"
)

// Delete removes a bank statement and its line items.
func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := parseUint(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "INVALID_ID", "Invalid statement ID")
	}

	if err := h.bankStatementUC.Delete(c.UserContext(), id); err != nil {
		h.l.Error(err, "handlers - http - v1 - bankstatement - Delete")
		return response.InternalError(c)
	}

	return response.NoContent(c)
}
