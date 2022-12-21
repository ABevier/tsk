package batch

import (
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	failIfNoPanic := func(f func()) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("")
			}
		}()

		f()
	}

	opts := Opts{MaxSize: 1, MaxLinger: 10 * time.Millisecond}
	failIfNoPanic(opts.validate)

	opts = Opts{MaxSize: 3}
	failIfNoPanic(opts.validate)
}
