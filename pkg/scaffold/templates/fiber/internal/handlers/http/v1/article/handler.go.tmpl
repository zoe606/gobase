package article

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

// Handler handles article endpoints.
type Handler struct {
	articleUC  usecase.Article
	jwtService jwt.Service
	l          logger.Interface
	v          *validator.Validate
}

// New creates a new article handler.
func New(articleUC usecase.Article, jwtService jwt.Service, l logger.Interface) *Handler {
	return &Handler{
		articleUC:  articleUC,
		jwtService: jwtService,
		l:          l,
		v:          validator.New(validator.WithRequiredStructEnabled()),
	}
}

// RegisterRoutes sets up routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	articles := router.Group("/articles")

	// Public routes (no auth required)
	articles.Get("/", h.List)
	articles.Get("/:id", h.GetByID)

	// Protected routes (JWT auth required)
	protected := articles.Group("", middleware.JWTAuth(h.jwtService, h.l))
	protected.Post("/", h.Create)
	protected.Put("/:id", h.Update)
	protected.Delete("/:id", h.Delete)
}
