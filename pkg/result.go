package tsk

type Result[T any] struct {
	Val T
	Err error
}

func NewSuccess[T any](val T) Result[T] {
	return Result[T]{Val: val}
}

func NewFailure[T any](err error) Result[T] {
	return Result[T]{Err: err}
}
