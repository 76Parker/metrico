package memstorage

import (
	"sync"
	"testing"

	"github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/stretchr/testify/assert"
)

/*
 * Тесты с постфиксом _Valid проверяют корректно ли код обрабатывает валидные входные данные
 * Тесты с постфиксом _Invalid проверяют корректно ли код обрабатывает невалидные входные данные, метод должен возвращать ошибку
 */

// TestUpdateGauge_Valid: валидные входные данные для обновления Gauge
func TestUpdateGauge_Valid(t *testing.T) {

	testCases := []struct {
		name       string
		metricName string
		metric     metrics.Metrics
	}{
		// NilDelta: значение `Delta` для Gauge может быть nil
		{
			name:       "NilDelta",
			metricName: "test_gauge_1",
			metric: metrics.Metrics{
				ID:    "test_gauge",
				Type:  metrics.Gauge,
				Delta: nil,
				Value: new(42.0),
				Hash:  "",
			},
		},
		// NonNilDelta: значение `Delta` для Gauge может быть не nil
		{
			name:       "NonNilDelta",
			metricName: "test_gauge_2",
			metric: metrics.Metrics{
				ID:    "test_gauge",
				Type:  metrics.Gauge,
				Delta: new(int64(123)),
				Value: new(0.00),
				Hash:  "123",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := newTestMemStorage(t)
			err := storage.UpdateOrCreateMetricByName(t.Context(), tc.metricName, tc.metric)
			assert.NoError(t, err)
		})
	}

}

// TestUpdateGauge_Invalid: невалидные входные данные для обновления метрики типа Gauge
func TestUpdateGauge_Invalid(t *testing.T) {
	testCases := []struct {
		name        string
		requiredErr error
		metricName  string
		metric      metrics.Metrics
	}{
		// NilValue: значение `Value` для Gauge не может быть nil
		{
			name:        "NilValue",
			metricName:  "test_name",
			requiredErr: metrics.ErrGaugeValueIsNil,
			metric: metrics.Metrics{
				ID:    "test_id",
				Type:  metrics.Gauge,
				Delta: new(int64(1)),
				Value: nil,
				Hash:  "",
			},
		},
		// InvalidMetricType: невалидный тип метрики
		{
			name:        "InvalidMetricType",
			requiredErr: metrics.ErrInvalidMetricType,
			metricName:  "test_gauge",
			metric: metrics.Metrics{
				Type:  metrics.MetricType("invalid_type"),
				Delta: nil,
				Value: new(0.0),
				Hash:  "",
			},
		},
		// EmptyMetricName: имя метрики не может быть пустым
		{
			name:        "EmptyMetricName",
			requiredErr: metrics.ErrMetricNameIsEmpty,
			metricName:  "",
			metric: metrics.Metrics{
				ID:    "123",
				Type:  metrics.Gauge,
				Delta: nil,
				Value: new(1.1),
				Hash:  "123",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := newTestMemStorage(t)
			err := storage.UpdateOrCreateMetricByName(t.Context(), tc.metricName, tc.metric)
			assert.ErrorIs(t, err, tc.requiredErr)
		})
	}
}

// TestUpdateCounter_Valid: валидные входные данные для обновления Counter
func TestUpdateCounter_Valid(t *testing.T) {
	testCases := []struct {
		name       string
		metricName string
		metric     metrics.Metrics
	}{
		// NilValue: значение `Value` для Counter может быть nil
		{
			name:       "NilValue",
			metricName: "test_counter",
			metric: metrics.Metrics{
				ID:    "test_id",
				Type:  metrics.Counter,
				Delta: new(int64(1)),
				Value: nil,
				Hash:  "",
			},
		},
		// NotNilValue: значение `Value` для Counter может быть не nil
		{
			name:       "NotNilValue",
			metricName: "test_counter",
			metric: metrics.Metrics{
				ID:    "test_id",
				Type:  metrics.Counter,
				Delta: new(int64(5)),
				Value: new(1.0),
				Hash:  "123",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := newTestMemStorage(t)
			err := storage.UpdateOrCreateMetricByName(t.Context(), tc.metricName, tc.metric)
			assert.NoError(t, err)
		})
	}
}

// TestUpdateCounter_Invalid: невалидные входные данные для обновления метрики типа Counter
func TestUpdateCounter_Invalid(t *testing.T) {
	testCases := []struct {
		name        string
		requiredErr error
		metricName  string
		metric      metrics.Metrics
	}{
		// NilDelta: delta для Counter не может быть nil
		{
			name:        "NilDelta",
			requiredErr: metrics.ErrCounterValueIsNil,
			metricName:  "test_gauge",
			metric: metrics.Metrics{
				ID:    "test_id",
				Type:  metrics.Counter,
				Delta: nil,
				Value: new(1.0),
				Hash:  "123",
			},
		},
		// EmptyMetricName: имя метрики не может быть пустым
		{
			name:        "EmptyMetricName",
			requiredErr: metrics.ErrMetricNameIsEmpty,
			metricName:  "",
			metric: metrics.Metrics{
				ID:    "test_id",
				Type:  metrics.Counter,
				Delta: new(int64(1)),
				Value: new(1.0),
				Hash:  "123",
			},
		},
		// InvalidMetricType: тип метрики не может быть невалидным
		{
			name:        "InvalidMetricType",
			requiredErr: metrics.ErrInvalidMetricType,
			metricName:  "test_name",
			metric: metrics.Metrics{
				ID:    "test_id",
				Type:  metrics.MetricType("invalid metric type"),
				Delta: new(int64(1)),
				Value: new(1.0),
				Hash:  "123",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := newTestMemStorage(t)
			err := storage.UpdateOrCreateMetricByName(t.Context(), tc.metricName, tc.metric)
			assert.ErrorIs(t, err, tc.requiredErr)
		})
	}
}

func newTestMemStorage(t *testing.T) *MemStorage {
	t.Helper()
	return &MemStorage{
		mu:      &sync.Mutex{},
		metrics: make(map[string]metrics.Metrics),
	}
}
