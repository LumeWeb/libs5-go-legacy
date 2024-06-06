package db

import (
	"github.com/hashicorp/golang-lru"
)

// LRUCache is an implementation of the Cache interface using an LRU cache.
type LRUCache struct {
	cache *lru.Cache
}

// NewLRUCache creates a new instance of LRUCache with the specified size.
func NewLRUCache(size int) (*LRUCache, error) {
	cache, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	return &LRUCache{cache: cache}, nil
}

// Get retrieves the value associated with the given key from the cache.
// It returns the value and a boolean indicating whether the key was found.
func (c *LRUCache) Get(key []byte) ([]byte, bool) {
	value, ok := c.cache.Get(string(key))
	if !ok {
		return nil, false
	}
	return value.([]byte), true
}

// Put adds a key-value pair to the cache.
func (c *LRUCache) Put(key []byte, value []byte) {
	c.cache.Add(string(key), value)
}

// Delete removes the value associated with the given key from the cache.
func (c *LRUCache) Delete(key []byte) {
	c.cache.Remove(string(key))
}
