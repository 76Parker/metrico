package metrics

import (
	"testing"

	"github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCoreService_update(t *testing.T) {
	t.Run("valid/create_gauge", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockStorage := NewMockRepository(ctrl)

		expectedMetricName := "test_metric"
		expectedValue := 10.0
		expectedChange := []Change{{
			name:  expectedMetricName,
			kind:  ChangeKindSetGauge,
			value: expectedValue,
		}}
		expectedMetric := metrics.Metrics{
			ID:    expectedMetricName,
			Type:  metrics.MetricTypeGauge,
			Value: &expectedValue,
		}
		cmd := UpdateCommand{
			Name:       expectedMetricName,
			MetricType: metrics.MetricTypeGauge,
			Value:      &expectedValue,
		}

		mockStorage.EXPECT().Apply(gomock.Any(), expectedChange).Return(nil)
		mockStorage.EXPECT().Get(gomock.Any(), expectedMetricName).Return(expectedMetric, nil)

		service := NewCoreService(mockStorage)
		err := service.update(t.Context(), cmd)
		require.NoError(t, err)

		actualMetric, err := service.getByName(t.Context(), expectedMetricName)
		require.NoError(t, err)
		require.Equal(t, expectedMetric, actualMetric)
	})

	t.Run("valid/create_counter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockStorage := NewMockRepository(ctrl)

		expectedMetricName := "test_counter"
		expectedDelta := int64(10)
		expectedChange := []Change{{
			name:  expectedMetricName,
			kind:  ChangeKindAddCounter,
			delta: expectedDelta,
		}}
		expectedMetric := metrics.Metrics{
			ID:    expectedMetricName,
			Type:  metrics.MetricTypeCounter,
			Delta: &expectedDelta,
		}
		cmd := UpdateCommand{
			Name:       expectedMetricName,
			MetricType: metrics.MetricTypeCounter,
			Delta:      &expectedDelta,
		}

		mockStorage.EXPECT().Apply(gomock.Any(), expectedChange).Return(nil)
		mockStorage.EXPECT().Get(gomock.Any(), expectedMetricName).Return(expectedMetric, nil)

		service := NewCoreService(mockStorage)
		err := service.update(t.Context(), cmd)
		require.NoError(t, err)

		actualMetric, err := service.getByName(t.Context(), expectedMetricName)
		require.NoError(t, err)
		require.Equal(t, expectedMetric, actualMetric)
	})

	t.Run("invalid/create_gauge_without_value", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockStorage := NewMockRepository(ctrl)

		expectedMetricName := "test_gauge"
		expectedMetric := metrics.Metrics{}
		cmd := UpdateCommand{
			Name:       expectedMetricName,
			MetricType: metrics.MetricTypeGauge,
		}

		mockStorage.EXPECT().Get(gomock.Any(), expectedMetricName).Return(expectedMetric, metrics.ErrMetricNotFound)

		service := NewCoreService(mockStorage)
		err := service.update(t.Context(), cmd)
		require.ErrorIs(t, err, metrics.ErrInvalidValueForGauge)

		actualMetric, err := service.getByName(t.Context(), expectedMetricName)
		require.ErrorIs(t, err, metrics.ErrMetricNotFound)
		require.Equal(t, expectedMetric, actualMetric)
	})

	t.Run("invalid/create_counter_without_delta", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockStorage := NewMockRepository(ctrl)

		expectedMetricName := "test_counter"
		expectedMetric := metrics.Metrics{}
		cmd := UpdateCommand{
			Name:       expectedMetricName,
			MetricType: metrics.MetricTypeCounter,
		}

		mockStorage.EXPECT().Get(gomock.Any(), expectedMetricName).Return(expectedMetric, metrics.ErrMetricNotFound)

		service := NewCoreService(mockStorage)
		err := service.update(t.Context(), cmd)
		require.ErrorIs(t, err, metrics.ErrInvalidValueForCounter)

		actualMetric, err := service.getByName(t.Context(), expectedMetricName)
		require.ErrorIs(t, err, metrics.ErrMetricNotFound)
		require.Equal(t, expectedMetric, actualMetric)
	})

	t.Run("invalid/unknown_metric_type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockStorage := NewMockRepository(ctrl)

		cmd := UpdateCommand{
			Name:       "test_unknown",
			MetricType: metrics.MetricType("unknown"),
		}

		service := NewCoreService(mockStorage)
		err := service.update(t.Context(), cmd)
		require.ErrorIs(t, err, metrics.ErrInvalidMetricType)
	})
}

func TestCoreService_getByName(t *testing.T) {
	t.Run("valid/get_created_gauge", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockStorage := NewMockRepository(ctrl)

		expectedMetricName := "test_gauge"
		expectedValue := 10.0
		expectedChange := []Change{{
			name:  expectedMetricName,
			kind:  ChangeKindSetGauge,
			value: expectedValue,
		}}
		expectedMetric := metrics.Metrics{
			ID:    expectedMetricName,
			Type:  metrics.MetricTypeGauge,
			Value: &expectedValue,
		}
		cmd := UpdateCommand{
			Name:       expectedMetricName,
			MetricType: metrics.MetricTypeGauge,
			Value:      &expectedValue,
		}

		mockStorage.EXPECT().Apply(gomock.Any(), expectedChange).Return(nil)
		mockStorage.EXPECT().Get(gomock.Any(), expectedMetricName).Return(expectedMetric, nil)

		service := NewCoreService(mockStorage)
		err := service.update(t.Context(), cmd)
		require.NoError(t, err)

		actualMetric, err := service.getByName(t.Context(), expectedMetricName)
		require.NoError(t, err)
		require.Equal(t, expectedMetric, actualMetric)
	})
}

func TestCoreService_getAll(t *testing.T) {
	t.Run("valid/one_created_metric", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockStorage := NewMockRepository(ctrl)

		expectedMetricName := "test_gauge"
		expectedValue := 10.0
		expectedChange := []Change{{
			name:  expectedMetricName,
			kind:  ChangeKindSetGauge,
			value: expectedValue,
		}}
		expectedMetric := metrics.Metrics{
			ID:    expectedMetricName,
			Type:  metrics.MetricTypeGauge,
			Value: &expectedValue,
		}
		cmd := UpdateCommand{
			Name:       expectedMetricName,
			MetricType: metrics.MetricTypeGauge,
			Value:      &expectedValue,
		}

		mockStorage.EXPECT().Apply(gomock.Any(), expectedChange).Return(nil)
		mockStorage.EXPECT().GetAll(gomock.Any()).Return([]metrics.Metrics{expectedMetric}, nil)

		service := NewCoreService(mockStorage)
		err := service.update(t.Context(), cmd)
		require.NoError(t, err)

		actualMetrics, err := service.getAll(t.Context())
		require.NoError(t, err)
		require.Len(t, actualMetrics, 1)
		require.Equal(t, expectedMetric, actualMetrics[0])
	})
}
