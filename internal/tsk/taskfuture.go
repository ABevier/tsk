package tsk

import (
	"context"

	"github.com/abevier/tsk/futures"
)

type TaskFuture[T any, R any] struct {
	Ctx    context.Context
	Task   T
	Future *futures.Future[R]
}

func NewTaskFuture[T any, R any](ctx context.Context, task T) TaskFuture[T, R] {
	return TaskFuture[T, R]{
		Ctx:    ctx,
		Task:   task,
		Future: futures.New[R](),
	}
}
