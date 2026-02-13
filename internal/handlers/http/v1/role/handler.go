package role

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

// Handler handles role management endpoints.
type Handler struct {
	roleUC     usecase.Role
	jwtService jwt.Service
	l          logger.Interface
	v          *validator.Validate
}

// New creates a new role handler.
func New(roleUC usecase.Role, jwtService jwt.Service, l logger.Interface) *Handler {
	return &Handler{
		roleUC:     roleUC,
		jwtService: jwtService,
		l:          l,
		v:          validator.New(validator.WithRequiredStructEnabled()),
	}
}

// RegisterRoutes sets up role management routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	roles := router.Group("/roles")
	roles.Use(middleware.JWTAuth(h.jwtService, h.l))
	{
		roles.Get("/", middleware.RequirePermission("roles:read"), h.List)
		roles.Post("/", middleware.RequirePermission("roles:write"), h.Create)
		roles.Get("/:id", middleware.RequirePermission("roles:read"), h.GetByID)
		roles.Put("/:id", middleware.RequirePermission("roles:write"), h.Update)
		roles.Delete("/:id", middleware.RequirePermission("roles:delete"), h.Delete)
		roles.Put("/:id/permissions", middleware.RequirePermission("roles:write"), h.AssignPermissions)
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
