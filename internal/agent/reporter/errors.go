package reporter

import (
	"errors"
	"net"
)

func isRetryable(err error) bool {
	var opErr *net.OpError
	return errors.As(err, &opErr) && opErr.Op == "dial"
}
