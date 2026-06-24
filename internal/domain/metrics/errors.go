// `metrics/errors.go`
// содержит в себе ошибки, связанные с метриками
package metrics

import "errors"

var (
	ErrInvalidMetricType      = errors.New("invalid metric type")
	ErrInvalidValueForCounter = errors.New("invalid value for counter metric")
	ErrInvalidValueForGauge   = errors.New("invalid value for gauge metric")

	ErrGaugeValueIsNil   = errors.New("gauge value is nil")
	ErrCounterValueIsNil = errors.New("counter value is nil")
)
