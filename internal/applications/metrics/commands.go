// `metrics/commands.go`
// содержит в себе usecase-команды для работы с метриками
package metrics

type BatchUpdateCommand = []UpdateCommand

type UpdateCommand struct {
	Name       string
	MetricType string
	Value      string
}
