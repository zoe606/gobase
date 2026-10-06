package article

import (
	"github.com/gofiber/fiber/v2"

	articledto "go-boilerplate/internal/dto/article"
	v1 "go-boilerplate/internal/handlers/http/v1"
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
// @Param       limit query int false "Page size" default(20)
// @Param       status query string false "Article status; omit or leave empty for all statuses" Enums(draft,published)
// @Success     200 {object} response.Response[articledto.ListResponse]
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /articles [get]
func (h *Handler) List(ctx *fiber.Ctx) error {
	var req articledto.ListRequest
	if err := ctx.QueryParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_QUERY", "Invalid query parameters")
	}
	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(ctx, v1.ParseValidationErrors(err))
	}

	result, err := h.articleUC.List(ctx.UserContext(), req)
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - article - List")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
