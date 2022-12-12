package futures

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFuture(t *testing.T) {
	require := require.New(t)

	f := New[int]()

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.Complete(1)
		f.Complete(2)
		f.Complete(3)
	}()

	v, err := f.Get(context.TODO())
	require.NoError(err)
	require.Equal(1, v)
}

func TestComplete(t *testing.T) {
	require := require.New(t)

	f := New[int]()

	for i := 0; i <= 1000; i++ {
		go func() {
			f.Complete(42)
		}()
	}

	v, err := f.Get(context.TODO())
	require.NoError(err)
	require.Equal(42, v)
}

func TestCancel(t *testing.T) {
	require := require.New(t)

	f := New[int]()

	for i := 0; i <= 1000; i++ {
		go func() {
			time.Sleep(10 * time.Millisecond)
			f.Cancel()
		}()
	}

	_, err := f.Get(context.TODO())
	require.ErrorIs(err, ErrCanceled)
}

func TestFail(t *testing.T) {
	require := require.New(t)

	testErr := errors.New("test error")

	f := New[int]()

	for i := 0; i <= 1000; i++ {
		go func() {
			time.Sleep(10 * time.Millisecond)
			f.Fail(testErr)
		}()
	}

	_, err := f.Get(context.TODO())
	require.ErrorIs(err, testErr)
}

func TestCancelOnGet(t *testing.T) {
	require := require.New(t)

	f := New[int]()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	_, err := f.Get(ctx)
	require.ErrorIs(err, context.Canceled)
}

func TestCancelingContextOnFuture(t *testing.T) {
	require := require.New(t)

	ctx, cancel := context.WithCancel(context.Background())

	f := NewWithContext[int](ctx)

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	_, err := f.Get(ctx)
	require.ErrorIs(err, context.Canceled)
}
