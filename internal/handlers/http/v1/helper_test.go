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
	URL      string `validate:"url"`
	UUID     string `validate:"uuid"`
	Status   string `validate:"oneof=active inactive pending"`
	Username string `validate:"alphanum"`
	Count    int    `validate:"numeric"`
	Active   bool   `validate:"boolean"`
	Age      int    `validate:"gte=18"`
	MaxAge   int    `validate:"lte=120"`
	Code     string `validate:"len=6"`
}

func TestParseValidationErrors(t *testing.T) {
	t.Parallel()

	validate := validator.New(validator.WithRequiredStructEnabled())

	t.Run("required validation", func(t *testing.T) {
		t.Parallel()
		err := validate.Struct(helperTestStruct{})
		result := v1.ParseValidationErrors(err)
		require.Equal(t, "This field is required", result["Email"])
		require.Equal(t, "This field is required", result["Name"])
		require.Equal(t, "This field is required", result["Password"])
	})

	t.Run("email validation", func(t *testing.T) {
		t.Parallel()
		err := validate.Struct(helperTestStruct{
			Email:    "invalid",
			Name:     "ValidName",
			Password: "pass",
		})
		result := v1.ParseValidationErrors(err)
		require.Equal(t, "Invalid email format", result["Email"])
	})

	t.Run("min validation with context", func(t *testing.T) {
		t.Parallel()
		err := validate.Struct(helperTestStruct{
			Email:    "a@b.com",
			Name:     "A",
			Password: "pass",
		})
		result := v1.ParseValidationErrors(err)
		require.Contains(t, result["Name"], "at least 2")
	})

	t.Run("max validation with context", func(t *testing.T) {
		t.Parallel()
		err := validate.Struct(helperTestStruct{
			Email:    "a@b.com",
			Name:     "A" + string(make([]byte, 50)),
			Password: "pass",
		})
		result := v1.ParseValidationErrors(err)
		require.Contains(t, result["Name"], "at most 50")
	})

	t.Run("url validation", func(t *testing.T) {
		t.Parallel()
		err := validate.Struct(helperTestStruct{
			Email:    "a@b.com",
			Name:     "ValidName",
			Password: "pass",
			URL:      "not-a-url",
		})
		result := v1.ParseValidationErrors(err)
		require.Equal(t, "Must be a valid URL", result["URL"])
	})

	t.Run("uuid validation", func(t *testing.T) {
		t.Parallel()
		err := validate.Struct(helperTestStruct{
			Email:    "a@b.com",
			Name:     "ValidName",
			Password: "pass",
			UUID:     "not-a-uuid",
		})
		result := v1.ParseValidationErrors(err)
		require.Equal(t, "Must be a valid UUID", result["UUID"])
	})

	t.Run("oneof validation with context", func(t *testing.T) {
		t.Parallel()
		err := validate.Struct(helperTestStruct{
			Email:    "a@b.com",
			Name:     "ValidName",
			Password: "pass",
			Status:   "invalid_status",
		})
		result := v1.ParseValidationErrors(err)
		require.Contains(t, result["Status"], "one of")
		require.Contains(t, result["Status"], "active inactive pending")
	})

	t.Run("alphanum validation", func(t *testing.T) {
		t.Parallel()
		err := validate.Struct(helperTestStruct{
			Email:    "a@b.com",
			Name:     "ValidName",
			Password: "pass",
			Username: "user@123",
		})
		result := v1.ParseValidationErrors(err)
		require.Equal(t, "Must contain only letters and numbers", result["Username"])
	})

	t.Run("numeric validation", func(t *testing.T) {
		t.Parallel()
		// numeric tag applies to fields, we need a string field for this test
		type numTestStruct struct {
			Value string `validate:"numeric"`
		}
		err := validate.Struct(numTestStruct{Value: "abc"})
		result := v1.ParseValidationErrors(err)
		require.Equal(t, "Must be a numeric value", result["Value"])
	})

	t.Run("boolean validation", func(t *testing.T) {
		t.Parallel()
		type boolTestStruct struct {
			Value string `validate:"boolean"`
		}
		err := validate.Struct(boolTestStruct{Value: "maybe"})
		result := v1.ParseValidationErrors(err)
		require.Equal(t, "Must be true or false", result["Value"])
	})

	t.Run("gte validation with context", func(t *testing.T) {
		t.Parallel()
		err := validate.Struct(helperTestStruct{
			Email:    "a@b.com",
			Name:     "ValidName",
			Password: "pass",
			Age:      16,
		})
		result := v1.ParseValidationErrors(err)
		require.Contains(t, result["Age"], "at least 18")
	})

	t.Run("lte validation with context", func(t *testing.T) {
		t.Parallel()
		err := validate.Struct(helperTestStruct{
			Email:    "a@b.com",
			Name:     "ValidName",
			Password: "pass",
			MaxAge:   150,
		})
		result := v1.ParseValidationErrors(err)
		require.Contains(t, result["MaxAge"], "at most 120")
	})

	t.Run("len validation with context", func(t *testing.T) {
		t.Parallel()
		err := validate.Struct(helperTestStruct{
			Email:    "a@b.com",
			Name:     "ValidName",
			Password: "pass",
			Code:     "abc",
		})
		result := v1.ParseValidationErrors(err)
		require.Contains(t, result["Code"], "exactly 6")
	})

	t.Run("nil error input", func(t *testing.T) {
		t.Parallel()
		result := v1.ParseValidationErrors(nil)
		require.Empty(t, result)
	})

	t.Run("non-validation error", func(t *testing.T) {
		t.Parallel()
		result := v1.ParseValidationErrors(errors.New("some other error"))
		require.Empty(t, result)
	})

	t.Run("valid struct returns empty", func(t *testing.T) {
		t.Parallel()
		err := validate.Struct(helperTestStruct{
			Email:    "valid@example.com",
			Name:     "ValidName",
			Password: "securepass",
			URL:      "https://example.com",
			UUID:     "550e8400-e29b-41d4-a716-446655440000",
			Status:   "active",
			Username: "user123",
			Count:    42,
			Active:   true,
			Age:      25,
			MaxAge:   50,
			Code:     "123456",
		})
		result := v1.ParseValidationErrors(err)
		require.Empty(t, result)
	})
}
