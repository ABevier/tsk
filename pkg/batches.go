package tasks

import (
	"context"
	"sync"
	"time"
)

//TODO:
// - is there a better way to return results??
// - context cancellation
// - configuration
// - comments

type ExecuteFunction[T any, R any] func(items []T) ([]Result[R], error)

type batch[T any, R any] struct {
	id          int
	items       []T
	resultChans []chan<- Result[R]
}

type Result[R any] struct {
	Val R
	Err error
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

func NewBatchExecutor[T any, R any](maxSize int, maxLinger time.Duration, execute ExecuteFunction[T, R]) *BatchExecutor[T, R] {
	return &BatchExecutor[T, R]{
		m:           &sync.Mutex{},
		sequenceNum: 0,
		execute:     execute,
		maxSize:     maxSize,
		maxLinger:   maxLinger,
	}
}

func (be *BatchExecutor[T, R]) Submit(ctx context.Context, item T) (R, error) {
	resultChan := make(chan Result[R])
	be.addItem(item, resultChan)

	res := <-resultChan
	return res.Val, res.Err
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
			c <- Result[R]{Err: err}
			close(c)
		}
	}

	for i, r := range res {
		c := b.resultChans[i]
		c <- r
		close(c)
	}
}
