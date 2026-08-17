package provider

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// MetricProvider собирает и обновляет Runtime-метрики
type MetricProvider struct {
	metrics           runtime.MemStats
	mu                sync.Mutex
	pollCount         int64
	previousPollCount int64
	pollInterval      time.Duration
}

func NewMetricProvider(pollInterval time.Duration) *MetricProvider {
	return NewMetricProviderWithContext(context.Background(), pollInterval)
}

func NewMetricProviderWithContext(ctx context.Context, pollInterval time.Duration) *MetricProvider {

	p := &MetricProvider{
		pollInterval: pollInterval,
	}
	p.Start(ctx)
	return p
}

// Start запускает цикл сбора метрик в отдельной горутине (не блокирующая операция)
func (mp *MetricProvider) Start(contexts ...context.Context) {
	ctx := context.Background()
	if len(contexts) > 0 && contexts[0] != nil {
		ctx = contexts[0]
	}
	timer := time.NewTicker(mp.pollInterval)
	go func() {
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				mp.mu.Lock()
				mp.pollCount++
				runtime.ReadMemStats(&mp.metrics)
				mp.mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Metrics возвращает snapshot текущих метрик
func (mp *MetricProvider) Metrics() (runtime.MemStats, int64) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	if mp.pollCount == mp.previousPollCount {
		return mp.metrics, 0
	}
	pollCount := mp.pollCount - mp.previousPollCount
	mp.previousPollCount = mp.pollCount
	return mp.metrics, pollCount
}
