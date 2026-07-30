// Пакет содержит в себе HTTP-обработчики запросов для работы с метриками
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/76Parker/metrico/internal/api/apierrs"
	metricsapp "github.com/76Parker/metrico/internal/applications/metrics"
	"github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/gin-gonic/gin"
	goccyjson "github.com/goccy/go-json"
)

const batchReservationSize = 50

type metricsApplication interface {
	Update(ctx context.Context, cmd metricsapp.UpdateCommand) error
	BatchUpdate(ctx context.Context, cmd metricsapp.BatchUpdateCommand) error
	GetByName(ctx context.Context, name string) (metrics.Metrics, error)
	GetAll(ctx context.Context) ([]metrics.Metrics, error)
}
type MetricsHandler struct {
	metricSvc       metricsApplication
	maxPathParamLen int
}

func NewMetricsHandler(metricSvc metricsApplication) *MetricsHandler {
	maxPathParamLen := 64
	return &MetricsHandler{
		metricSvc:       metricSvc,
		maxPathParamLen: maxPathParamLen,
	}
}

func (h *MetricsHandler) Update(c *gin.Context) {
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
	metricKind := metrics.MetricType(metricType)
	var delta *int64
	var value *float64
	switch metricKind {
	case metrics.MetricTypeGauge:
		parsedValue, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			c.Error(apierrs.NewError(metrics.ErrInvalidValueForGauge.Error(), http.StatusBadRequest))
			c.Status(http.StatusBadRequest)
			return
		}
		value = &parsedValue
	case metrics.MetricTypeCounter:
		parsedDelta, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			c.Error(apierrs.NewError(metrics.ErrInvalidValueForCounter.Error(), http.StatusBadRequest))
			c.Status(http.StatusBadRequest)
			return
		}
		delta = &parsedDelta
	default:
		c.Error(apierrs.NewError(metrics.ErrInvalidMetricType.Error(), http.StatusBadRequest))
		c.Status(http.StatusBadRequest)
		return
	}
	cmd := metricsapp.UpdateCommand{
		Name:       metricName,
		MetricType: metricKind,
		Delta:      delta,
		Value:      value,
	}
	if err := h.metricSvc.Update(c.Request.Context(), cmd); err != nil {
		apiErr := apierrs.NewErrorFromService(err)
		c.Error(err)
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

func (h *MetricsHandler) UpdateFromJSON(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	decoder := goccyjson.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var metric metrics.Metrics
	if err := decoder.Decode(&metric); err != nil {
		c.Error(apierrs.NewError("invalid input JSON", http.StatusBadRequest))
		c.Status(http.StatusBadRequest)
		return
	}
	cmd := metricsapp.UpdateCommand{
		Name:       metric.ID,
		MetricType: metric.Type,
		Delta:      metric.Delta,
		Value:      metric.Value,
	}
	if err := h.metricSvc.Update(c.Request.Context(), cmd); err != nil {
		apiErr := apierrs.NewErrorFromService(err)
		c.Error(err)
		c.Status(apiErr.Status)
		return
	}
	updatedMetric, err := h.metricSvc.GetByName(c.Request.Context(), metric.ID)
	if err != nil {
		apiErr := apierrs.NewErrorFromService(err)
		c.Error(err)
		c.Status(apiErr.Status)
		return
	}

	response, err := goccyjson.Marshal(updatedMetric)
	if err != nil {
		c.Error(apierrs.NewError("failed to serialize metric", http.StatusInternalServerError))
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json", response)
}

func (h *MetricsHandler) BatchUpdateFromJSON(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	decoder := goccyjson.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	metricsBatch := make([]metrics.Metrics, 0, batchReservationSize)
	if err := decoder.Decode(&metricsBatch); err != nil {
		c.Error(apierrs.NewError("invalid input JSON", http.StatusBadRequest))
		c.Status(http.StatusBadRequest)
		return
	}
	batchCommand := make(metricsapp.BatchUpdateCommand, 0, len(metricsBatch))
	for _, metric := range metricsBatch {
		batchCommand = append(batchCommand, metricsapp.UpdateCommand{
			Name:       metric.ID,
			MetricType: metrics.MetricType(strings.TrimSpace(string(metric.Type))),
			Delta:      metric.Delta,
			Value:      metric.Value,
		})
	}
	if err := h.metricSvc.BatchUpdate(c.Request.Context(), batchCommand); err != nil {
		apiErr := apierrs.NewErrorFromService(err)
		c.Error(err)
		c.Status(apiErr.Status)
		return
	}
	c.Status(http.StatusOK)
}

func (h *MetricsHandler) GetFromJSON(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	decoder := goccyjson.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	var request metrics.Metrics
	if err := decoder.Decode(&request); err != nil {
		c.Error(apierrs.NewError("invalid input JSON", http.StatusBadRequest))
		c.Status(http.StatusBadRequest)
		return
	}

	request.ID = strings.TrimSpace(request.ID)
	request.Type = metrics.MetricType(strings.TrimSpace(string(request.Type)))
	if request.ID == "" {
		c.Error(apierrs.NewError(metrics.ErrMetricNameIsEmpty.Error(), http.StatusBadRequest))
		c.Status(http.StatusBadRequest)
		return
	}
	if request.Type != metrics.MetricTypeCounter && request.Type != metrics.MetricTypeGauge {
		c.Error(apierrs.NewError(metrics.ErrInvalidMetricType.Error(), http.StatusBadRequest))
		c.Status(http.StatusBadRequest)
		return
	}

	metric, err := h.metricSvc.GetByName(c.Request.Context(), request.ID)
	if err != nil {
		apiErr := apierrs.NewErrorFromService(err)
		c.Error(err)
		c.Status(apiErr.Status)
		return
	}

	response, err := goccyjson.Marshal(metric)
	if err != nil {
		c.Error(apierrs.NewError("failed to serialize metric", http.StatusInternalServerError))
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", response)
}

func (h *MetricsHandler) GetByName(c *gin.Context) {
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

	metric, err := h.metricSvc.GetByName(c.Request.Context(), metricName)
	if err != nil {
		apiErr := apierrs.NewErrorFromService(err)
		c.Error(err)
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

func (h *MetricsHandler) GetAll(c *gin.Context) {
	metricsSnapshot, err := h.metricSvc.GetAll(c.Request.Context())
	if err != nil {
		c.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	resp := make([]metricResponse, 0, len(metricsSnapshot))
	for _, metric := range metricsSnapshot {
		switch metric.Type {
		case metrics.MetricTypeGauge:
			resp = append(resp, metricResponse{
				Name:  metric.ID,
				Value: *metric.Value,
			})
		case metrics.MetricTypeCounter:
			resp = append(resp, metricResponse{
				Name:  metric.ID,
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
