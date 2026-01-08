package v1

import (
	"github.com/gofiber/fiber/v2"
	"go-boilerplate/internal/controller/restapi/v1/request"
	translationdto "go-boilerplate/internal/dto/translation"
	"go-boilerplate/pkg/response"
)

// @Summary     Show history
// @Description Show all translation history
// @ID          history
// @Tags  	    translation
// @Accept      json
// @Produce     json
// @Success     200 {object} response.Response[translation.HistoryResponse]
// @Failure     500 {object} response.ErrorResponse
// @Router      /translation/history [get]
func (r *V1) history(ctx *fiber.Ctx) error {
	result, err := r.t.History(ctx.UserContext())
	if err != nil {
		r.l.Error(err, "restapi - v1 - history")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result.ToResponse())
}

// @Summary     Translate
// @Description Translate a text
// @ID          do-translate
// @Tags  	    translation
// @Accept      json
// @Produce     json
// @Param       request body request.Translate true "Set up translation"
// @Success     200 {object} response.Response[translation.TranslationResponse]
// @Failure     400 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /translation/do-translate [post]
func (r *V1) doTranslate(ctx *fiber.Ctx) error {
	var body request.Translate

	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - doTranslate")
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		return response.ValidationError(ctx, parseValidationErrors(err))
	}

	result, err := r.t.Translate(ctx.UserContext(), translationdto.TranslateInput{
		Source:      body.Source,
		Destination: body.Destination,
		Original:    body.Original,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - doTranslate")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result.ToResponse())
}
