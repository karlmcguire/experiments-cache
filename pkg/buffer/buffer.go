package buffer

import (
	"fmt"

	"github.com/karlmcguire/experiments-cache/pkg/store"
	"github.com/karlmcguire/experiments-cache/pkg/try"
)

type Buffer struct {
	try.Mutex

	In      chan string
	Data    store.Store
	Workers []*Worker
}

func NewBuffer(workerCount, workerSize, workerThreshold uint64, data store.Store) *Buffer {
	buffer := &Buffer{
		In:      make(chan string),
		Data:    data,
		Workers: make([]*Worker, workerCount),
	}

	// start each worker listening on the single consumption channel
	for id, _ := range buffer.Workers {
		buffer.Workers[id] = NewWorker(
			id,
			workerSize,
			workerThreshold,
			buffer,
		)
		// start worker
		go buffer.Workers[id].Run(buffer.In)
	}

	return buffer
}

// Add records an access in the buffer.
func (b *Buffer) Add(key string) {
	b.In <- key
}

type Worker struct {
	Id        int
	Queue     chan string
	Threshold int
	Buffer    *Buffer
}

func NewWorker(id int, size, threshold uint64, buffer *Buffer) *Worker {
	return &Worker{
		Id:        id,
		Queue:     make(chan string, size),
		Threshold: int(threshold),
		Buffer:    buffer,
	}
}

func (w *Worker) Run(in chan string) {
	for key := range in {
		select {
		case w.Queue <- key:
			if len(w.Queue) >= w.Threshold {
				// attempt to drain
				w.Drain(false)
			}
		default:
			// queue is full, required drain
			w.Drain(true)
		}
	}

	close(w.Queue)
}

func (w *Worker) Drain(required bool) {
	if !required {
		if w.Buffer.TryLock() {
			fmt.Printf("%d threshold draining\n", w.Id)

			// drain queue
			for key := range w.Queue {
				fmt.Printf("\t%s\n", key)
			}
			w.Buffer.Unlock()

		} else {
			// since the drain isn't required, we'll stop here and let it
			// attempt again when its called later
		}
	}

	// drain is required
	w.Buffer.Lock()
	fmt.Printf("%d required draining\n", w.Id)
	for key := range w.Queue {
		fmt.Printf("\t%s\n", key)
	}
	w.Buffer.Unlock()
}