package clhm

import (
	"fmt"
	"testing"
)

func TestCLHM(t *testing.T) {
	m := New(&Config{
		BufferCount: 4,
		BufferSize:  4,
		MapSize:     16,
	})

	for i := 0; i < 10; i++ {
		m.Set([]byte(fmt.Sprintf("%d", i)), i)
	}

	fmt.Println(m.Get([]byte("1")))
}
