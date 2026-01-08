package hasher_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/hasher"
)

func TestHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: "password123",
		},
		{
			name:     "complex password",
			password: "P@ssw0rd!#$%^&*()",
		},
		{
			name:     "unicode password",
			password: "密码测试",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, err := hasher.Hash(tt.password)

			require.NoError(t, err)
			require.NotEmpty(t, hash)
			require.NotEqual(t, tt.password, hash)
			require.True(t, strings.HasPrefix(hash, "$2a$"))
		})
	}
}

func TestHash_DifferentSalt(t *testing.T) {
	t.Parallel()

	password := "samepassword"

	hash1, err1 := hasher.Hash(password)
	require.NoError(t, err1)

	hash2, err2 := hasher.Hash(password)
	require.NoError(t, err2)

	// Same password should produce different hashes due to salt
	require.NotEqual(t, hash1, hash2)
}

func TestCheck(t *testing.T) {
	t.Parallel()

	password := "testpassword123"
	hash, err := hasher.Hash(password)
	require.NoError(t, err)

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "correct password",
			password: password,
			want:     true,
		},
		{
			name:     "wrong password",
			password: "wrongpassword",
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			want:     false,
		},
		{
			name:     "similar password",
			password: "testpassword12",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := hasher.Check(tt.password, hash)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCheckWithError(t *testing.T) {
	t.Parallel()

	password := "testpassword123"
	hash, err := hasher.Hash(password)
	require.NoError(t, err)

	t.Run("correct password", func(t *testing.T) {
		t.Parallel()
		err := hasher.CheckWithError(password, hash)
		require.NoError(t, err)
	})

	t.Run("wrong password", func(t *testing.T) {
		t.Parallel()
		err := hasher.CheckWithError("wrongpassword", hash)
		require.Error(t, err)
	})
}

func TestHashWithCost(t *testing.T) {
	t.Parallel()

	password := "testpassword"

	// Low cost for testing
	hash, err := hasher.HashWithCost(password, 4)
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	// Verify the hash works
	require.True(t, hasher.Check(password, hash))
}

func BenchmarkHash(b *testing.B) {
	password := "benchmarkpassword"
	for b.Loop() {
		_, _ = hasher.HashWithCost(password, 4) //nolint:errcheck // benchmark
	}
}

func BenchmarkCheck(b *testing.B) {
	password := "benchmarkpassword"
	hash, _ := hasher.HashWithCost(password, 4) //nolint:errcheck // benchmark setup
	for b.Loop() {
		_ = hasher.Check(password, hash)
	}
}
