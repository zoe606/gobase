// Package translation provides DTOs for translation operations.
package translation

// TranslateInput represents translation input.
type TranslateInput struct {
	Source      string
	Destination string
	Original    string
}
