package ratelimiter

import (
	"context"

	"github.com/abevier/tsk/futures"
	"golang.org/x/time/rate"
)

type RunFunction[T any, R any] func(ctx context.Context, task T) (R, error)

type RateLimiter[T any, R any] struct {
	limiter  *rate.Limiter
	taskChan chan (taskFuture[T, R])

	run RunFunction[T, R]
}

type taskFuture[T any, R any] struct {
	ctx    context.Context
	task   T
	future *futures.Future[R]
}

func New[T any, R any](limit rate.Limit, burst int, run RunFunction[T, R]) *RateLimiter[T, R] {
	rl := &RateLimiter[T, R]{
		limiter:  rate.NewLimiter(limit, burst),
		taskChan: make(chan taskFuture[T, R]),
		run:      run,
	}

	rl.worker()

	return rl
}

func (rl *RateLimiter[T, R]) worker() {
	go func() {
		for {
			tf := <-rl.taskChan

			if err := rl.limiter.Wait(tf.ctx); err != nil {
				tf.future.Fail(err)
				continue
			}

			rl.runTask(tf)
		}
	}()
}

func (rl *RateLimiter[T, R]) runTask(tf taskFuture[T, R]) {
	go func() {
		r, err := rl.run(tf.ctx, tf.task)
		if err != nil {
			tf.future.Fail(err)
			return
		}
		tf.future.Complete(r)
	}()
}

func (rl *RateLimiter[T, R]) Submit(ctx context.Context, task T) (R, error) {
	f := rl.SubmitF(ctx, task)
	return f.Get(ctx)
}

func (rl *RateLimiter[T, R]) SubmitF(ctx context.Context, task T) *futures.Future[R] {
	future := futures.New[R]()
	tf := taskFuture[T, R]{ctx: ctx, task: task, future: future}

	rl.taskChan <- tf
	return future
}
