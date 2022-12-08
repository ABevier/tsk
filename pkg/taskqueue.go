package tsk

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
)

//TODO: rename this lol
type taskWrapper[T any, R any] struct {
	task       T
	resultChan chan<- Result[R]
}

type TaskQueue[T any, R any] struct {
	isStopping uint32

	taskChan chan taskWrapper[T, R]

	// Protecting channels with N writers is hard :-(
	waitSend *sync.WaitGroup
	waitStop *sync.WaitGroup
	stopOnce *sync.Once
}

type RunFunction[T any, R any] func(task T) (R, error)

func NewTaskQueue[T any, R any](maxWorkers int, run RunFunction[T, R]) *TaskQueue[T, R] {
	taskChan := make(chan taskWrapper[T, R])
	waitStop := sync.WaitGroup{}

	for i := 0; i < maxWorkers; i++ {

		go func(n int) {
			waitStop.Add(1)
			defer waitStop.Done()

			for {
				select {
				case task, ok := <-taskChan:
					if !ok {
						return
					}
					log.Printf("Running task on worker %d", n)

					res, err := run(task.task)
					task.resultChan <- Result[R]{Val: res, Err: err}
					close(task.resultChan)
				}
			}
		}(i)
	}

	return &TaskQueue[T, R]{
		taskChan: taskChan,
		waitSend: &sync.WaitGroup{},
		waitStop: &waitStop,
		stopOnce: &sync.Once{},
	}
}

func (tq *TaskQueue[T, R]) Submit(task T) (R, error) {
	tq.waitSend.Add(1)
	defer tq.waitSend.Done() // Consider releasing this immediately after the write??

	if atomic.LoadUint32(&tq.isStopping) == 1 {
		//TODO: fix this error
		return *new(R), errors.New("fix me")
	}

	resultChan := make(chan Result[R])
	tw := taskWrapper[T, R]{task: task, resultChan: resultChan}

	// TODO: this blocks, but add a strategy to return an error
	tq.taskChan <- tw

	// TODO: check okay and return an error?
	result := <-resultChan
	return result.Val, result.Err
}

func (tq *TaskQueue[T, R]) Stop() {
	tq.stopOnce.Do(func() {
		atomic.StoreUint32(&tq.isStopping, 1)
		tq.waitSend.Wait()
		close(tq.taskChan)
	})

	// TODO: allow a timeout
	tq.waitStop.Wait()
}
