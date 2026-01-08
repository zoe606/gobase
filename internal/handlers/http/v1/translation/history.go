package translation

import (
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/pkg/response"
)

// History godoc
// @Summary     Show history
// @Description Show all translation history
// @ID          history
// @Tags  	    translation
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[translationdto.HistoryResponse]
// @Failure     500 {object} response.ErrorResponse
// @Router      /translation/history [get]
func (h *Handler) History(ctx *fiber.Ctx) error {
	result, err := h.translationUC.History(ctx.UserContext())
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - translation - History")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
