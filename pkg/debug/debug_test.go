package debug_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/debug"
)

type testStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestDump(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		contains string
	}{
		{
			name:     "string",
			input:    "hello world",
			contains: "hello world",
		},
		{
			name:     "byte slice",
			input:    []byte("byte data"),
			contains: "byte data",
		},
		{
			name:     "struct",
			input:    testStruct{Name: "test", Value: 42},
			contains: `"name": "test"`,
		},
		{
			name:     "map",
			input:    map[string]int{"a": 1, "b": 2},
			contains: `"a": 1`,
		},
		{
			name:     "slice",
			input:    []int{1, 2, 3},
			contains: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := debug.Dump(tt.input)
			require.Contains(t, result, tt.contains)
		})
	}
}

func TestToFile(t *testing.T) {
	t.Parallel()

	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "debug.log")

	// Write to file
	data := testStruct{Name: "file test", Value: 123}
	err := debug.ToFile(data, tmpFile)
	require.NoError(t, err)

	// Verify file contents
	content, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	require.Contains(t, string(content), "file test")
	require.Contains(t, string(content), "123")
}

func TestToFile_Append(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "debug.log")

	// Write multiple times
	err := debug.ToFile("first", tmpFile)
	require.NoError(t, err)

	err = debug.ToFile("second", tmpFile)
	require.NoError(t, err)

	// Verify both entries exist
	content, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	require.Contains(t, string(content), "first")
	require.Contains(t, string(content), "second")
}

func TestToFile_InvalidPath(t *testing.T) {
	t.Parallel()

	err := debug.ToFile("test", "/nonexistent/path/file.log")
	require.Error(t, err)
}
