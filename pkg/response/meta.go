package response

// Meta contains pagination metadata for list responses.
type Meta struct {
	Page       int   `json:"page,omitempty"`
	Limit      int   `json:"limit,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// NewMeta creates a new Meta with calculated total pages.
func NewMeta(page, limit int, total int64) *Meta {
	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
