package article

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	_ "go-boilerplate/internal/dto/article" // swagger type resolution
	articleuc "go-boilerplate/internal/usecase/article"
	"go-boilerplate/pkg/response"
)

// GetByID godoc
// @Summary     Get article by ID
// @Description Get a article by its ID
// @ID          article-get-by-id
// @Tags        articles
// @Accept      json
// @Produce     json
// @Param       id path int true "Article ID"
// @Success     200 {object} response.Response[articledto.Response]
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /articles/{id} [get]
func (h *Handler) GetByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		return response.BadRequest(ctx, "INVALID_ID", "Invalid article ID")
	}

	result, err := h.articleUC.GetByID(ctx.UserContext(), uint(id))
	if err != nil {
		if errors.Is(err, articleuc.ErrNotFound) {
			return response.NotFound(ctx, "Article not found")
		}
		h.l.Error(err, "handlers - http - v1 - article - GetByID")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
