package json_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/json"
)

type testStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestMarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{
			name:    "struct",
			input:   testStruct{Name: "test", Value: 42},
			wantErr: false,
		},
		{
			name:    "map",
			input:   map[string]string{"key": "value"},
			wantErr: false,
		},
		{
			name:    "slice",
			input:   []int{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "nil",
			input:   nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, data)
		})
	}
}

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   []byte
		target  any
		wantErr bool
	}{
		{
			name:    "valid struct",
			input:   []byte(`{"name":"test","value":42}`),
			target:  &testStruct{},
			wantErr: false,
		},
		{
			name:    "valid map",
			input:   []byte(`{"key":"value"}`),
			target:  &map[string]string{},
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   []byte(`{invalid}`),
			target:  &testStruct{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := json.Unmarshal(tt.input, tt.target)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	original := testStruct{Name: "roundtrip", Value: 123}

	// Marshal
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal
	var result testStruct
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Compare
	require.Equal(t, original, result)
}

func TestCodec(t *testing.T) {
	t.Parallel()

	codec := json.NewGoJSONCodec()

	// Test Marshal
	data, err := codec.Marshal(testStruct{Name: "codec", Value: 456})
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Test Unmarshal
	var result testStruct
	err = codec.Unmarshal(data, &result)
	require.NoError(t, err)
	require.Equal(t, "codec", result.Name)
	require.Equal(t, 456, result.Value)
}

func BenchmarkMarshal(b *testing.B) {
	data := testStruct{Name: "benchmark", Value: 999}
	for b.Loop() {
		_, _ = json.Marshal(data) //nolint:errcheck // benchmark
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	data := []byte(`{"name":"benchmark","value":999}`)
	var result testStruct
	for b.Loop() {
		_ = json.Unmarshal(data, &result) //nolint:errcheck // benchmark
	}
}
