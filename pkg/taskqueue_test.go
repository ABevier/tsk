package tsk

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTaskQueue(t *testing.T) {
	require := require.New(t)

	run := func(task int) (int, error) {
		log.Printf("i am task %d", task)
		time.Sleep(100 * time.Millisecond)
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
