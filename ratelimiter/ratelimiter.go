package ratelimiter

import (
	"context"

	"github.com/abevier/tsk/futures"
	"github.com/abevier/tsk/internal/tsk"
	"golang.org/x/time/rate"
)

type RunFunction[T any, R any] func(ctx context.Context, task T) (R, error)

type RateLimiter[T any, R any] struct {
	limiter  *rate.Limiter
	taskChan chan (tsk.TaskFuture[T, R])

	submit tsk.SubmitFunction[T, R]
	run    RunFunction[T, R]
}

func New[T any, R any](opts RateLimiterOpts, run RunFunction[T, R]) *RateLimiter[T, R] {
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

func (rl *RateLimiter[T, R]) Submit(ctx context.Context, task T) (R, error) {
	f := rl.SubmitF(ctx, task)
	return f.Get(ctx)
}

func (rl *RateLimiter[T, R]) SubmitF(ctx context.Context, task T) *futures.Future[R] {
	tf := tsk.NewTaskFuture[T, R](ctx, task)
	rl.submit(rl.taskChan, tf)
	return tf.Future
}

// WARNING If this is called twice or Submit is called after calling Close it will panic
func (rl *RateLimiter[T, R]) Close() {
	close(rl.taskChan)
}
