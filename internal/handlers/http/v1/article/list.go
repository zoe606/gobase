package article

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/dto/article"
	"go-boilerplate/pkg/response"
)

// List godoc
// @Summary     List articles
// @Description Get a paginated list of articles
// @ID          article-list
// @Tags        articles
// @Accept      json
// @Produce     json
// @Param       page query int false "Page number" default(1)
// @Param       page_size query int false "Page size" default(20)
// @Success     200 {object} response.Response[articledto.ListResponse]
// @Failure     500 {object} response.ErrorResponse
// @Router      /articles [get]
func (h *Handler) List(ctx *fiber.Ctx) error {
	var req articledto.ListRequest
	if err := ctx.QueryParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_QUERY", "Invalid query parameters")
	}

	result, err := h.articleUC.List(ctx.UserContext(), req)
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - article - List")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
