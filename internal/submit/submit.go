package submit

import (
	"context"
	"errors"
	"log"
)

var (
	ErrQueueFull = errors.New("task queue is full")
)

type FullQueueStrategy int

const (
	BlockWhenFull FullQueueStrategy = iota
	ErrorWhenFull
)

type SubmitFunction[T any, R any] func(taskChan chan<- TaskFuture[T, R], tf TaskFuture[T, R]) error

func GetSubmitFunction[T any, R any](s FullQueueStrategy) SubmitFunction[T, R] {
	switch s {
	case BlockWhenFull:
		return blockWhenFullStrategy[T, R]
	case ErrorWhenFull:
		return errorWhenFullStrategy[T, R]
	default:
		log.Panicf("invalid submit strategy value %d", s)
	}
	return blockWhenFullStrategy[T, R]
}

func blockWhenFullStrategy[T any, R any](taskChan chan<- TaskFuture[T, R], t TaskFuture[T, R]) error {
	select {
	case taskChan <- t:
		return nil
	case <-t.Ctx.Done():
		return context.Canceled
	}
}

func errorWhenFullStrategy[T any, R any](taskChan chan<- TaskFuture[T, R], t TaskFuture[T, R]) error {
	select {
	case taskChan <- t:
		return nil
	default:
		return ErrQueueFull
	}
}
