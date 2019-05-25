package snap

import (
	"sync"
)

type Map interface {
	Get(string) interface{}
	Set(string, interface{})
	Del(string)
}

////////////////////////////////////////////////////////////////////////////////

type SyncMap struct {
	data *sync.Map
}

func NewSyncMap() *SyncMap {
	return &SyncMap{
		data: &sync.Map{},
	}
}

func (m *SyncMap) Get(key string) interface{} {
	value, _ := m.data.Load(key)
	return value
}

func (m *SyncMap) Set(key string, value interface{}) {
	m.data.Store(key, value)
}

func (m *SyncMap) Del(key string) {
	m.data.Delete(key)
}
