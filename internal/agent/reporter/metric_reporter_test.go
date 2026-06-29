package reporter

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := "http://localhost:8080"
	cmd := exec.CommandContext(ctx, "go", "run", "/Users/parkersec/go-projects/go-musthave-metrics-tpl/cmd/server/main.go")
	cmd.Dir = "/Users/parkersec/go-projects/go-musthave-metrics-tpl"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	waitServerReady(t, server)

	go func() {
		time.Sleep(10 * time.Second)
		cancel()
	}()

	provider := provider.NewMetricProvider(2 * time.Second)
	reporter := NewMetricReporter(server, &http.Client{}, provider, 5*time.Second)
	err := reporter.Run(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		assert.NoError(t, err)
	}
}

func waitServerReady(t *testing.T, server string) {
	t.Helper()
	path := "/update/counter/TestCounter/1"
	retry := 0
	for {
		time.Sleep(1 * time.Second)
		resp, err := http.Post(server+path, "text/plain", nil)
		if err == nil && resp.StatusCode == http.StatusOK {
			log.Printf("metrics server is ready: %s", server)
			return
		}
		retry++
		if retry > 5 {
			t.Fatal("failed to connect to metrics server")
		}
	}
}
