package futures

import (
	"context"

	"github.com/abevier/tsk/results"
)

// ResolveAll waits for all of the provided Futures to complete and returns a results.Result for each
// future at the index corresponding to the provided slice.
// If the provided context is canceled, the cancellation error will be returned as an error by this function.
func ResolveAll[T any](ctx context.Context, fs []*Future[T]) ([]results.Result[T], error) {
	res := make([]results.Result[T], 0, len(fs))

	for _, f := range fs {
		r, err := f.Get(ctx)
		res = append(res, results.New(r, err))
		// check for error at the end of the loop to avoid the race of cancelling while Getting the last value in the list
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}

	return res, nil
}
