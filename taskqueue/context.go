package taskqueue

import (
	"context"
	"strconv"
)

type workerIDKey struct{}

func withWorkerID(ctx context.Context, id int) context.Context {
	return context.WithValue(ctx, workerIDKey{}, "worker-"+strconv.Itoa(id))
}

func WorkerIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(workerIDKey{}).(string)
	return v, ok
}
