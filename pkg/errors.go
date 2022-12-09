package tsk

import "errors"

var (
	ErrQueueFull = errors.New("task queue is full")
	ErrStopped   = errors.New("task queue has been stopped")
)
