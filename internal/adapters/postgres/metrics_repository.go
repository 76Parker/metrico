package postgres

import (
	"context"
	"fmt"

	"github.com/76Parker/metrico/internal/adapters/postgres/pgen"
	metricapp "github.com/76Parker/metrico/internal/applications/metrics"
	"github.com/76Parker/metrico/internal/domain/metrics"
	goccyjson "github.com/goccy/go-json"
)

type metricsRepo struct {
	q *pgen.Queries
}

func NewMetricsRepository(q *pgen.Queries) *metricsRepo {
	return &metricsRepo{q: q}
}

func (t *metricsRepo) Apply(ctx context.Context, changes []metricapp.Change) error {
	batch := make(pgen.BatchUpsertParam, 0, len(changes))
	for _, change := range changes {
		batchElem := pgen.BatchUpsertMetric{
			Name: change.Name(),
			Type: string(change.MetricType()),
		}
		switch change.MetricType() {
		case metrics.MetricTypeGauge:
			value := change.Value()
			batchElem.Value = &value
		case metrics.MetricTypeCounter:
			delta := change.Delta()
			batchElem.Delta = &delta
		}
		batch = append(batch, batchElem)
	}
	batchByte, err := goccyjson.Marshal(batch)
	if err != nil {
		return fmt.Errorf("unmarshal batch: %w", err)
	}
	err = t.q.BatchUpsert(ctx, batchByte)
	if err != nil {
		return fmt.Errorf("batch upsert: %w", err)
	}
	return nil
}

func (r *metricsRepo) Get(ctx context.Context, metricName string) (metrics.Metrics, error) {

	metric, err := r.q.GetByName(ctx, metricName)
	if err != nil {
		return metrics.Metrics{}, err
	}
	return metrics.Metrics{
		ID:    metric.Name,
		Type:  metrics.MetricType(metric.Type),
		Value: &metric.Value.Float64,
		Delta: &metric.Delta.Int64,
	}, nil
}

func (r *metricsRepo) GetAll(ctx context.Context) ([]metrics.Metrics, error) {
	listedMetrics, err := r.q.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]metrics.Metrics, 0, len(listedMetrics))
	for _, metric := range listedMetrics {
		result = append(result, metrics.Metrics{
			ID:    metric.Name,
			Type:  metrics.MetricType(metric.Type),
			Value: &metric.Value.Float64,
			Delta: &metric.Delta.Int64,
		})
	}
	return result, nil

}
