// `metrics/service.go`
// сервисный слой для работы с метриками.
// Описывает пользовательские сценарии взаимодействия с метриками.
package metrics

import (
	"context"

	"github.com/76Parker/metrico/internal/domain/metrics"
)

type Service struct {
	metricsStorage metricStorage
}

func NewService(storage metricStorage) *Service {
	return &Service{
		metricsStorage: storage,
	}
}

func (s *Service) UpdateOrCreateMetric(ctx context.Context, cmd UpdateCommand) error {
	switch cmd.MetricType {
	case metrics.Gauge:
		if cmd.Value == nil {
			return metrics.ErrInvalidValueForGauge
		}
		metric := metrics.Metrics{
			ID:    cmd.Name,
			Type:  metrics.Gauge,
			Delta: nil,
			Value: cmd.Value,
		}
		return s.metricsStorage.UpdateOrCreate(ctx, cmd.Name, metric)
	case metrics.Counter:
		if cmd.Delta == nil {
			return metrics.ErrInvalidValueForCounter
		}
		metric := metrics.Metrics{
			ID:    cmd.Name,
			Type:  metrics.Counter,
			Value: nil,
			Delta: cmd.Delta,
		}
		return s.metricsStorage.UpdateOrCreate(ctx, cmd.Name, metric)
	default:
		return metrics.ErrInvalidMetricType
	}
}

func (s *Service) GetMetricByName(ctx context.Context, cmd GetCommand) (metrics.Metrics, error) {
	metric, err := s.metricsStorage.Get(ctx, cmd.Name)
	if err != nil {
		return metrics.Metrics{}, err
	}
	if metric.Type != cmd.MetricType {
		return metrics.Metrics{}, metrics.ErrMetricNotFound
	}
	return metric, nil
}

// GetAllMetrics Возвращает все метрики из хранилища
func (s *Service) GetAllMetrics(ctx context.Context) ([]metrics.Metrics, error) {
	return s.metricsStorage.GetAll(ctx)
}
