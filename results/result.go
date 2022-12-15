package results

type Result[R any] struct {
	Val R
	Err error
}

func New[R any](val R, err error) Result[R] {
	return Result[R]{Val: val, Err: err}
}

func Success[T any](val T) Result[T] {
	return Result[T]{Val: val}
}

func Failure[T any](err error) Result[T] {
	return Result[T]{Err: err}
}
