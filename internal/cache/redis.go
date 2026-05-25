package cache

import (
	"context"
	"sync"
	"time"
)

type Item struct {
	Value     []byte
	ExpiresAt time.Time
}

type Cache struct {
	mu    sync.RWMutex
	items map[string]Item
}

func NewCache() *Cache {
	return &Cache{items: map[string]Item{}}
}

func (c *Cache) Get(_ context.Context, key string) ([]byte, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if !item.ExpiresAt.IsZero() && item.ExpiresAt.Before(time.Now().UTC()) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, false
	}
	return item.Value, true
}

func (c *Cache) Set(_ context.Context, key string, value []byte, ttl time.Duration) {
	expiresAt := time.Time{}
	if ttl > 0 {
		expiresAt = time.Now().UTC().Add(ttl)
	}
	c.mu.Lock()
	c.items[key] = Item{Value: value, ExpiresAt: expiresAt}
	c.mu.Unlock()
}

func (c *Cache) Delete(_ context.Context, key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}
