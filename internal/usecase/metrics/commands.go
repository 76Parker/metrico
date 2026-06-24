// `metrics/commands.go`
// содержит в себе usecase-команды для работы с метриками
package metrics

import "github.com/76Parker/metrico/internal/domain/metrics"

type UpdateMetricCommand struct {
	Name       string
	MetricType metrics.MetricType
	Value      string
}
