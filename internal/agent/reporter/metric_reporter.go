package reporter

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"mime"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"time"

	"github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/76Parker/metrico/internal/signature"
	goccyjson "github.com/goccy/go-json"
)

// gaugeMetricNames содержит список имен Gauge-метрик которые необходимо отправлять на сервер
var gaugeMetricNames = []string{
	"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
	"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC",
	"Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs",
	"NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse",
	"StackSys", "Sys", "TotalAlloc", "RandomValue",
}

type metricProvider interface {
	Metrics() (gaugeMetrics runtime.MemStats, pollCount int64)
}

// MetricReporter отправляет собранные метрики от MetricsProvider'a на сервер
type MetricReporter struct {
	client         *http.Client
	key            string
	reportInterval time.Duration
	url            *url.URL
	provider       metricProvider
}

// NewMetricReporter создает новый MetricReporter с заданным URL и клиентом
func NewMetricReporter(
	serverAddr string,
	client *http.Client,
	provider metricProvider,
	reportInterval time.Duration,
	key string,
) *MetricReporter {
	baseURL, err := url.Parse(serverAddr)
	if err != nil {
		return nil
	}
	return &MetricReporter{
		client:         client,
		key:            key,
		url:            baseURL,
		provider:       provider,
		reportInterval: reportInterval,
	}
}

// Run запускает бесконеный цикл отправки метрик на сервер (блокирующая операция)
func (r *MetricReporter) Run(ctx context.Context) error {
	reportTicker := time.NewTicker(r.reportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case <-reportTicker.C:
			metrics, pollCount := r.provider.Metrics()
			if err := r.sendMetrics(metrics, pollCount); err != nil {
				log.Print(err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (r *MetricReporter) sendMetrics(runtimeMetrics runtime.MemStats, pollCount int64) error {
	v := reflect.ValueOf(runtimeMetrics)
	t := v.Type()
	metricsBatch := make([]metrics.Metrics, 0, len(gaugeMetricNames)+1)

	// Собирает runtime-метрики из MemStats
	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name // Имя поля берем из типа
		fieldVal := v.Field(i)       // Значение поля берем из value
		for _, gaugeMetricName := range gaugeMetricNames {
			if fieldName == gaugeMetricName {
				var metricValue float64
				switch fieldVal.Kind() {
				case reflect.Uint64, reflect.Uint32:
					metricValue = float64(fieldVal.Uint())
				case reflect.Float64:
					metricValue = fieldVal.Float()
				}
				metricsBatch = append(metricsBatch, newGaugeMetric(metricValue, fieldName))
			}
		}
	}
	// Добавляем одну Counter-метрику - PollCount
	metricsBatch = append(metricsBatch, newCounterMetric(pollCount, "PollCount"))
	// Добавляем кастомную Gauge-метрику - RandomValue
	metricsBatch = append(metricsBatch, newGaugeMetric(rand.Float64(), "RandomValue"))

	return r.sendBatch(metricsBatch)
}

func newGaugeMetric(metricValue float64, metricName string) metrics.Metrics {
	return metrics.Metrics{
		ID:    metricName,
		Type:  metrics.MetricTypeGauge,
		Value: &metricValue,
	}
}

func newCounterMetric(metricValue int64, metricName string) metrics.Metrics {
	return metrics.Metrics{
		ID:    metricName,
		Type:  metrics.MetricTypeCounter,
		Delta: &metricValue,
	}
}

func (r *MetricReporter) sendBatch(metricsBatch []metrics.Metrics) error {
	body, err := goccyjson.Marshal(metricsBatch)
	if err != nil {
		return fmt.Errorf("marshal metric batch: %w", err)
	}

	var response *http.Response
	defer func() {
		if response != nil {
			response.Body.Close()
		}
	}()
	err = withRetry(func() error {
		request, err := http.NewRequest(http.MethodPost, r.batchUpdateURL(), bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("create batch update request: %w", err)
		}
		request.Header.Set("Content-Type", "application/json")
		if r.key != "" {
			request.Header.Set(signature.HeaderName, signature.Sum(body, r.key))
		}

		resp, err := r.client.Do(request)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		response = resp
		return nil
	})
	if err != nil {
		return err
	}
	contentType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil {
		return fmt.Errorf("parse response content type: %w", err)
	}
	if contentType != "application/json" {
		return fmt.Errorf("unexpected response content type: %s", contentType)
	}

	return nil
}

func (r *MetricReporter) batchUpdateURL() string {
	requestURL := *r.url
	requestURL.Path = "/updates"
	requestURL.RawQuery = ""
	return requestURL.String()
}
