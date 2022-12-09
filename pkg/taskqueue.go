package tsk

import (
	"log"
	"sync"
	"sync/atomic"
)

//TODO: rename this lol
type taskWrapper[T any, R any] struct {
	task       T
	resultChan chan<- Result[R]
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

func NewTaskQueue[T any, R any](maxWorkers int, maxQueueSize int, run RunFunction[T, R]) *TaskQueue[T, R] {
	taskChan := make(chan taskWrapper[T, R], maxQueueSize)
	waitStop := sync.WaitGroup{}

	tq := &TaskQueue[T, R]{
		run: run,
		//submit:   submitBlockWhenFull[T, R],
		submit:   submitErrorWhenFull[T, R],
		taskChan: taskChan,
		waitSend: &sync.WaitGroup{},
		waitStop: &waitStop,
		stopOnce: &sync.Once{},
	}

	for i := 0; i < maxWorkers; i++ {
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
			//TODO: how to log in a library
			log.Printf("Running task on worker %d", workerNum)

			res, err := tq.run(task.task)
			task.resultChan <- NewResult(res, err)
			close(task.resultChan)
		}
	}
}

// TODO: accept a context and allow cancel
func (tq *TaskQueue[T, R]) Submit(task T) (R, error) {
	tq.waitSend.Add(1)
	defer tq.waitSend.Done() // Consider releasing this immediately after the write??

	if atomic.LoadUint32(&tq.isStopping) == 1 {
		return *new(R), ErrStopped
	}

	resultChan := make(chan Result[R])
	tw := taskWrapper[T, R]{task: task, resultChan: resultChan}

	if err := tq.submit(tq.taskChan, tw); err != nil {
		return *new(R), err
	}

	result := <-resultChan
	return result.Val, result.Err
}

func submitBlockWhenFull[T any, R any](taskChan chan<- taskWrapper[T, R], t taskWrapper[T, R]) error {
	taskChan <- t
	return nil
}

func submitErrorWhenFull[T any, R any](taskChan chan<- taskWrapper[T, R], t taskWrapper[T, R]) error {
	log.Println("FFF")
	select {
	case taskChan <- t:
		return nil
	default:
		log.Println("ZZZ")
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
