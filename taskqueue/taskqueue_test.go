package taskqueue

import (
	"context"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTaskQueue(t *testing.T) {
	require := require.New(t)

	maxWorkers := 3
	wg := sync.WaitGroup{}

	run := func(ctx context.Context, task int) (int, error) {
		workerId, ok := WorkerIDFromContext(ctx)
		require.True(ok)
		require.True(isValidWorkerID(workerId, maxWorkers))
		return task * 2, nil
	}

	tq := New(Opts{MaxWorkers: maxWorkers, MaxQueueDepth: 10}, run)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			val, err := tq.Submit(context.Background(), n)
			require.NoError(err)
			require.Equal(n*2, val)
		}(i)
	}

	wg.Wait()
	tq.Close()
}

func TestTaskQueueContextCancellation(t *testing.T) {
	require := require.New(t)

	run := func(ctx context.Context, task int) (int, error) {
		<-ctx.Done()
		return 0, context.Canceled
	}

	tq := New(Opts{MaxWorkers: 3, MaxQueueDepth: 10, FullQueueStrategy: BlockWhenFull}, run)

	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := tq.Submit(ctx, i)
		require.ErrorIs(err, context.Canceled)
	}
}

func isValidWorkerID(id string, maxWorkers int) bool {
	for i := 0; i < maxWorkers; i++ {
		if id == "worker-"+strconv.Itoa(i) {
			return true
		}
	}
	return false
}
