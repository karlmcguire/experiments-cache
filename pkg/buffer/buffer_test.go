package buffer

import (
	"fmt"
	"testing"
	"time"

	"github.com/karlmcguire/experiments-cache/pkg/store"
)

func TestBuffer(t *testing.T) {
	buffer := NewBuffer(
		// number of workers
		8,
		// worker size
		5,
		// worker threshold (to attempt drain)
		3,
		// data store
		store.NewMapStore(16),
	)

	for i := 0; i < 64; i++ {
		buffer.Add(fmt.Sprintf("%d", i))
	}

	<-time.After(time.Second)
}
