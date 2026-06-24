// `metrics/repository.go`
// содержит в себе интерфейсы для работы с адаптерами (внешними системами)
package metrics

import (
	"context"

	"github.com/76Parker/metrico/internal/domain/metrics"
)

type metricStorage interface {
	UpdateOrCreateMetricByName(ctx context.Context, metricName string, metric metrics.Metrics) error
}
