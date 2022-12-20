package taskqueue

import "github.com/abevier/tsk/internal/tsk"

type FullQueueStrategy tsk.FullQueueStrategy

const (
	BlockWhenFull FullQueueStrategy = FullQueueStrategy(tsk.BlockWhenFull)
	ErrorWhenFull FullQueueStrategy = FullQueueStrategy(tsk.ErrorWhenFull)
)

type TaskQueueOpts struct {
	MaxWorkers        int
	MaxQueueDepth     int
	FullQueueStrategy FullQueueStrategy
}

func (o TaskQueueOpts) validate() {

}
