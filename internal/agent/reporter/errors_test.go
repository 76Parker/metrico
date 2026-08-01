package reporter

import (
	"errors"
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRetryable(t *testing.T) {
	t.Run("valid/dial_error", func(t *testing.T) {
		err := &url.Error{
			Op: "Post",
			Err: &net.OpError{
				Op:  "dial",
				Err: errors.New("connection refused"),
			},
		}

		assert.True(t, isRetryable(err))
	})

	t.Run("invalid/non_dial_error", func(t *testing.T) {
		assert.False(t, isRetryable(errors.New("unexpected error")))
	})
}
