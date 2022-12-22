// Package taskqueue provides a taskqueue implementation that limits the concurrency of tasks to a maximum value
// usually to avoid overwhelming external systems.
//
// A trivial example that limits the number of concurrent invocations of a squaring function to 3 at a time:
//   var run := func(ctx context.Context, n int) (int, error) {
//     time.Sleep(10 * time.Milliseconds)
//     return n * n
//   }
//   te := taskqueue.New(taskqueue.Opts{ MaxWorkers: 3, MaxQueueDepth:10 }, run)
package taskqueue

import (
	"context"

	"github.com/abevier/tsk/futures"
	"github.com/abevier/tsk/internal/tsk"
)

// RunFunction is a function signature for a function that can be invoked by the TaskQueue
type RunFunction[T any, R any] func(ctx context.Context, task T) (R, error)

// TaskQueue is a concurrency limiter that controls the number of concurrent invocations of the
// provided RunFunction.  Concurrent calls to Submit that exceed MaxWorkers will be queued up
// to the MaxQueueDepth.  If the MaxQueueDepth is exceeded then the behavior is determined by
// the cofigured FullQueueStrategy - the call to submit is either blocked or will immediately return
// an error.
// A TaskQueue must be created by calling New.
type TaskQueue[T any, R any] struct {
	run      RunFunction[T, R]
	taskChan chan tsk.TaskFuture[T, R]

	submit tsk.SubmitFunction[T, R]
}

// New creates a new TaskQueue with the provided options that limits concurrent calls
// to the provided run function.
func New[T any, R any](opts Opts, run RunFunction[T, R]) *TaskQueue[T, R] {
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

// Submit adds a task to the TaskQueue and blocks until a result is returned from the run function
func (tq *TaskQueue[T, R]) Submit(ctx context.Context, task T) (R, error) {
	f := tq.SubmitF(ctx, task)
	return f.Get(ctx)
}

// Submit adds a task to the TaskQueue and returns a futures.Future once the task has been successfully added.
// The returned future will contain the result of the run function once the function has been invoked and returns a value.
//
// Note that if a call to this funtion would cause MaxQueueDepth to be exceeded then depending on the configured
// FullQueueBehavior it will either block until the queue can accept the task OR this function wil immediately
// return a failed future with an ErrQueueFull.
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
