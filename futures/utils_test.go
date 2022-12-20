package futures

import (
	"context"
	"testing"
	"time"

	"github.com/abevier/tsk/results"
	"github.com/stretchr/testify/require"
)

func TestResolveAll(t *testing.T) {
	require := require.New(t)

	f1 := FromFunc(func() (int, error) {
		time.Sleep(6 * time.Millisecond)
		return 1, nil
	})

	f2 := FromFunc(func() (int, error) {
		time.Sleep(4 * time.Millisecond)
		return 2, nil
	})

	f3 := FromFunc(func() (int, error) {
		time.Sleep(2 * time.Millisecond)
		return 3, nil
	})

	rs, err := ResolveAll(context.Background(), []*Future[int]{f1, f2, f3})
	require.NoError(err)

	expected := []results.Result[int]{
		results.Success(1),
		results.Success(2),
		results.Success(3),
	}

	require.Equal(expected, rs)
}

func TestResolveAllCancellation(t *testing.T) {
	require := require.New(t)

	f1 := New[int]()
	f2 := New[int]()
	f3 := New[int]()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	_, err := ResolveAll(ctx, []*Future[int]{f1, f2, f3})
	require.ErrorIs(err, context.Canceled)
}
