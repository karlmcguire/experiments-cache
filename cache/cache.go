package cache

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/allegro/bigcache"
	"github.com/karlmcguire/experiments-cache/ring"
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

////////////////////////////////////////////////////////////////////////////////

type (
	MapCache struct {
		sync.RWMutex
		data map[string]*list.Element
		lru  *list.List
		size int
	}
)

func NewMapCache(size int) *MapCache {
	return &MapCache{
		data: make(map[string]*list.Element),
		lru:  list.New(),
		size: size,
	}
}

func (c *MapCache) Get(key string) *Value {
	c.RLock()
	defer c.RUnlock()

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

func (c *MapCache) Set(key string, data interface{}) {
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

func (c *MapCache) Del(key string) {
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

func (c *MapCache) candidate() string {
	return c.lru.Back().Value.(*Value).Key
}

////////////////////////////////////////////////////////////////////////////////

type (
	MapWrapCache struct {
		sync.RWMutex
		data   map[string]*list.Element
		lru    *list.List
		lruMu  sync.Mutex
		access *ring.Buffer
		size   int
	}
)

func NewMapWrapCache(size int) *MapWrapCache {
	cache := &MapWrapCache{
		data: make(map[string]*list.Element, size),
		lru:  list.New(),
		size: size,
	}
	cache.access = ring.NewBuffer(ring.LOSSY, &ring.Config{
		Consumer: cache,
		Capacity: size * 64,
	})
	return cache
}

func (c *MapWrapCache) Push(keys []ring.Element) {
	c.lruMu.Lock()
	defer c.lruMu.Unlock()

	for _, key := range keys {
		if element, exists := c.data[string(key)]; exists {
			c.lru.MoveToFront(element)
		}
	}
}

func (c *MapWrapCache) Get(key string) *Value {
	c.RLock()
	defer c.RUnlock()

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

func (c *MapWrapCache) Set(key string, data interface{}) {
	c.lruMu.Lock()
	defer c.lruMu.Unlock()

	// check if eviction is needed
	if c.lru.Len() == c.size {
		// eviction is needed, get the victim
		victim := c.lru.Back()
		// remove the victim from lru list
		c.lru.Remove(victim)
		// remove the victim from data store
		delete(c.data, victim.Value.(*Value).Key)
	}

	// add new element to store
	c.data[key] = c.lru.PushFront(&Value{key, data})
}

func (c *MapWrapCache) Del(key string) {
	c.Lock()
	defer c.Unlock()

	element, exists := c.data[key]
	if !exists {
		return
	}
	delete(c.data, key)

	c.lruMu.Lock()
	defer c.lruMu.Unlock()
	// remove from list
	c.lru.Remove(element)
}

func (c *MapWrapCache) candidate() string {
	return c.lru.Back().Value.(*Value).Key
}

////////////////////////////////////////////////////////////////////////////////

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

////////////////////////////////////////////////////////////////////////////////

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

////////////////////////////////////////////////////////////////////////////////

type (
	SyncMapValue struct {
		Value *Value
		Count uint64
	}
	SyncMap struct {
		data *sync.Map
	}
)

func NewSyncMap(size int) *SyncMap {
	return &SyncMap{
		data: &sync.Map{},
	}
}

func (c *SyncMap) Get(key string) *Value {
	raw, _ := c.data.Load(key)
	if raw == nil {
		return nil
	}

	value := raw.(*SyncMapValue)
	value.Count++
	return value.Value
}

func (c *SyncMap) Set(key string, data interface{}) {
	c.data.Store(key, &SyncMapValue{
		Value: &Value{key, data},
		Count: 0,
	})
}

func (c *SyncMap) Del(key string) {
	c.data.Delete(key)
}

func (c *SyncMap) candidate() string {
	return ""
}

////////////////////////////////////////////////////////////////////////////////

type (
	SyncMapWrap struct {
		sync.Mutex
		buffer *ring.Buffer
		data   *sync.Map
		counts map[string]uint64
	}
)

func NewSyncMapWrap(size int) *SyncMapWrap {
	cache := &SyncMapWrap{
		data:   &sync.Map{},
		counts: make(map[string]uint64, size),
	}
	cache.buffer = ring.NewBuffer(ring.LOSSY, &ring.Config{
		Consumer: cache,
		Capacity: size * 64,
	})
	return cache
}

func (c *SyncMapWrap) Push(keys []ring.Element) {
	c.Lock()
	defer c.Unlock()
	for _, key := range keys {
		c.counts[string(key)]++
	}
}

func (c *SyncMapWrap) Get(key string) *Value {
	value, _ := c.data.Load(key)
	c.buffer.Push(ring.Element(key))
	return value.(*Value)
}

func (c *SyncMapWrap) Set(key string, data interface{}) {
	c.data.Store(key, &Value{key, data})
}

func (c *SyncMapWrap) Del(key string) {
	c.data.Delete(key)
}

func (c *SyncMapWrap) candidate() string {
	return ""
}

////////////////////////////////////////////////////////////////////////////////

type (
	FastCache struct {
		cache *fastcache.Cache
	}
)

func NewFastCache(size int) *FastCache {
	return &FastCache{
		cache: fastcache.New(size),
	}
}

func (c *FastCache) Get(key string) *Value {
	data := make([]byte, 1)
	c.cache.Get(data, []byte(key))
	return &Value{key, data}
}

func (c *FastCache) Set(key string, data interface{}) {
	// TODO
	c.cache.Set([]byte(key), nil)
}

func (c *FastCache) Del(key string) {
	// TODO
}

func (c *FastCache) candidate() string {
	// TODO
	return ""
}

////////////////////////////////////////////////////////////////////////////////

type (
	BigCache struct {
		cache *bigcache.BigCache
	}
)

func NewBigCache(size int) *BigCache {
	bc, err := bigcache.NewBigCache(
		bigcache.DefaultConfig(time.Second * 30),
	)
	if err != nil {
		panic(err)
	}
	return &BigCache{
		cache: bc,
	}
}

func (c *BigCache) Get(key string) *Value {
	data, _ := c.cache.Get(key)
	return &Value{key, data}
}

func (c *BigCache) Set(key string, data interface{}) {
	// TODO
	if err := c.cache.Set(key, nil); err != nil {
		panic(err)
	}
}

func (c *BigCache) Del(key string) {
	// TODO
}

func (c *BigCache) candidate() string {
	// TODO
	return ""
}
