package apierrs

import (
	"errors"
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
	switch {
	case errors.Is(err, metrics.ErrMetricNotFound):
		return NewError(metrics.ErrMetricNotFound.Error(), http.StatusNotFound)
	case errors.Is(err, metrics.ErrMetricNameIsEmpty),
		errors.Is(err, metrics.ErrInvalidMetricType),
		errors.Is(err, metrics.ErrMetricTypeConflict),
		errors.Is(err, metrics.ErrInvalidValueForCounter),
		errors.Is(err, metrics.ErrInvalidValueForGauge),
		errors.Is(err, metrics.ErrGaugeValueIsNil),
		errors.Is(err, metrics.ErrCounterValueIsNil):
		return NewError(clientErrorMessage(err), http.StatusBadRequest)
	default:
		return NewError("unexpected error", http.StatusInternalServerError)
	}
}

func clientErrorMessage(err error) string {
	for _, domainErr := range []error{
		metrics.ErrMetricNameIsEmpty,
		metrics.ErrInvalidMetricType,
		metrics.ErrMetricTypeConflict,
		metrics.ErrInvalidValueForCounter,
		metrics.ErrInvalidValueForGauge,
		metrics.ErrGaugeValueIsNil,
		metrics.ErrCounterValueIsNil,
	} {
		if errors.Is(err, domainErr) {
			return domainErr.Error()
		}
	}

	return "unexpected error"
}
