package translation

// TranslationResponse represents a translation in API responses.
type TranslationResponse struct {
	ID          uint   `json:"id"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Original    string `json:"original"`
	Translation string `json:"translation"`
}

// HistoryResponse represents translation history in API responses.
type HistoryResponse struct {
	History []TranslationResponse `json:"history"`
}
