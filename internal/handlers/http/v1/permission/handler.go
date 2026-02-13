package permission

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

// Handler handles permission endpoints.
type Handler struct {
	permissionUC usecase.Permission
	jwtService   jwt.Service
	l            logger.Interface
	v            *validator.Validate
}

// New creates a new permission handler.
func New(permissionUC usecase.Permission, jwtService jwt.Service, l logger.Interface) *Handler {
	return &Handler{
		permissionUC: permissionUC,
		jwtService:   jwtService,
		l:            l,
		v:            validator.New(validator.WithRequiredStructEnabled()),
	}
}

// RegisterRoutes sets up permission routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	permissions := router.Group("/permissions")
	permissions.Use(middleware.JWTAuth(h.jwtService, h.l))
	{
		permissions.Get("/", middleware.RequirePermission("permissions:read"), h.List)
		permissions.Post("/", middleware.RequirePermission("permissions:write"), h.Create)
		permissions.Delete("/:id", middleware.RequirePermission("permissions:delete"), h.Delete)
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
