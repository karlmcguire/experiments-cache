package fcl

// TODO: FreeCache-like cache implementation

type Cache struct {
}

type Segment struct {
}

func New(segments uint64) *Cache {
	return &Cache{}
}

func (c *Cache) Get(key string) []byte {
	return nil
}

func (c *Cache) Set(key string, val []byte) (string, []byte) {
	return "", nil
}

func (c *Cache) Del(key string) []byte {
	return nil
}
