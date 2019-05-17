package ring

import (
	"sync"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

type SmallConsumer struct{}

func (c *SmallConsumer) Wrap(consume func())      { consume() }
func (c *SmallConsumer) Push(id int, block Block) {}

func BenchmarkRing(b *testing.B) {
	ring := &Buffer{Consumer: &SmallConsumer{}}

	for n := 0; n < b.N; n++ {
		ring.Add(Block(n))
	}
}

func BenchmarkStriped(b *testing.B) {
	striped := NewStriped(&SmallConsumer{})

	for n := 0; n < b.N; n++ {
		striped.Add(Block(n))
	}
}

type MockConsumer struct {
	sync.Mutex

	data []Block
}

func (c *MockConsumer) Wrap(consume func()) {
	c.Lock()
	defer c.Unlock()

	consume()
}

func (c *MockConsumer) Push(id int, block Block) {
	// safely write to the data store
	c.data = append(c.data, block)
}

func TestRing(t *testing.T) {
	var (
		num      = 32
		consumer = &MockConsumer{data: make([]Block, 0, num)}
		buffer   = &Buffer{Consumer: consumer}
	)

	for i := 1; i <= num; i++ {
		block := Block(uint32(i))

		// try twice to add
		if !buffer.Add(block) {
			buffer.Add(block)
		}
	}

	spew.Dump(consumer.data)
}

// TODO: sometimes this hangs -- not sure why
func TestStriped(t *testing.T) {
	var (
		num      = 64
		consumer = &MockConsumer{data: make([]Block, 0)}
		striped  = NewStriped(consumer)
		routines = 8
		wg       sync.WaitGroup
	)

	for i := 0; i < num; i += routines {
		wg.Add(1)
		go func(i int) {
			for b := i; b < i+routines; b++ {
				striped.Add(Block(b))
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
	spew.Dump(consumer.data)
}
