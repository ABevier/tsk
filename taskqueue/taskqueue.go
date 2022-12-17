package taskqueue

import (
	"context"

	"github.com/abevier/tsk/futures"
)

type FullQueueBehavior int

const (
	BlockWhenFull FullQueueBehavior = iota
	ErrorWhenFull
)

type TaskQueueOpts struct {
	MaxWorkers    int
	MaxQueueDepth int
	//TODO: better name?
	FullQueueBehavior FullQueueBehavior
}

type taskFuture[T any, R any] struct {
	ctx    context.Context
	task   T
	future *futures.Future[R]
}

type submitStrategy[T any, R any] func(taskChan chan<- taskFuture[T, R], tf taskFuture[T, R]) error

type TaskQueue[T any, R any] struct {
	run      RunFunction[T, R]
	taskChan chan taskFuture[T, R]

	submit submitStrategy[T, R]
}

type RunFunction[T any, R any] func(ctx context.Context, task T) (R, error)

func NewTaskQueue[T any, R any](opts TaskQueueOpts, run RunFunction[T, R]) *TaskQueue[T, R] {
	taskChan := make(chan taskFuture[T, R], opts.MaxQueueDepth)

	tq := &TaskQueue[T, R]{
		run:      run,
		taskChan: taskChan,
	}

	if opts.FullQueueBehavior == BlockWhenFull {
		tq.submit = submitBlockWhenFull[T, R]
	} else {
		tq.submit = submitErrorWhenFull[T, R]
	}

	for i := 0; i < opts.MaxWorkers; i++ {
		go tq.worker(i)
	}

	return tq
}

func (tq *TaskQueue[T, R]) worker(workerNum int) {
	for {
		select {
		case tf, ok := <-tq.taskChan:
			if !ok {
				return
			}

			//TODO: instead of trying to log in the library add TSK_WORKER_ID to the context
			//log.Printf("Running task on worker %d", workerNum)

			res, err := tq.run(tf.ctx, tf.task)
			if err != nil {
				tf.future.Fail(err)
			}
			tf.future.Complete(res)
		}
	}
}

func (tq *TaskQueue[T, R]) Submit(ctx context.Context, task T) (R, error) {
	future, err := tq.SubmitF(ctx, task)
	if err != nil {
		return *new(R), err
	}
	return future.Get(ctx)
}

func (tq *TaskQueue[T, R]) SubmitF(ctx context.Context, task T) (*futures.Future[R], error) {
	future := futures.New[R]()
	tf := taskFuture[T, R]{ctx: ctx, task: task, future: future}

	if err := tq.submit(tq.taskChan, tf); err != nil {
		return nil, err
	}

	return future, nil
}

func submitBlockWhenFull[T any, R any](taskChan chan<- taskFuture[T, R], t taskFuture[T, R]) error {
	taskChan <- t
	return nil
}

func submitErrorWhenFull[T any, R any](taskChan chan<- taskFuture[T, R], t taskFuture[T, R]) error {
	select {
	case taskChan <- t:
		return nil
	default:
		return ErrQueueFull
	}
}
