package provider

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemMetricProvider struct {
	mu       sync.RWMutex
	metrics  []metrics.Metrics
	interval time.Duration
}

func NewSystemMetricProvider(interval time.Duration) *SystemMetricProvider {
	return NewSystemMetricProviderWithContext(context.Background(), interval)
}

func NewSystemMetricProviderWithContext(ctx context.Context, interval time.Duration) *SystemMetricProvider {
	provider := &SystemMetricProvider{interval: interval}
	provider.Start(ctx)
	return provider
}

func (p *SystemMetricProvider) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				collected, err := collectSystemMetrics()
				if err != nil {
					continue
				}
				p.mu.Lock()
				p.metrics = collected
				p.mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (p *SystemMetricProvider) Metrics() []metrics.Metrics {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]metrics.Metrics, len(p.metrics))
	copy(result, p.metrics)
	return result
}

func collectSystemMetrics() ([]metrics.Metrics, error) {
	memory, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("read memory metrics: %w", err)
	}
	cpuUsage, err := cpu.Percent(0, true)
	if err != nil {
		return nil, fmt.Errorf("read CPU metrics: %w", err)
	}
	cpuCount := runtime.NumCPU()
	if len(cpuUsage) > cpuCount {
		cpuUsage = cpuUsage[:cpuCount]
	}
	for len(cpuUsage) < cpuCount {
		cpuUsage = append(cpuUsage, 0)
	}
	result := make([]metrics.Metrics, 0, len(cpuUsage)+2)
	result = append(result,
		newSystemGauge("TotalMemory", float64(memory.Total)),
		newSystemGauge("FreeMemory", float64(memory.Free)),
	)
	for index, utilization := range cpuUsage {
		result = append(result, newSystemGauge("CPUutilization"+strconv.Itoa(index+1), utilization))
	}
	return result, nil
}

func newSystemGauge(name string, value float64) metrics.Metrics {
	return metrics.Metrics{ID: name, Type: metrics.MetricTypeGauge, Value: &value}
}
