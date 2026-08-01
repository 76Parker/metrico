package reporter

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/76Parker/metrico/internal/domain/metrics"
	goccyjson "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
 * Тесты с постфиксом _Valid проверяют корректно ли код обрабатывает валидные входные данные
 * Тесты с постфиксом _Invalid проверяют корректно ли код обрабатывает невалидные входные данные, метод должен возвращать ошибку
 */

func TestSendMetrics_Valid(t *testing.T) {
	requestCount := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/updates", r.URL.Path)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var got []metrics.Metrics
		require.NoError(t, goccyjson.NewDecoder(r.Body).Decode(&got))
		require.Len(t, got, len(gaugeMetricNames)+1)

		gotByName := make(map[string]metrics.Metrics, len(got))
		for _, metric := range got {
			gotByName[metric.ID] = metric
		}
		require.Equal(t, metrics.Metrics{
			ID:    "Alloc",
			Type:  metrics.MetricTypeGauge,
			Value: float64Pointer(42),
		}, gotByName["Alloc"])
		require.Equal(t, metrics.Metrics{
			ID:    "GCCPUFraction",
			Type:  metrics.MetricTypeGauge,
			Value: float64Pointer(0.5),
		}, gotByName["GCCPUFraction"])
		require.Equal(t, metrics.Metrics{
			ID:    "PollCount",
			Type:  metrics.MetricTypeCounter,
			Delta: int64Pointer(7),
		}, gotByName["PollCount"])
		randomValue, ok := gotByName["RandomValue"]
		require.True(t, ok)
		require.Equal(t, metrics.MetricTypeGauge, randomValue.Type)
		require.NotNil(t, randomValue.Value)
		assert.GreaterOrEqual(t, *randomValue.Value, 0.0)
		assert.Less(t, *randomValue.Value, 1.0)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	reporter := NewMetricReporter(testServer.URL, testServer.Client(), nil, time.Second)
	err := reporter.sendMetrics(runtime.MemStats{
		Alloc:         42,
		GCCPUFraction: 0.5,
	}, 7)
	require.NoError(t, err)
	assert.Equal(t, 1, requestCount)
}

func TestSendMetrics_Invalid(t *testing.T) {
	requestCount := 0
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer testServer.Close()

	reporter := NewMetricReporter(testServer.URL, testServer.Client(), nil, time.Second)
	err := reporter.sendMetrics(runtime.MemStats{}, 0)
	assert.Error(t, err)
	assert.Equal(t, 1, requestCount)
}

func TestSendMetrics_InvalidResponseContentType(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	reporter := NewMetricReporter(testServer.URL, testServer.Client(), nil, time.Second)

	assert.Error(t, reporter.sendMetrics(runtime.MemStats{}, 0))
}

func float64Pointer(value float64) *float64 {
	return &value
}

func int64Pointer(value int64) *int64 {
	return &value
}

// func TestIntegration(t *testing.T) {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	server := "http://localhost:8080"
// 	cmd := exec.CommandContext(ctx, "go", "run", "/Users/parkersec/go-projects/go-musthave-metrics-tpl/cmd/server/main.go")
// 	cmd.Dir = "/Users/parkersec/go-projects/go-musthave-metrics-tpl"
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	if err := cmd.Start(); err != nil {
// 		t.Fatal(err)
// 	}
// 	defer cmd.Process.Kill()

// 	waitServerReady(t, server)

// 	go func() {
// 		time.Sleep(10 * time.Second)
// 		cancel()
// 	}()

// 	provider := provider.NewMetricProvider(2 * time.Second)
// 	reporter := NewMetricReporter(server, &http.Client{}, provider, 5*time.Second)
// 	err := reporter.Run(ctx)
// 	if err != nil && !errors.Is(err, context.Canceled) {
// 		assert.NoError(t, err)
// 	}
// }

// func waitServerReady(t *testing.T, server string) {
// 	t.Helper()
// 	path := "/update/counter/TestCounter/1"
// 	retry := 0
// 	for {
// 		time.Sleep(1 * time.Second)
// 		resp, err := http.Post(server+path, "text/plain", nil)
// 		if err == nil && resp.StatusCode == http.StatusOK {
// 			log.Printf("metrics server is ready: %s", server)
// 			return
// 		}
// 		retry++
// 		if retry > 5 {
// 			t.Fatal("failed to connect to metrics server")
// 		}
// 	}
// }
