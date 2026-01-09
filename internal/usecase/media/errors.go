package media

import "errors"

// Errors.
var (
	ErrInvalidMimeType = errors.New("invalid mime type")
	ErrFileTooLarge    = errors.New("file too large")
)
