package bcl

// TODO: BigCache-like cache implementation

import (
	"github.com/karlmcguire/experiments-cache/util"
)

type Cache struct {
	Shards []*Shard
}

func New() *Cache {
	return &Cache{
		Shards: make([]*Shard, util.Near(n)),
	}
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
