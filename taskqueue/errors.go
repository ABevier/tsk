package taskqueue

import (
	"errors"

	"github.com/abevier/tsk/internal/tsk"
)

var (
	ErrQueueFull = tsk.ErrQueueFull
	ErrStopped   = errors.New("task queue has been stopped")
)
