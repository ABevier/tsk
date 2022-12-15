package futures

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/abevier/tsk/result"
)

var (
	ErrCanceled = errors.New("future canceled")
)

type FutureFunc[T any] func() (T, error)

type Future[T any] struct {
	isCompleted uint32
	completed   chan struct{}

	value T
	err   error
}

func New[T any](ctx context.Context) *Future[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	f := &Future[T]{
		completed: make(chan struct{}),
	}

	go func() {
		select {
		case <-ctx.Done():
			f.internalComplete(*new(T), context.Canceled)
		case <-f.completed:
		}
	}()

	return f
}

func FromFunc[T any](ctx context.Context, do FutureFunc[T]) *Future[T] {
	f := New[T](ctx)

	go func() {
		t, err := do()
		if err != nil {
			f.Fail(err)
		}
		f.Complete(t)
	}()

	return f
}

func (f *Future[T]) Complete(value T) {
	f.internalComplete(value, nil)
}

func (f *Future[T]) CompleteWithResult(r result.Result[T]) {
	f.internalComplete(r.Val, r.Err)
}

func (f *Future[T]) Cancel() {
	f.Fail(ErrCanceled)
}

func (f *Future[T]) Fail(err error) {
	f.internalComplete(*new(T), err)
}

func (f *Future[T]) internalComplete(val T, err error) {
	if atomic.CompareAndSwapUint32(&f.isCompleted, 0, 1) {
		f.value = val
		f.err = err
		close(f.completed)
	}
}

func (f *Future[T]) Get(ctx context.Context) (T, error) {
	select {
	case <-f.completed:
		return f.value, f.err
	case <-ctx.Done():
		return *new(T), context.Canceled
	}
}
