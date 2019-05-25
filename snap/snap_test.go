package snap

import (
	"testing"
)

func GenerateTests(create func() Map) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("get", func(t *testing.T) {
			m := create()
			m.Set("1", 1)
			if m.Get("1").(int) != 1 {
				t.Fatalf("set/get fail")
			}
		})
		t.Run("set", func(t *testing.T) {
			m := create()
			m.Set("1", 1)
			m.Set("1", 2)
			if m.Get("1").(int) != 2 {
				t.Fatal("set/get fail")
			}
		})
		t.Run("del", func(t *testing.T) {
			m := create()
			m.Set("1", 1)
			m.Del("1")
			if m.Get("1") != nil {
				t.Fatal("set/del fail")
			}
		})
	}
}

func TestSyncMap(t *testing.T) {
	GenerateTests(func() Map { return NewSyncMap() })(t)
}

func GenerateBenchmarks(create func() Map) func(b *testing.B) {
	return func(b *testing.B) {
		b.Run("get", func(b *testing.B) {
			m := create()
			m.Set("1", 1)
			b.SetBytes(1)
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				m.Get("1")
			}
		})
		b.Run("get_para", func(b *testing.B) {
			m := create()
			m.Set("1", 1)
			b.SetBytes(1)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					m.Get("1")
				}
			})
		})
		b.Run("set", func(b *testing.B) {
			m := create()
			b.SetBytes(1)
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				m.Set("1", 1)
			}
		})
		b.Run("set_para", func(b *testing.B) {
			m := create()
			m.Set("1", 1)
			b.SetBytes(1)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					m.Set("1", 1)
				}
			})
		})
		/*
			b.Run("del", func(b *testing.B) {
				m := create()
				b.SetBytes(1)
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					m.Del("1")
				}
			})
			b.Run("del_multi", func(b *testing.B) {
				m := create()
				b.SetBytes(1)
				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						m.Del("1")
					}
				})
			})
		*/
	}
}

func BenchmarkSyncmap(b *testing.B) {
	GenerateBenchmarks(func() Map { return NewSyncMap() })(b)
}
