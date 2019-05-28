package cache

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/karlmcguire/experiments-cache/ring"
	"github.com/karlmcguire/experiments-cache/snap"
)

type (
	Cache struct {
		sync.Mutex
		data     snap.Map
		buffer   *ring.Buffer
		size     uint64
		capacity uint64
		sample   uint64
		test     []byte
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
	// calculate eviction sample size
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
	value := c.data.Get(key).(*Value)
	value.meta.count++
	return value
}

func (c *Cache) Evict() {
	var (
		minKey   interface{}
		minScore float64
		i        uint64
	)

	c.data.Range(func(key, value interface{}) bool {
		var (
			meta    = &value.(*Value).meta
			count   = float64(atomic.LoadUint64(&meta.count))
			created = float64(meta.created)
			score   = count / created
		)

		// keep track of the smallest element score (potential victim)
		if i == 0 || score < minScore {
			minKey = key
			minScore = score
		}

		i++
		// keep iterating until we reach our sample size maximum set at init
		return i < c.sample
	})

	// delete victim
	c.data.Del(minKey.(string))
}

func (c *Cache) Set(key string, data interface{}) {
	if c.data.Get(key) == nil {
		if atomic.AddUint64(&c.size, 1) == c.capacity {
			// we're at full capacity so evict an element
			c.Evict()
		}
	}

	// add to the cache
	c.data.Set(key, &Value{
		key:  key,
		data: data,
		meta: metadata{
			count:   1,
			created: uint64(time.Now().UnixNano()),
		},
	})
}
