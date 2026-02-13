package bankstatement

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/pkg/response"
)

// GetByID returns a bank statement with its line items.
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, err := parseUint(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "INVALID_ID", "Invalid statement ID")
	}

	result, err := h.bankStatementUC.GetByID(c.UserContext(), id)
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - bankstatement - GetByID")
		return response.InternalError(c)
	}

	return response.OK(c, result)
}
