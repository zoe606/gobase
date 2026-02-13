package bankstatement

import (
	"github.com/gofiber/fiber/v2"

	bankstatementdto "go-boilerplate/internal/dto/bankstatement"
	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/pkg/response"
)

// List returns a paginated list of bank statements.
func (h *Handler) List(c *fiber.Ctx) error {
	var req bankstatementdto.ListRequest
	if err := c.QueryParser(&req); err != nil {
		return response.BadRequest(c, "INVALID_QUERY", "Invalid query parameters")
	}

	req.UserID = middleware.GetUserID(c)

	result, err := h.bankStatementUC.List(c.UserContext(), req)
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - bankstatement - List")
		return response.InternalError(c)
	}

	return response.OK(c, result)
}
