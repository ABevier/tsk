package batch

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/abevier/tsk/results"
	"github.com/stretchr/testify/require"
)

var ErrTest = errors.New("unit test error")

func TestBatch(t *testing.T) {
	require := require.New(t)

	var actualCount uint32 = 0
	itemCount := 10

	wg := sync.WaitGroup{}
	wg.Add(itemCount)

	run := func(items []int) ([]results.Result[int], error) {
		var rs []results.Result[int]

		for _, n := range items {
			r := results.Success(n * 2)
			rs = append(rs, r)
			atomic.AddUint32(&actualCount, 1)
		}

		return rs, nil
	}

	be := NewExecutor(BatchOpts{MaxSize: 3, MaxLinger: 100 * time.Millisecond}, run)

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

	run := func(items []int) ([]results.Result[int], error) {
		return nil, ErrTest
	}

	be := NewExecutor(BatchOpts{MaxSize: 3, MaxLinger: 100 * time.Millisecond}, run)

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

	run := func(items []int) ([]results.Result[int], error) {
		var rs []results.Result[int]
		for _, n := range items {
			rs = append(rs, results.Success(n*2))
		}
		return rs, nil
	}

	be := NewExecutor(BatchOpts{MaxSize: 3, MaxLinger: 100 * time.Millisecond}, run)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel the context before submitting

	_, err := be.Submit(ctx, 5)
	require.ErrorIs(err, context.Canceled)
}
