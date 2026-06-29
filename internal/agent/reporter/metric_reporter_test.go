package reporter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/76Parker/metrico/internal/agent/provider"
	"github.com/stretchr/testify/assert"
)

/*
 * Тесты с постфиксом _Valid проверяют корректно ли код обрабатывает валидные входные данные
 * Тесты с постфиксом _Invalid проверяют корректно ли код обрабатывает невалидные входные данные, метод должен возвращать ошибку
 */

func TestSendGauge_Valid(t *testing.T) {

	var path string
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	provider := provider.NewMetricProvider(2 * time.Second)
	reporter := NewMetricReporter(testServer.URL, testServer.Client(), provider, 10*time.Second)
	err := reporter.sendGaugeMetric(1.000000, "TestGauge")
	assert.NoError(t, err)
	assert.Equal(t, "/update/gauge/TestGauge/1.000000", path)
}

func TestSendGauge_Invalid(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer testServer.Close()

	provider := provider.NewMetricProvider(2 * time.Second)
	reporter := NewMetricReporter(testServer.URL, testServer.Client(), provider, 10*time.Second)
	err := reporter.sendGaugeMetric(1.0, "") // empty metric name
	assert.Error(t, err)
}

func TestSendCounter_Valid(t *testing.T) {
	var path string
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	provider := provider.NewMetricProvider(2 * time.Second)
	reporter := NewMetricReporter(testServer.URL, testServer.Client(), provider, 10*time.Second)
	err := reporter.sendCounterMetric(1, "TestCounter")
	assert.NoError(t, err)
	assert.Equal(t, "/update/counter/TestCounter/1", path)
}

func TestSendCounter_Invalid(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer testServer.Close()

	provider := provider.NewMetricProvider(2 * time.Second)
	reporter := NewMetricReporter(testServer.URL, testServer.Client(), provider, 10*time.Second)
	err := reporter.sendCounterMetric(1, "") // empty metric name
	assert.Error(t, err)
}

func TestIntegration(t *testing.T) {
	var requestCount atomic.Int64
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	reporter := NewMetricReporter(testServer.URL, testServer.Client(), testMetricProvider{}, 5*time.Millisecond)
	err := reporter.Run(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Positive(t, requestCount.Load())
}

type testMetricProvider struct{}

func (testMetricProvider) Metrics() (runtime.MemStats, int64) {
	return runtime.MemStats{Alloc: 1}, 1
}
