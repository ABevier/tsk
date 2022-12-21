package taskqueue

import "testing"

func TestConfig(t *testing.T) {
	failIfNoPanic := func(f func()) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("")
			}
		}()

		f()
	}

	opts := Opts{MaxWorkers: 0, MaxQueueDepth: 10}
	failIfNoPanic(opts.validate)

	opts = Opts{MaxWorkers: 3, MaxQueueDepth: -1}
	failIfNoPanic(opts.validate)
}
