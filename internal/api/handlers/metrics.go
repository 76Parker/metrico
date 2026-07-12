// Пакет содержит в себе HTTP-обработчики запросов для работы с метриками
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/76Parker/metrico/internal/api/apierrs"
	"github.com/76Parker/metrico/internal/domain/metrics"
	metricsusecase "github.com/76Parker/metrico/internal/usecase/metrics"
	"github.com/gin-gonic/gin"
)

type metricService interface {
	UpdateOrCreateMetric(ctx context.Context, cmd metricsusecase.UpdateMetricCommand) error
	GetMetricByName(ctx context.Context, cmd metricsusecase.GetMetricByNameCommand) (metrics.Metrics, error)
	GetAllMetrics(ctx context.Context) (map[string]metrics.Metrics, error)
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

func (h *MetricsHandler) UpdateMetric(c *gin.Context) {
	metricType := strings.TrimSpace(c.Param("metricType"))
	metricName := strings.TrimSpace(c.Param("metricName"))
	metricValue := strings.TrimSpace(c.Param("metricValue"))
	if metricName == "" {
		c.Error(apierrs.NewError("metric name is empty", http.StatusNotFound))
		c.Status(http.StatusNotFound)
		return
	}
	if err := h.validateParams(metricType, metricName, metricValue); err != nil {
		c.Error(apierrs.NewError(err.Error(), http.StatusBadRequest))
		c.Status(http.StatusBadRequest)
		return
	}
	cmd := metricsusecase.UpdateMetricCommand{
		Name:       metricName,
		MetricType: metrics.MetricType(strings.ToLower(strings.TrimSpace(metricType))),
		Value:      metricValue,
	}
	if err := h.svc.UpdateOrCreateMetric(c.Request.Context(), cmd); err != nil {
		apiErr := apierrs.NewErrorFromService(err)
		c.Error(apiErr)
		c.Status(apiErr.Status)
		return
	}
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Status(http.StatusOK)
}
func (h *MetricsHandler) validateParams(metricType, metricName, metricValue string) error {
	if utf8.RuneCountInString(metricType) > h.maxPathParamLen {
		return fmt.Errorf("metric type is too long: max len is %d", h.maxPathParamLen)
	}
	if utf8.RuneCountInString(metricName) > h.maxPathParamLen {
		return fmt.Errorf("metric name is too long: max len is %d", h.maxPathParamLen)
	}
	if utf8.RuneCountInString(metricValue) > h.maxPathParamLen {
		return fmt.Errorf("metric value is too long: max len is %d", h.maxPathParamLen)
	}
	if metricType == "" {
		return fmt.Errorf("metric type cannot be empty")
	}
	if metricValue == "" {
		return fmt.Errorf("metric value cannot be empty")
	}
	return nil
}

func (h *MetricsHandler) GetMetricByName(c *gin.Context) {
	metricName := strings.TrimSpace(c.Param("metricName"))
	if metricName == "" {
		c.Error(apierrs.NewError("metric name is empty", http.StatusNotFound))
		c.Status(http.StatusNotFound)
		return
	}
	metricType := strings.TrimSpace(c.Param("metricType"))
	if metricType == "" {
		c.Error(apierrs.NewError("metric type is empty", http.StatusBadRequest))
		c.Status(http.StatusBadRequest)
		return
	}
	cmd := metricsusecase.GetMetricByNameCommand{
		Name:       metricName,
		MetricType: metrics.MetricType(metricType),
	}
	metric, err := h.svc.GetMetricByName(c.Request.Context(), cmd)
	if err != nil {
		apiErr := apierrs.NewErrorFromService(err)
		c.Error(apiErr)
		c.Status(apiErr.Status)
		return
	}
	c.Header("Content-Type", "text/plain; charset=utf-8")
	switch metricType {
	case "counter":
		c.String(http.StatusOK, "%d", *metric.Delta)
	case "gauge":
		c.String(http.StatusOK, "%g", *metric.Value)
	default:
		c.Status(http.StatusNotFound)
	}
}

type metricResponse struct {
	Name  string
	Value any
}

func (h *MetricsHandler) GetAllMetrics(c *gin.Context) {
	metricsSnapshot, err := h.svc.GetAllMetrics(c.Request.Context())
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	resp := make([]metricResponse, 0, len(metricsSnapshot))
	for name, metric := range metricsSnapshot {
		switch metric.Type {
		case metrics.Gauge:
			resp = append(resp, metricResponse{
				Name:  name,
				Value: *metric.Value,
			})
		case metrics.Counter:
			resp = append(resp, metricResponse{
				Name:  name,
				Value: *metric.Delta,
			})
		}
	}
	// Так как map не гарантирует порядок, сортируем срез по имени для одинаково ответа
	// Без сортировки значения всегда в разном порядке
	sort.Slice(resp, func(i, j int) bool {
		return resp[i].Name < resp[j].Name
	})
	c.HTML(http.StatusOK, "index.tmpl", resp)
}
