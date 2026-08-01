// Пакет содержит в себе HTTP-обработчики запросов для работы с метриками
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"

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
	params := [3]string{
		strings.TrimSpace(c.Param("metricType")),
		strings.TrimSpace(c.Param("metricName")),
		strings.TrimSpace(c.Param("metricValue")),
	}
	for _, param := range params {
		if utf8.RuneCountInString(param) > h.maxPathParamLen {
			c.Error(ErrFieldTooLong)
			return
		}
	}
	cmd := metricsapp.UpdateCommand{
		Name:       params[1],
		MetricType: params[0],
		Value:      params[2],
	}
	if err := h.metricSvc.Update(c.Request.Context(), cmd); err != nil {
		c.Error(err)
		return
	}
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Status(http.StatusOK)
}

type updateRequest struct {
	Name       string          `json:"name"`
	MetricType string          `json:"type"`
	Value      json.RawMessage `json:"value"`
	Delta      json.RawMessage `json:"delta"`
}

func (r updateRequest) toCommand() metricsapp.UpdateCommand {
	rawValue := r.Value
	if r.MetricType == string(metrics.MetricTypeCounter) {
		rawValue = r.Delta
	}

	return metricsapp.UpdateCommand{
		Name:       r.Name,
		MetricType: r.MetricType,
		Value:      string(rawValue),
	}
}

func (h *MetricsHandler) UpdateFromJSON(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	decoder := goccyjson.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var req updateRequest
	if err := decoder.Decode(&req); err != nil {
		c.Error(ErrInvalidJSON)
		return
	}
	cmd := req.toCommand()
	if err := h.metricSvc.Update(c.Request.Context(), cmd); err != nil {
		c.Error(err)
		return
	}
	updatedMetric, err := h.metricSvc.GetByName(c.Request.Context(), cmd.Name)
	if err != nil {
		c.Error(err)
		return
	}

	response, err := goccyjson.Marshal(updatedMetric)
	if err != nil {
		c.Error(fmt.Errorf("marshal updated metric: %w", err))
		return
	}
	c.Data(http.StatusOK, "application/json", response)
}

func (h *MetricsHandler) BatchUpdateFromJSON(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	decoder := goccyjson.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	batchRequest := make([]updateRequest, 0, batchReservationSize)
	if err := decoder.Decode(&batchRequest); err != nil {
		c.Error(ErrInvalidJSON)
		return
	}
	batchCommand := make(metricsapp.BatchUpdateCommand, 0, len(batchRequest))
	for _, req := range batchRequest {
		batchCommand = append(batchCommand, req.toCommand())
	}
	if err := h.metricSvc.BatchUpdate(c.Request.Context(), batchCommand); err != nil {
		c.Error(err)
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
		c.Error(ErrInvalidJSON)
		return
	}
	request.ID = strings.TrimSpace(request.ID)
	metric, err := h.metricSvc.GetByName(c.Request.Context(), request.ID)
	if err != nil {
		c.Error(err)
		return
	}
	response, err := goccyjson.Marshal(metric)
	if err != nil {
		c.Error(fmt.Errorf("marshal metric: %w", err))
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", response)
}

func (h *MetricsHandler) GetByName(c *gin.Context) {
	metricName := strings.TrimSpace(c.Param("metricName"))
	metricType := strings.TrimSpace(c.Param("metricType"))
	metric, err := h.metricSvc.GetByName(c.Request.Context(), metricName)
	if err != nil {
		c.Error(err)
		return
	}
	c.Header("Content-Type", "text/plain; charset=utf-8")
	switch metricType {
	case "counter":
		c.String(http.StatusOK, "%d", *metric.Delta)
	case "gauge":
		c.String(http.StatusOK, "%g", *metric.Value)
	default:
		c.Error(ErrMetricTypeNotFound)
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
