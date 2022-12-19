package ratelimiter

import (
	"context"
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRateLimiter(t *testing.T) {
	require := require.New(t)

	wg := sync.WaitGroup{}

	run := func(ctx context.Context, n int) (int, error) {
		log.Printf("processing request: %d", n)
		return n * 2, nil
	}

	rl := New(10, 1, run)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			r, err := rl.Submit(context.Background(), n)
			require.NoError(err)
			require.Equal(n*2, r)
		}(i)
	}

	wg.Wait()
}
