package ring

// Concurrent-safe ring buffer using non-buffered input channel and buffered
// output channel. Optionally, you can use the TryAdd and TryGet for
// non-blocking attempts of Adding and Getting values.
type Buffer struct {
	In  chan interface{}
	Out chan interface{}
}

func New(in, out chan interface{}) *Buffer {
	return &Buffer{in, out}
}

func (b *Buffer) Run() {
	for value := range b.In {
		select {
		case b.Out <- value:
		default:
			<-b.Out
			b.Out <- value
		}
	}

	close(b.Out)
}

func (b *Buffer) Add(v interface{}) { b.In <- v }
func (b *Buffer) Get() interface{}  { return <-b.Out }

func (b *Buffer) TryAdd(v interface{}) bool {
	select {
	case b.In <- v:
		return true
	default:
		return false
	}
}

func (b *Buffer) TryGet() (interface{}, bool) {
	select {
	case v := <-b.Out:
		return v, true
	default:
		return nil, false
	}
}
