package futures

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	ErrTest = errors.New("test error")
)

func TestFuture(t *testing.T) {
	require := require.New(t)

	f := New[int](nil)

	go func() {
		time.Sleep(10 * time.Millisecond)
		f.Complete(1)
		f.Complete(2)
		f.Complete(3)
	}()

	v, err := f.Get(context.Background())
	require.NoError(err)
	require.Equal(1, v)
}

func TestFromFunc(t *testing.T) {
	require := require.New(t)

	f := FromFunc(context.Background(), func() (int, error) {
		time.Sleep(10 * time.Millisecond)
		return 42, nil
	})

	r, err := f.Get(context.Background())
	require.NoError(err)
	require.Equal(42, r)

	f = FromFunc(context.Background(), func() (int, error) {
		time.Sleep(10 * time.Millisecond)
		return 0, ErrTest
	})

	r, err = f.Get(context.Background())
	require.ErrorIs(err, ErrTest)
}

func TestComplete(t *testing.T) {
	require := require.New(t)

	f := New[int](context.Background())

	for i := 0; i <= 1000; i++ {
		go func() {
			f.Complete(42)
		}()
	}

	v, err := f.Get(context.Background())
	require.NoError(err)
	require.Equal(42, v)
}

func TestCancel(t *testing.T) {
	require := require.New(t)

	f := New[int](context.Background())

	for i := 0; i <= 1000; i++ {
		go func() {
			time.Sleep(10 * time.Millisecond)
			f.Cancel()
		}()
	}

	_, err := f.Get(context.Background())
	require.ErrorIs(err, ErrCanceled)
}

func TestFail(t *testing.T) {
	require := require.New(t)

	f := New[int](context.Background())

	for i := 0; i <= 1000; i++ {
		go func() {
			time.Sleep(10 * time.Millisecond)
			f.Fail(ErrTest)
		}()
	}

	_, err := f.Get(context.Background())
	require.ErrorIs(err, ErrTest)
}

func TestCancelOnGet(t *testing.T) {
	require := require.New(t)

	f := New[int](context.Background())
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
	f := New[int](ctx)

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	_, err := f.Get(ctx)
	require.ErrorIs(err, context.Canceled)
}
