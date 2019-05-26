package cache

import (
	"testing"
)

func TestCache(t *testing.T) {
	c := NewCache(4)

	c.Set("1", 1)
	c.Set("2", 2)
	c.Set("3", 3)
	c.Set("4", 4)
}

func BenchmarkCache(b *testing.B) {
	c := NewCache(16)
	c.Set("1", 1)

	b.SetBytes(1)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Get("1")
		}
	})
}
