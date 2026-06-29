package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/76Parker/metrico/internal/adapters/memstorage"
	"github.com/76Parker/metrico/internal/usecase/metrics"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

/*
 * Тесты с постфиксом _Valid проверяют корректно ли код обрабатывает валидные входные данные
 * Тесты с постфиксом _Invalid проверяют корректно ли код обрабатывает невалидные входные данные, метод должен возвращать ошибку
 */

func TestUpdateMetric_Valid(t *testing.T) {
	router := createTestRouter()
	testCases := []struct {
		name         string
		path         string
		expectedCode int
	}{
		{
			name:         "ValidUpdate",
			path:         "/update/counter/Test/1",
			expectedCode: http.StatusOK,
		},
		{
			name:         "ValidUpdate_2",
			path:         "/update/counter/tt/500",
			expectedCode: http.StatusOK,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
	router := createTestRouter()
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
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			res := rec.Result()
			assert.Equal(t, tc.expectedCode, res.StatusCode)
			res.Body.Close()
		})
	}
}

func createTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	metricStorage := memstorage.NewMemStorage()
	metricService := metrics.NewService(metricStorage)
	metricHandler := NewMetricsHandler(metricService)

	router := gin.New()
	router.POST("/update/:metricType/:metricName/:metricValue", metricHandler.UpdateMetric)
	return router
}
