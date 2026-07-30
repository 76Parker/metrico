package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	metricsapp "github.com/76Parker/metrico/internal/applications/metrics"
	domainmetrics "github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/gin-gonic/gin"
	goccyjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

/*
 * Тесты с постфиксом _Valid проверяют корректно ли код обрабатывает валидные входные данные
 * Тесты с постфиксом _Invalid проверяют корректно ли код обрабатывает невалидные входные данные, метод должен возвращать ошибку
 */

func TestUpdateMetric_Valid(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		expectedCode int
		command      metricsapp.UpdateCommand
	}{
		{
			name:         "ValidUpdate",
			path:         "/update/counter/Test/1",
			expectedCode: http.StatusOK,
			command: metricsapp.UpdateCommand{
				Name:       "Test",
				MetricType: domainmetrics.MetricTypeCounter,
				Delta:      int64Pointer(1),
			},
		},
		{
			name:         "ValidUpdate_2",
			path:         "/update/counter/tt/500",
			expectedCode: http.StatusOK,
			command: metricsapp.UpdateCommand{
				Name:       "tt",
				MetricType: domainmetrics.MetricTypeCounter,
				Delta:      int64Pointer(500),
			},
		},
		{
			name:         "ValidGaugeUpdate",
			path:         "/update/gauge/temperature/23.5",
			expectedCode: http.StatusOK,
			command: metricsapp.UpdateCommand{
				Name:       "temperature",
				MetricType: domainmetrics.MetricTypeGauge,
				Value:      float64Pointer(23.5),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metricService := newMockMetricsApplication(t)
			metricService.EXPECT().Update(gomock.Any(), tc.command).Return(nil)
			router := createTestRouter(metricService)

			req := httptest.NewRequest("POST", tc.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)
			res := rec.Result()
			assert.Equal(t, tc.expectedCode, res.StatusCode)
			res.Body.Close()
		})
	}
}

func TestUpdateMetric_Invalid(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		expectedCode int
	}{
		{
			name:         "InvalidGaugeValue",
			path:         "/update/gauge/test/12f",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "InvalidCounterValue",
			path:         "/update/counter/test/12f",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "InvalidMetricName",
			path:         "/update/counter//1",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "InvalidMetricType",
			path:         "/update/unknown/test/1",
			expectedCode: http.StatusBadRequest,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metricService := newMockMetricsApplication(t)
			router := createTestRouter(metricService)

			req := httptest.NewRequest("POST", tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			res := rec.Result()
			assert.Equal(t, tc.expectedCode, res.StatusCode)
			res.Body.Close()
		})
	}
}

func TestGetMetricByNameJSON(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantStatus   int
		wantResponse *domainmetrics.Metrics
		metricName   string
		metric       domainmetrics.Metrics
		serviceErr   error
	}{
		{
			name:       "returns gauge metric",
			body:       `{"id":"LastGC","type":"gauge"}`,
			wantStatus: http.StatusOK,
			metricName: "LastGC",
			metric: domainmetrics.Metrics{
				ID:    "LastGC",
				Type:  domainmetrics.MetricTypeGauge,
				Value: float64Pointer(1744184459),
			},
			wantResponse: &domainmetrics.Metrics{
				ID:    "LastGC",
				Type:  domainmetrics.MetricTypeGauge,
				Value: float64Pointer(1744184459),
			},
		},
		{
			name:       "rejects invalid JSON",
			body:       `{"id":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects empty metric ID",
			body:       `{"id":"  ","type":"gauge"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects unknown metric type",
			body:       `{"id":"LastGC","type":"histogram"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns not found for missing metric",
			body:       `{"id":"missing","type":"gauge"}`,
			wantStatus: http.StatusNotFound,
			metricName: "missing",
			serviceErr: domainmetrics.ErrMetricNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metricService := newMockMetricsApplication(t)
			if tt.metricName != "" {
				metricService.EXPECT().GetByName(gomock.Any(), tt.metricName).Return(tt.metric, tt.serviceErr)
			}
			router := createTestRouter(metricService)

			request := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(tt.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			require.Equal(t, tt.wantStatus, response.Code)
			if tt.wantResponse == nil {
				return
			}

			require.Contains(t, response.Header().Get("Content-Type"), "application/json")
			var got domainmetrics.Metrics
			require.NoError(t, goccyjson.Unmarshal(response.Body.Bytes(), &got))
			require.Equal(t, *tt.wantResponse, got)
		})
	}
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
		MetricType: domainmetrics.MetricTypeGauge,
		Value:      float64Pointer(1.5),
	}).Return(nil)
	metricService.EXPECT().GetByName(gomock.Any(), "RandomValue").Return(updatedMetric, nil)
	router := createTestRouter(metricService)
	request := httptest.NewRequest(
		http.MethodPost,
		"/update",
		strings.NewReader(`{"id":"RandomValue","type":"gauge","value":1.5}`),
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

func float64Pointer(value float64) *float64 {
	return &value
}

func int64Pointer(value int64) *int64 {
	return &value
}

func newMockMetricsApplication(t *testing.T) *MockMetricsApplication {
	t.Helper()

	return NewMockMetricsApplication(gomock.NewController(t))
}

func createTestRouter(metricService metricsApplication) *gin.Engine {
	gin.SetMode(gin.TestMode)

	metricHandler := NewMetricsHandler(metricService)

	router := gin.New()
	router.POST("/update/:metricType/:metricName/:metricValue", metricHandler.Update)
	router.POST("/update", metricHandler.UpdateFromJSON)
	router.POST("/value", metricHandler.GetFromJSON)
	return router
}
