package v1

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

// ParseValidationErrors extracts field-level validation errors.
func ParseValidationErrors(err error) map[string]string {
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
