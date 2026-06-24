// Пакет содержит в себе HTTP-обработчики запросов для работы с метриками
package handlers

import (
	"context"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/76Parker/metrico/internal/api/apierrs"
	"github.com/76Parker/metrico/internal/domain/metrics"
	metricsusecase "github.com/76Parker/metrico/internal/usecase/metrics"
)

type metricService interface {
	UpdateOrCreateMetric(ctx context.Context, cmd metricsusecase.UpdateMetricCommand) error
}

type MetricsHandler struct {
	svc             metricService
	maxPathParamLen int
}

func NewMetricsHandler(svc metricService) *MetricsHandler {
	maxPathParamLen := 64
	return &MetricsHandler{
		svc:             svc,
		maxPathParamLen: maxPathParamLen,
	}
}

func (h *MetricsHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	metricType := strings.TrimSpace(r.PathValue("metricType"))
	metricName := strings.TrimSpace(r.PathValue("metricName"))
	metricValue := strings.TrimSpace(r.PathValue("metricValue"))
	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := h.validateParams(metricType, metricName, metricValue); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cmd := metricsusecase.UpdateMetricCommand{
		Name:       metricName,
		MetricType: metrics.MetricType(strings.ToLower(strings.TrimSpace(metricType))),
		Value:      metricValue,
	}
	if err := h.svc.UpdateOrCreateMetric(context.Background(), cmd); err != nil {
		_, code := apierrs.ToAPIError(err)
		w.WriteHeader(code)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
func (h *MetricsHandler) validateParams(metricType, metricName, metricValue string) error {
	if utf8.RuneCountInString(metricType) > h.maxPathParamLen {
		return apierrs.ErrMetricTypeTooLong
	}
	if utf8.RuneCountInString(metricName) > h.maxPathParamLen {
		return apierrs.ErrMetricNameTooLong
	}
	if utf8.RuneCountInString(metricValue) > h.maxPathParamLen {
		return apierrs.ErrMetricValueTooLong
	}
	if metricType == "" {
		return apierrs.ErrMetricTypeCannotBeEmpty
	}
	if metricValue == "" {
		return apierrs.ErrMetricValueCannotBeEmpty
	}
	return nil
}
