package ring

// TODO

type Buffer struct {
	data  []byte
	size  uint64
	read  uint64
	write uint64
}

func New(size uint64) *Buffer {
	return &Buffer{
		data:  make([]byte, size),
		size:  size,
		read:  0,
		write: 0,
	}
}
