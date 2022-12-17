package batch

import (
	"context"
	"math"
	"time"

	"github.com/abevier/tsk/futures"
	"github.com/abevier/tsk/results"
)

type BatchOpts struct {
	MaxSize   int
	MaxLinger time.Duration
}

type RunBatchFunction[T any, R any] func(tasks []T) ([]results.Result[R], error)

type batch[T any, R any] struct {
	id      int
	tasks   []T
	futures []*futures.Future[R]
}

func (b *batch[T, R]) add(task T, future *futures.Future[R]) int {
	b.tasks = append(b.tasks, task)
	b.futures = append(b.futures, future)
	return len(b.tasks)
}

type taskFuture[T any, R any] struct {
	task   T
	future *futures.Future[R]
}

type BatchExecutor[T any, R any] struct {
	maxSize   int
	maxLinger time.Duration
	taskChan  chan taskFuture[T, R]
	run       RunBatchFunction[T, R]
}

func NewExecutor[T any, R any](opts BatchOpts, run RunBatchFunction[T, R]) *BatchExecutor[T, R] {
	be := &BatchExecutor[T, R]{
		maxSize:   opts.MaxSize,
		maxLinger: opts.MaxLinger,
		taskChan:  make(chan taskFuture[T, R]),
		run:       run,
	}

	be.do(be.taskChan)

	return be
}

func (be *BatchExecutor[T, R]) Submit(ctx context.Context, task T) (R, error) {
	f := be.SubmitF(ctx, task)
	return f.Get(ctx)
}

func (be *BatchExecutor[T, R]) SubmitF(ctx context.Context, task T) *futures.Future[R] {
	future := futures.New[R](ctx)
	be.taskChan <- taskFuture[T, R]{task: task, future: future}
	return future
}

func (be *BatchExecutor[T, R]) do(taskChan <-chan taskFuture[T, R]) {
	go func() {
		var currentBatch *batch[T, R]
		t := time.NewTimer(math.MaxInt64)

		for {
			select {
			case <-t.C:
				// batch expired due to time
				if currentBatch != nil {
					go be.runBatch(currentBatch)
					currentBatch = nil
					t.Reset(be.maxLinger)
				}

			case ft := <-taskChan:
				if currentBatch == nil {
					// open a new batch since once doesn't exist
					currentBatch = &batch[T, R]{
						tasks: make([]T, 0, be.maxSize),
					}

					if !t.Stop() {
						<-t.C
					}
					t.Reset(be.maxLinger)
				}

				size := currentBatch.add(ft.task, ft.future)
				if size >= be.maxSize {
					// flush the batch due to size
					go be.runBatch(currentBatch)
					currentBatch = nil

					if !t.Stop() {
						<-t.C
					}
					t.Reset(math.MaxInt64)
				}
			}
		}
	}()
}

func (be *BatchExecutor[T, R]) runBatch(b *batch[T, R]) {
	res, err := be.run(b.tasks)
	if err != nil {
		for _, f := range b.futures {
			f.Fail(err)
		}
		return
	}

	if len(res) != len(b.tasks) {
		for _, f := range b.futures {
			f.Fail(ErrBatchResultMismatch)
		}
		return
	}

	for i, r := range res {
		if r.Err != nil {
			b.futures[i].Fail(err)
		} else {
			b.futures[i].Complete(r.Val)
		}
	}
}
