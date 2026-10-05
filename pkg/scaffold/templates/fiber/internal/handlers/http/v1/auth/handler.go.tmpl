package auth

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

// Handler handles authentication endpoints.
type Handler struct {
	authUC     usecase.Auth
	jwtService jwt.Service
	l          logger.Interface
	v          *validator.Validate
}

// New creates a new auth handler.
func New(authUC usecase.Auth, jwtService jwt.Service, l logger.Interface) *Handler {
	return &Handler{
		authUC:     authUC,
		jwtService: jwtService,
		l:          l,
		v:          validator.New(validator.WithRequiredStructEnabled()),
	}
}

// RegisterRoutes sets up auth routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	auth := router.Group("/auth")
	{
		// Public routes
		auth.Post("/register", h.Register)
		auth.Post("/login", h.Login)
		auth.Post("/refresh", h.Refresh)

		// Protected routes
		auth.Post("/logout", middleware.JWTAuth(h.jwtService, h.l), h.Logout)
		auth.Get("/me", middleware.JWTAuth(h.jwtService, h.l), h.Me)
	}
}
