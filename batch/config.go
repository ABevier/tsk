package batch

import "time"

type BatchOpts struct {
	MaxSize   int
	MaxLinger time.Duration
}

func (o BatchOpts) validate() {
	// TODO
}
