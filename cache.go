package cache

import (
	"sync/atomic"
	"time"

	"github.com/karlmcguire/experiments-cache/snap"
)

type (
	Cache struct {
		data     snap.Map
		size     uint64
		capacity uint64
	}

	value struct {
		key  string
		data interface{}
		meta metadata
	}

	metadata struct {
		count   uint64
		created uint64
	}
)

func NewCache(capacity uint64) *Cache {
	return &Cache{
		data:     snap.NewSyncMap(),
		capacity: capacity,
	}
}

func (c *Cache) Evict() func(interface{}, interface{}) bool {
	i := 0

	return func(key, value interface{}) bool {
		if i == 3 {
			return false
		}

		// do stuff

		i++
		return true
	}
}

func (c *Cache) Set(key string, data interface{}) {
	if c.data.Get(key) == nil {
		if atomic.AddUint64(&c.size, 1) == c.capacity {
			// evict elements
			c.data.Range(c.Evict())
		}
	}

	c.data.Set(key, &value{
		key:  key,
		data: data,
		meta: metadata{
			count:   1,
			created: uint64(time.Now().UnixNano()),
		},
	})
}
