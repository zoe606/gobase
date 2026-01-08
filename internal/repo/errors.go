// Package repo contains repository interfaces and common errors.
package repo

import "errors"

// ErrNotFound is returned when a record is not found in the database.
var ErrNotFound = errors.New("record not found")
