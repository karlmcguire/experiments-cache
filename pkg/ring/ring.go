package ring

import (
	"sync/atomic"
)

const (
	GET BlockType = iota
	SET
	DEL

	SIZE      = 16
	MASK      = SIZE - 1
	THRESHOLD = 12
)

type (
	Block     uint64
	BlockType int
)

func (b Block) Type() BlockType {
	return BlockType(b >> 62)
}

type Buffer struct {
	push func(int, Block)
	busy uint32
	head uint32
	data [SIZE]Block
}

func (b *Buffer) Add(block Block) bool {
	head, full := b.next()
	if full {
		// attempt to drain
		if atomic.CompareAndSwapUint32(&b.busy, 0, 1) {
			for id, block := range b.data {
				if block != 0 {
					b.push(id, block)

					// clear block
					b.data[id] = 0
				}
			}

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
	prev := atomic.LoadUint32(&b.head)

	for {
		head := (prev + 1) & MASK

		// if we should drain
		if head >= THRESHOLD {
			return head, true
		}

		// attempt to increment
		if atomic.CompareAndSwapUint32(&b.head, prev, head) {
			return head, false
		}
	}
}
