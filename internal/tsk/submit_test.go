package tsk

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSubmitFunction(t *testing.T) {
	req := require.New(t)

	f := GetSubmitFunction[int, int](BlockWhenFull)
	req.NotNil(f)

	f = GetSubmitFunction[int, int](ErrorWhenFull)
	req.NotNil(f)
}

func TestGetSubmitFunctionPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("GetSubmitFunction did not panic")
		}
	}()

	GetSubmitFunction[int, int](-1)
}

func TestBlockWhenFullStrategy(t *testing.T) {
	req := require.New(t)

	c := make(chan TaskFuture[int, int])

	// Test cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tf := NewTaskFuture[int, int](ctx, 1)
	err := blockWhenFullStrategy(c, tf)
	req.ErrorIs(context.Canceled, err)

	// Test consumption
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			v, ok := <-c
			if !ok {
				return
			}
			v.Future.Complete(42)
		}
	}()

	ctx = context.Background()
	tf = NewTaskFuture[int, int](ctx, 1)

	err = blockWhenFullStrategy(c, tf)
	req.NoError(err)

	v, err := tf.Future.Get(ctx)
	req.NoError(err)
	req.Equal(42, v)

	close(c)
	wg.Wait()
}

func TestErrorWhenFull(t *testing.T) {
	req := require.New(t)

	c := make(chan TaskFuture[int, int])

	startConsumer := func() {
		go func() {
			for {
				v, ok := <-c
				if !ok {
					return
				}
				v.Future.Complete(42)
			}
		}()
	}

	wg := sync.WaitGroup{}

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx := context.Background()
			tf := NewTaskFuture[int, int](ctx, 1)
			err := errorWhenFullStrategy(c, tf)
			if err == ErrQueueFull {
				startConsumer()
			} else {
				v, err := tf.Future.Get(ctx)
				req.NoError(err)
				req.Equal(42, v)
			}
		}()
	}

	wg.Wait()
	close(c)
}
