package snapshot

import (
	"context"

	"github.com/76Parker/metrico/internal/domain/metrics"
)

type metricsStore interface {
	GetAll(ctx context.Context) ([]metrics.Metrics, error)
	Load(ctx context.Context, metrics []metrics.Metrics) error
}

type snapshotStore interface {
	Save(ctx context.Context, metrics []metrics.Metrics) error
	Restore(ctx context.Context) ([]metrics.Metrics, error)
}
