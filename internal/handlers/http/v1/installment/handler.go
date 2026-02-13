package installment

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

// Handler handles installment endpoints.
type Handler struct {
	installmentUC usecase.Installment
	jwtService    jwt.Service
	l             logger.Interface
	v             *validator.Validate
}

// New creates a new installment handler.
func New(installmentUC usecase.Installment, jwtService jwt.Service, l logger.Interface) *Handler {
	return &Handler{
		installmentUC: installmentUC,
		jwtService:    jwtService,
		l:             l,
		v:             validator.New(validator.WithRequiredStructEnabled()),
	}
}

// RegisterRoutes sets up installment routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	inst := router.Group("/installments")
	inst.Use(middleware.JWTAuth(h.jwtService, h.l))
	{
		inst.Post("/", middleware.RequirePermission("installment:write"), h.Create)
		inst.Get("/", middleware.RequirePermission("installment:read"), h.List)
		inst.Get("/:id", middleware.RequirePermission("installment:read"), h.GetByID)
		inst.Put("/:id", middleware.RequirePermission("installment:write"), h.Update)
		inst.Delete("/:id", middleware.RequirePermission("installment:delete"), h.Delete)
		inst.Post("/:id/link", middleware.RequirePermission("installment:write"), h.LinkItems)
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
