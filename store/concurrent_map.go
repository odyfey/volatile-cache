package store

import (
	"encoding/gob"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type ConcurrentTTLMap struct {
	sync.RWMutex
	defaultExpiration time.Duration
	items             map[string]Item
	stopWatcher       chan bool
}

type Item struct {
	Value      string
	Expiration int64
}

func NewConcurrentMap(exp time.Duration) *ConcurrentTTLMap {
	c := &ConcurrentTTLMap{
		defaultExpiration: exp,
		items:             make(map[string]Item),
		stopWatcher:       make(chan bool),
	}

	go c.runWatcher(50 * time.Millisecond) // todo: env config
	return c
}

func (c *ConcurrentTTLMap) Set(key, value string, exp time.Duration) {
	if exp == 0 {
		exp = c.defaultExpiration
	}

	c.Lock()
	c.items[key] = Item{
		Value:      value,
		Expiration: time.Now().Add(exp).UnixNano(),
	}
	c.Unlock()
}

func (c *ConcurrentTTLMap) SetPreparedItem(key string, item Item) {
	c.Lock()
	c.items[key] = item
	c.Unlock()
}

func (c *ConcurrentTTLMap) Get(key string) (string, bool) {
	var result string

	c.RLock()
	item, ok := c.items[key]
	c.RUnlock()

	if ok {
		if time.Now().UnixNano() <= item.Expiration {
			result = item.Value
		}
	}
	return result, ok
}

func (c *ConcurrentTTLMap) Delete(key string) {
	c.Lock()
	if _, ok := c.items[key]; ok {
		delete(c.items, key)
	}
	c.Unlock()
}

func (c *ConcurrentTTLMap) Save(w io.Writer) (err error) {
	enc := gob.NewEncoder(w)

	c.RLock()
	if err := enc.Encode(&c.items); err != nil {
		err = errors.Wrap(err, "can't encode items")
	}
	c.RUnlock()
	return
}

func (c *ConcurrentTTLMap) Load(r io.Reader) error {
	dec := gob.NewDecoder(r)
	items := make(map[string]Item)
	if err := dec.Decode(&items); err != nil {
		return errors.Wrapf(err, "can't decode items")
	}

	c.Lock()
	for key, value := range items {
		_, ok := c.items[key]
		if !ok {
			c.items[key] = value
		}

	}
	c.Unlock()
	return nil
}

func (c *ConcurrentTTLMap) Items() map[string]Item {
	result := make(map[string]Item, len(c.items))
	now := time.Now().UnixNano()
	for key, value := range c.items {
		if now > value.Expiration {
			continue
		}
		result[key] = value
	}
	return result
}

func (c *ConcurrentTTLMap) len() int {
	return len(c.items)
}

func (c *ConcurrentTTLMap) deleteExpired() {
	now := time.Now().UnixNano()

	c.Lock()
	for key, value := range c.items {
		if now > value.Expiration {
			delete(c.items, key)
		}
	}
	c.Unlock()
}

func (c *ConcurrentTTLMap) runWatcher(interval time.Duration) {
	tick := time.NewTicker(interval)
	for {
		select {
		case <-tick.C:
			c.deleteExpired()
		case <-c.stopWatcher:
			tick.Stop()
			return
		}
	}
}

func (c *ConcurrentTTLMap) pauseWatcher() {
	c.stopWatcher <- true
}
