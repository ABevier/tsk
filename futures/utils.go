package futures

import (
	"context"

	"github.com/abevier/tsk/results"
)

func ResolveAll[T any](ctx context.Context, fs []*Future[T]) ([]results.Result[T], error) {
	res := make([]results.Result[T], 0, len(fs))

	for _, f := range fs {
		r, err := f.Get(ctx)
		res = append(res, results.New(r, err))
	}

	return res, nil
}
