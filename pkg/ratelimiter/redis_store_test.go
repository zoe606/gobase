package ratelimiter_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/ratelimiter"
)

type mockStorage struct {
	store map[string][]byte
}

func newMockStorage() *mockStorage {
	return &mockStorage{store: make(map[string][]byte)}
}

func (m *mockStorage) Get(key string) ([]byte, error) {
	v, ok := m.store[key]
	if !ok {
		return nil, nil
	}

	return v, nil
}

func (m *mockStorage) Set(key string, val []byte, _ time.Duration) error {
	m.store[key] = val
	return nil
}

func (m *mockStorage) Delete(key string) error {
	delete(m.store, key)
	return nil
}

func (m *mockStorage) Reset() error {
	m.store = make(map[string][]byte)
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func TestRedisStore_GetSetDelete(t *testing.T) {
	t.Parallel()

	mock := newMockStorage()
	store := ratelimiter.NewRedisStore(mock)

	// Get non-existent key returns nil.
	val, err := store.Get("key1")
	require.NoError(t, err)
	assert.Nil(t, val)

	// Set stores a value.
	err = store.Set("key1", []byte("value1"), time.Minute)
	require.NoError(t, err)

	// Get existing key returns the value.
	val, err = store.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, []byte("value1"), val)

	// Delete removes the key.
	err = store.Delete("key1")
	require.NoError(t, err)

	val, err = store.Get("key1")
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestRedisStore_Reset(t *testing.T) {
	t.Parallel()

	mock := newMockStorage()
	store := ratelimiter.NewRedisStore(mock)

	_ = store.Set("a", []byte("1"), time.Minute)
	_ = store.Set("b", []byte("2"), time.Minute)

	err := store.Reset()
	require.NoError(t, err)

	val, _ := store.Get("a")
	assert.Nil(t, val)

	val, _ = store.Get("b")
	assert.Nil(t, val)
}

func TestRedisStore_Close(t *testing.T) {
	t.Parallel()

	mock := newMockStorage()
	store := ratelimiter.NewRedisStore(mock)

	err := store.Close()
	require.NoError(t, err)
}
