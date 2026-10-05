package translation

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/logger"
)

// Handler handles translation endpoints.
type Handler struct {
	translationUC usecase.Translation
	l             logger.Interface
	v             *validator.Validate
}

// New creates a new translation handler.
func New(translationUC usecase.Translation, l logger.Interface) *Handler {
	return &Handler{
		translationUC: translationUC,
		l:             l,
		v:             validator.New(validator.WithRequiredStructEnabled()),
	}
}

// RegisterRoutes sets up translation routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	translation := router.Group("/translation")
	{
		translation.Get("/history", h.History)
		translation.Post("/do-translate", h.Translate)
	}
}
