package tsk

type Result[R any] struct {
	Val R
	Err error
}

func NewResult[R any](val R, err error) Result[R] {
	return Result[R]{Val: val, Err: err}
}

func NewSuccess[T any](val T) Result[T] {
	return Result[T]{Val: val}
}

func NewFailure[T any](err error) Result[T] {
	return Result[T]{Err: err}
}
