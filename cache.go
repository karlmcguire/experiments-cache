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
		sample   uint64
	}

	Value struct {
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
	sample := uint64(4)
	if sample < capacity {
		sample = capacity
	}

	return &Cache{
		data:     snap.NewSyncMap(),
		capacity: capacity,
		sample:   sample,
	}
}

func (c *Cache) Get(key string) interface{} {
	value := c.data.Get(key)
	// increment request counter
	atomic.AddUint64(&value.(*Value).meta.count, 1)
	return value
}

func (c *Cache) Evict() {
	min := struct {
		key   interface{}
		score float64
	}{}
	i := uint64(0)

	c.data.Range(func(key, value interface{}) bool {
		var (
			meta    = &value.(*Value).meta
			count   = float64(atomic.LoadUint64(&meta.count))
			created = float64(meta.created)
			score   = count / created
		)

		if i == 0 || score < min.score {
			min.key = key
			min.score = score
		}

		i++
		return i < c.sample
	})

	// delete victim
	c.data.Del(min.key.(string))
}

func (c *Cache) Set(key string, data interface{}) {
	if c.data.Get(key) == nil {
		if atomic.AddUint64(&c.size, 1) == c.capacity {
			c.Evict()
		}
	}

	c.data.Set(key, &Value{
		key:  key,
		data: data,
		meta: metadata{
			count:   1,
			created: uint64(time.Now().UnixNano()),
		},
	})
}
