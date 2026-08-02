// `memstorage/metric_storage.go`
// содержит в себе реализацию хранилища метрик - InMemoryStorage
package memstorage

import (
	"context"
	"maps"
	"sync"

	metricsapp "github.com/76Parker/metrico/internal/applications/metrics"
	"github.com/76Parker/metrico/internal/domain/metrics"
)

type MemStorage struct {
	mu      *sync.Mutex
	metrics map[string]metrics.Metrics
}

var _ metricsapp.Repository = (*MemStorage)(nil)

func NewMemStorage() *MemStorage {
	reservation := 1024 // Резервируем место в памяти для 1024 метрик, для избежания лишних аллокаций
	return &MemStorage{
		mu:      &sync.Mutex{},
		metrics: make(map[string]metrics.Metrics, reservation),
	}
}

// Apply применяет пакет изменений метрик.
func (s *MemStorage) Apply(_ context.Context, changes []metricsapp.Change) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	updated := s.cloneMetrics()
	for _, change := range changes {
		if err := s.applyChange(updated, change); err != nil {
			return err
		}
	}
	s.metrics = updated

	return nil
}

// Get Возвращает метрику по metricName
func (s *MemStorage) Get(_ context.Context, metricName string) (metrics.Metrics, error) {
	if metricName == "" {
		return metrics.Metrics{}, metrics.ErrMetricNameIsEmpty
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	metric, ok := s.metrics[metricName]
	if !ok {
		return metrics.Metrics{}, metrics.ErrMetricNotFound
	}
	return metric, nil
}

func (s *MemStorage) Load(_ context.Context, metricsSlice []metrics.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics = make(map[string]metrics.Metrics, len(metricsSlice))
	for _, metric := range metricsSlice {
		s.metrics[metric.ID] = metric
	}
	return nil
}

func (s *MemStorage) applyChange(items map[string]metrics.Metrics, change metricsapp.Change) error {
	metricType := change.MetricType()
	if metricType == metrics.MetricTypeInvalid {
		return metrics.ErrInvalidMetricType
	}
	if change.Name() == "" {
		return metrics.ErrMetricNameIsEmpty
	}
	switch metricType {
	case metrics.MetricTypeGauge:
		s.applyGauge(items, change.Name(), change.Value())
	case metrics.MetricTypeCounter:
		s.applyCounter(items, change.Name(), change.Delta())
	}

	return nil
}

func (s *MemStorage) applyGauge(items map[string]metrics.Metrics, metricName string, value float64) {
	if v, ok := items[metricName]; ok {
		v.Value = &value
		items[metricName] = v
	} else {
		items[metricName] = metrics.Metrics{
			ID:    metricName,
			Type:  metrics.MetricTypeGauge,
			Value: &value,
		}
	}
}

func (s *MemStorage) applyCounter(items map[string]metrics.Metrics, metricName string, delta int64) {
	if v, ok := items[metricName]; ok {
		if v.Delta == nil {
			v.Delta = &delta
		} else {
			newValue := *v.Delta + delta
			v.Delta = &newValue
		}
		items[metricName] = v
	} else {
		items[metricName] = metrics.Metrics{
			ID:    metricName,
			Type:  metrics.MetricTypeCounter,
			Delta: &delta,
		}
	}
}

func (s *MemStorage) cloneMetrics() map[string]metrics.Metrics {
	clone := make(map[string]metrics.Metrics, len(s.metrics))
	maps.Copy(clone, s.metrics)
	return clone
}

// GetAll Возвращает все метрики из хранилища
func (s *MemStorage) GetAll(ctx context.Context) ([]metrics.Metrics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := s.createMetricSnapshot()
	return result, nil
}

func (s *MemStorage) createMetricSnapshot() []metrics.Metrics {
	result := make([]metrics.Metrics, 0, len(s.metrics))
	for _, v := range s.metrics {
		result = append(result, v)
	}
	return result
}
