package buffer

import (
	"fmt"
	"testing"
	"time"

	"github.com/karlmcguire/experiments-cache/pkg/store"
)

func TestBuffer(t *testing.T) {
	buffer := NewBuffer(
		// worker count
		8,
		// worker queue size
		2,
		// data store
		store.NewMapStore(16),
	)

	for i := 0; i < 16; i++ {
		buffer.Add(fmt.Sprintf("%d", i))
	}

	<-time.After(time.Second / 2)
}
