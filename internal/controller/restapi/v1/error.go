package v1

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go-boilerplate/internal/controller/restapi/v1/response"
)

func errorResponse(ctx *fiber.Ctx, code int, msg string) error {
	return ctx.Status(code).JSON(response.Error{Error: msg})
}

// parseValidationErrors extracts field-level validation errors.
func parseValidationErrors(err error) map[string]string {
	fieldErrors := make(map[string]string)

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, e := range validationErrors {
			field := e.Field()
			switch e.Tag() {
			case "required":
				fieldErrors[field] = "This field is required"
			case "email":
				fieldErrors[field] = "Invalid email format"
			case "min":
				fieldErrors[field] = "Value is too short"
			case "max":
				fieldErrors[field] = "Value is too long"
			default:
				fieldErrors[field] = "Invalid value"
			}
		}
	}

	return fieldErrors
}
