package tsk

import (
	"context"
	"sync"
	"time"
)

//TODO:
// Batching:
// - configuration
// - comments
// - readme
// - examples: SQS, SQL
// - server context to shutdown the BatchExecutor
// ---------
// Worker Queue (control max concurrency):
// - initial implementation
// - backpressure?

type ExecuteFunction[T any, R any] func(items []T) ([]Result[R], error)

type batch[T any, R any] struct {
	id          int
	items       []T
	resultChans []chan<- Result[R]
}

func (b *batch[T, R]) add(item T, resultChan chan<- Result[R]) {
	b.items = append(b.items, item)
	b.resultChans = append(b.resultChans, resultChan)
}

type BatchExecutor[T any, R any] struct {
	m            *sync.Mutex
	sequenceNum  int
	currentBatch *batch[T, R]
	execute      ExecuteFunction[T, R]
	maxSize      int
	maxLinger    time.Duration
}

func NewBatchExecutor[T any, R any](execute ExecuteFunction[T, R], opts BatchOpts) *BatchExecutor[T, R] {
	return &BatchExecutor[T, R]{
		m:           &sync.Mutex{},
		sequenceNum: 0,
		execute:     execute,
		maxSize:     opts.MaxSize,
		maxLinger:   opts.MaxLinger,
	}
}

func (be *BatchExecutor[T, R]) Submit(ctx context.Context, item T) (R, error) {
	resultChan := make(chan Result[R])
	be.addItem(item, resultChan)

	select {
	case res := <-resultChan:
		return res.Val, res.Err

	case <-ctx.Done():
		return *new(R), context.Canceled
	}
}

func (be *BatchExecutor[T, R]) addItem(item T, resultChan chan<- Result[R]) {
	be.m.Lock()
	defer be.m.Unlock()

	if be.currentBatch == nil {
		be.currentBatch = be.newBatch()
	}
	be.currentBatch.add(item, resultChan)

	if len(be.currentBatch.items) >= be.maxSize {
		go be.executeBatch(be.currentBatch)
		be.currentBatch = nil
	}
}

func (be *BatchExecutor[T, R]) newBatch() *batch[T, R] {
	be.sequenceNum++

	b := &batch[T, R]{
		id:    be.sequenceNum,
		items: make([]T, 0, be.maxSize),
	}

	go be.expireBatch(b.id)
	return b
}

func (be *BatchExecutor[T, R]) expireBatch(batchId int) {
	time.Sleep(be.maxLinger)

	be.m.Lock()
	defer be.m.Unlock()

	if be.currentBatch != nil && be.currentBatch.id == batchId {
		go be.executeBatch(be.currentBatch)
		be.currentBatch = nil
	}
}

func (be *BatchExecutor[T, R]) executeBatch(b *batch[T, R]) {
	res, err := be.execute(b.items)
	if err != nil {
		for _, c := range b.resultChans {
			c <- NewFailure[R](err)
			close(c)
		}
	}

	for i, r := range res {
		c := b.resultChans[i]
		c <- r
		close(c)
	}
}
