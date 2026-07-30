package metrics

import "github.com/76Parker/metrico/internal/domain/metrics"

type ChangeKind int

const (
	ChangeKindInvalid    ChangeKind = iota // Неизвестный тип мутации метрики
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

func newSetGaugeMutation(name string, value float64) Change {
	return Change{
		name:  name,
		kind:  ChangeKindSetGauge,
		value: value,
	}
}

// NewSetGaugeChange creates a change that sets a gauge value.
func NewSetGaugeChange(name string, value float64) Change {
	return newSetGaugeMutation(name, value)
}

func newAddCounterMutation(name string, delta int64) Change {
	return Change{
		name:  name,
		kind:  ChangeKindAddCounter,
		delta: delta,
	}
}

// NewAddCounterChange creates a change that adds delta to a counter.
func NewAddCounterChange(name string, delta int64) Change {
	return newAddCounterMutation(name, delta)
}
