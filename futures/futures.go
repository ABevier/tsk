package futures

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrCanceled = errors.New("future canceled")
)

type Future[T any] struct {
	m           *sync.Mutex
	isCompleted uint32
	completed   chan struct{}

	value T
	err   error
}

func New[T any]() *Future[T] {
	return &Future[T]{
		m:         &sync.Mutex{},
		completed: make(chan struct{}),
	}
}

func NewWithContext[T any](ctx context.Context) *Future[T] {
	f := New[T]()

	go func() {
		select {
		case <-ctx.Done():
			f.internalComplete(*new(T), context.Canceled)
		case <-f.completed:
		}
	}()

	return f
}

func (f *Future[T]) Complete(value T) {
	f.internalComplete(value, nil)
}

func (f *Future[T]) Cancel() {
	f.Fail(ErrCanceled)
}

func (f *Future[T]) Fail(err error) {
	f.internalComplete(*new(T), err)
}

func (f *Future[T]) internalComplete(val T, err error) {
	if atomic.LoadUint32(&f.isCompleted) != 0 {
		return
	}

	f.m.Lock()
	defer f.m.Unlock()

	if f.isCompleted == 0 {
		f.value = val
		f.err = err
		atomic.StoreUint32(&f.isCompleted, 1)
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