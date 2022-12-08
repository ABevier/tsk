package tsk

import "time"

type BatchOpts struct {
	MaxSize   int
	MaxLinger time.Duration
}
