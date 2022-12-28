package taskqueue

import (
	"context"
	"strconv"
)

type workerIDKey struct{}

func withWorkerID(ctx context.Context, id int) context.Context {
	return context.WithValue(ctx, workerIDKey{}, "worker-"+strconv.Itoa(id))
}

// WorkerIDFromContext attempts to retrieve a worker id string from the current context.
// The worker id string is added to the current context by the TaskQueue before invoking the run
// function. This id can be useful for logging.
func WorkerIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(workerIDKey{}).(string)
	return v, ok
}
