package cache

import (
	"container/list"
	"sync"

	"github.com/karlmcguire/ring"
)

type (
	Cache interface {
		Get(string) *Value
		Set(string, interface{})
		Del(string)
	}

	Value struct {
		Key  string
		Data interface{}
	}
)

type (
	NaiveCache struct {
		sync.Mutex
		data map[string]*list.Element
		lru  *list.List
		size int
	}
)

func NewNaiveCache(size int) *NaiveCache {
	return &NaiveCache{
		data: make(map[string]*list.Element),
		lru:  list.New(),
		size: size,
	}
}

func (c *NaiveCache) Get(key string) *Value {
	c.Lock()
	defer c.Unlock()

	// check if list element exists in data store
	element, exists := c.data[key]
	if !exists {
		return nil
	}

	// maintain access order
	c.lru.MoveToFront(element)

	// return actual value inside list element
	return element.Value.(*Value)
}

func (c *NaiveCache) Set(key string, data interface{}) {
	c.Lock()
	defer c.Unlock()

	// element already exists
	if _, exists := c.data[key]; exists {
		return
	}

	// check if eviction is needed
	if len(c.data) == c.size {
		// eviction is needed, get the victim
		victim := c.lru.Back()
		// remove the victim from lru list
		c.lru.Remove(victim)
		// remove the victim from data store
		delete(c.data, victim.Value.(*Value).Key)
	}

	// add new element
	c.data[key] = c.lru.PushFront(&Value{key, data})
}

func (c *NaiveCache) Del(key string) {
	c.Lock()
	defer c.Unlock()

	element, exists := c.data[key]
	if !exists {
		return
	}

	// remove from list
	c.lru.Remove(element)
	// remove from data store
	delete(c.data, key)
}

func (c *NaiveCache) candidate() string {
	return c.lru.Back().Value.(*Value).Key
}

type (
	WrappedCache struct {
		data   map[string]*list.Element
		dataMu sync.RWMutex
		lru    *list.List
		lruMu  sync.Mutex
		access *ring.Buffer
		size   int
	}
)

func NewWrappedCache(size int) *WrappedCache {
	cache := &WrappedCache{
		data: make(map[string]*list.Element),
		lru:  list.New(),
		size: size,
	}

	cache.access = ring.NewBuffer(ring.LOSSY, &ring.Config{
		Consumer: cache,
		Capacity: size / 16,
	})

	return cache
}

func (c *WrappedCache) Push(keys []ring.Element) {
	c.lruMu.Lock()
	defer c.lruMu.Unlock()

	for _, key := range keys {
		if element, exists := c.data[string(key)]; exists {
			c.lru.MoveToFront(element)
		}
	}
}

func (c *WrappedCache) Get(key string) *Value {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()

	element, exists := c.data[key]
	if !exists {
		return nil
	}

	// get value from list element
	value := element.Value.(*Value)

	// record access in buffer
	c.access.Push(ring.Element(value.Key))

	return value
}

func (c *WrappedCache) Set(key string, data interface{}) {
	c.dataMu.Lock()
	defer c.dataMu.Unlock()

	if _, exists := c.data[key]; exists {
		return
	}

	c.lruMu.Lock()
	defer c.lruMu.Unlock()
	// check if eviction is needed
	if len(c.data) == c.size {
		// eviction is needed, get the victim
		victim := c.lru.Back()
		// remove the victim from lru list
		c.lru.Remove(victim)
		// remove the victim from data store
		delete(c.data, victim.Value.(*Value).Key)
	}

	c.data[key] = c.lru.PushFront(&Value{key, data})
}

func (c *WrappedCache) Del(key string) {
	c.dataMu.Lock()
	defer c.dataMu.Unlock()
	c.lruMu.Lock()
	defer c.lruMu.Unlock()

	element, exists := c.data[key]
	if !exists {
		return
	}

	// remove from list
	c.lru.Remove(element)
	// remove from data store
	delete(c.data, key)
}

func (c *WrappedCache) candidate() string {
	return c.lru.Back().Value.(*Value).Key
}
