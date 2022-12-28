package results

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResult(t *testing.T) {
	req := require.New(t)

	r := New(1, nil)
	req.Equal(1, r.Val)
	req.NoError(r.Err)

	r = Success(2)
	req.Equal(2, r.Val)
	req.NoError(r.Err)

	errTest := errors.New("test err")
	r = Failure[int](errTest)
	req.Equal(0, r.Val)
	req.ErrorIs(r.Err, errTest)
}
