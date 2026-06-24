// `memstorage/metric_storage.go`
// содержит в себе реализацию хранилища метрик - InMemoryStorage
package memstorage

import (
	"context"
	"sync"

	"github.com/76Parker/metrico/internal/domain/metrics"
)

type MemStorage struct {
	mu      *sync.Mutex
	metrics map[string]metrics.Metrics
}

func NewMemStorage() *MemStorage {
	reservation := 1024 // Резервируем место в памяти для 1024 метрик, для избежания лишних аллокаций
	return &MemStorage{
		mu:      &sync.Mutex{},
		metrics: make(map[string]metrics.Metrics, reservation),
	}
}

func (s *MemStorage) UpdateOrCreateMetricByName(_ context.Context, metricName string, metric metrics.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch metric.Type {
	case metrics.Gauge:
		return s.updateOrCreateGauge(metricName, metric)
	case metrics.Counter:
		return s.updateOrCreateCounter(metricName, metric)
	default:
		return metrics.ErrInvalidMetricType
	}
}

// updateOrCreateGauge Логика обновления/создания для Gauge-метрик: замещение значения на newValue
func (s *MemStorage) updateOrCreateGauge(metricName string, metric metrics.Metrics) error {
	if metric.Value == nil {
		return metrics.ErrGaugeValueIsNil
	}
	if v, ok := s.metrics[metricName]; ok {
		v.Value = metric.Value
		s.metrics[metricName] = v
	} else {
		s.metrics[metricName] = metric
	}
	return nil
}

// updateCounter Логика обновления/создания для Counter-метрик: увеличение текущего значения на delta
func (s *MemStorage) updateOrCreateCounter(metricName string, metric metrics.Metrics) error {
	if metric.Delta == nil {
		return metrics.ErrCounterValueIsNil
	}
	if v, ok := s.metrics[metricName]; ok {
		switch v.Delta {
		case nil: // На случай если уже существующий счетчик каким-то образом == nil
			v.Delta = metric.Delta
		default:
			newValue := *v.Delta + *metric.Delta
			v.Delta = &newValue
		}
		s.metrics[metricName] = v
	} else {
		s.metrics[metricName] = metric
	}
	return nil
}
