// Package debug provides development-only debugging utilities.
// These functions are designed for quick debugging during development
// and should not be used in production code.
package debug

import (
	"bytes"
	"fmt"
	"os"

	"github.com/goccy/go-json"
)

const separator = "════════════════════════════════════════════════════════"

// Dump serializes any value to a pretty-printed JSON string.
// For strings and byte slices, returns them as-is.
func Dump(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	default:
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetIndent("", "  ")
		if err := enc.Encode(v); err != nil {
			return fmt.Sprintf("debug.Dump error: %v", err)
		}
		return buf.String()
	}
}

// Print outputs a value with visible markers to stdout.
// Useful for quick debugging during development.
//
//nolint:forbidigo // This is a debug utility, fmt.Print is intentional.
func Print(v any) {
	fmt.Println(separator)
	fmt.Print(Dump(v))
	fmt.Println(separator)
}

// ToFile appends a value to a debug file.
// Creates the file if it doesn't exist.
func ToFile(v any, path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("debug.ToFile: %w", err)
	}
	defer f.Close()

	_, err = f.WriteString(Dump(v))
	return err
}

// PrintLabeled outputs a value with a label for easier identification.
//
//nolint:forbidigo // This is a debug utility, fmt.Print is intentional.
func PrintLabeled(label string, v any) {
	fmt.Println(separator)
	fmt.Printf("[%s]\n", label)
	fmt.Print(Dump(v))
	fmt.Println(separator)
}
