package v1

import (
	"github.com/go-playground/validator/v10"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/logger"
)

// V1 -.
type V1 struct {
	t usecase.Translation
	l logger.Interface
	v *validator.Validate
}
