package cache

import "github.com/karlmcguire/experiments-cache/pkg/store"

type (
	Cache struct {
		Store store.Store
	}

	Config struct {
	}
)

func New(config *Config) (*Cache, error) {
	return nil, nil
}
