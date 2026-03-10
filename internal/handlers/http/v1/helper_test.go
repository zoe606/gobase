package v1_test

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"

	v1 "go-boilerplate/internal/handlers/http/v1"
)

type helperTestStruct struct {
	Email    string `validate:"required,email"`
	Name     string `validate:"required,min=2,max=50"`
	Password string `validate:"required"`
}

func TestParseValidationErrors(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	t.Run("required fields", func(t *testing.T) {
		t.Parallel()
		err := validate.Struct(helperTestStruct{})
		result := v1.ParseValidationErrors(err)
		require.Equal(t, "This field is required", result["Email"])
		require.Equal(t, "This field is required", result["Name"])
		require.Equal(t, "This field is required", result["Password"])
	})

	t.Run("email validation", func(t *testing.T) {
		t.Parallel()
		err := validate.Struct(helperTestStruct{Email: "invalid", Name: "Ab", Password: "pass"})
		result := v1.ParseValidationErrors(err)
		require.Equal(t, "Invalid email format", result["Email"])
	})

	t.Run("min validation", func(t *testing.T) {
		t.Parallel()
		err := validate.Struct(helperTestStruct{Email: "a@b.com", Name: "A", Password: "pass"})
		result := v1.ParseValidationErrors(err)
		require.Equal(t, "Value is too short", result["Name"])
	})

	t.Run("non-validation error", func(t *testing.T) {
		t.Parallel()
		result := v1.ParseValidationErrors(errors.New("some other error"))
		require.Empty(t, result)
	})
}
