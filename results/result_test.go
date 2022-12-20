package results

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResult(t *testing.T) {
	require := require.New(t)

	r := New(1, nil)
	require.Equal(1, r.Val)
	require.NoError(r.Err)

	r = Success(2)
	require.Equal(2, r.Val)
	require.NoError(r.Err)

	errTest := errors.New("test err")
	r = Failure[int](errTest)
	require.Equal(0, r.Val)
	require.ErrorIs(r.Err, errTest)
}
