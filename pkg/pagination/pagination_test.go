package pagination_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

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

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)
	return gormDB
}

func TestApply(t *testing.T) {
	t.Parallel()

	t.Run("with allowed sort", func(t *testing.T) {
		t.Parallel()
		gormDB := newTestDB(t)
		params := pagination.Params{Page: 2, Limit: 10, Sort: "name", Order: "asc"}
		result := params.Apply(gormDB, []string{"name", "created_at"})
		require.NotNil(t, result)
	})

	t.Run("sort not in allowed list is ignored", func(t *testing.T) {
		t.Parallel()
		gormDB := newTestDB(t)
		params := pagination.Params{Page: 1, Limit: 10, Sort: "password", Order: "asc"}
		result := params.Apply(gormDB, []string{"name", "created_at"})
		require.NotNil(t, result)
	})

	t.Run("no sort", func(t *testing.T) {
		t.Parallel()
		gormDB := newTestDB(t)
		params := pagination.Params{Page: 1, Limit: 5}
		result := params.Apply(gormDB, []string{"name"})
		require.NotNil(t, result)
	})
}

func TestNewMeta(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		page      int
		limit     int
		total     int64
		wantPages int
	}{
		{"normal", 1, 10, 25, 3},
		{"exact division", 2, 10, 20, 2},
		{"zero limit", 1, 0, 10, 0},
		{"zero total", 1, 10, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			meta := pagination.NewMeta(tt.page, tt.limit, tt.total)
			require.Equal(t, tt.page, meta.Page)
			require.Equal(t, tt.limit, meta.Limit)
			require.Equal(t, tt.total, meta.Total)
			require.Equal(t, tt.wantPages, meta.TotalPages)
		})
	}
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
