// `metrics/repository.go`
// содержит в себе интерфейсы для работы с адаптерами (внешними системами)
package metrics

import (
	"context"

	"github.com/76Parker/metrico/internal/domain/metrics"
)

// metricStorage интерфейс для работы хранилищем метрик
type metricStorage interface {
	UpdateOrCreate(ctx context.Context, metricName string, metric metrics.Metrics) error
	Get(ctx context.Context, metricName string) (metrics.Metrics, error)
	GetAll(ctx context.Context) ([]metrics.Metrics, error)
}
