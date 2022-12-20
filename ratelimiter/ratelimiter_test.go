package ratelimiter

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

func TestRateLimiter(t *testing.T) {
	require := require.New(t)

	wg := sync.WaitGroup{}

	run := func(ctx context.Context, n int) (int, error) {
		if n == 5 {
			return -1, ErrTest
		}
		return n * 2, nil
	}

	rl := New(RateLimiterOpts{Limit: 100, Burst: 1, FullQueueStrategy: BlockWhenFull}, run)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			r, err := rl.Submit(context.Background(), n)
			if n == 5 {
				require.ErrorIs(err, ErrTest)
				return
			}
			require.NoError(err)
			require.Equal(n*2, r)
		}(i)
	}

	wg.Wait()
	rl.Close()
}

func TestRateLimiterDeadlineError(t *testing.T) {
	require := require.New(t)

	var errCount int32
	var successCount int32
	wg := sync.WaitGroup{}

	run := func(ctx context.Context, n int) (int, error) {
		return n * 2, nil
	}

	// Rate limit is extremely slow so deadlines will expire
	rl := New(RateLimiterOpts{Limit: Every(10 * time.Minute), Burst: 1, FullQueueStrategy: BlockWhenFull}, run)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			r, err := rl.Submit(ctx, n)
			if err != nil {
				atomic.AddInt32(&errCount, 1)
			} else if r == n*2 {
				atomic.AddInt32(&successCount, 1)
			} else {
				t.Fail()
			}
		}(i)
	}

	wg.Wait()
	rl.Close()

	require.Equal(1, int(successCount))
	require.Equal(4, int(errCount))
}
