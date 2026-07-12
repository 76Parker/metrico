package apierrs

import (
	"net/http"

	"github.com/76Parker/metrico/internal/domain/metrics"
)

type Error struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e Error) Error() string {
	return e.Message
}

func NewError(message string, status int) Error {
	return Error{
		Message: message,
		Status:  status,
	}
}

func NewErrorFromService(err error) Error {
	switch err {
	case metrics.ErrMetricNotFound:
		return NewError(err.Error(), http.StatusNotFound)
	case metrics.ErrMetricNameIsEmpty,
		metrics.ErrInvalidMetricType,
		metrics.ErrInvalidValueForCounter,
		metrics.ErrInvalidValueForGauge,
		metrics.ErrGaugeValueIsNil,
		metrics.ErrCounterValueIsNil:
		return NewError(err.Error(), http.StatusBadRequest)
	default:
		return NewError("unexpected error", http.StatusInternalServerError)
	}
}
