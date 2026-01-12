//nolint:revive // json package name is intentional for drop-in replacement
package json

import (
	gojson "github.com/goccy/go-json"
)

// GoJSONCodec implements Codec using goccy/go-json library.
// This is the fastest JSON library for Go, making it the default choice.
type GoJSONCodec struct{}

// NewGoJSONCodec creates a new GoJSONCodec instance.
func NewGoJSONCodec() *GoJSONCodec {
	return &GoJSONCodec{}
}

// Marshal encodes v to JSON bytes using goccy/go-json.
func (c *GoJSONCodec) Marshal(v any) ([]byte, error) {
	return gojson.Marshal(v)
}

// Unmarshal decodes JSON bytes into v using goccy/go-json.
func (c *GoJSONCodec) Unmarshal(data []byte, v any) error {
	return gojson.Unmarshal(data, v)
}

// MarshalIndent encodes v to indented JSON bytes using goccy/go-json.
func (c *GoJSONCodec) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return gojson.MarshalIndent(v, prefix, indent)
}

// Package-level convenience functions using goccy/go-json directly.
// These provide a drop-in replacement for encoding/json.

// Marshal encodes v to JSON bytes.
func Marshal(v any) ([]byte, error) {
	return gojson.Marshal(v)
}

// Unmarshal decodes JSON bytes into v.
func Unmarshal(data []byte, v any) error {
	return gojson.Unmarshal(data, v)
}

// MarshalIndent encodes v to indented JSON bytes.
func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return gojson.MarshalIndent(v, prefix, indent)
}
