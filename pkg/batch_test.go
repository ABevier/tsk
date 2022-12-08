package tsk

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var ErrTest = errors.New("unit test error")

func TestBatch(t *testing.T) {
	require := require.New(t)

	var actualCount uint32 = 0
	itemCount := 10

	wg := sync.WaitGroup{}
	wg.Add(itemCount)

	run := func(items []int) ([]Result[int], error) {
		var results []Result[int]

		for _, n := range items {
			r := NewSuccess(n * 2)
			results = append(results, r)
			atomic.AddUint32(&actualCount, 1)
		}

		return results, nil
	}

	be := NewBatchExecutor(BatchOpts{MaxSize: 3, MaxLinger: 100 * time.Millisecond}, run)

	for i := 0; i < itemCount; i++ {
		go func(val int) {
			res, err := be.Submit(context.TODO(), val)
			require.NoError(err)
			require.Equal(2*val, res)

			wg.Done()
		}(i)
	}

	wg.Wait()

	require.Equal(itemCount, int(actualCount))
}

func TestBatchFailure(t *testing.T) {
	require := require.New(t)

	itemCount := 10
	wg := sync.WaitGroup{}

	run := func(items []int) ([]Result[int], error) {
		return nil, ErrTest
	}

	be := NewBatchExecutor(BatchOpts{MaxSize: 3, MaxLinger: 100 * time.Millisecond}, run)

	for i := 0; i < itemCount; i++ {
		wg.Add(1)
		go func(val int) {
			_, err := be.Submit(context.TODO(), val)
			require.ErrorIs(err, ErrTest)
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestSubmitCancellation(t *testing.T) {
	require := require.New(t)

	run := func(items []int) ([]Result[int], error) {
		var results []Result[int]
		for _, n := range items {
			results = append(results, NewSuccess(n*2))
		}
		return results, nil
	}

	be := NewBatchExecutor(BatchOpts{MaxSize: 3, MaxLinger: 100 * time.Millisecond}, run)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel the context before submitting

	_, err := be.Submit(ctx, 5)
	require.ErrorIs(err, context.Canceled)
}
