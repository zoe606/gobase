package bankstatement

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/pkg/response"
)

// ListBanks returns all available banks.
func (h *Handler) ListBanks(c *fiber.Ctx) error {
	result, err := h.bankStatementUC.ListBanks(c.UserContext())
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - bankstatement - ListBanks")
		return response.InternalError(c)
	}

	return response.OK(c, result)
}
