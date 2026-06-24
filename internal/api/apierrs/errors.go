// Пакет содержит в себе ошибки API которые не раскрывают внутренние детали приложения
package apierrs

import (
	"errors"
	"net/http"

	"github.com/76Parker/metrico/internal/domain/metrics"
)

var (
	ErrMetricNameNotFound       = errors.New("metric name not found in path parameters")
	ErrMetricTypeCannotBeEmpty  = errors.New("metric type cannot be empty")
	ErrMetricValueCannotBeEmpty = errors.New("metric value cannot be empty")

	ErrMetricNameTooLong  = errors.New("metric name is too long")
	ErrMetricValueTooLong = errors.New("metric value is too long")
	ErrMetricTypeTooLong  = errors.New("metric type is too long")
)

// ToAPIError Функция для маппинга ошибок сервисного слоя в API ошибки.
// Эта функция нужна тк не все ошибки сервисного слоя мы должны раскрывать пользователю напрямую.
// В case-случаях будут обрабатываться ошибки, детали которых мы хотим скрыть от пользователя.
// Результат этой функции - ошибка, которая не раскрывает внутренних деталей приложения.
func ToAPIError(err error) (e error, code int) {
	switch err {
	case nil:
		return nil, http.StatusOK
	// пользователь не должен знать что есть такой тип nil и у нас произошло такое преобразование
	// отдаем просто понятную ошибку, без раскрытия внутренних деталей.
	case metrics.ErrGaugeValueIsNil:
		return errors.New("invalid gauge value"), http.StatusBadRequest
	case metrics.ErrCounterValueIsNil:
		return errors.New("invalid counter value"), http.StatusBadRequest
	case metrics.ErrInvalidValueForCounter:
		return errors.New("invalid value for counter"), http.StatusBadRequest
	case metrics.ErrInvalidValueForGauge:
		return errors.New("invalid value for gauge"), http.StatusBadRequest
	case metrics.ErrInvalidMetricType:
		return errors.New("invalid metric type"), http.StatusBadRequest
	default:
		return err, http.StatusInternalServerError
	}
}
