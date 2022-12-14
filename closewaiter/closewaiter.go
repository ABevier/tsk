package closewaiter

import (
	"errors"
	"runtime"
	"sync/atomic"
)

const (
	open     = 0
	closed   = 1
	minusOne = ^uint32(0)
)

var (
	ErrClosed = errors.New("closed")
)

type CloseWaiter struct {
	isClosed  uint32
	activeCnt uint32

	closed chan struct{}
}

func New() *CloseWaiter {
	return &CloseWaiter{
		closed: make(chan struct{}),
	}
}

func (c *CloseWaiter) Do(f func()) error {
	atomic.AddUint32(&c.activeCnt, 1)
	defer atomic.AddUint32(&c.activeCnt, minusOne)

	if atomic.LoadUint32(&c.isClosed) == closed {
		return ErrClosed
	}

	f()
	return nil
}

func (c *CloseWaiter) Close(f func()) {
	if atomic.CompareAndSwapUint32(&c.isClosed, open, closed) {
		go func() {
			for atomic.LoadUint32(&c.activeCnt) != 0 {
				// busy wait while yielding until all calls to Do have exited
				runtime.Gosched()
			}

			f()

			close(c.closed)
		}()
	}

	<-c.closed
}
