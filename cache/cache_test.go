package cache

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"

	"github.com/xba/stress"
)

const (
	KEYS       = 100000
	CACHE_SIZE = 256
	PARA_MULTI = 1
)

func init() {
	fmt.Printf("goroutines: %d\n", runtime.GOMAXPROCS(0)*PARA_MULTI)
}

func zipfKeys() []string {
	size := uint64(KEYS)
	keys := make([]string, size)
	zipf := stress.GenerateZipf(2.1, 1, size)

	i := 0
	for n, err := zipf(); err != stress.ErrDone; n, err = zipf() {
		keys[i] = fmt.Sprintf("%d", n)
	}

	return keys
}

////////////////////////////////////////////////////////////////////////////////

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

func TestMapCache(t *testing.T) {
	GenerateTests(func() TestCache { return NewMapCache(CACHE_SIZE) })(t)
}

func TestMapWrapCache(t *testing.T) {
	GenerateTests(func() TestCache { return NewMapWrapCache(CACHE_SIZE) })(t)
}

////////////////////////////////////////////////////////////////////////////////

func GenerateBenchmarks(create func() Cache) func(b *testing.B) {
	return func(b *testing.B) {
		b.Run("singular", func(b *testing.B) {
			cache := create()
			cache.Set("1", 1)

			b.SetBytes(1)
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				cache.Get("1")
			}
		})
		b.Run("parallel", func(b *testing.B) {
			cache := create()
			cache.Set("1", 1)

			b.SetParallelism(PARA_MULTI)
			b.SetBytes(1)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					cache.Get("1")
				}
			})
		})
	}
}

func GenerateBenchmarksZipf(create func() Cache) func(b *testing.B) {
	return func(b *testing.B) {
		b.Run("singular", func(b *testing.B) {
			cache := create()
			keys := zipfKeys()
			for _, key := range keys {
				cache.Set(key, nil)
			}
			mask := len(keys) - 1

			b.SetBytes(1)
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				cache.Get(keys[n&mask])
			}
		})
		b.Run("parallel", func(b *testing.B) {
			cache := create()
			keys := zipfKeys()
			for _, key := range keys {
				cache.Set(key, nil)
			}
			mask := len(keys) - 1

			b.SetParallelism(PARA_MULTI)
			b.SetBytes(1)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				index := rand.Int() & mask

				for pb.Next() {
					cache.Get(keys[index&mask])
					index++
				}
			})
		})
	}
}

////////////////////////////////////////////////////////////////////////////////

func BenchmarkMapCache(b *testing.B) {
	GenerateBenchmarks(func() Cache {
		return NewMapCache(CACHE_SIZE)
	})(b)
}

func BenchmarkMapCacheZipf(b *testing.B) {
	GenerateBenchmarksZipf(func() Cache {
		return NewMapCache(CACHE_SIZE)
	})(b)
}

////////////////////////////////////////////////////////////////////////////////

func BenchmarkMapWrapCache(b *testing.B) {
	GenerateBenchmarks(func() Cache {
		return NewMapWrapCache(CACHE_SIZE)
	})(b)
}

func BenchmarkMapWrapCacheZipf(b *testing.B) {
	GenerateBenchmarksZipf(func() Cache {
		return NewMapWrapCache(CACHE_SIZE)
	})(b)
}

////////////////////////////////////////////////////////////////////////////////

func BenchmarkSyncMap(b *testing.B) {
	GenerateBenchmarks(func() Cache {
		return NewSyncMap(CACHE_SIZE)
	})(b)
}

func BenchmarkSyncMapZipf(b *testing.B) {
	GenerateBenchmarksZipf(func() Cache {
		return NewSyncMap(CACHE_SIZE)
	})(b)
}

////////////////////////////////////////////////////////////////////////////////

func BenchmarkSyncMapWrap(b *testing.B) {
	GenerateBenchmarks(func() Cache {
		return NewSyncMapWrap(CACHE_SIZE)
	})(b)
}

func BenchmarkSyncMapWrapZipf(b *testing.B) {
	GenerateBenchmarksZipf(func() Cache {
		return NewSyncMapWrap(CACHE_SIZE)
	})(b)
}

////////////////////////////////////////////////////////////////////////////////

func BenchmarkFastCache(b *testing.B) {
	GenerateBenchmarks(func() Cache {
		return NewFastCache(CACHE_SIZE)
	})(b)
}

func BenchmarkFastCacheZipf(b *testing.B) {
	GenerateBenchmarksZipf(func() Cache {
		return NewFastCache(CACHE_SIZE)
	})(b)
}

////////////////////////////////////////////////////////////////////////////////

func BenchmarkBigCache(b *testing.B) {
	GenerateBenchmarks(func() Cache {
		return NewBigCache(CACHE_SIZE)
	})(b)
}

func BenchmarkBigCacheZipf(b *testing.B) {
	GenerateBenchmarksZipf(func() Cache {
		return NewBigCache(CACHE_SIZE)
	})(b)
}

////////////////////////////////////////////////////////////////////////////////
