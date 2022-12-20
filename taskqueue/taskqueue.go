package taskqueue

import (
	"context"

	"github.com/abevier/tsk/futures"
	"github.com/abevier/tsk/internal/tsk"
)

type RunFunction[T any, R any] func(ctx context.Context, task T) (R, error)

type TaskQueue[T any, R any] struct {
	run      RunFunction[T, R]
	taskChan chan tsk.TaskFuture[T, R]

	submit tsk.SubmitFunction[T, R]
}

func NewTaskQueue[T any, R any](opts TaskQueueOpts, run RunFunction[T, R]) *TaskQueue[T, R] {
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

func (tq *TaskQueue[T, R]) Submit(ctx context.Context, task T) (R, error) {
	future, err := tq.SubmitF(ctx, task)
	if err != nil {
		return *new(R), err
	}
	return future.Get(ctx)
}

func (tq *TaskQueue[T, R]) SubmitF(ctx context.Context, task T) (*futures.Future[R], error) {
	tf := tsk.NewTaskFuture[T, R](ctx, task)

	if err := tq.submit(tq.taskChan, tf); err != nil {
		return nil, err
	}

	return tf.Future, nil
}

// WARNING If this is called twice or Submit is called after calling Close it will panic
func (tq *TaskQueue[T, R]) Close() {
	close(tq.taskChan)
}
