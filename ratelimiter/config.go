package ratelimiter

import (
	"time"

	"github.com/abevier/tsk/internal/tsk"
	"golang.org/x/time/rate"
)

var (
	ErrQueueFull = tsk.ErrQueueFull
)

// FullQueueStategy is the type of behavior that should occcur when too many items are submitted to the rate limiter
type FullQueueStrategy tsk.FullQueueStrategy

const (
	// BlockWhenFull exerts back pressure by blocking the caller when too many items have been submitted.
	BlockWhenFull FullQueueStrategy = FullQueueStrategy(tsk.BlockWhenFull)
	// ErrorWhenFull immediately returns an error when too many items have been submitted.
	ErrorWhenFull FullQueueStrategy = FullQueueStrategy(tsk.ErrorWhenFull)
)

// A rate limit expressed as N requests per second
type Limit = rate.Limit

// Every converts the provided duration into a number of requests per second
// for instance Every(100 * time.Milliseconds) will yield 10 requests per second
func Every(interval time.Duration) Limit {
	return rate.Every(interval)
}

// Opts is used to configure a RateLimiter via the New function.
type Opts struct {
	// Limit is the rate limit expressed in requests per second.
	Limit Limit
	// Burst is the size of the Token Bucket
	Burst int
	// MaxQueueDepth controls the number of outstanding tasks that can be submitted to the rate limiter.
	MaxQueueDepth int
	// FullQueueStategy determines the rate limiter's behavior when the MaxQueueDepth is exceeded.
	// By default the rate limiter will block the caller.
	FullQueueStrategy FullQueueStrategy
}

func (o Opts) validate() {
	if o.Limit < 0 {
		panic("rate limiter limit must be 0 or greater")
	}

	if o.Burst < 1 {
		panic("rate limiter burst must be 1 or greater")
	}

	if o.MaxQueueDepth < 0 {
		panic("rate limiter max queue depth must be 0 or greater")
	}
}
