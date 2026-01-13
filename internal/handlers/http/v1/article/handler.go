package article

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/logger"
)

// Handler handles article endpoints.
type Handler struct {
	articleUC usecase.Article
	l         logger.Interface
	v         *validator.Validate
}

// New creates a new article handler.
func New(articleUC usecase.Article, l logger.Interface) *Handler {
	return &Handler{
		articleUC: articleUC,
		l:         l,
		v:         validator.New(validator.WithRequiredStructEnabled()),
	}
}

// RegisterRoutes sets up routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	articles := router.Group("/articles")
	{
		articles.Post("/", h.Create)
		articles.Get("/", h.List)
		articles.Get("/:id", h.GetByID)
		articles.Put("/:id", h.Update)
		articles.Delete("/:id", h.Delete)
	}
}
