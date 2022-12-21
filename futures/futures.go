// Package futures provides an implementation of a Future which represents an asynchronous computation.
// A Future can be created and then passed around and read by multiple consumers.  This is the key difference
// between a Future and using a channel for an asynchronous computation as a channel value can only be read once.
package futures

import (
	"context"
	"errors"
	"sync/atomic"
)

var (
	// ErrCanceled is the error reported when a future is completed by calling Cancel
	ErrCanceled = errors.New("future canceled")
)

// FutureFunc is the function signature required to create a Future via FromFunc
type FutureFunc[T any] func() (T, error)

// Future is a structure that represents an asynchronous computation.
// A Future should be created by calling New() or using the FromFunc convience function.
// Once a future has been created it can be completed exactly once.  The first completion value
// wins and all other completions are silently ignored.
//
// The functions Complete, Cancel and Fail will all complete a future.
// Complete is used in the success case
// Fail is used for signaling that the Future failed with an error
// Cancel is used to signal that the asynchronous computation was canceled
//
// Get is used to extract the value and an error from the Future.  If the future has not been
// completed calling Get will block until the future completes or until the context is canceled.
// Get can be called by multiple go routines simultaneously and they will all receive the same value.
type Future[T any] struct {
	isCompleted uint32
	completed   chan struct{}

	value T
	err   error
}

// New creates a new uncompleted Future that will eventually contain a value of type T which can be anything.
// This future must be manually completed by calling Complete, Fail, or Cancel
func New[T any]() *Future[T] {
	return &Future[T]{
		completed: make(chan struct{}),
	}
}

// FromFunc creates a new uncompleted Future that will eventually contain the return value of the provided function.
// The provided function is run asynchronously when this function is invoked.
func FromFunc[T any](do FutureFunc[T]) *Future[T] {
	f := New[T]()

	go func() {
		t, err := do()
		if err != nil {
			f.Fail(err)
		}
		f.Complete(t)
	}()

	return f
}

// Complete completes this Future with the provided value.  If the future has already been completed this call is ignored.
func (f *Future[T]) Complete(value T) {
	f.internalComplete(value, nil)
}

// Cancel completes this Future with the ErrCanceled error.  If the future has already been completed this call is ignored.
func (f *Future[T]) Cancel() {
	f.Fail(ErrCanceled)
}

// Fail completes this Future with the provided error.  If the future has already been completed this call is ignored.
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

// Get retrieves the value of this Future.  If the future is not yet completed this call will block until the future is
// completed or until the provided context is canceled.
func (f *Future[T]) Get(ctx context.Context) (T, error) {
	select {
	case <-f.completed:
		return f.value, f.err
	case <-ctx.Done():
		return *new(T), context.Canceled
	}
}
