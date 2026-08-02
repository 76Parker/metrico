package memstorage

import (
	"context"
	"sync"
	"testing"

	metricsapp "github.com/76Parker/metrico/internal/applications/metrics"
	"github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_Apply(t *testing.T) {
	t.Run("valid/empty_batch", func(t *testing.T) {
		storage := newTestMemStorage(t)

		require.NoError(t, storage.Apply(t.Context(), nil))

		items, err := storage.GetAll(t.Context())
		require.NoError(t, err)
		require.Empty(t, items)
	})

	t.Run("valid/create_and_update_gauge", func(t *testing.T) {
		storage := newTestMemStorage(t)

		require.NoError(t, storage.Apply(t.Context(), []metricsapp.Change{
			metricsapp.NewSetGaugeChange("temperature", 42.5),
		}))
		require.NoError(t, storage.Apply(t.Context(), []metricsapp.Change{
			metricsapp.NewSetGaugeChange("temperature", 0),
		}))

		metric, err := storage.Get(t.Context(), "temperature")
		require.NoError(t, err)
		require.Equal(t, gaugeMetric("temperature", 0), metric)
	})

	t.Run("valid/create_and_increment_counter", func(t *testing.T) {
		storage := newTestMemStorage(t)

		require.NoError(t, storage.Apply(t.Context(), []metricsapp.Change{
			metricsapp.NewAddCounterChange("requests", 0),
			metricsapp.NewAddCounterChange("requests", 5),
			metricsapp.NewAddCounterChange("requests", -2),
		}))

		metric, err := storage.Get(t.Context(), "requests")
		require.NoError(t, err)
		require.Equal(t, counterMetric("requests", 3), metric)
	})

	t.Run("valid/mixed_batch", func(t *testing.T) {
		storage := newTestMemStorage(t)

		require.NoError(t, storage.Apply(t.Context(), []metricsapp.Change{
			metricsapp.NewSetGaugeChange("temperature", 18.5),
			metricsapp.NewAddCounterChange("requests", 3),
			metricsapp.NewAddCounterChange("requests", 2),
		}))

		items, err := storage.GetAll(t.Context())
		require.NoError(t, err)
		require.ElementsMatch(t, []metrics.Metrics{
			gaugeMetric("temperature", 18.5),
			counterMetric("requests", 5),
		}, items)
	})

	t.Run("valid/concurrent_counter_updates", func(t *testing.T) {
		storage := newTestMemStorage(t)
		const updates = 100

		var wg sync.WaitGroup
		errs := make(chan error, updates)
		for range updates {
			wg.Add(1)
			go func() {
				defer wg.Done()
				errs <- storage.Apply(context.Background(), []metricsapp.Change{
					metricsapp.NewAddCounterChange("requests", 1),
				})
			}()
		}
		wg.Wait()
		close(errs)

		for err := range errs {
			require.NoError(t, err)
		}

		metric, err := storage.Get(t.Context(), "requests")
		require.NoError(t, err)
		require.Equal(t, counterMetric("requests", updates), metric)
	})

	t.Run("invalid/change_kind", func(t *testing.T) {
		storage := newTestMemStorage(t)

		err := storage.Apply(t.Context(), []metricsapp.Change{{}})

		require.ErrorIs(t, err, metrics.ErrInvalidMetricType)
	})

	t.Run("invalid/empty_metric_name", func(t *testing.T) {
		storage := newTestMemStorage(t)

		err := storage.Apply(t.Context(), []metricsapp.Change{
			metricsapp.NewSetGaugeChange("", 1),
		})

		require.ErrorIs(t, err, metrics.ErrMetricNameIsEmpty)
	})
}

func TestMemStorage_Get(t *testing.T) {
	t.Run("valid/existing_metric", func(t *testing.T) {
		storage := newTestMemStorage(t)
		require.NoError(t, storage.Apply(t.Context(), []metricsapp.Change{
			metricsapp.NewSetGaugeChange("temperature", 42.5),
		}))

		metric, err := storage.Get(t.Context(), "temperature")

		require.NoError(t, err)
		require.Equal(t, gaugeMetric("temperature", 42.5), metric)
	})

	t.Run("invalid/empty_metric_name", func(t *testing.T) {
		storage := newTestMemStorage(t)

		_, err := storage.Get(t.Context(), "")

		require.ErrorIs(t, err, metrics.ErrMetricNameIsEmpty)
	})

	t.Run("invalid/metric_not_found", func(t *testing.T) {
		storage := newTestMemStorage(t)

		_, err := storage.Get(t.Context(), "missing")

		require.ErrorIs(t, err, metrics.ErrMetricNotFound)
	})
}

func TestMemStorage_Load(t *testing.T) {
	t.Run("valid/replaces_existing_metrics", func(t *testing.T) {
		storage := newTestMemStorage(t)
		require.NoError(t, storage.Apply(t.Context(), []metricsapp.Change{
			metricsapp.NewSetGaugeChange("stale", 1),
		}))

		require.NoError(t, storage.Load(t.Context(), []metrics.Metrics{
			gaugeMetric("temperature", 18.5),
			counterMetric("requests", 3),
		}))

		items, err := storage.GetAll(t.Context())
		require.NoError(t, err)
		require.ElementsMatch(t, []metrics.Metrics{
			gaugeMetric("temperature", 18.5),
			counterMetric("requests", 3),
		}, items)
	})

	t.Run("valid/empty_snapshot_clears_storage", func(t *testing.T) {
		storage := newTestMemStorage(t)
		require.NoError(t, storage.Apply(t.Context(), []metricsapp.Change{
			metricsapp.NewSetGaugeChange("temperature", 18.5),
		}))

		require.NoError(t, storage.Load(t.Context(), nil))

		items, err := storage.GetAll(t.Context())
		require.NoError(t, err)
		require.NotNil(t, items)
		require.Empty(t, items)
	})
}

func TestMemStorage_GetAll(t *testing.T) {
	t.Run("valid/empty_storage_returns_non_nil_empty_slice", func(t *testing.T) {
		storage := newTestMemStorage(t)

		items, err := storage.GetAll(t.Context())

		require.NoError(t, err)
		require.NotNil(t, items)
		require.Empty(t, items)
	})

	t.Run("valid/returns_all_metrics", func(t *testing.T) {
		storage := newTestMemStorage(t)
		require.NoError(t, storage.Apply(t.Context(), []metricsapp.Change{
			metricsapp.NewSetGaugeChange("temperature", 18.5),
			metricsapp.NewAddCounterChange("requests", 3),
		}))

		items, err := storage.GetAll(t.Context())

		require.NoError(t, err)
		require.ElementsMatch(t, []metrics.Metrics{
			gaugeMetric("temperature", 18.5),
			counterMetric("requests", 3),
		}, items)
	})
}

func newTestMemStorage(t *testing.T) *MemStorage {
	t.Helper()

	return NewMemStorage()
}

func gaugeMetric(name string, value float64) metrics.Metrics {
	return metrics.Metrics{
		ID:    name,
		Type:  metrics.MetricTypeGauge,
		Value: &value,
	}
}

func counterMetric(name string, delta int64) metrics.Metrics {
	return metrics.Metrics{
		ID:    name,
		Type:  metrics.MetricTypeCounter,
		Delta: &delta,
	}
}
