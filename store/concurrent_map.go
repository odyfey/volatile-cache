package store

import (
	"encoding/gob"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type concurrentTTLMap struct {
	sync.RWMutex
	defaultExpiration time.Duration
	items             map[string]Item
	stopWatcher       chan bool
}

type Item struct {
	Value      string
	Expiration int64
}

func newConcurrentTTLMap(exp time.Duration) *concurrentTTLMap {
	c := &concurrentTTLMap{
		defaultExpiration: exp,
		items:             make(map[string]Item),
		stopWatcher:       make(chan bool),
	}
	go c.runWatcher(50 * time.Millisecond) // todo: env config
	return c
}

func (c *concurrentTTLMap) Set(key, value string, exp time.Duration) {
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

func (c *concurrentTTLMap) setPreparedItem(key string, item Item) {
	c.Lock()
	c.items[key] = item
	c.Unlock()
}

func (c *concurrentTTLMap) Get(key string) (string, bool) {
	var result string

	c.RLock()
	item, ok := c.items[key]
	if !ok {
		c.RUnlock()
		return result, ok
	}

	if time.Now().UnixNano() <= item.Expiration {
		result = item.Value
	}
	c.RUnlock()
	return result, ok
}

func (c *concurrentTTLMap) Delete(key string) {
	c.Lock()
	if _, ok := c.items[key]; ok {
		delete(c.items, key)
	}
	c.Unlock()
}

func (c *concurrentTTLMap) Save(w io.Writer) error {
	enc := gob.NewEncoder(w)
	err := c.saveToStream(enc)
	return err
}

func (c *concurrentTTLMap) saveToStream(enc *gob.Encoder) (err error) {
	c.RLock()
	if err = enc.Encode(&c.items); err != nil {
		err = errors.Wrap(err, "can't encode items")
	}
	c.RUnlock()
	return
}

func (c *concurrentTTLMap) Load(r io.Reader) error {
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

func (c *concurrentTTLMap) itemsCopy() map[string]Item {
	c.RLock()
	defer c.RUnlock()
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

func (c *concurrentTTLMap) len() int {
	return len(c.items)
}

func (c *concurrentTTLMap) deleteExpired() {
	now := time.Now().UnixNano()

	c.Lock()
	for key, value := range c.items {
		if now > value.Expiration {
			delete(c.items, key)
		}
	}
	c.Unlock()
}

func (c *concurrentTTLMap) runWatcher(interval time.Duration) {
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

func (c *concurrentTTLMap) pauseWatcher() {
	c.stopWatcher <- true
}
