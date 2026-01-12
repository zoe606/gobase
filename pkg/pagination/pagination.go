// Package pagination provides helpers for paginating database queries.
package pagination

import (
	"gorm.io/gorm"
)

// Default values for pagination.
const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// Params holds pagination parameters from query string.
type Params struct {
	Page  int    `query:"page"`
	Limit int    `query:"limit"`
	Sort  string `query:"sort"`
	Order string `query:"order"` // "asc" or "desc"
}

// NewParams creates Params with default values.
func NewParams() Params {
	return Params{
		Page:  DefaultPage,
		Limit: DefaultLimit,
		Order: "desc",
	}
}

// Normalize applies defaults and enforces limits.
func (p *Params) Normalize() {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.Limit < 1 {
		p.Limit = DefaultLimit
	}
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
	if p.Order != "asc" && p.Order != "desc" {
		p.Order = "desc"
	}
}

// Offset calculates the offset for the current page.
func (p *Params) Offset() int {
	return (p.Page - 1) * p.Limit
}

// Apply applies pagination to a GORM query.
// allowedSorts is a whitelist of allowed sort fields to prevent SQL injection.
func (p *Params) Apply(db *gorm.DB, allowedSorts []string) *gorm.DB {
	p.Normalize()

	query := db.Offset(p.Offset()).Limit(p.Limit)

	if p.Sort != "" && isAllowedSort(p.Sort, allowedSorts) {
		query = query.Order(p.Sort + " " + p.Order)
	}

	return query
}

// isAllowedSort checks if the sort field is in the allowed list.
func isAllowedSort(field string, allowed []string) bool {
	for _, f := range allowed {
		if f == field {
			return true
		}
	}
	return false
}

// TotalPages calculates the total number of pages.
func TotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := int(total) / limit
	if int(total)%limit > 0 {
		pages++
	}
	return pages
}

// Meta holds pagination metadata for API responses.
type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewMeta creates pagination metadata from params and total count.
func NewMeta(page, limit int, total int64) *Meta {
	return &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: TotalPages(total, limit),
	}
}

// Result holds paginated results.
type Result[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewResult creates a new paginated result.
func NewResult[T any](items []T, params Params, total int64) *Result[T] {
	params.Normalize()
	return &Result[T]{
		Items:      items,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: TotalPages(total, params.Limit),
	}
}
