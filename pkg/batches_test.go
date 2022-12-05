package tasks

import (
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBatches(t *testing.T) {
	var actualCount uint32 = 0
	itemCount := 10

	wg := sync.WaitGroup{}
	wg.Add(itemCount)

	run := func(items []string) {
		for range items {
			atomic.AddUint32(&actualCount, 1)
			wg.Done()
		}
	}

	be := NewBatchExecutor(3, 100*time.Millisecond, run)

	for i := 0; i < itemCount; i++ {
		be.AddItem(strconv.Itoa(i))
	}

	wg.Wait()

	require.Equal(t, itemCount, int(actualCount))
}
