package ring

import (
	"fmt"
	"testing"
)

func Consumer() func(int, Block) {
	return func(id int, block Block) {
		fmt.Printf("got %02d: %d\n", id, block)
	}
}

func TestRing(t *testing.T) {
	buffer := &Buffer{
		push: Consumer(),
	}

	for i := 0; i < 32; i++ {
		buffer.Add(Block(uint32(i)))
	}
}
