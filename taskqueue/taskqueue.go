package taskqueue

import (
	"context"
	"log"
	"sync"
	"sync/atomic"

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
	task   T
	future *futures.Future[R]
}

type submitStrategy[T any, R any] func(taskChan chan<- taskFuture[T, R], tf taskFuture[T, R]) error

type TaskQueue[T any, R any] struct {
	isStopping uint32

	run      RunFunction[T, R]
	taskChan chan taskFuture[T, R]

	submit submitStrategy[T, R]

	waitSend *sync.WaitGroup
	waitStop *sync.WaitGroup
	stopOnce *sync.Once
}

type RunFunction[T any, R any] func(task T) (R, error)

func NewTaskQueue[T any, R any](opts TaskQueueOpts, run RunFunction[T, R]) *TaskQueue[T, R] {
	taskChan := make(chan taskFuture[T, R], opts.MaxQueueDepth)
	waitStop := sync.WaitGroup{}

	tq := &TaskQueue[T, R]{
		run:      run,
		taskChan: taskChan,
		waitSend: &sync.WaitGroup{},
		waitStop: &waitStop,
		stopOnce: &sync.Once{},
	}

	if opts.FullQueueBehavior == BlockWhenFull {
		tq.submit = submitBlockWhenFull[T, R]
	} else {
		tq.submit = submitErrorWhenFull[T, R]
	}

	for i := 0; i < opts.MaxWorkers; i++ {
		waitStop.Add(1)
		go tq.worker(i)
	}

	return tq
}

func (tq *TaskQueue[T, R]) worker(workerNum int) {
	defer tq.waitStop.Done()

	for {
		select {
		case tf, ok := <-tq.taskChan:
			if !ok {
				return
			}

			//TODO: check if context was cancelled before calling the function

			//TODO: instead of trying to log in the library add TSK_WORKER_ID to the context
			log.Printf("Running task on worker %d", workerNum)

			res, err := tq.run(tf.task)
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
	tq.waitSend.Add(1)
	defer tq.waitSend.Done()

	if atomic.LoadUint32(&tq.isStopping) == 1 {
		return nil, ErrStopped
	}

	future := futures.New[R](ctx)
	tf := taskFuture[T, R]{task: task, future: future}

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

func (tq *TaskQueue[T, R]) Stop() {
	tq.stopOnce.Do(func() {
		atomic.StoreUint32(&tq.isStopping, 1)
		tq.waitSend.Wait()
		close(tq.taskChan)
	})

	tq.waitStop.Wait()
}
