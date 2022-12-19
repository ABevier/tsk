package taskqueue

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTaskQueue(t *testing.T) {
	require := require.New(t)

	run := func(ctx context.Context, task int) (int, error) {
		return task * 2, nil
	}

	tq := NewTaskQueue(TaskQueueOpts{MaxWorkers: 3, MaxQueueDepth: 10, FullQueueStrategy: BlockWhenFull}, run)

	for i := 0; i < 100; i++ {
		go func(n int) {
			val, err := tq.Submit(context.Background(), n)
			require.NoError(err)
			require.Equal(n*2, val)
		}(i)
	}
}

//TODO: REDO THIS TEST, THERE IS A RACE
// func TestTaskQueueErrorWhenFull(t *testing.T) {
// 	require := require.New(t)

// 	var successCount int32
// 	var errCount int32

// 	numTasks := 10

// 	taskWaiter := sync.WaitGroup{}

// 	runWaiter := sync.WaitGroup{}
// 	runWaiter.Add(numTasks - 6)

// 	run := func(ctx context.Context, task int) (int, error) {
// 		// first 3 tasks will sleep, the next 3 will fill the queue, the rest will error, then the 3 tasks in the queue will be processed
// 		runWaiter.Wait()
// 		return task * 2, nil
// 	}

// 	tq := NewTaskQueue(TaskQueueOpts{MaxWorkers: 3, MaxQueueDepth: 3, FullQueueStrategy: ErrorWhenFull}, run)

// 	for i := 0; i < numTasks; i++ {
// 		taskWaiter.Add(1)
// 		go func(n int) {
// 			v, err := tq.Submit(context.TODO(), n)
// 			if err != nil {
// 				runWaiter.Done()
// 				atomic.AddInt32(&errCount, 1)
// 				require.ErrorIs(err, ErrQueueFull)
// 			} else {
// 				atomic.AddInt32(&successCount, 1)
// 				require.Equal(n*2, v)
// 			}
// 			taskWaiter.Done()
// 		}(i)
// 	}

// 	taskWaiter.Wait()

// 	require.Equal(int32(6), atomic.LoadInt32(&successCount))
// 	require.Equal(int32(numTasks-6), atomic.LoadInt32(&errCount))
// }

func TestTaskQueueContextCancellation(t *testing.T) {
	require := require.New(t)

	run := func(ctx context.Context, task int) (int, error) {
		select {
		case <-ctx.Done():
			return 0, context.Canceled
		}
	}

	tq := NewTaskQueue(TaskQueueOpts{MaxWorkers: 3, MaxQueueDepth: 10, FullQueueStrategy: BlockWhenFull}, run)

	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := tq.Submit(ctx, i)
		require.ErrorIs(err, context.Canceled)
	}
}

func randWait(min, max int) time.Duration {
	n := rand.Intn(max-min) + min
	return time.Duration(n) * time.Millisecond
}
