package taskqueue

import (
	"context"
	"log"
	"sync"
	"sync/atomic"

	"github.com/ABevier/tsk/result"
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

//TODO: rename this lol
type taskWrapper[T any, R any] struct {
	ctx        context.Context
	task       T
	resultChan chan result.Result[R]
}

func newTaskWrapper[T any, R any](ctx context.Context, task T) taskWrapper[T, R] {
	resultChan := make(chan result.Result[R])
	return taskWrapper[T, R]{
		ctx:        ctx,
		task:       task,
		resultChan: resultChan,
	}
}

type submitStrategy[T any, R any] func(taskChan chan<- taskWrapper[T, R], t taskWrapper[T, R]) error

type TaskQueue[T any, R any] struct {
	isStopping uint32

	run      RunFunction[T, R]
	taskChan chan taskWrapper[T, R]

	submit submitStrategy[T, R]

	waitSend *sync.WaitGroup
	waitStop *sync.WaitGroup
	stopOnce *sync.Once
}

type RunFunction[T any, R any] func(task T) (R, error)

func NewTaskQueue[T any, R any](opts TaskQueueOpts, run RunFunction[T, R]) *TaskQueue[T, R] {
	taskChan := make(chan taskWrapper[T, R], opts.MaxQueueDepth)
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
		case task, ok := <-tq.taskChan:
			if !ok {
				return
			}

			//TODO: check if context was cancelled before calling the function

			//TODO: instead of trying to log in the library add TSK_WORKER_ID to the context
			log.Printf("Running task on worker %d", workerNum)

			res, err := tq.run(task.task)
			select {
			case <-task.ctx.Done():
			case task.resultChan <- result.NewResult(res, err):
			}

			close(task.resultChan)
		}
	}
}

func (tq *TaskQueue[T, R]) Submit(ctx context.Context, task T) (R, error) {
	tq.waitSend.Add(1)
	defer tq.waitSend.Done()

	if atomic.LoadUint32(&tq.isStopping) == 1 {
		return *new(R), ErrStopped
	}

	tw := newTaskWrapper[T, R](ctx, task)

	if err := tq.submit(tq.taskChan, tw); err != nil {
		return *new(R), err
	}

	select {
	case res := <-tw.resultChan:
		return res.Val, res.Err

	case <-ctx.Done():
		return *new(R), context.Canceled
	}
}

func submitBlockWhenFull[T any, R any](taskChan chan<- taskWrapper[T, R], t taskWrapper[T, R]) error {
	taskChan <- t
	return nil
}

func submitErrorWhenFull[T any, R any](taskChan chan<- taskWrapper[T, R], t taskWrapper[T, R]) error {
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
