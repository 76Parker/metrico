package provider

import (
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

	p := &MetricProvider{
		pollInterval: pollInterval,
	}
	p.Start()
	return p
}

// Start запускает цикл сбора метрик в отдельной горутине (не блокирующая операция)
func (mp *MetricProvider) Start() {
	timer := time.NewTicker(mp.pollInterval)
	go func() {
		for {
			<-timer.C
			mp.mu.Lock()
			mp.pollCount++
			runtime.ReadMemStats(&mp.metrics)
			mp.mu.Unlock()
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
