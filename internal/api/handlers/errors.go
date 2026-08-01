package handlers

import "errors"

// Ошибки для простой валидации на транспортном уровне
var (
	ErrFieldTooLong       = errors.New("field is too long")
	ErrInvalidJSON        = errors.New("invalid input json")
	ErrMetricTypeNotFound = errors.New("metric type not found")
)
