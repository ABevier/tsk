package batch

import "errors"

var (
	ErrBatchResultMismatch = errors.New("the batch function did not return the correct number of responses")
)
