package media

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

// Handler handles media endpoints.
type Handler struct {
	mediaUC    usecase.Media
	jwtService jwt.Service
	l          logger.Interface
	v          *validator.Validate
	maxSize    int64
}

// New creates a new media handler.
func New(mediaUC usecase.Media, jwtService jwt.Service, l logger.Interface, maxSize int64) *Handler {
	return &Handler{
		mediaUC:    mediaUC,
		jwtService: jwtService,
		l:          l,
		v:          validator.New(validator.WithRequiredStructEnabled()),
		maxSize:    maxSize,
	}
}

// RegisterRoutes sets up media routes.
func (h *Handler) RegisterRoutes(router fiber.Router) {
	media := router.Group("/media")
	media.Use(middleware.JWTAuth(h.jwtService, h.l))
	{
		media.Post("/upload", h.Upload)
		media.Post("/presigned-url", h.GetPresignedURL)
		media.Get("/:id", h.GetByID)
		media.Get("/:id/url", h.GetURL)
		media.Delete("/:id", h.Delete)
		media.Get("", h.GetByAttachable)
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
