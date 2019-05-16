package try

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type Mutex struct {
	sync.Mutex
}

func (m *Mutex) TryLock() bool {
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(m)), 0, 1)
}
