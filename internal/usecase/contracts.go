// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	translationdto "go-boilerplate/internal/dto/translation"
)

//go:generate mockgen -source=contracts.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// Translation defines the translation use case interface.
	Translation interface {
		Translate(context.Context, translationdto.TranslateInput) (*translationdto.TranslateOutput, error)
		History(context.Context) (*translationdto.HistoryOutput, error)
	}
)
