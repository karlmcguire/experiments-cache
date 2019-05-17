package clhm

import (
	"container/list"

	"github.com/karlmcguire/experiments-cache/pkg/try"
	"github.com/karlmcguire/experiments-cache/pkg/util"
)

const (
	GET uint64 = iota
	SET
	DEL
)

type (
	Config struct {
		BufferCount     uint32
		BufferSize      uint32
		BufferThreshold uint32
		MapSize         uint32
	}

	Buffer struct {
		config *Config
		id     int
		reads  int
		writes int
		blocks []uint64
	}

	Map struct {
		try.Mutex

		config   *Config
		elem     map[uint32]*list.Element
		data     *list.List
		buffers  []*Buffer
		buffMask uint32
	}
)

func New(config *Config) *Map {
	m := &Map{
		elem:     make(map[uint32]*list.Element, config.MapSize),
		data:     list.New(),
		buffers:  make([]*Buffer, config.BufferCount),
		buffMask: config.BufferCount - 1,
	}

	// initialize buffers
	for id := range m.buffers {
		m.buffers[id] = &Buffer{
			config: config,
			id:     id,
			blocks: make([]uint64, config.BufferSize),
		}
	}

	return m
}

func (m *Map) Record(kind uint64, key []byte, hash uint32) {
	// set the MSB to the corresponding access type
	//
	// 00 -> GET
	// 01 -> SET
	// 10 -> DEL
	//
	// this leaves the 32 LSB being the hashed key for later lookup
	block := uint64(hash) | (kind << 62)

	// record the access in the buffer
	m.buffers[hash&m.buffMask].Add(block)
}

func (m *Map) Get(key []byte) interface{} {
	hash := util.Hash32(key)

	if elem, exists := m.elem[hash]; !exists {
		return nil
	} else {
		// record the access to the buffers
		defer m.Record(GET, key, hash)
		return elem.Value
	}
}

func (m *Map) Set(key []byte, value interface{}) ([]byte, interface{}) {
	m.elem[util.Hash32(key)] = m.data.PushFront(value)
	return nil, nil
}

func (m *Map) Del(key []byte) interface{} {
	return nil
}

func (b *Buffer) Add(block uint64) {
	// TODO
	// add to the buffer
	// check if drain is needed
}
