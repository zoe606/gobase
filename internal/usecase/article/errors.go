package article

import "errors"

// Error definitions.
var (
	// ErrNotFound indicates that the article was not found.
	ErrNotFound = errors.New("article not found")

	// ErrAlreadyExists indicates that the article already exists.
	ErrAlreadyExists = errors.New("article already exists")

	// ErrInvalid indicates invalid Article data.
	ErrInvalid = errors.New("invalid article data")
)
