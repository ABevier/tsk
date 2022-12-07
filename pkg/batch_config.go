package tsk

import "time"

type BatchOpts struct {
	MaxSize   int
	MaxLinger time.Duration
}

//TODO: validate config and set defaults
