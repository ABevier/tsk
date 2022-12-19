package ratelimiter

import (
	"time"

	"github.com/abevier/tsk/internal/submit"
	"golang.org/x/time/rate"
)

var (
	ErrQueueFull = submit.ErrQueueFull
)

type FullQueueStrategy submit.FullQueueStrategy

const (
	BlockWhenFull FullQueueStrategy = FullQueueStrategy(submit.BlockWhenFull)
	ErrorWhenFull FullQueueStrategy = FullQueueStrategy(submit.ErrorWhenFull)
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
