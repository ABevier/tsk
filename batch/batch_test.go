package batch

import (
	"context"
	"errors"
	"math"
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

	run := func(items []int) ([]results.Result[int], error) {
		var rs []results.Result[int]

		for _, n := range items {
			if n == 5 {
				rs = append(rs, results.Failure[int](ErrTest))
			} else {
				rs = append(rs, results.Success(n*2))
			}
			atomic.AddUint32(&actualCount, 1)
		}

		return rs, nil
	}

	be := NewExecutor(Opts{MaxSize: 3, MaxLinger: 10 * time.Millisecond}, run)

	for i := 0; i < itemCount; i++ {
		wg.Add(1)

		go func(n int) {
			defer wg.Done()

			res, err := be.Submit(context.TODO(), n)
			if n == 5 {
				require.ErrorIs(err, ErrTest)
				return
			}
			require.NoError(err)
			require.Equal(2*n, res)
		}(i)
	}

	wg.Wait()
	be.Close()

	require.Equal(itemCount, int(actualCount))
}

func TestBatchFailure(t *testing.T) {
	require := require.New(t)

	itemCount := 10
	wg := sync.WaitGroup{}

	run := func(items []int) ([]results.Result[int], error) {
		return nil, ErrTest
	}

	be := NewExecutor(Opts{MaxSize: 3, MaxLinger: 10 * time.Millisecond}, run)

	for i := 0; i < itemCount; i++ {
		wg.Add(1)
		go func(val int) {
			_, err := be.Submit(context.TODO(), val)
			require.ErrorIs(err, ErrTest)
			wg.Done()
		}(i)
	}

	wg.Wait()
	be.Close()
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

	be := NewExecutor(Opts{MaxSize: 3, MaxLinger: math.MaxInt64}, run)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel the context before submitting

	_, err := be.Submit(ctx, 5)
	require.ErrorIs(err, context.Canceled)

	be.Close()
}

func TestBadRunFunction(t *testing.T) {
	require := require.New(t)

	wg := sync.WaitGroup{}

	run := func(items []int) ([]results.Result[int], error) {
		return []results.Result[int]{}, nil
	}

	be := NewExecutor(Opts{MaxSize: 3, MaxLinger: 10 * time.Millisecond}, run)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(n int) {
			_, err := be.Submit(context.Background(), n)
			require.ErrorIs(err, ErrBatchResultMismatch)
			wg.Done()
		}(i)
	}

	wg.Wait()
	be.Close()
}
