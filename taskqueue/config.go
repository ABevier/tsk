package taskqueue

import "github.com/abevier/tsk/internal/submit"

type FullQueueStrategy submit.FullQueueStrategy

const (
	BlockWhenFull FullQueueStrategy = FullQueueStrategy(submit.BlockWhenFull)
	ErrorWhenFull FullQueueStrategy = FullQueueStrategy(submit.ErrorWhenFull)
)

type TaskQueueOpts struct {
	MaxWorkers        int
	MaxQueueDepth     int
	FullQueueStrategy FullQueueStrategy
}

func (o TaskQueueOpts) validate() {

}
