// `metrics/commands.go`
// содержит в себе usecase-команды для работы с метриками
package metrics

import "github.com/76Parker/metrico/internal/domain/metrics"

type UpdateCommand struct {
	Name       string
	MetricType metrics.MetricType
	Delta      *int64
	Value      *float64
}

type GetCommand struct {
	MetricType metrics.MetricType
	Name       string
}
