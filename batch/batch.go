package batch

import (
	"context"
	"sync"
	"time"

	"github.com/abevier/tsk/result"
)

type BatchOpts struct {
	MaxSize   int
	MaxLinger time.Duration
}

type RunBatchFunction[T any, R any] func(tasks []T) ([]result.Result[R], error)

type batch[T any, R any] struct {
	id          int
	tasks       []T
	resultChans []chan<- result.Result[R]
}

func (b *batch[T, R]) add(task T, resultChan chan<- result.Result[R]) {
	b.tasks = append(b.tasks, task)
	b.resultChans = append(b.resultChans, resultChan)
}

type BatchExecutor[T any, R any] struct {
	m            *sync.Mutex
	sequenceNum  int
	currentBatch *batch[T, R]
	run          RunBatchFunction[T, R]
	maxSize      int
	maxLinger    time.Duration
}

func NewBatchExecutor[T any, R any](opts BatchOpts, run RunBatchFunction[T, R]) *BatchExecutor[T, R] {
	return &BatchExecutor[T, R]{
		m:           &sync.Mutex{},
		sequenceNum: 0,
		run:         run,
		maxSize:     opts.MaxSize,
		maxLinger:   opts.MaxLinger,
	}
}

func (be *BatchExecutor[T, R]) Submit(ctx context.Context, task T) (R, error) {
	resultChan := make(chan result.Result[R])
	be.addTask(task, resultChan)

	select {
	case res := <-resultChan:
		return res.Val, res.Err

	case <-ctx.Done():
		return *new(R), context.Canceled
	}
}

func (be *BatchExecutor[T, R]) addTask(task T, resultChan chan<- result.Result[R]) {
	be.m.Lock()
	defer be.m.Unlock()

	if be.currentBatch == nil {
		be.currentBatch = be.newBatch()
	}
	be.currentBatch.add(task, resultChan)

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
		for _, c := range b.resultChans {
			c <- result.NewFailure[R](err)
			close(c)
		}
	}

	for i, r := range res {
		c := b.resultChans[i]
		c <- r
		close(c)
	}
}
