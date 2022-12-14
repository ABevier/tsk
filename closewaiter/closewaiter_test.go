package closewaiter

import (
	"sync"
	"testing"
)

func TestCloseWaiter(t *testing.T) {
	cw := New()

	testChan := make(chan int)
	shutdownSignal := make(chan struct{})

	wg := sync.WaitGroup{}

	// start 3 writers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			var err error
			for err == nil {
				err = cw.Do(func() {
					testChan <- 1
				})
			}
			wg.Done()
		}()
	}

	// single reader
	cnt := 0
	go func() {
		for {
			<-testChan
			cnt++
			if cnt == 100 {
				// simulate a shutdown, but don't stop reading or else the publishers will block in Do
				// normal shutdown sequence stops writers before readers to allow thigs to drain
				close(shutdownSignal)
			}
		}
	}()

	// Let the work flow until a shutdown is signaled
	<-shutdownSignal

	// should not panic
	cw.Close(func() {
		close(testChan)
	})

	// all writers should have closed "gracefully" during the shutdown sequence
	wg.Wait()
}
