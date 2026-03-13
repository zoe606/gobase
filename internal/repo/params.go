package repo

import "go-boilerplate/pkg/pagination"

// ArticleListParams holds query parameters for listing articles.
type ArticleListParams struct {
	pagination.Params
	Status string
	UserID uint
	Search string
}

// TranslationHistoryParams holds query parameters for translation history.
type TranslationHistoryParams struct {
	pagination.Params
	Search string
}
