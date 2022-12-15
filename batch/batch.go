package batch

import (
	"context"
	"sync"
	"time"

	"github.com/abevier/tsk/futures"
	"github.com/abevier/tsk/result"
)

type BatchOpts struct {
	MaxSize   int
	MaxLinger time.Duration
}

type RunBatchFunction[T any, R any] func(tasks []T) ([]result.Result[R], error)

type batch[T any, R any] struct {
	id      int
	tasks   []T
	futures []*futures.Future[R]
}

func (b *batch[T, R]) add(task T, future *futures.Future[R]) {
	b.tasks = append(b.tasks, task)
	b.futures = append(b.futures, future)
}

type BatchExecutor[T any, R any] struct {
	m            *sync.Mutex
	sequenceNum  int
	currentBatch *batch[T, R]
	run          RunBatchFunction[T, R]
	maxSize      int
	maxLinger    time.Duration
}

func NewExecutor[T any, R any](opts BatchOpts, run RunBatchFunction[T, R]) *BatchExecutor[T, R] {
	return &BatchExecutor[T, R]{
		m:           &sync.Mutex{},
		sequenceNum: 0,
		run:         run,
		maxSize:     opts.MaxSize,
		maxLinger:   opts.MaxLinger,
	}
}

func (be *BatchExecutor[T, R]) Submit(ctx context.Context, task T) (R, error) {
	f := be.SubmitF(ctx, task)
	return f.Get(ctx)
}

func (be *BatchExecutor[T, R]) SubmitF(ctx context.Context, task T) *futures.Future[R] {
	future := futures.New[R](ctx)
	be.addTask(task, future)
	return future
}

func (be *BatchExecutor[T, R]) addTask(task T, future *futures.Future[R]) {
	be.m.Lock()
	defer be.m.Unlock()

	if be.currentBatch == nil {
		be.currentBatch = be.newBatch()
	}
	be.currentBatch.add(task, future)

	if len(be.currentBatch.tasks) >= be.maxSize {
		go be.runBatch(be.currentBatch)
		be.currentBatch = nil
	}
}

func (be *BatchExecutor[T, R]) newBatch() *batch[T, R] {
	be.sequenceNum++

	b := &batch[T, R]{
		id:    be.sequenceNum,
		tasks: make([]T, 0, be.maxSize),
	}

	go be.expireBatch(b.id)
	return b
}

func (be *BatchExecutor[T, R]) expireBatch(batchId int) {
	time.Sleep(be.maxLinger)

	be.m.Lock()
	defer be.m.Unlock()

	if be.currentBatch != nil && be.currentBatch.id == batchId {
		go be.runBatch(be.currentBatch)
		be.currentBatch = nil
	}
}

func (be *BatchExecutor[T, R]) runBatch(b *batch[T, R]) {
	res, err := be.run(b.tasks)
	if err != nil {
		for _, f := range b.futures {
			f.Fail(err)
		}
	}

	//TODO: verify that result length is the same as the task length

	for i, r := range res {
		b.futures[i].CompleteWithResult(r)
	}
}

// func (b *batch[T, R]) sendResult(idx int, result result.Result[R]) {
// 	ctx := b.contexts[idx]
// 	resultChan := b.resultChans[idx]

// 	select {
// 	case resultChan <- result:
// 	case <-ctx.Done():
// 	}

// 	close(resultChan)
// }
