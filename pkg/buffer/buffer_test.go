package buffer

import (
	"fmt"
	"testing"
)

func TestBuffer(t *testing.T) {
	b := &Buffer{
		In:  make(chan interface{}),
		Out: make(chan interface{}, 8),
	}
	go b.Run()

	for i := 0; i < 10; i++ {
		b.Add(i)
	}

	close(b.In)

	for i := range b.Out {
		fmt.Println(i.(int))
	}
}
