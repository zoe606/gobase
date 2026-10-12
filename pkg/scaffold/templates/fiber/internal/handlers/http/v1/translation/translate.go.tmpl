package translation

import (
	"github.com/gofiber/fiber/v2"

	translationdto "go-boilerplate/internal/dto/translation"
	v1 "go-boilerplate/internal/handlers/http/v1"
	"go-boilerplate/pkg/response"
)

// Translate godoc
// @Summary     Translate
// @Description Translate a text
// @ID          do-translate
// @Tags  	    translation
// @Accept      json
// @Produce     json
// @Param       request body translationdto.TranslateRequest true "Set up translation"
// @Success     200 {object} response.Response[translationdto.TranslationResponse]
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /translation/do-translate [post]
func (h *Handler) Translate(ctx *fiber.Ctx) error {
	var req translationdto.TranslateRequest

	if err := ctx.BodyParser(&req); err != nil {
		h.l.Error(err, "handlers - http - v1 - translation - Translate")
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return response.ValidationError(ctx, v1.ParseValidationErrors(err))
	}

	result, err := h.translationUC.Translate(ctx.UserContext(), req)
	if err != nil {
		h.l.Error(err, "handlers - http - v1 - translation - Translate")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result)
}
