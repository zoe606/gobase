package profile

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

// Handler handles profile endpoints.
type Handler struct {
	profileUC  usecase.Profile
	jwtService jwt.Service
	l          logger.Interface
	v          *validator.Validate
}

// New creates a new profile handler.
func New(profileUC usecase.Profile, jwtService jwt.Service, l logger.Interface) *Handler {
	return &Handler{
		profileUC:  profileUC,
		jwtService: jwtService,
		l:          l,
		v:          validator.New(validator.WithRequiredStructEnabled()),
	}
}

// RegisterRoutes sets up profile routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	profile := router.Group("/profile")
	profile.Use(middleware.JWTAuth(h.jwtService, h.l))
	{
		profile.Get("", h.GetProfile)
		profile.Patch("", h.UpdateProfile)
	}
}
