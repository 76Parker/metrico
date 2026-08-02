package postgres

import (
	"testing"

	metricsapp "github.com/76Parker/metrico/internal/applications/metrics"
	"github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/stretchr/testify/require"
)

func TestMakeBatch(t *testing.T) {
	t.Run("valid/replaces_counter_with_gauge", func(t *testing.T) {
		batch, err := makeBatch([]metricsapp.Change{
			metricsapp.NewAddCounterChange("metric", 3),
			metricsapp.NewSetGaugeChange("metric", 42.5),
		})

		require.NoError(t, err)
		require.Len(t, batch, 1)
		require.Equal(t, "metric", batch[0].Name)
		require.Equal(t, string(metrics.MetricTypeGauge), batch[0].Type)
		require.NotNil(t, batch[0].Value)
		require.Equal(t, 42.5, *batch[0].Value)
		require.Nil(t, batch[0].Delta)
	})

	t.Run("valid/replaces_gauge_with_counter", func(t *testing.T) {
		batch, err := makeBatch([]metricsapp.Change{
			metricsapp.NewSetGaugeChange("metric", 42.5),
			metricsapp.NewAddCounterChange("metric", 3),
		})

		require.NoError(t, err)
		require.Len(t, batch, 1)
		require.Equal(t, "metric", batch[0].Name)
		require.Equal(t, string(metrics.MetricTypeCounter), batch[0].Type)
		require.Nil(t, batch[0].Value)
		require.NotNil(t, batch[0].Delta)
		require.Equal(t, int64(3), *batch[0].Delta)
	})
}
