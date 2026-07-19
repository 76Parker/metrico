package handlers

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/76Parker/metrico/internal/adapters/filestorage"
	"github.com/76Parker/metrico/internal/adapters/memstorage"
	domainmetrics "github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/76Parker/metrico/internal/usecase/metrics"
	"github.com/76Parker/metrico/internal/usecase/snapshot"
	"github.com/gin-gonic/gin"
	goccyjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		{
			name:         "ValidGaugeUpdate",
			path:         "/update/gauge/temperature/23.5",
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
		{
			name:         "InvalidMetricType",
			path:         "/update/unknown/test/1",
			expectedCode: http.StatusBadRequest,
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

func TestGetMetricByNameJSON(t *testing.T) {
	router := createTestRouter()

	seedRequest := httptest.NewRequest(http.MethodPost, "/update/gauge/LastGC/1744184459", nil)
	seedResponse := httptest.NewRecorder()
	router.ServeHTTP(seedResponse, seedRequest)
	require.Equal(t, http.StatusOK, seedResponse.Code)

	tests := []struct {
		name         string
		body         string
		wantStatus   int
		wantResponse *domainmetrics.Metrics
	}{
		{
			name:       "returns gauge metric",
			body:       `{"id":"LastGC","type":"gauge"}`,
			wantStatus: http.StatusOK,
			wantResponse: &domainmetrics.Metrics{
				ID:    "LastGC",
				Type:  domainmetrics.Gauge,
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	router := createTestRouter()
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
		Type:  domainmetrics.Gauge,
		Value: float64Pointer(1.5),
	}, got)
}

func float64Pointer(value float64) *float64 {
	return &value
}

func createTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	storagePath := "/Users/parkersec/go-projects/go-musthave-metrics-tpl/metrics.json"

	metricStorage := memstorage.NewMemStorage()
	metricService := metrics.NewService(metricStorage)
	snapshotStorage, err := filestorage.NewStorage(storagePath)
	if err != nil {
		log.Fatal("failed to create snapshot storage: %w", err)
	}
	snapshotService := snapshot.NewService(metricStorage, snapshotStorage, 0*time.Second)
	ctx := context.Background()
	snapshotService.Run(ctx)
	metricHandler := NewMetricsHandler(metricService, snapshotService)

	router := gin.New()
	router.POST("/update/:metricType/:metricName/:metricValue", metricHandler.Update)
	router.POST("/update", metricHandler.UpdateFromJSON)
	router.POST("/value", metricHandler.GetFromJSON)
	return router
}
