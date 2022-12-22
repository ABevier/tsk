// Package batch provides a batch executor implementation that allows for multiple writers
// to submit tasks which will be batched and flushed based on configurable values.
//
// An example implementation that simply squares each number in the batch and returns them:
//   var runBatch := func (tasks []int) ([]results.Result[int], error) {
//     var res []results.Result[int]
//     for _, n := range tasks {
//		   res = append(res, results.Success(n * n))
//     }
//     return res, nil
//   }
//   be := batch.New(batch.Opts{ MaxSize: 10, MaxLinger: 250 * time.Milliseconds }, runBatch)
//
// The above example flushes the batch when 10 items are in the batch OR after the oldest item in
// the batch is 250ms old.
package batch

import (
	"context"
	"math"
	"time"

	"github.com/abevier/tsk/futures"
	"github.com/abevier/tsk/internal/tsk"
	"github.com/abevier/tsk/results"
)

// RunBatchFunction is a function that handles a flushed batch of task items.
// Each time the batch executor flushes a batch, a slice of these task is provided
// to a function with this signature.
//
// This function must return a slice of results.Result of equal size to the number of
// tasks passed in. Each result in the returned slice should correspond to the task with
// the same index i.e. the returned result at index i corresponds to the provided task at index i.
//
// If the RunBatchFunction returns an error every item in the batch will complete with that error.
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

// BatchExecutor batches values of type T submited via multiple producers and invokes the provided run function
// when a batch flushes either due to size or a timeout.  Results of type R are returned to the caller of Submit
// when the batch finishes execution.
// A BatchExecutor must be created by calling New
type BatchExecutor[T any, R any] struct {
	maxSize   int
	maxLinger time.Duration
	taskChan  chan tsk.TaskFuture[T, R]
	run       RunBatchFunction[T, R]
}

// New creates a new BatchExecutor with the specified options that invokes the provided run function
// when a batch of items flushes due to either size or time.
func New[T any, R any](opts Opts, run RunBatchFunction[T, R]) *BatchExecutor[T, R] {
	be := &BatchExecutor[T, R]{
		maxSize:   opts.MaxSize,
		maxLinger: opts.MaxLinger,
		taskChan:  make(chan tsk.TaskFuture[T, R]),
		run:       run,
	}

	be.startWorker(be.taskChan)

	return be
}

// Submit adds an item to a batch and then blocks until the batch has been processed and a result
// has been returned.
func (be *BatchExecutor[T, R]) Submit(ctx context.Context, task T) (R, error) {
	f := be.SubmitF(task)
	return f.Get(ctx)
}

// SubmitF adds an item to a batch without blocking and then returns a futures.Future that will
// contain the result of the batch computation when it is completed
func (be *BatchExecutor[T, R]) SubmitF(task T) *futures.Future[R] {
	future := futures.New[R]()
	be.taskChan <- tsk.TaskFuture[T, R]{Task: task, Future: future}
	return future
}

func (be *BatchExecutor[T, R]) startWorker(taskChan <-chan tsk.TaskFuture[T, R]) {
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

			case ft, ok := <-taskChan:
				if !ok {
					if !t.Stop() {
						<-t.C
					}
					return
				}

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

				size := currentBatch.add(ft.Task, ft.Future)
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
			b.futures[i].Fail(r.Err)
		} else {
			b.futures[i].Complete(r.Val)
		}
	}
}

// Close closes the BatchExecutor's underlying channel.  It is the responsibility of the caller to ensure
// that no writers are still calling Submit or SubmitF as this will cause a panic.
func (be *BatchExecutor[T, R]) Close() {
	close(be.taskChan)
}
