package translation

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/internal/dto/translation"
	"go-boilerplate/pkg/response"
)

// History godoc
// @Summary     Show history
// @Description Show all translation history with pagination
// @ID          history
// @Tags  	    translation
// @Accept      json
// @Produce     json
// @Param       page  query int false "Page number" default(1)
// @Param       limit query int false "Items per page" default(20)
// @Param       sort  query string false "Sort field" Enums(created_at, id)
// @Param       order query string false "Sort order" Enums(asc, desc) default(desc)
// @Success     200 {object} response.Response[[]translationdto.TranslationResponse]
// @Failure     500 {object} response.ErrorResponse
// @Router      /translation/history [get]
func (h *Handler) History(ctx *fiber.Ctx) error {
	var req translationdto.HistoryRequest
	if err := ctx.QueryParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_PARAMS", "invalid pagination parameters")
	}

	result, err := h.translationUC.History(ctx.UserContext(), req)
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - translation - History")
		return response.InternalError(ctx)
	}

	return response.OKWithMeta(ctx, result.Items, &response.Meta{
		Page:       result.Meta.Page,
		Limit:      result.Meta.Limit,
		Total:      result.Meta.Total,
		TotalPages: result.Meta.TotalPages,
	})
}
