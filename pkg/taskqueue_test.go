package tsk

import (
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTaskQueue(t *testing.T) {
	require := require.New(t)

	run := func(task int) (int, error) {
		log.Printf("i am task %d", task)
		time.Sleep(randWait(10, 50))
		return task * 2, nil
	}

	tq := NewTaskQueue(3, run)

	for i := 0; i < 100; i++ {
		val, err := tq.Submit(i)
		require.NoError(err)
		require.Equal(i*2, val)
	}

	tq.Stop()
}

func TestTaskQueueCloseMultipleTimes(t *testing.T) {
	wg := sync.WaitGroup{}

	run := func(task int) (int, error) {
		time.Sleep(randWait(10, 100))
		return task * 2, nil
	}

	tq := NewTaskQueue(3, run)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			time.Sleep(randWait(30, 100))
			tq.Submit(n)
			tq.Stop()
		}(i)
	}

	wg.Wait()
	//Test should not block or panic
}

func randWait(min, max int) time.Duration {
	n := rand.Intn(max-min) + min
	return time.Duration(n) * time.Millisecond
}
