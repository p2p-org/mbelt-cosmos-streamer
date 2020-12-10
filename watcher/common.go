package watcher

import (
	"reflect"
	"sync"
	"time"
)

const cacheTicker = time.Second * 31
const cacheExpiration = time.Minute * 2

type CacheWatcher struct {
	mu          sync.RWMutex
	heights     []int64
	storage     map[int64]time.Time
	chanHeights chan int64
}

func (c *CacheWatcher) InitCache() {
	c.storage = make(map[int64]time.Time)
	c.chanHeights = make(chan int64, 10000)
	go c.observeStorage()
}

func (c *CacheWatcher) observeStorage() {
	for {
		select {
		case <-time.Tick(cacheTicker):
			keys := reflect.ValueOf(c.storage).MapKeys()
			for _, key := range keys {
				c.mu.Lock()
				value, _ := c.storage[key.Int()]
				if value.Before(time.Now()) {
					delete(c.storage, key.Int())
				}
				c.mu.Unlock()
			}
		}
	}
}

func (c *CacheWatcher) Store(height int64) {
	c.mu.Lock()
	c.storage[height] = time.Now().Add(cacheExpiration)
	c.mu.Unlock()
	c.chanHeights <- height
}

func (c *CacheWatcher) Subscribe() <-chan int64 {
	return c.chanHeights
}

func (c *CacheWatcher) Count() int {
	return reflect.ValueOf(c.storage).Len()
}
