package buffer

import (
	"fmt"

	"github.com/karlmcguire/experiments-cache/pkg/store"
)

type Buffer struct {
	In      chan string
	Data    store.Store
	Workers []*Worker
}

func NewBuffer(workerCount, workerSize uint64, data store.Store) *Buffer {
	buffer := &Buffer{
		In:      make(chan string),
		Data:    data,
		Workers: make([]*Worker, workerCount),
	}

	// start each worker listening on the single consumption channel
	for id, _ := range buffer.Workers {
		buffer.Workers[id] = NewWorker(workerSize)
		go buffer.Workers[id].Run(id, buffer.In)
	}

	return buffer
}

// Add records an access in the buffer.
func (b *Buffer) Add(key string) {
	b.In <- key
}

type Worker struct {
	Queue chan string
}

func NewWorker(size uint64) *Worker {
	return &Worker{
		Queue: make(chan string, size),
	}
}

func (w *Worker) Run(id int, in chan string) {
	for key := range in {
		select {
		case w.Queue <- key:
			fmt.Printf("%d added: %s\n", id, key)
		default:
			// queue is full
			fmt.Printf("%d full\n", id)
			// TODO: - try for a lock on w.Data
			//       - do the underlying LRU operations
		}
	}

	close(w.Queue)
}
