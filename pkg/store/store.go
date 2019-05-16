package store

import "errors"

var ErrNoValue = errors.New("no value associated with key")

type Store interface {
	Get(string) ([]byte, error)
	Set(string, []byte) error
	Del(string) error
}

type mapStore struct {
	data map[string][]byte
}

func NewMapStore(size int) Store {
	return &mapStore{
		data: make(map[string][]byte, size),
	}
}

func (s *mapStore) Get(key string) ([]byte, error) {
	if value, exists := s.data[key]; !exists {
		return nil, ErrNoValue
	} else {
		return value, nil
	}
}

func (s *mapStore) Set(key string, value []byte) error {
	s.data[key] = value
	return nil
}

func (s *mapStore) Del(key string) error {
	delete(s.data, key)
	return nil
}
