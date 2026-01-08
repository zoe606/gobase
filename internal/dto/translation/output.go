package translation

import "go-boilerplate/internal/entity"

// TranslateOutput represents translation output.
type TranslateOutput struct {
	Translation entity.Translation
}

// ToResponse converts output to API response.
func (o *TranslateOutput) ToResponse() TranslationResponse {
	return TranslationResponse{
		ID:          o.Translation.ID,
		Source:      o.Translation.Source,
		Destination: o.Translation.Destination,
		Original:    o.Translation.Original,
		Translation: o.Translation.Translation,
	}
}

// HistoryOutput represents translation history output.
type HistoryOutput struct {
	History []entity.Translation
}

// ToResponse converts output to API response.
func (o *HistoryOutput) ToResponse() HistoryResponse {
	history := make([]TranslationResponse, len(o.History))

	for i := range o.History {
		history[i] = TranslationResponse{
			ID:          o.History[i].ID,
			Source:      o.History[i].Source,
			Destination: o.History[i].Destination,
			Original:    o.History[i].Original,
			Translation: o.History[i].Translation,
		}
	}

	return HistoryResponse{History: history}
}
