package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/76Parker/metrico/internal/adapters/memstorage"
	"github.com/76Parker/metrico/internal/usecase/metrics"
	"github.com/stretchr/testify/assert"
)

/*
 * Тесты с постфиксом _Valid проверяют корректно ли код обрабатывает валидные входные данные
 * Тесты с постфиксом _Invalid проверяют корректно ли код обрабатывает невалидные входные данные, метод должен возвращать ошибку
 */

func TestUpdateMetric_Valid(t *testing.T) {
	metricStorage := memstorage.NewMemStorage()
	metricService := metrics.NewService(metricStorage)
	metricHandler := NewMetricsHandler(metricService)
	testCases := []struct {
		name         string
		path         string
		expectedCode int
	}{
		{
			name:         "ValidUpdate",
			path:         "/update/counter/Test/1",
			expectedCode: 200,
		},
		{
			name:         "ValidUpdate_2",
			path:         "/update/counter/tt/500",
			expectedCode: 200,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tc.path, nil)
			rec := httptest.NewRecorder()

			res := rec.Result()
			metricHandler.UpdateMetric(rec, req)
			assert.Equal(t, tc.expectedCode, res.StatusCode)
			res.Body.Close()
		})
	}
}

func TestUpdateMetric_Invalid(t *testing.T) {
	metricStorage := memstorage.NewMemStorage()
	metricService := metrics.NewService(metricStorage)
	metricHandler := NewMetricsHandler(metricService)
	testCases := []struct {
		name         string
		metricName   string
		metricType   string
		metricValue  string
		expectedCode int
	}{
		{
			name:         "InvalidGaugeValue",
			metricName:   "test",
			metricType:   "gauge",
			metricValue:  "12f",
			expectedCode: 400,
		},
		{
			name:         "InvalidCounterValue",
			metricName:   "test",
			metricType:   "counter",
			metricValue:  "12f",
			expectedCode: 400,
		},
		{
			name:         "InvalidMetricName",
			metricName:   "",
			metricType:   "counter",
			metricValue:  "1",
			expectedCode: 404,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/update", nil)
			req.SetPathValue("metricType", tc.metricType)
			req.SetPathValue("metricName", tc.metricName)
			req.SetPathValue("metricValue", tc.metricValue)
			rec := httptest.NewRecorder()
			metricHandler.UpdateMetric(rec, req)
			res := rec.Result()
			assert.Equal(t, tc.expectedCode, res.StatusCode)
			res.Body.Close()
		})
	}
}
