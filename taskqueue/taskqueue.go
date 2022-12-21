// Package taskqueue provides a taskqueue implementation that limits the concurrency of tasks to a maximum value
// usually to avoid overwhelming external systems.
package taskqueue

import (
	"context"

	"github.com/abevier/tsk/futures"
	"github.com/abevier/tsk/internal/tsk"
)

//
type RunFunction[T any, R any] func(ctx context.Context, task T) (R, error)

//
type TaskQueue[T any, R any] struct {
	run      RunFunction[T, R]
	taskChan chan tsk.TaskFuture[T, R]

	submit tsk.SubmitFunction[T, R]
}

//
func NewTaskQueue[T any, R any](opts Opts, run RunFunction[T, R]) *TaskQueue[T, R] {
	taskChan := make(chan tsk.TaskFuture[T, R], opts.MaxQueueDepth)

	tq := &TaskQueue[T, R]{
		run:      run,
		taskChan: taskChan,
		submit:   tsk.GetSubmitFunction[T, R](tsk.FullQueueStrategy(opts.FullQueueStrategy)),
	}

	for i := 0; i < opts.MaxWorkers; i++ {
		tq.startWorker(i)
	}

	return tq
}

func (tq *TaskQueue[T, R]) startWorker(workerNum int) {
	go func() {
		for {
			tf, ok := <-tq.taskChan
			if !ok {
				return
			}

			ctx := withWorkerID(tf.Ctx, workerNum)
			res, err := tq.run(ctx, tf.Task)
			if err != nil {
				tf.Future.Fail(err)
			}
			tf.Future.Complete(res)
		}
	}()
}

//
func (tq *TaskQueue[T, R]) Submit(ctx context.Context, task T) (R, error) {
	f := tq.SubmitF(ctx, task)
	return f.Get(ctx)
}

//
func (tq *TaskQueue[T, R]) SubmitF(ctx context.Context, task T) *futures.Future[R] {
	tf := tsk.NewTaskFuture[T, R](ctx, task)
	if err := tq.submit(tq.taskChan, tf); err != nil {
		tf.Future.Fail(err)
	}
	return tf.Future
}

// Close closes the TaskQueue's underlying channel.  It is the responsibility of the caller to ensure
// that no writers are still calling Submit or SubmitF as this will cause a panic.
func (tq *TaskQueue[T, R]) Close() {
	close(tq.taskChan)
}
