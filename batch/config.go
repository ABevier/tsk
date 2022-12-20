package batch

import (
	"errors"
	"time"
)

var (
	ErrBatchResultMismatch = errors.New("the batch function did not return the correct number of responses")
)

type BatchOpts struct {
	MaxSize   int
	MaxLinger time.Duration
}

func (o BatchOpts) validate() {
	if o.MaxSize <= 1 {
		panic("maximum batch size must be greater than 1")
	}

	if o.MaxLinger <= 0 {
		panic("batch linger must be greater than 0")
	}
}
