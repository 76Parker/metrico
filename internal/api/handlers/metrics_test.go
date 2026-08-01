package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/76Parker/metrico/internal/api/handlers"
	"github.com/76Parker/metrico/internal/api/middleware"
	metricsapp "github.com/76Parker/metrico/internal/applications/metrics"
	domainmetrics "github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/gin-gonic/gin"
	goccyjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

/*
 * Тесты с постфиксом _Valid проверяют корректно ли код обрабатывает валидные входные данные
 * Тесты с постфиксом _Invalid проверяют корректно ли код обрабатывает невалидные входные данные, метод должен возвращать ошибку
 */

func TestUpdateMetric(t *testing.T) {
	t.Run("valid/updates_counter_metric", func(t *testing.T) {
		metricService := newMockMetricsApplication(t)
		metricService.EXPECT().Update(gomock.Any(), metricsapp.UpdateCommand{
			Name:       "Test",
			MetricType: string(domainmetrics.MetricTypeCounter),
			Value:      "1",
		}).Return(nil)

		response := serveRequest(t, createTestRouter(metricService), http.MethodPost, "/update/counter/Test/1", "")
		require.Equal(t, http.StatusOK, response.Code)
	})

	t.Run("valid/updates_counter_metric_with_another_value", func(t *testing.T) {
		metricService := newMockMetricsApplication(t)
		metricService.EXPECT().Update(gomock.Any(), metricsapp.UpdateCommand{
			Name:       "tt",
			MetricType: string(domainmetrics.MetricTypeCounter),
			Value:      "500",
		}).Return(nil)

		response := serveRequest(t, createTestRouter(metricService), http.MethodPost, "/update/counter/tt/500", "")
		require.Equal(t, http.StatusOK, response.Code)
	})

	t.Run("valid/updates_gauge_metric", func(t *testing.T) {
		metricService := newMockMetricsApplication(t)
		metricService.EXPECT().Update(gomock.Any(), metricsapp.UpdateCommand{
			Name:       "temperature",
			MetricType: string(domainmetrics.MetricTypeGauge),
			Value:      "23.5",
		}).Return(nil)

		response := serveRequest(t, createTestRouter(metricService), http.MethodPost, "/update/gauge/temperature/23.5", "")
		require.Equal(t, http.StatusOK, response.Code)
	})

	t.Run("invalid/gauge_value", func(t *testing.T) {
		metricService := newMockMetricsApplication(t)
		metricService.EXPECT().Update(gomock.Any(), metricsapp.UpdateCommand{
			Name:       "test",
			MetricType: string(domainmetrics.MetricTypeGauge),
			Value:      "12f",
		}).Return(domainmetrics.ErrInvalidValueForGauge)

		response := serveRequest(t, createTestRouter(metricService), http.MethodPost, "/update/gauge/test/12f", "")
		require.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("invalid/counter_value", func(t *testing.T) {
		metricService := newMockMetricsApplication(t)
		metricService.EXPECT().Update(gomock.Any(), metricsapp.UpdateCommand{
			Name:       "test",
			MetricType: string(domainmetrics.MetricTypeCounter),
			Value:      "12f",
		}).Return(domainmetrics.ErrInvalidValueForCounter)

		response := serveRequest(t, createTestRouter(metricService), http.MethodPost, "/update/counter/test/12f", "")
		require.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("invalid/missing_metric_name", func(t *testing.T) {
		metricService := newMockMetricsApplication(t)
		metricService.EXPECT().Update(gomock.Any(), metricsapp.UpdateCommand{
			MetricType: string(domainmetrics.MetricTypeCounter),
			Value:      "1",
		}).Return(domainmetrics.ErrMetricNameIsEmpty)

		response := serveRequest(t, createTestRouter(metricService), http.MethodPost, "/update/counter//1", "")
		require.Equal(t, http.StatusNotFound, response.Code)
	})

	t.Run("invalid/metric_type", func(t *testing.T) {
		metricService := newMockMetricsApplication(t)
		metricService.EXPECT().Update(gomock.Any(), metricsapp.UpdateCommand{
			Name:       "test",
			MetricType: "unknown",
			Value:      "1",
		}).Return(domainmetrics.ErrInvalidMetricType)

		response := serveRequest(t, createTestRouter(metricService), http.MethodPost, "/update/unknown/test/1", "")
		require.Equal(t, http.StatusBadRequest, response.Code)
	})
}

func TestGetMetricByNameJSON(t *testing.T) {
	t.Run("valid/returns_gauge_metric", func(t *testing.T) {
		metricService := newMockMetricsApplication(t)
		expected := domainmetrics.Metrics{
			ID:    "LastGC",
			Type:  domainmetrics.MetricTypeGauge,
			Value: float64Pointer(1744184459),
		}
		metricService.EXPECT().GetByName(gomock.Any(), "LastGC").Return(expected, nil)

		response := serveRequest(t, createTestRouter(metricService), http.MethodPost, "/value", `{"id":"LastGC","type":"gauge"}`)
		require.Equal(t, http.StatusOK, response.Code)
		require.Contains(t, response.Header().Get("Content-Type"), "application/json")

		var got domainmetrics.Metrics
		require.NoError(t, goccyjson.Unmarshal(response.Body.Bytes(), &got))
		require.Equal(t, expected, got)
	})

	t.Run("invalid/json", func(t *testing.T) {
		response := serveRequest(t, createTestRouter(newMockMetricsApplication(t)), http.MethodPost, "/value", `{"id":`)
		require.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("invalid/empty_metric_id", func(t *testing.T) {
		metricService := newMockMetricsApplication(t)
		metricService.EXPECT().GetByName(gomock.Any(), "").Return(domainmetrics.Metrics{}, domainmetrics.ErrMetricNameIsEmpty)

		response := serveRequest(t, createTestRouter(metricService), http.MethodPost, "/value", `{"id":"  ","type":"gauge"}`)
		require.Equal(t, http.StatusNotFound, response.Code)
	})

	t.Run("invalid/unknown_metric_type", func(t *testing.T) {
		metricService := newMockMetricsApplication(t)
		metricService.EXPECT().GetByName(gomock.Any(), "LastGC").Return(domainmetrics.Metrics{}, domainmetrics.ErrInvalidMetricType)

		response := serveRequest(t, createTestRouter(metricService), http.MethodPost, "/value", `{"id":"LastGC","type":"histogram"}`)
		require.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("invalid/metric_not_found", func(t *testing.T) {
		metricService := newMockMetricsApplication(t)
		metricService.EXPECT().GetByName(gomock.Any(), "missing").Return(domainmetrics.Metrics{}, domainmetrics.ErrMetricNotFound)

		response := serveRequest(t, createTestRouter(metricService), http.MethodPost, "/value", `{"id":"missing","type":"gauge"}`)
		require.Equal(t, http.StatusNotFound, response.Code)
	})
}

func TestUpdateMetricJSONSetsJSONContentType(t *testing.T) {
	metricService := newMockMetricsApplication(t)
	updatedMetric := domainmetrics.Metrics{
		ID:    "RandomValue",
		Type:  domainmetrics.MetricTypeGauge,
		Value: float64Pointer(1.5),
	}
	metricService.EXPECT().Update(gomock.Any(), metricsapp.UpdateCommand{
		Name:       "RandomValue",
		MetricType: string(domainmetrics.MetricTypeGauge),
		Value:      "1.5",
	}).Return(nil)
	metricService.EXPECT().GetByName(gomock.Any(), "RandomValue").Return(updatedMetric, nil)
	router := createTestRouter(metricService)
	request := httptest.NewRequest(
		http.MethodPost,
		"/update",
		strings.NewReader(`{"name":"RandomValue","type":"gauge","value":1.5}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Header().Get("Content-Type"), "application/json")
	var got domainmetrics.Metrics
	require.NoError(t, goccyjson.Unmarshal(response.Body.Bytes(), &got))
	require.Equal(t, domainmetrics.Metrics{
		ID:    "RandomValue",
		Type:  domainmetrics.MetricTypeGauge,
		Value: float64Pointer(1.5),
	}, got)
}

func TestBatchUpdateMetricJSON(t *testing.T) {
	metricService := newMockMetricsApplication(t)
	batch := metricsapp.BatchUpdateCommand{
		{
			Name:       "Alloc",
			MetricType: string(domainmetrics.MetricTypeGauge),
			Value:      "42.5",
		},
		{
			Name:       "PollCount",
			MetricType: string(domainmetrics.MetricTypeCounter),
			Value:      "7",
		},
	}
	metricService.EXPECT().BatchUpdate(gomock.Any(), batch).Return(nil)
	router := createTestRouter(metricService)
	request := httptest.NewRequest(
		http.MethodPost,
		"/updates",
		strings.NewReader(`[{"name":"Alloc","type":"gauge","value":42.5},{"name":"PollCount","type":"counter","delta":7}]`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Header().Get("Content-Type"), "application/json")
}

func serveRequest(
	t *testing.T,
	router *gin.Engine,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	return response
}

func float64Pointer(value float64) *float64 {
	return &value
}

func int64Pointer(value int64) *int64 {
	return &value
}

func newMockMetricsApplication(t *testing.T) *handlers.MockMetricsApplication {
	t.Helper()

	return handlers.NewMockMetricsApplication(gomock.NewController(t))
}

func createTestRouter(metricService *handlers.MockMetricsApplication) *gin.Engine {
	gin.SetMode(gin.TestMode)

	metricHandler := handlers.NewMetricsHandler(metricService)

	router := gin.New()
	errHandlerMW := middleware.ErrorHandler()
	router.POST("/update/:metricType/:metricName/:metricValue", errHandlerMW, metricHandler.Update)
	router.POST("/update", errHandlerMW, metricHandler.UpdateFromJSON)
	router.POST("/updates", errHandlerMW, metricHandler.BatchUpdateFromJSON)
	router.POST("/value", errHandlerMW, metricHandler.GetFromJSON)
	return router
}
