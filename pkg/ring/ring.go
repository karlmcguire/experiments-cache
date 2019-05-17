package ring

import (
	"sync/atomic"
)

const (
	GET BlockType = iota
	SET
	DEL

	BUFFER_SIZE      = 16
	BUFFER_MASK      = BUFFER_SIZE - 1
	BUFFER_THRESHOLD = 12

	STRIPE_COUNT = 4
	STRIPE_MASK  = STRIPE_COUNT - 1
)

type (
	Block     uint64
	BlockType int
)

func (b Block) Type() BlockType {
	return BlockType(b >> 62)
}

type Consumer interface {
	Push(int, Block)
	Wrap(func())
}

type Striped struct {
	buffers [STRIPE_COUNT]*Buffer
}

func NewStriped(consumer Consumer) *Striped {
	striped := &Striped{}

	for id := range striped.buffers {
		striped.buffers[id] = &Buffer{Consumer: consumer}
	}

	return striped
}

func (s *Striped) Add(block Block) (int, bool) {
	var (
		tries = 0
		id    = int(block) & STRIPE_MASK
	)

	for {
		if s.buffers[id].Add(block) {
			return tries, true
		}

		tries++
		id = (id + 1) & STRIPE_MASK
	}
}

type Buffer struct {
	Consumer Consumer

	busy uint32
	head uint32
	data [BUFFER_SIZE]Block
}

func (b *Buffer) Add(block Block) bool {
	head, full := b.next()
	if full {
		// attempt to drain
		//
		// (this is essentially a try lock)
		if atomic.CompareAndSwapUint32(&b.busy, 0, 1) {
			b.Consumer.Wrap(func() {
				// push each non-nil block to the consumer and reset the block
				for id, block := range b.data {
					if block != 0 {
						b.Consumer.Push(id, block)

						// clear block
						b.data[id] = 0
					}
				}
			})

			// finish
			atomic.StoreUint32(&b.head, 0)
			atomic.StoreUint32(&b.busy, 0)
		}

		return false
	}

	b.data[head] = block
	return true
}

func (b *Buffer) next() (uint32, bool) {
	head := atomic.LoadUint32(&b.head)

	for {
		// check if we're passed the draining threshold, the caller of this
		// function will handle the draining
		if head >= BUFFER_THRESHOLD {
			return head, true
		}

		// attempt to increment
		if atomic.CompareAndSwapUint32(&b.head, head, (head+1)&BUFFER_MASK) {
			return head, false
		}
	}
}
