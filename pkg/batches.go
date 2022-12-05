package tasks

import (
	"sync"
	"time"
)

type ExecuteFunction[T any] func(items []T)

type BatchExecutor[T any] struct {
	m            *sync.Mutex
	sequenceNum  int
	currentBatch *batch[T]
	execute      ExecuteFunction[T]
	maxSize      int
	maxLinger    time.Duration
}

type batch[T any] struct {
	id    int
	items []T
}

func NewBatchExecutor[T any](maxSize int, maxLinger time.Duration, execute ExecuteFunction[T]) *BatchExecutor[T] {
	return &BatchExecutor[T]{
		m:           &sync.Mutex{},
		sequenceNum: 0,
		execute:     execute,
		maxSize:     maxSize,
		maxLinger:   maxLinger,
	}
}

func (be *BatchExecutor[T]) AddItem(item T) {
	be.m.Lock()
	defer be.m.Unlock()

	if be.currentBatch == nil {
		be.currentBatch = be.newBatch()
	}
	be.currentBatch.items = append(be.currentBatch.items, item)

	if len(be.currentBatch.items) >= be.maxSize {
		go be.execute(be.currentBatch.items)
		be.currentBatch = nil
	}
}

func (be *BatchExecutor[T]) newBatch() *batch[T] {
	be.sequenceNum++

	b := &batch[T]{
		id:    be.sequenceNum,
		items: make([]T, 0, be.maxSize),
	}

	go be.expireBatch(b.id)
	return b
}

func (be *BatchExecutor[T]) expireBatch(batchId int) {
	time.Sleep(be.maxLinger)

	be.m.Lock()
	defer be.m.Unlock()

	if be.currentBatch != nil && be.currentBatch.id == batchId {
		go be.execute(be.currentBatch.items)
		be.currentBatch = nil
	}
}
