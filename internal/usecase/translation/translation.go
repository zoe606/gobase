// Package translation provides translation use cases.
package translation

//go:generate mockgen -source=../../repo/contracts.go -destination=mocks_test.go -package=translation_test

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
