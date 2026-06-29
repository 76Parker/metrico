// `metrics/service.go`
// сервисный слой для работы с метриками.
// Описывает пользовательские сценарии взаимодействия с метриками.
package metrics

import (
	"context"
	"strconv"

	"github.com/76Parker/metrico/internal/domain/metrics"
)

type Service struct {
	storage metricStorage
}

func NewService(storage metricStorage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) UpdateOrCreateMetric(ctx context.Context, cmd UpdateMetricCommand) error {
	switch cmd.MetricType {
	case metrics.Gauge:
		v, err := strconv.ParseFloat(cmd.Value, 64)
		if err != nil {
			return metrics.ErrInvalidValueForGauge
		}
		metric := metrics.Metrics{
			ID:    cmd.Name,
			Type:  metrics.Gauge,
			Delta: nil,
			Value: &v,
		}
		return s.storage.UpdateOrCreateMetricByName(ctx, cmd.Name, metric)
	case metrics.Counter:
		v, err := strconv.ParseInt(cmd.Value, 10, 64)
		if err != nil {
			return metrics.ErrInvalidValueForCounter
		}
		metric := metrics.Metrics{
			ID:    cmd.Name,
			Type:  metrics.Counter,
			Value: nil,
			Delta: &v,
		}
		return s.storage.UpdateOrCreateMetricByName(ctx, cmd.Name, metric)
	default:
		return metrics.ErrInvalidMetricType
	}
}

func (s *Service) GetMetricByName(ctx context.Context, cmd GetMetricByNameCommand) (metrics.Metrics, error) {
	metric, err := s.storage.GetMetricByName(ctx, cmd.Name)
	if err != nil {
		return metrics.Metrics{}, err
	}
	if metric.Type != cmd.MetricType {
		return metrics.Metrics{}, metrics.ErrMetricNotFound
	}
	return metric, nil
}

// GetAllMetrics Возвращает все метрики из хранилища
func (s *Service) GetAllMetrics(ctx context.Context) (map[string]metrics.Metrics, error) {
	return s.storage.GetAllMetrics(ctx)
}
