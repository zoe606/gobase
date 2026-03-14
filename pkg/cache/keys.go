package cache

import "fmt"

// KeyBuilder constructs consistent cache keys for an entity type.
// Pattern: entity:scope:id (e.g., "article:list:page=1&size=10").
type KeyBuilder struct {
	entity string
}

// NewKeyBuilder creates a KeyBuilder for the given entity type.
func NewKeyBuilder(entity string) *KeyBuilder {
	return &KeyBuilder{entity: entity}
}

// ID returns a key for a specific entity instance.
func (kb *KeyBuilder) ID(id uint) string {
	return fmt.Sprintf("%s:%d", kb.entity, id)
}

// List returns a key for a list query with the given qualifier.
func (kb *KeyBuilder) List(qualifier string) string {
	return fmt.Sprintf("%s:list:%s", kb.entity, qualifier)
}

// ListPrefix returns the prefix for all list keys.
func (kb *KeyBuilder) ListPrefix() string {
	return kb.entity + ":list:"
}

// Prefix returns the prefix for all keys of this entity.
func (kb *KeyBuilder) Prefix() string {
	return kb.entity + ":"
}
