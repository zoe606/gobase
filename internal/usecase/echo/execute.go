package echo

import (
	"context"
	"strings"
	"unicode/utf8"
)

func (*UseCase) Echo(ctx context.Context, message string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if !utf8.ValidString(message) || strings.TrimSpace(message) == "" || len(message) > 4096 {
		return "", ErrInvalidMessage
	}
	return message, nil
}
