package v1

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ParseValidationErrors extracts field-level validation errors.
func ParseValidationErrors(err error) map[string]string {
	fieldErrors := make(map[string]string)

	if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
		for _, e := range validationErrors {
			field := e.Field()
			fieldErrors[field] = validationMessage(e)
		}
	}

	return fieldErrors
}

// validationMessage returns a user-friendly message for a validation error.
func validationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Must be at least %s characters", e.Param())
	case "max":
		return fmt.Sprintf("Must be at most %s characters", e.Param())
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", e.Param())
	case "gte":
		return fmt.Sprintf("Must be at least %s", e.Param())
	case "lte":
		return fmt.Sprintf("Must be at most %s", e.Param())
	case "len":
		return fmt.Sprintf("Must be exactly %s characters", e.Param())
	case "url":
		return "Must be a valid URL"
	case "uuid":
		return "Must be a valid UUID"
	case "alphanum":
		return "Must contain only letters and numbers"
	case "numeric":
		return "Must be a numeric value"
	case "boolean":
		return "Must be true or false"
	default:
		return "Invalid value"
	}
}
