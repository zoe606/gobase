// Package translation provides DTOs for translation operations.
package translation

import "go-boilerplate/pkg/pagination"

// TranslateRequest represents translation request.
type TranslateRequest struct {
	Source      string `json:"source" validate:"required" example:"auto"`
	Destination string `json:"destination" validate:"required" example:"en"`
	Original    string `json:"original" validate:"required" example:"text to translate"`
}

// HistoryRequest represents translation history request with pagination and search.
type HistoryRequest struct {
	pagination.Params
	Search string `query:"search"` // Search in original or translated text
}
