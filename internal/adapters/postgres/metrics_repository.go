package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/76Parker/metrico/internal/adapters/postgres/pgen"
	metricapp "github.com/76Parker/metrico/internal/applications/metrics"
	"github.com/76Parker/metrico/internal/domain/metrics"
	goccyjson "github.com/goccy/go-json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type metricsRepo struct {
	q *pgen.Queries
}

func NewMetricsRepository(q *pgen.Queries) *metricsRepo {
	return &metricsRepo{q: q}
}

func (t *metricsRepo) Apply(ctx context.Context, changes []metricapp.Change) error {
	batch, err := makeBatch(changes)
	if err != nil {
		return err
	}

	batchByte, err := goccyjson.Marshal(batch)
	if err != nil {
		return fmt.Errorf("marshal batch: %w", err)
	}
	if err := t.q.BatchUpsert(ctx, batchByte); err != nil {
		return fmt.Errorf("batch upsert: %w", err)
	}

	return nil
}

func makeBatch(changes []metricapp.Change) (pgen.BatchUpsertParam, error) {
	batch := make(pgen.BatchUpsertParam, 0, len(changes))
	indices := make(map[string]int, len(changes))
	for _, change := range changes {
		if index, ok := indices[change.Name()]; ok {
			if batch[index].Type != string(change.MetricType()) {
				return nil, metrics.ErrMetricTypeConflict
			}

			switch change.MetricType() {
			case metrics.MetricTypeGauge:
				value := change.Value()
				batch[index].Value = &value
			case metrics.MetricTypeCounter:
				delta := *batch[index].Delta + change.Delta()
				batch[index].Delta = &delta
			}

			continue
		}

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
		indices[change.Name()] = len(batch) - 1
	}

	return batch, nil
}

func (r *metricsRepo) Get(ctx context.Context, metricName string) (metrics.Metrics, error) {

	metric, err := r.q.GetByName(ctx, metricName)
	if err != nil {
		return metrics.Metrics{}, newGetMetricError(err)
	}
	return metrics.Metrics{
		ID:    metric.Name,
		Type:  metrics.MetricType(metric.Type),
		Value: nullableFloat64(metric.Value),
		Delta: nullableInt64(metric.Delta),
	}, nil
}

func newGetMetricError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return metrics.ErrMetricNotFound
	}

	return fmt.Errorf("get metric by name: %w", err)
}

func (r *metricsRepo) GetAll(ctx context.Context) ([]metrics.Metrics, error) {
	listedMetrics, err := r.q.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list metrics: %w", err)
	}
	result := make([]metrics.Metrics, 0, len(listedMetrics))
	for _, metric := range listedMetrics {
		result = append(result, metrics.Metrics{
			ID:    metric.Name,
			Type:  metrics.MetricType(metric.Type),
			Value: nullableFloat64(metric.Value),
			Delta: nullableInt64(metric.Delta),
		})
	}
	return result, nil

}

func nullableFloat64(value pgtype.Float8) *float64 {
	if !value.Valid {
		return nil
	}

	return &value.Float64
}

func nullableInt64(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}

	return &value.Int64
}
