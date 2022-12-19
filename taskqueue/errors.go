package taskqueue

import (
	"errors"

	"github.com/abevier/tsk/internal/submit"
)

var (
	ErrQueueFull = submit.ErrQueueFull
	ErrStopped   = errors.New("task queue has been stopped")
)
