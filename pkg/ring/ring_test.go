package ring

import (
	"fmt"
	"sync"
	"testing"
)

type MockConsumer struct {
	sync.Mutex
}

func (c *MockConsumer) Init() {
	c.Lock()
	println("locked")
}

func (c *MockConsumer) Done() {
	c.Unlock()
	println("unlocked")
}

func (c *MockConsumer) Push(id int, block Block) {
	fmt.Printf("got %02d: %d\n", id, block)
}

func TestRing(t *testing.T) {
	buffer := &Buffer{Consumer: &MockConsumer{}}

	for i := 0; i < 32; i++ {
		buffer.Add(Block(uint32(i)))
	}
}
