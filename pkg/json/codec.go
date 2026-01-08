// Package json provides a swappable JSON encoder/decoder abstraction.
// Following SOLID principles, the Codec interface allows dependency injection
// and easy swapping of JSON libraries without changing application code.
package json

// Codec defines the JSON encoding/decoding interface.
// This abstraction allows swapping JSON libraries (goccy, sonic, stdlib)
// by implementing this interface and injecting via dependency injection.
type Codec interface {
	// Marshal encodes v to JSON bytes.
	Marshal(v any) ([]byte, error)

	// Unmarshal decodes JSON bytes into v.
	Unmarshal(data []byte, v any) error

	// MarshalIndent encodes v to indented JSON bytes.
	MarshalIndent(v any, prefix, indent string) ([]byte, error)
}
