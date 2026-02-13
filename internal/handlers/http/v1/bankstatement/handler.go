package bankstatement

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

// Handler handles bank statement endpoints.
type Handler struct {
	bankStatementUC usecase.BankStatement
	jwtService      jwt.Service
	l               logger.Interface
	v               *validator.Validate
}

// New creates a new bank statement handler.
func New(bankStatementUC usecase.BankStatement, jwtService jwt.Service, l logger.Interface) *Handler {
	return &Handler{
		bankStatementUC: bankStatementUC,
		jwtService:      jwtService,
		l:               l,
		v:               validator.New(validator.WithRequiredStructEnabled()),
	}
}

// RegisterRoutes sets up bank statement routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	stmts := router.Group("/bank-statements")
	stmts.Use(middleware.JWTAuth(h.jwtService, h.l))
	{
		stmts.Post("/upload", middleware.RequirePermission("bank-statement:write"), h.Upload)
		stmts.Get("/", middleware.RequirePermission("bank-statement:read"), h.List)
		stmts.Get("/:id", middleware.RequirePermission("bank-statement:read"), h.GetByID)
		stmts.Patch("/:id/items/:itemId", middleware.RequirePermission("bank-statement:write"), h.UpdateLineItem)
		stmts.Delete("/:id", middleware.RequirePermission("bank-statement:delete"), h.Delete)
	}

	banks := router.Group("/banks")
	banks.Use(middleware.JWTAuth(h.jwtService, h.l))
	{
		banks.Get("/", middleware.RequirePermission("bank-statement:read"), h.ListBanks)
	}
}

// parseUint parses a string to uint.
func parseUint(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
