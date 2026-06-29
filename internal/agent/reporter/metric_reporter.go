package reporter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"time"
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
) *MetricReporter {
	baseURL, err := url.Parse(serverAddr)
	if err != nil {
		return nil
	}
	return &MetricReporter{
		client:         client,
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

func (r *MetricReporter) sendMetrics(metrics runtime.MemStats, pollCount int64) error {
	v := reflect.ValueOf(metrics)
	t := v.Type()
	errs := make([]error, 0, 10)

	// Отправляет runtime-метрики из MemStats
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
				if err := r.sendGaugeMetric(metricValue, fieldName); err != nil {
					errs = append(errs, fmt.Errorf("failed to send gauge metric: name: %s error: %w", fieldName, err))
				}
			}
		}
	}
	// Отправляем одну Counter-метрику - PollCount
	if err := r.sendCounterMetric(pollCount, "PollCount"); err != nil {
		errs = append(errs, fmt.Errorf("failed to send counter metric: name: PollCount error: %w", err))
	}
	// Отправляем кастомную Gauge-метрику - RandomValue
	if err := r.sendGaugeMetric(rand.Float64(), "RandomValue"); err != nil {
		errs = append(errs, fmt.Errorf("failed to send gauge metric: name: RandomValue error: %w", err))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (r *MetricReporter) sendGaugeMetric(metricValue float64, metricName string) error {
	requestURL := r.metricURL("gauge", metricName, fmt.Sprintf("%f", metricValue))
	resp, err := r.client.Post(requestURL, "text/plain", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (r *MetricReporter) sendCounterMetric(metricValue int64, metricName string) error {
	requestURL := r.metricURL("counter", metricName, fmt.Sprintf("%d", metricValue))
	resp, err := r.client.Post(requestURL, "text/plain", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (r *MetricReporter) metricURL(metricType, metricName, metricValue string) string {
	requestURL := *r.url
	requestURL.Path = fmt.Sprintf(
		"/update/%s/%s/%s",
		url.PathEscape(metricType),
		url.PathEscape(metricName),
		url.PathEscape(metricValue),
	)
	requestURL.RawQuery = ""
	return requestURL.String()
}
