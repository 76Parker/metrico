package health

import "context"

type dbPinger interface {
	Ping(ctx context.Context) error
}
