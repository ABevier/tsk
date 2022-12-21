// Package ratelimiter provides an implementation of a rate limiter that controls the number
// of calls to a provided function based on a Token Bucket algorithm
//
// A trivial example that allows for squaring 10 numbers per second (roughly every 100ms):
//   opts := rateLimiter.Opts{Limit: 10, Burst: 1}
//   rl := ratelimter.New(opts, func(ctx context.Context, n int) (int, error){
//	   return n * n, nil
//   })
package ratelimiter

import (
	"context"

	"github.com/abevier/tsk/futures"
	"github.com/abevier/tsk/internal/tsk"
	"golang.org/x/time/rate"
)

// RunFunction is a function signature for a function that can be rate limited by the rate limiter
type RunFunction[T any, R any] func(ctx context.Context, task T) (R, error)

// RateLimiter limits the number of calls to the provided function based on the configured limit and burst values.
// A RateLimiter must be created by calling New
type RateLimiter[T any, R any] struct {
	limiter  *rate.Limiter
	taskChan chan (tsk.TaskFuture[T, R])

	submit tsk.SubmitFunction[T, R]
	run    RunFunction[T, R]
}

// New creates a new RateLimiter with the provided options that limits invocations to the provided run function.
func New[T any, R any](opts Opts, run RunFunction[T, R]) *RateLimiter[T, R] {
	rl := &RateLimiter[T, R]{
		limiter:  rate.NewLimiter(rate.Limit(opts.Limit), opts.Burst),
		taskChan: make(chan tsk.TaskFuture[T, R], opts.MaxQueueDepth),
		submit:   tsk.GetSubmitFunction[T, R](tsk.FullQueueStrategy(opts.FullQueueStrategy)),
		run:      run,
	}

	rl.startWorker()

	return rl
}

func (rl *RateLimiter[T, R]) startWorker() {
	go func() {
		for {
			tf, ok := <-rl.taskChan
			if !ok {
				return
			}

			if err := rl.limiter.Wait(tf.Ctx); err != nil {
				tf.Future.Fail(err)
				continue
			}

			rl.runTask(tf)
		}
	}()
}

func (rl *RateLimiter[T, R]) runTask(tf tsk.TaskFuture[T, R]) {
	go func() {
		r, err := rl.run(tf.Ctx, tf.Task)
		if err != nil {
			tf.Future.Fail(err)
			return
		}
		tf.Future.Complete(r)
	}()
}

// Submit adds a task to the rate limiter and blocks until a result is returned from the run function.
func (rl *RateLimiter[T, R]) Submit(ctx context.Context, task T) (R, error) {
	f := rl.SubmitF(ctx, task)
	return f.Get(ctx)
}

// SubmitF adds a task to the rate limiter and returns a futures.Future once the task has been successfully added.
// The returned future will contain the result of the run function once the function has been invoked and returns a value
func (rl *RateLimiter[T, R]) SubmitF(ctx context.Context, task T) *futures.Future[R] {
	tf := tsk.NewTaskFuture[T, R](ctx, task)
	if err := rl.submit(rl.taskChan, tf); err != nil {
		tf.Future.Fail(err)
	}
	return tf.Future
}

// Close closes the RateLimiter's underlying channel.  It is the responsibility of the caller to ensure
// that no writers are still calling Submit or SubmitF as this will cause a panic.
func (rl *RateLimiter[T, R]) Close() {
	close(rl.taskChan)
}
