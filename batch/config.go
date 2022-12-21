package batch

import (
	"errors"
	"time"
)

var (
	// ErrBatchResultMismatch is returned when a RunBatchFunction does not return exactly one Result for each provided task.
	ErrBatchResultMismatch = errors.New("the batch function did not return the correct number of responses")
)

// Opts is used to configure a BatchExecutor via the New function.
type Opts struct {
	// MaxSize is the maximum batch size allowed before a batch is flushed for processing.  It must be greater than 1.
	MaxSize int
	// MaxLinger is the maximum amount of time a batch will remain open before it is flushed due to a timeout.
	// This timer is started when the first item is added to a batch.
	MaxLinger time.Duration
}

func (o Opts) validate() {
	if o.MaxSize <= 1 {
		panic("maximum batch size must be greater than 1")
	}

	if o.MaxLinger <= 0 {
		panic("batch linger must be greater than 0")
	}
}
