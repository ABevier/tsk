package futures

import (
	"context"
	"testing"
	"time"

	"github.com/abevier/tsk/results"
	"github.com/stretchr/testify/require"
)

func TestResolveAll(t *testing.T) {
	require := require.New(t)

	f1 := FromFunc(func() (int, error) {
		time.Sleep(6 * time.Millisecond)
		return 1, nil
	})

	f2 := FromFunc(func() (int, error) {
		time.Sleep(4 * time.Millisecond)
		return 2, nil
	})

	f3 := FromFunc(func() (int, error) {
		time.Sleep(2 * time.Millisecond)
		return 3, nil
	})

	rs, err := ResolveAll(context.Background(), []*Future[int]{f1, f2, f3})
	require.NoError(err)

	expected := []results.Result[int]{
		results.Success(1),
		results.Success(2),
		results.Success(3),
	}

	require.Equal(expected, rs)
}
