package installment

import (
	"github.com/gofiber/fiber/v2"

	installmentdto "go-boilerplate/internal/dto/installment"
	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/pkg/response"
)

// List returns a paginated list of installments.
func (h *Handler) List(c *fiber.Ctx) error {
	var req installmentdto.ListRequest
	if err := c.QueryParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_QUERY", "Invalid query parameters")
	}

	req.UserID = middleware.GetUserID(c)

	result, err := h.installmentUC.List(c.UserContext(), req)
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - installment - List")
		return response.InternalError(c)
	}

	return response.OK(c, result)
}
