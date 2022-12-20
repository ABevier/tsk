package ratelimiter

import (
	"time"

	"github.com/abevier/tsk/internal/tsk"
	"golang.org/x/time/rate"
)

var (
	ErrQueueFull = tsk.ErrQueueFull
)

type FullQueueStrategy tsk.FullQueueStrategy

const (
	BlockWhenFull FullQueueStrategy = FullQueueStrategy(tsk.BlockWhenFull)
	ErrorWhenFull FullQueueStrategy = FullQueueStrategy(tsk.ErrorWhenFull)
)

type Limit = rate.Limit

func Every(interval time.Duration) Limit {
	return rate.Every(interval)
}

type RateLimiterOpts struct {
	Limit Limit
	Burst int

	MaxQueueDepth     int
	FullQueueStrategy FullQueueStrategy
}

func (o RateLimiterOpts) validate() {
	//TODO:
}
