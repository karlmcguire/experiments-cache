package cache

import (
	"container/list"
	"sync"
	"sync/atomic"

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
		data   *sync.Map
		lru    *list.List
		lruMu  sync.Mutex
		access *ring.Buffer
		size   int
	}
)

func NewWrappedCache(size int) *WrappedCache {
	cache := &WrappedCache{
		data: &sync.Map{},
		lru:  list.New(),
		size: size,
	}

	cache.access = ring.NewBuffer(ring.LOSSY, &ring.Config{
		Consumer: cache,
		Capacity: size * 64,
	})

	return cache
}

func (c *WrappedCache) Push(keys []ring.Element) {
	c.lruMu.Lock()
	defer c.lruMu.Unlock()

	for _, key := range keys {
		if element, exists := c.data.Load(string(key)); exists {
			c.lru.MoveToFront(element.(*list.Element))
		}
	}
}

func (c *WrappedCache) Get(key string) *Value {
	element, exists := c.data.Load(key)
	if !exists {
		return nil
	}

	// get value from list element
	value := element.(*list.Element).Value.(*Value)

	// record access in buffer
	c.access.Push(ring.Element(value.Key))

	return value
}

func (c *WrappedCache) Set(key string, data interface{}) {
	c.lruMu.Lock()
	defer c.lruMu.Unlock()
	// check if eviction is needed
	if c.lru.Len() == c.size {
		// eviction is needed, get the victim
		victim := c.lru.Back()
		// remove the victim from lru list
		c.lru.Remove(victim)
		// remove the victim from data store
		c.data.Delete(victim.Value.(*Value).Key)
	}

	c.data.Store(key, c.lru.PushFront(&Value{key, data}))
}

func (c *WrappedCache) Del(key string) {
	element, exists := c.data.Load(key)
	if !exists {
		return
	}

	c.data.Delete(key)

	c.lruMu.Lock()
	defer c.lruMu.Unlock()
	// remove from list
	c.lru.Remove(element.(*list.Element))
}

func (c *WrappedCache) candidate() string {
	return c.lru.Back().Value.(*Value).Key
}

type (
	LockFreeNode struct {
		key  string
		next *LockFreeNode
		prev *LockFreeNode
	}

	LockFreeList struct {
		head *LockFreeNode
		tail *LockFreeNode
	}

	LockFreeCache struct {
		sync.Mutex
		data map[string]atomic.Value
		lru  *LockFreeList
		size int
	}
)

func (l *LockFreeList) Offer(key string) {
}

func (l *LockFreeList) Candidate() string {
	return ""
}

func (l *LockFreeList) Clean() {
}

func (l *LockFreeList) Delete(key string) {
}

func (l *LockFreeList) Evict(victims int) []string {
	return nil
}

func NewLockFreeCache(size int) *LockFreeCache {
	return &LockFreeCache{
		data: make(map[string]atomic.Value),
		lru:  &LockFreeList{},
		size: size,
	}
}

func (c *LockFreeCache) Get(key string) *Value {
	node, exists := c.data[key]
	if !exists {
		return nil
	}

	// atomically load and make sure != nil before conversion
	data := node.Load()
	if data == nil {
		return nil
	}

	value := data.(*Value)

	// offer value to the tail of list (MRU position)
	c.lru.Offer(key)

	return value
}

func (c *LockFreeCache) Set(key string, data interface{}) {
	c.Lock()
	defer c.Unlock()

	if _, exists := c.data[key]; exists {
		return
	}

	// offer value to tail of list (MRU position)
	c.lru.Offer(key)

	// check if eviction needed
	if len(c.data) >= c.size {
		victims := c.lru.Evict(len(c.data) - c.size)
		// delete victims from map
		for _, victim := range victims {
			delete(c.data, victim)
		}
	}

	// create new atomic value and save to map
	var value atomic.Value
	value.Store(&Value{key, data})
	c.data[key] = value
}

func (c *LockFreeCache) Del(key string) {
	c.Lock()
	defer c.Unlock()

	if _, exists := c.data[key]; !exists {
		return
	}

	delete(c.data, key)
	c.lru.Delete(key)
}

func (c *LockFreeCache) candidate() string {
	return c.lru.Candidate()
}

type (
	SampledValue struct {
		Value *Value
		Used  int64
	}

	SampledCache struct {
		key  []byte
		data map[string]*Value
	}
)

func NewSampledCache(size int) *SampledCache {
	return &SampledCache{
		key: []byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		},
		data: make(map[string]*Value),
	}
}

func (c *SampledCache) Get(key string) *Value {
	//hash := highwayhash.Sum64([]byte(key), c.key)
	//atomic.LoadUint64(&hash)

	return c.data[key]
}

func (c *SampledCache) Set(key string, data interface{}) {
	c.data[key] = &Value{key, data}
}

func (c *SampledCache) Del(key string) {
	delete(c.data, key)
}

func (c *SampledCache) candidate() string {
	return ""
}
