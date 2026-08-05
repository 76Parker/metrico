package errmap

import (
	"errors"
	"net/http"

	"github.com/76Parker/metrico/internal/api/handlers"
	"github.com/76Parker/metrico/internal/domain/metrics"
)

type registry map[error]Error

var Registry = registry{
	// Transport-specific errors
	handlers.ErrFieldTooLong:       newError("metric value is too long", CodeInvalidRequest, http.StatusBadRequest),
	handlers.ErrInvalidJSON:        newError("invalid input json", CodeInvalidJSON, http.StatusBadRequest),
	handlers.ErrMetricTypeNotFound: newError("metric type not found", CodeNotFound, http.StatusNotFound),

	// Metrics domain-specific errors
	metrics.ErrMetricNotFound:         newError("metric not found", CodeNotFound, http.StatusNotFound),
	metrics.ErrMetricNameIsEmpty:      newError("metric name is empty", CodeEmpty, http.StatusNotFound),
	metrics.ErrInvalidMetricType:      newError("invalid metric type", CodeInvalidMetricType, http.StatusBadRequest),
	metrics.ErrInvalidValueForCounter: newError("invalid value for counter metric", CodeInvalidMetricValue, http.StatusBadRequest),
	metrics.ErrInvalidValueForGauge:   newError("invalid value for gauge metric", CodeInvalidMetricValue, http.StatusBadRequest),
	metrics.ErrGaugeValueIsNil:        newError("gauge value is nil", CodeInvalidMetricValue, http.StatusBadRequest),
	metrics.ErrCounterValueIsNil:      newError("counter value is nil", CodeInvalidMetricValue, http.StatusBadRequest),
}

var UnexpectedError = newError("unexpected error", CodeInternal, http.StatusInternalServerError)

func (r registry) Resolve(err error) (Error, bool) {
	for target, resolved := range r {
		if errors.Is(err, target) {
			return resolved, true
		}
	}
	return Error{}, false
}

// Структура ошибки возвращаемая клиенту
type Error struct {
	// Публичное сообщение для клиента
	Message string `json:"message"`
	// Код ошибки для удобного отображения на Frontend
	Code string `json:"code"`
	// HTTP статус код
	Status int `json:"status"`
}

func newError(message, code string, status int) Error {
	return Error{
		Message: message,
		Code:    code,
		Status:  status,
	}
}
