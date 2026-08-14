package metrics

import "github.com/76Parker/metrico/internal/domain/metrics"

type ChangeKind int

const (
	ChangeKindInvalid    ChangeKind = iota // Неизвестный тип изменения метрики
	ChangeKindSetGauge                     // устанавливает значение для gauge-метрики
	ChangeKindAddCounter                   // увеличивает значение counter-метрики
)

// Change представляет собой операцию изменения метрики, которая может быть применена к физическому хранилищу метрик,
// структуру мы передаем если нам необходимо произвести изменения в физическом хранилище метрик
type Change struct {
	name  string
	kind  ChangeKind
	value float64
	delta int64
}

func (c Change) Kind() ChangeKind {
	return c.kind
}

func (c Change) Name() string {
	return c.name
}

func (c Change) Value() float64 {
	return c.value
}

func (c Change) Delta() int64 {
	return c.delta
}

func (c Change) MetricType() metrics.MetricType {
	switch c.kind {
	case ChangeKindSetGauge:
		return metrics.MetricTypeGauge
	case ChangeKindAddCounter:
		return metrics.MetricTypeCounter
	default:
		return metrics.MetricTypeInvalid
	}
}

func newSetGaugeChange(name string, value float64) Change {
	return Change{
		name:  name,
		kind:  ChangeKindSetGauge,
		value: value,
	}
}

// NewSetGaugeChange creates a change that sets a gauge value.
func NewSetGaugeChange(name string, value float64) Change {
	return newSetGaugeChange(name, value)
}

func newAddCounterChange(name string, delta int64) Change {
	return Change{
		name:  name,
		kind:  ChangeKindAddCounter,
		delta: delta,
	}
}
func NewAddCounterChange(name string, delta int64) Change {
	return newAddCounterChange(name, delta)
}
