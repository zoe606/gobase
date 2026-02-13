package installment

import (
	"github.com/gofiber/fiber/v2"

	installmentdto "go-boilerplate/internal/dto/installment"
	"go-boilerplate/pkg/response"
)

// LinkItems links line items to an installment.
func (h *Handler) LinkItems(c *fiber.Ctx) error {
	id, err := parseUint(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "INVALID_ID", "Invalid installment ID")
	}

	var req installmentdto.LinkItemsRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_JSON", "Invalid request body")
	}

	if err := h.installmentUC.LinkItems(c.UserContext(), id, req); err != nil {
		h.l.Error(err, "handlers - http - v1 - installment - LinkItems")
		return response.InternalError(c)
	}

	return response.OK(c, fiber.Map{"message": "Items linked successfully"})
}
