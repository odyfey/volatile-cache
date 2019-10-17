package store

import (
	"sync"
	"time"
)

// todo: Save & Load methods
type Store interface {
	Set(key, value string, exp time.Duration)
	Get(key string) (string, bool)
	Delete(key string)
	DeleteExpired()
}

type VolatileCache struct {
	sync.RWMutex
	defaultExpiration time.Duration
	items             map[string]Item
	watcher           Watcher
}

type Item struct {
	Value      string
	Expiration int64
}

func NewCache(exp time.Duration) *VolatileCache {
	w := NewWatcher(50 * time.Millisecond) // todo: env config
	c := &VolatileCache{
		defaultExpiration: exp,
		items:             make(map[string]Item),
		watcher:           w,
	}

	go c.watcher.Start(c)
	return c
}

func (c *VolatileCache) Set(key, value string, exp time.Duration) {
	c.Lock()
	defer c.Unlock()

	if exp == 0 {
		exp = c.defaultExpiration
	}

	c.items[key] = Item{
		Value:      value,
		Expiration: time.Now().Add(exp).UnixNano(),
	}
}

func (c *VolatileCache) Get(key string) (string, bool) {
	c.RLock()
	defer c.Unlock()

	var result string
	item, ok := c.items[key]
	if ok {
		if time.Now().UnixNano() <= item.Expiration {
			result = item.Value
		}
	}
	return result, ok
}

func (c *VolatileCache) Delete(key string) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.items[key]; ok {
		delete(c.items, key)
	}
}

// ?
func (i Item) Expired() bool {
	return time.Now().UnixNano() > i.Expiration
}

func (c *VolatileCache) DeleteExpired() {
	c.Lock()
	defer c.Unlock()

	now := time.Now().UnixNano()
	for key, value := range c.items {
		if now > value.Expiration {
			delete(c.items, key)
		}
	}
}
