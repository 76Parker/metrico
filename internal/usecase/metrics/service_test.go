package metrics

import (
	"testing"

	"github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUpdateOrCreateMetric(t *testing.T) {

	t.Run("valid/create_gauge", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockStorage := NewMockmetricStorage(ctrl)
		expectedMetricName := "test_metric"
		expectedMetric := metrics.Metrics{
			ID:    expectedMetricName,
			Type:  metrics.Gauge,
			Value: new(10.0),
		}
		// На валидном случае не должна возвращаться ошибка
		mockStorage.EXPECT().UpdateOrCreate(gomock.Any(), expectedMetricName, expectedMetric).Return(nil)
		// Добавленная метрика должна быть доступна для получения после создания/обновления
		mockStorage.EXPECT().Get(gomock.Any(), expectedMetricName).Return(expectedMetric, nil)
		service := NewService(mockStorage)
		cmd := UpdateCommand{
			Name:       "test_metric",
			MetricType: metrics.Gauge,
			Value:      new(10.0),
		}
		err := service.UpdateOrCreateMetric(t.Context(), cmd)
		require.NoError(t, err)

		getCmd := GetCommand{
			Name:       "test_metric",
			MetricType: metrics.Gauge,
		}
		actualMetric, err := service.GetMetricByName(t.Context(), getCmd)
		require.NoError(t, err)
		require.Equal(t, expectedMetric, actualMetric)
	})
	t.Run("valid/create_counter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockStorage := NewMockmetricStorage(ctrl)
		expectedMetricName := "test_counter"
		expectedMetric := metrics.Metrics{
			ID:    expectedMetricName,
			Type:  metrics.Counter,
			Delta: new(int64(10)),
		}
		// На валидном случае не должна возвращаться ошибка
		mockStorage.EXPECT().UpdateOrCreate(gomock.Any(), expectedMetricName, expectedMetric).Return(nil)
		// Добавленная метрика должна быть доступна для получения после создания/обновления
		mockStorage.EXPECT().Get(gomock.Any(), expectedMetricName).Return(expectedMetric, nil)
		service := NewService(mockStorage)
		cmd := UpdateCommand{
			Name:       "test_counter",
			MetricType: metrics.Counter,
			Delta:      new(int64(10)),
		}
		err := service.UpdateOrCreateMetric(t.Context(), cmd)
		require.NoError(t, err)

		getCmd := GetCommand{
			Name:       "test_counter",
			MetricType: metrics.Counter,
		}
		actualMetric, err := service.GetMetricByName(t.Context(), getCmd)
		require.NoError(t, err)
		require.Equal(t, expectedMetric, actualMetric)
	})

	t.Run("invalid/create_gauge_without_value", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockStorage := NewMockmetricStorage(ctrl)
		// При попытке получить метрику по имени должна возвращаться ошибка, так как она не была создана
		mockStorage.EXPECT().Get(gomock.Any(), "test_gauge").Return(metrics.Metrics{}, metrics.ErrMetricNotFound)
		service := NewService(mockStorage)
		cmd := UpdateCommand{
			Name:       "test_gauge",
			MetricType: metrics.Gauge,
			Value:      nil,
		}
		err := service.UpdateOrCreateMetric(t.Context(), cmd)
		require.Error(t, err)
		require.ErrorIs(t, err, metrics.ErrInvalidValueForGauge)
		getCmd := GetCommand{
			Name:       "test_gauge",
			MetricType: metrics.Gauge,
		}
		actualMetric, err := service.GetMetricByName(t.Context(), getCmd)
		require.Error(t, err)
		require.ErrorIs(t, err, metrics.ErrMetricNotFound)
		require.Equal(t, metrics.Metrics{}, actualMetric)
	})
	t.Run("invalid/create_counter_without_delta", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockStorage := NewMockmetricStorage(ctrl)
		// При попытке получить метрику по имени должна возвращаться ошибка, так как она не была создана
		mockStorage.EXPECT().Get(gomock.Any(), "test_counter").Return(metrics.Metrics{}, metrics.ErrMetricNotFound)
		service := NewService(mockStorage)
		cmd := UpdateCommand{
			Name:       "test_counter",
			MetricType: metrics.Counter,
			Delta:      nil,
		}
		err := service.UpdateOrCreateMetric(t.Context(), cmd)
		require.Error(t, err)
		require.ErrorIs(t, err, metrics.ErrInvalidValueForCounter)
		getCmd := GetCommand{
			Name:       "test_counter",
			MetricType: metrics.Counter,
		}
		actualMetric, err := service.GetMetricByName(t.Context(), getCmd)
		require.Error(t, err)
		require.ErrorIs(t, err, metrics.ErrMetricNotFound)
		require.Equal(t, metrics.Metrics{}, actualMetric)
	})
	t.Run("invalid/unknown_metric_type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockStorage := NewMockmetricStorage(ctrl)
		service := NewService(mockStorage)
		cmd := UpdateCommand{
			Name:       "test_unknown",
			MetricType: metrics.MetricType("unknown"),
			Delta:      nil,
			Value:      nil,
		}
		err := service.UpdateOrCreateMetric(t.Context(), cmd)
		require.Error(t, err)
		require.ErrorIs(t, err, metrics.ErrInvalidMetricType)
	})
}

func TestGetMetricByName(t *testing.T) {

	t.Run("valid/get_created_gauge", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockStorage := NewMockmetricStorage(ctrl)
		service := NewService(mockStorage)
		expectedMetric := metrics.Metrics{
			ID:    "test_gauge",
			Type:  metrics.Gauge,
			Value: new(float64(10.0)),
		}
		cmd := UpdateCommand{
			Name:       "test_gauge",
			MetricType: metrics.Gauge,
			Value:      new(float64(10.0)),
		}
		mockStorage.EXPECT().UpdateOrCreate(gomock.Any(), "test_gauge", expectedMetric).Return(nil)
		mockStorage.EXPECT().Get(gomock.Any(), "test_gauge").Return(expectedMetric, nil)
		err := service.UpdateOrCreateMetric(t.Context(), cmd)
		require.NoError(t, err)
		getCmd := GetCommand{
			Name:       "test_gauge",
			MetricType: metrics.Gauge,
		}
		actualMetric, err := service.GetMetricByName(t.Context(), getCmd)

		require.NoError(t, err)
		require.Equal(t, expectedMetric, actualMetric)
	})
	t.Run("invalid/metric_type_mismatch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockStorage := NewMockmetricStorage(ctrl)
		service := NewService(mockStorage)
		createdMetric := metrics.Metrics{
			ID:    "test_gauge",
			Type:  metrics.Gauge,
			Value: new(float64(10.0)),
		}
		returnedMismatchMetric := metrics.Metrics{
			ID:    "test_gauge",
			Type:  metrics.Counter,
			Value: new(float64(10.0)),
		}
		// Создаем "test_gauge" метрику типа Gauge
		mockStorage.EXPECT().UpdateOrCreate(gomock.Any(), "test_gauge", createdMetric).Return(nil)

		/*
		 * Получаем "test_gauge" метрику но с другим типом,
		 * при `Get` должна вернуться ошибка так как несоответствие запрашиваемого типа и фактического типа
		 */
		mockStorage.EXPECT().Get(gomock.Any(), "test_gauge").Return(returnedMismatchMetric, nil)
		getCmd := GetCommand{
			Name:       "test_gauge",
			MetricType: metrics.Gauge,
		}
		createCmd := UpdateCommand{
			Name:       "test_gauge",
			MetricType: metrics.Gauge,
			Value:      new(float64(10.0)),
		}
		err := service.UpdateOrCreateMetric(t.Context(), createCmd)
		require.NoError(t, err)
		_, err = service.GetMetricByName(t.Context(), getCmd)
		require.Error(t, err)
		require.ErrorIs(t, err, metrics.ErrMetricNotFound)
	})
}

func TestGetAll(t *testing.T) {
	t.Run("valid/one_created_metric", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockStorage := NewMockmetricStorage(ctrl)
		service := NewService(mockStorage)
		expectedMetric := metrics.Metrics{
			ID:    "test_gauge",
			Type:  metrics.Gauge,
			Value: new(float64(10.0)),
		}
		mockStorage.EXPECT().UpdateOrCreate(gomock.Any(), "test_gauge", expectedMetric).Return(nil)
		mockStorage.EXPECT().GetAll(gomock.Any()).Return([]metrics.Metrics{expectedMetric}, nil)
		err := service.UpdateOrCreateMetric(t.Context(), UpdateCommand{
			Name:       expectedMetric.ID,
			MetricType: expectedMetric.Type,
			Value:      expectedMetric.Value,
		})
		require.NoError(t, err)
		metrics, err := service.GetAllMetrics(t.Context())
		require.NoError(t, err)
		require.Len(t, metrics, 1)
		require.Equal(t, expectedMetric, metrics[0])
	})
}
