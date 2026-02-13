package user

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

// Handler handles user management endpoints.
type Handler struct {
	userUC     usecase.User
	jwtService jwt.Service
	l          logger.Interface
	v          *validator.Validate
}

// New creates a new user handler.
func New(userUC usecase.User, jwtService jwt.Service, l logger.Interface) *Handler {
	return &Handler{
		userUC:     userUC,
		jwtService: jwtService,
		l:          l,
		v:          validator.New(validator.WithRequiredStructEnabled()),
	}
}

// RegisterRoutes sets up user management routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	users := router.Group("/users")
	users.Use(middleware.JWTAuth(h.jwtService, h.l))
	{
		users.Get("/", middleware.RequirePermission("users:read"), h.List)
		users.Post("/", middleware.RequirePermission("users:write"), h.Create)
		users.Get("/:id", middleware.RequirePermission("users:read"), h.GetByID)
		users.Put("/:id", middleware.RequirePermission("users:write"), h.Update)
		users.Delete("/:id", middleware.RequirePermission("users:delete"), h.Delete)
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
