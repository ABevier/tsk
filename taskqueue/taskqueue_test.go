package taskqueue

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTaskQueue(t *testing.T) {
	require := require.New(t)

	wg := sync.WaitGroup{}

	run := func(task int) (int, error) {
		log.Printf("i am task %d", task)
		time.Sleep(randWait(10, 50))
		return task * 2, nil
	}

	tq := NewTaskQueue(TaskQueueOpts{MaxWorkers: 3, MaxQueueDepth: 10, FullQueueBehavior: BlockWhenFull}, run)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			val, err := tq.Submit(context.TODO(), n)
			require.NoError(err)
			require.Equal(n*2, val)
		}(i)
	}

	wg.Wait()
	tq.Stop()
}

func TestTaskQueueCloseMultipleTimes(t *testing.T) {
	wg := sync.WaitGroup{}

	run := func(task int) (int, error) {
		time.Sleep(randWait(10, 100))
		return task * 2, nil
	}

	tq := NewTaskQueue(TaskQueueOpts{MaxWorkers: 10, MaxQueueDepth: 1000, FullQueueBehavior: BlockWhenFull}, run)

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			time.Sleep(randWait(30, 100))
			tq.Submit(context.TODO(), n)
			tq.Stop()
		}(i)
	}

	wg.Wait()
	//Test should not block or panic
}

func TestTaskQueueErrorWhenFull(t *testing.T) {
	require := require.New(t)

	var successCount int32
	var errCount int32

	numTasks := 100
	wg := sync.WaitGroup{}

	run := func(task int) (int, error) {
		// first 3 tasks will sleep, the next 3 will fill the queue, the rest will error, then the 3 tasks in the queue will be processed
		time.Sleep(100 * time.Millisecond)
		return task * 2, nil
	}

	tq := NewTaskQueue(TaskQueueOpts{MaxWorkers: 3, MaxQueueDepth: 3, FullQueueBehavior: ErrorWhenFull}, run)

	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			v, err := tq.Submit(context.TODO(), n)
			if err != nil {
				atomic.AddInt32(&errCount, 1)
				require.ErrorIs(err, ErrQueueFull)
			} else {
				atomic.AddInt32(&successCount, 1)
				require.Equal(n*2, v)
			}
		}(i)
	}

	wg.Wait()
	tq.Stop()

	require.Equal(int32(6), atomic.LoadInt32(&successCount))
	require.Equal(int32(numTasks-6), atomic.LoadInt32(&errCount))
}

func TestTaskQueueContextCancellation(t *testing.T) {
	require := require.New(t)

	run := func(task int) (int, error) {
		return task * 2, nil
	}

	tq := NewTaskQueue(TaskQueueOpts{MaxWorkers: 3, MaxQueueDepth: 10, FullQueueBehavior: BlockWhenFull}, run)

	for i := 0; i < 100; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := tq.Submit(ctx, i)
		require.ErrorIs(err, context.Canceled)
	}

	tq.Stop()
}

func randWait(min, max int) time.Duration {
	n := rand.Intn(max-min) + min
	return time.Duration(n) * time.Millisecond
}
