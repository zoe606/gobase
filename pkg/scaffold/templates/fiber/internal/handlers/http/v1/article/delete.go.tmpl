package article

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/handlers/http/middleware"
	articleuc "go-boilerplate/internal/usecase/article"
	"go-boilerplate/pkg/response"
)

// Delete godoc
// @Summary     Delete article
// @Description Delete a article by ID
// @ID          article-delete
// @Tags        articles
// @Accept      json
// @Produce     json
// @Param       id path int true "Article ID"
// @Success     204 "No Content"
// @Failure     401 {object} response.ErrorResponse
// @Failure     403 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Security    BearerAuth
// @Router      /articles/{id} [delete]
func (h *Handler) Delete(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		return response.BadRequest(ctx, "INVALID_ID", "Invalid article ID")
	}

	userID := middleware.GetUserID(ctx)

	if err := h.articleUC.Delete(ctx.UserContext(), userID, uint(id)); err != nil {
		if errors.Is(err, articleuc.ErrNotFound) {
			return response.NotFound(ctx, "Article not found")
		}
		if errors.Is(err, articleuc.ErrForbidden) {
			return response.Forbidden(ctx, "Not authorized to modify this article")
		}
		h.l.Error(err, "handlers - http - v1 - article - Delete")
		return response.InternalError(ctx)
	}

	return response.NoContent(ctx)
}
