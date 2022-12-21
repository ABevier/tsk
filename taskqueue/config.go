package taskqueue

import (
	"github.com/abevier/tsk/internal/tsk"
)

var (
	// ErrQueueFull is the error returned when the FullQueueStrategy is ErrorWhenFull and MaxQueueDepth is exceeded
	ErrQueueFull = tsk.ErrQueueFull
)

// FullQueueStategy is the type of behavior that should occcur when too many items are submitted to the TaskQueue
type FullQueueStrategy tsk.FullQueueStrategy

const (
	// BlockWhenFull exerts back pressure by blocking the caller when too many items have been submitted.
	BlockWhenFull FullQueueStrategy = FullQueueStrategy(tsk.BlockWhenFull)
	// ErrorWhenFull immediately returns an error when too many items have been submitted.
	ErrorWhenFull FullQueueStrategy = FullQueueStrategy(tsk.ErrorWhenFull)
)

// Opts is used to configure a TaskQueue via the New function.
type Opts struct {
	// MaxWorkers controls the maximum number of concurrent tasks that the task queue will run
	MaxWorkers int
	// MaxQueueDepth controls the maximum number of outstanding tasks that can be submitted to the task queue.
	MaxQueueDepth int
	// FullQueueStategy determines the task queue's behavior when the MaxQueueDepth is exceeded.
	// By default the rate limiter will block the caller.
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
