package pagination_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/pagination"
)

func TestParams_Normalize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    pagination.Params
		expected pagination.Params
	}{
		{
			name:  "default values when all zero",
			input: pagination.Params{},
			expected: pagination.Params{
				Page:  pagination.DefaultPage,
				Limit: pagination.DefaultLimit,
				Order: "desc",
			},
		},
		{
			name: "respects valid values",
			input: pagination.Params{
				Page:  2,
				Limit: 50,
				Order: "asc",
			},
			expected: pagination.Params{
				Page:  2,
				Limit: 50,
				Order: "asc",
			},
		},
		{
			name: "caps limit at max",
			input: pagination.Params{
				Page:  1,
				Limit: 200,
				Order: "asc",
			},
			expected: pagination.Params{
				Page:  1,
				Limit: pagination.MaxLimit,
				Order: "asc",
			},
		},
		{
			name: "fixes invalid order",
			input: pagination.Params{
				Page:  1,
				Limit: 10,
				Order: "invalid",
			},
			expected: pagination.Params{
				Page:  1,
				Limit: 10,
				Order: "desc",
			},
		},
		{
			name: "fixes negative page",
			input: pagination.Params{
				Page:  -5,
				Limit: 10,
				Order: "asc",
			},
			expected: pagination.Params{
				Page:  pagination.DefaultPage,
				Limit: 10,
				Order: "asc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := tt.input
			params.Normalize()

			require.Equal(t, tt.expected.Page, params.Page)
			require.Equal(t, tt.expected.Limit, params.Limit)
			require.Equal(t, tt.expected.Order, params.Order)
		})
	}
}

func TestParams_Offset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		page     int
		limit    int
		expected int
	}{
		{"page 1", 1, 10, 0},
		{"page 2", 2, 10, 10},
		{"page 3", 3, 20, 40},
		{"page 5", 5, 15, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := pagination.Params{
				Page:  tt.page,
				Limit: tt.limit,
			}
			require.Equal(t, tt.expected, params.Offset())
		})
	}
}

func TestTotalPages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		total    int64
		limit    int
		expected int
	}{
		{"exact division", 100, 10, 10},
		{"with remainder", 101, 10, 11},
		{"less than limit", 5, 10, 1},
		{"zero total", 0, 10, 0},
		{"zero limit", 100, 0, 0},
		{"single item", 1, 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := pagination.TotalPages(tt.total, tt.limit)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestNewParams(t *testing.T) {
	t.Parallel()

	params := pagination.NewParams()

	require.Equal(t, pagination.DefaultPage, params.Page)
	require.Equal(t, pagination.DefaultLimit, params.Limit)
	require.Equal(t, "desc", params.Order)
}

func TestNewResult(t *testing.T) {
	t.Parallel()

	items := []string{"a", "b", "c"}
	params := pagination.Params{Page: 2, Limit: 10}
	total := int64(25)

	result := pagination.NewResult(items, params, total)

	require.Equal(t, items, result.Items)
	require.Equal(t, 2, result.Page)
	require.Equal(t, 10, result.Limit)
	require.Equal(t, int64(25), result.Total)
	require.Equal(t, 3, result.TotalPages)
}
