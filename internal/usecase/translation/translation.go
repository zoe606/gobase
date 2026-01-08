// Package translation provides translation use cases.
package translation

import (
	"go-boilerplate/internal/repo"
)

// UseCase handles translation business logic.
type UseCase struct {
	repo   repo.TranslationRepo
	webAPI repo.TranslationWebAPI
}

// New creates a new translation use case.
func New(r repo.TranslationRepo, w repo.TranslationWebAPI) *UseCase {
	return &UseCase{
		repo:   r,
		webAPI: w,
	}
}
