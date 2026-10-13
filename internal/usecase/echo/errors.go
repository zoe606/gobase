package echo

import "errors"

var ErrInvalidMessage = errors.New("message must contain non-empty UTF-8 text up to 4096 bytes")
