package translationdto

import (
	"go-boilerplate/internal/entity"
	"go-boilerplate/pkg/pagination"
)

// TranslationResponse represents a translation in API responses.
type TranslationResponse struct {
	ID          uint   `json:"id"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Original    string `json:"original"`
	Translation string `json:"translation"`
}

// NewTranslationResponse creates a TranslationResponse from entity.
func NewTranslationResponse(t *entity.Translation) *TranslationResponse {
	return &TranslationResponse{
		ID:          t.ID,
		Source:      t.Source,
		Destination: t.Destination,
		Original:    t.Original,
		Translation: t.Translation,
	}
}

// HistoryResponse represents translation history in API responses.
type HistoryResponse struct {
	Items []TranslationResponse
	Meta  *pagination.Meta
}

// NewHistoryResponse creates a HistoryResponse from entities with pagination.
func NewHistoryResponse(translations []entity.Translation, params pagination.Params, total int64) *HistoryResponse {
	items := make([]TranslationResponse, len(translations))
	for i := range translations {
		items[i] = TranslationResponse{
			ID:          translations[i].ID,
			Source:      translations[i].Source,
			Destination: translations[i].Destination,
			Original:    translations[i].Original,
			Translation: translations[i].Translation,
		}
	}

	params.Normalize()

	return &HistoryResponse{
		Items: items,
		Meta:  pagination.NewMeta(params.Page, params.Limit, total),
	}
}
