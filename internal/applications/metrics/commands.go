// `metrics/commands.go`
// содержит в себе usecase-команды для работы с метриками
package metrics

import "github.com/76Parker/metrico/internal/domain/metrics"

type BatchUpdateCommand = []UpdateCommand

type UpdateCommand struct {
	Name       string
	MetricType metrics.MetricType
	Delta      *int64
	Value      *float64
}
