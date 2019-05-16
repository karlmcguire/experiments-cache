package bcl

import "sync"

type Shard struct {
	sync.RWMutex

	data map[uint64]uint32
}

func (s *Shard) Get(key string, hashed uint64) ([]byte, error) {
	return nil, nil
}

func (s *Shard) Set(key string, hashed uint64, value []byte) error {
	return nil
}

func (s *Shard) Del(key string, hashed uint64) error {
	return nil
}
