package article

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/dto/article"
	v1 "go-boilerplate/internal/handlers/http/v1"
	"go-boilerplate/pkg/response"
)

// Create godoc
// @Summary     Create article
// @Description Create a new article
// @ID          article-create
// @Tags        articles
// @Accept      json
// @Produce     json
// @Param       request body articledto.CreateRequest true "Create Article request"
// @Success     201 {object} response.Response[articledto.Response]
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /articles [post]
func (h *Handler) Create(ctx *fiber.Ctx) error {
	var req articledto.CreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(ctx, v1.ParseValidationErrors(err))
	}

	result, err := h.articleUC.Create(ctx.UserContext(), req)
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - article - Create")
		return response.InternalError(ctx)
	}

	return response.Created(ctx, result)
}
