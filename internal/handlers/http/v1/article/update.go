package article

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	articledto "go-boilerplate/internal/dto/article"
	v1 "go-boilerplate/internal/handlers/http/v1"
	articleuc "go-boilerplate/internal/usecase/article"
	"go-boilerplate/pkg/response"
)

// Update godoc
// @Summary     Update article
// @Description Update an existing article
// @ID          article-update
// @Tags        articles
// @Accept      json
// @Produce     json
// @Param       id path int true "Article ID"
// @Param       request body articledto.UpdateRequest true "Update Article request"
// @Success     200 {object} response.Response[articledto.Response]
// @Failure     400 {object} response.ErrorResponse
// @Failure     404 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /articles/{id} [put]
func (h *Handler) Update(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		return response.BadRequest(ctx, "INVALID_ID", "Invalid article ID")
	}

	var req articledto.UpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(ctx, v1.ParseValidationErrors(err))
	}

	result, err := h.articleUC.Update(ctx.UserContext(), uint(id), req)
	if err != nil {
		if errors.Is(err, articleuc.ErrNotFound) {
			return response.NotFound(ctx, "Article not found")
		}
		h.l.Error(err, "handlers - http - v1 - article - Update")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
