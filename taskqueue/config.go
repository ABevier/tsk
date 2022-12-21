package taskqueue

import (
	"errors"

	"github.com/abevier/tsk/internal/tsk"
)

var (
	ErrQueueFull = tsk.ErrQueueFull
	ErrStopped   = errors.New("task queue has been stopped")
)

type FullQueueStrategy tsk.FullQueueStrategy

const (
	BlockWhenFull FullQueueStrategy = FullQueueStrategy(tsk.BlockWhenFull)
	ErrorWhenFull FullQueueStrategy = FullQueueStrategy(tsk.ErrorWhenFull)
)

type Opts struct {
	MaxWorkers        int
	MaxQueueDepth     int
	FullQueueStrategy FullQueueStrategy
}

func (o Opts) validate() {
	if o.MaxWorkers <= 0 {
		panic("task queue max workers must be greater than 0")
	}

	if o.MaxQueueDepth < 0 {
		panic("task queue max queue depth must be 0 or greater")
	}
}
