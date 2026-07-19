// `metrics/repository.go`
// содержит в себе интерфейсы для работы с адаптерами (внешними системами)
package metrics

import (
	"context"

	"github.com/76Parker/metrico/internal/domain/metrics"
)

// metricStorage интерфейс для работы с адаптерами (внешними системами)
// при необходимости вынести в общий порт если будет использоваться несколькими usecase-ами
type metricStorage interface {
	UpdateOrCreateMetricByName(ctx context.Context, metricName string, metric metrics.Metrics) error
	GetMetricByName(ctx context.Context, metricName string) (metrics.Metrics, error)
	GetAllMetrics(ctx context.Context) (map[string]metrics.Metrics, error)
}
