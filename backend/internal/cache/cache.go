package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Value      []byte
	Expiration time.Time
}

// InMemoryCache is a simple in-memory cache (fallback when Redis is not available)
// For production, replace with Redis implementation
type InMemoryCache struct {
	mu    sync.RWMutex
	items map[string]CacheItem
}

// CacheService interface for cache operations
type CacheService interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteByPrefix(ctx context.Context, prefix string) error
}

// Global cache instance
var cache *InMemoryCache
var once sync.Once

// GetCache returns the singleton cache instance
func GetCache() CacheService {
	once.Do(func() {
		cache = &InMemoryCache{
			items: make(map[string]CacheItem),
		}
		// Start cleanup goroutine
		go cache.cleanup()
	})
	return cache
}

// Get retrieves a value from cache
func (c *InMemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	if time.Now().After(item.Expiration) {
		return nil, fmt.Errorf("key expired: %s", key)
	}

	return item.Value, nil
}

// Set stores a value in cache with TTL
func (c *InMemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	c.items[key] = CacheItem{
		Value:      data,
		Expiration: time.Now().Add(ttl),
	}

	return nil
}

// Delete removes a key from cache
func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

// DeleteByPrefix removes all keys with given prefix
func (c *InMemoryCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.items, key)
		}
	}
	return nil
}

// cleanup removes expired items periodically
func (c *InMemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.Expiration) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// Cache key constants
const (
	KeyTags       = "cache:tags"
	KeyCategories = "cache:categories"
	KeyStats      = "cache:stats"
	KeyPrompts    = "cache:prompts"
)

// TTL constants
const (
	TTLTags       = 5 * time.Minute
	TTLCategories = 5 * time.Minute
	TTLStats      = 1 * time.Minute
	TTLPrompts    = 2 * time.Minute
)
