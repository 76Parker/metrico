package postgres

import (
	"context"

	"github.com/76Parker/metrico/internal/adapters/postgres/pgen"
	"github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/jackc/pgx/v5/pgtype"
)

type metricsRepo struct {
	q *pgen.Queries
}

func NewMetricsRepository(q *pgen.Queries) *metricsRepo {
	return &metricsRepo{q: q}
}

func (r *metricsRepo) UpdateOrCreate(ctx context.Context, metricName string, metric metrics.Metrics) error {
	var delta pgtype.Int8
	if metric.Delta != nil {
		delta = pgtype.Int8{
			Int64: *metric.Delta,
			Valid: true,
		}
	}
	var value pgtype.Float8
	if metric.Value != nil {
		value = pgtype.Float8{
			Float64: *metric.Value,
			Valid:   true,
		}
	}
	sqlInput := pgen.UpsertParams{
		Name:  metricName,
		Type:  string(metric.Type),
		Value: value,
		Delta: delta,
	}
	err := r.q.Upsert(ctx, sqlInput)
	return err
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
