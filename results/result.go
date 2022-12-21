// Package results provides an implementation of a generic result type.
// This type is typically used for storing the result of functions with the following signature:
//   func() (R, error)
// A result would commonly be used in one of the following ways:
//   value, err := myFunction()
//   result := results.New(value, err)
// Or
//   value, err := myFunction()
//   if err != nil {
//	   return results.Failure[T](err)
//   }
//   return results.Success(value)
package results

// Result combines a value of type R and an error.
type Result[R any] struct {
	Val R
	Err error
}

// New creates a new Result with the specified value and error
func New[R any](val R, err error) Result[R] {
	return Result[R]{Val: val, Err: err}
}

// Success creates a new Result that represents a successful function invocation that returned the provided value
func Success[T any](val T) Result[T] {
	return Result[T]{Val: val}
}

// Failure creates a new Result that represents a failed function invocation that returned the provided error
func Failure[T any](err error) Result[T] {
	return Result[T]{Err: err}
}
