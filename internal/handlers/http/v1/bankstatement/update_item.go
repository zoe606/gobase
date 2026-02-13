package bankstatement

import (
	"github.com/gofiber/fiber/v2"

	bankstatementdto "go-boilerplate/internal/dto/bankstatement"
	"go-boilerplate/pkg/response"
)

// UpdateLineItem updates a line item within a bank statement.
func (h *Handler) UpdateLineItem(c *fiber.Ctx) error {
	itemID, err := parseUint(c.Params("itemId"))
	if err != nil {
		return response.BadRequest(c, "INVALID_ID", "Invalid item ID")
	}

	var req bankstatementdto.UpdateLineItemRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_JSON", "Invalid request body")
	}

	result, err := h.bankStatementUC.UpdateLineItem(c.UserContext(), itemID, req)
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - bankstatement - UpdateLineItem")
		return response.InternalError(c)
	}

	return response.OK(c, result)
}
