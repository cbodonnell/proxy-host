// implement a simple, thread-safe key-value cache
package cache

import (
	"sync"
	"time"
)

// Cache is a simple, thread-safe key-value cache
type Cache struct {
	// items contains all the items stored in the cache
	items map[string]Item
	// mutex is used to synchronize access to the cache
	mutex sync.RWMutex
	// defaultExpiration specifies the default expiration time of an item
	defaultExpiration time.Duration
	// cleanupInterval specifies how often the cache should be cleaned
	cleanupInterval time.Duration
	// stopCleanup is used to stop the background cleanup process
	stopCleanup chan bool
}

// Item represents a cache item
type Item struct {
	// value is the value stored in the cache
	value interface{}
	// expiration specifies how long the item is valid
	expiration int64
}

// NewCache creates a new cache with the specified default expiration and cleanup interval
func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	items := make(map[string]Item)
	cache := Cache{
		items:             items,
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		stopCleanup:       make(chan bool),
	}
	cache.startCleanupTimer()
	return &cache
}

// Set adds a new item to the cache. If the item already exists, it will be overwritten
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	var expiration int64
	if duration == 0 {
		duration = c.defaultExpiration
	}
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}
	c.items[key] = Item{
		value:      value,
		expiration: expiration,
	}
}

// Get returns the value of the item with the specified key. If the item does not exist or is expired,
// nil will be returned instead
func (c *Cache) Get(key string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	item, found := c.items[key]
	if !found {
		return nil
	}
	if item.expiration > 0 {
		if time.Now().UnixNano() > item.expiration {
			return nil
		}
	}
	return item.value
}

// Delete removes the item with the specified key from the cache
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.items, key)
}

func (c *Cache) Extend(key string, duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	item, found := c.items[key]
	if !found {
		return
	}
	if duration == 0 {
		duration = c.defaultExpiration
	}
	if duration > 0 {
		item.expiration = time.Now().Add(duration).UnixNano()
	}
	c.items[key] = item
}

// startCleanupTimer starts a background goroutine that cleans up the cache at the specified
// cleanup interval
func (c *Cache) startCleanupTimer() {
	ticker := time.NewTicker(c.cleanupInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.deleteExpiredItems()
			case <-c.stopCleanup:
				ticker.Stop()
				return
			}
		}
	}()
}

// deleteExpiredItems deletes all expired items from the cache
func (c *Cache) deleteExpiredItems() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for key, item := range c.items {
		if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
			delete(c.items, key)
		}
	}
}

// StopCleanup stops the background cleanup process
func (c *Cache) StopCleanup() {
	c.stopCleanup <- true
}
