package cache

import (
	"fmt"
	"sync"
	"testing"
)

const (
	CACHE_SIZE = 256
)

type TestCache interface {
	Cache

	candidate() string
}

func GenerateTests(create func() TestCache) func(t *testing.T) {
	return func(t *testing.T) {
		cache := create()

		cache.Set("1", 1)
		if cache.Get("1").Key != "1" {
			t.Fatal("set/get error")
		}

		cache.Del("1")
		if cache.Get("1") != nil {
			t.Fatal("del error")
		}

		for i := 0; i < CACHE_SIZE*2; i++ {
			cache.Set(fmt.Sprintf("%d", i), i)
		}

		if cache.candidate() != fmt.Sprintf("%d", CACHE_SIZE) {
			t.Fatal("eviction error")
		}
	}
}

func TestNaiveCache(t *testing.T) {
	GenerateTests(func() TestCache { return NewNaiveCache(CACHE_SIZE) })(t)
}

func TestWrappedCache(t *testing.T) {
	GenerateTests(func() TestCache { return NewWrappedCache(CACHE_SIZE) })(t)
}

func GenerateBenchmarks(create func() Cache) func(b *testing.B) {
	return func(b *testing.B) {
		b.Run("singular", func(b *testing.B) {
			cache := create()
			cache.Set("1", 1)
			b.SetBytes(1)
			for n := 0; n < b.N; n++ {
				cache.Get("1")
			}
		})
		b.Run("parallel", func(b *testing.B) {
			cache := create()
			cache.Set("1", 1)
			b.SetParallelism(5)
			b.SetBytes(1)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					cache.Get("1")
				}
			})
		})
	}
}

func BenchmarkNaiveCache(b *testing.B) {
	GenerateBenchmarks(func() Cache { return NewNaiveCache(CACHE_SIZE) })(b)
}

func BenchmarkWrappedCache(b *testing.B) {
	GenerateBenchmarks(func() Cache { return NewWrappedCache(CACHE_SIZE) })(b)
}

func BenchmarkSampledCache(b *testing.B) {
	GenerateBenchmarks(func() Cache { return NewSampledCache(CACHE_SIZE) })(b)
}

func BenchmarkSyncMap(b *testing.B) {
	m := &sync.Map{}

	m.Store("1", 1)

	b.SetBytes(1)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Load("1")
		}
	})
}
