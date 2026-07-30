package apierrs

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/stretchr/testify/require"
)

func TestNewErrorServiceDomainErrors(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantMessage string
		wantStatus  int
	}{
		{name: "metric not found", err: metrics.ErrMetricNotFound, wantMessage: metrics.ErrMetricNotFound.Error(), wantStatus: http.StatusNotFound},
		{name: "wrapped metric not found", err: fmt.Errorf("get metric: %w", metrics.ErrMetricNotFound), wantMessage: metrics.ErrMetricNotFound.Error(), wantStatus: http.StatusNotFound},
		{name: "metric name is empty", err: metrics.ErrMetricNameIsEmpty, wantMessage: metrics.ErrMetricNameIsEmpty.Error(), wantStatus: http.StatusBadRequest},
		{name: "metric type is invalid", err: metrics.ErrInvalidMetricType, wantMessage: metrics.ErrInvalidMetricType.Error(), wantStatus: http.StatusBadRequest},
		{name: "metric type conflicts with existing metric", err: metrics.ErrMetricTypeConflict, wantMessage: metrics.ErrMetricTypeConflict.Error(), wantStatus: http.StatusBadRequest},
		{name: "counter value is invalid", err: metrics.ErrInvalidValueForCounter, wantMessage: metrics.ErrInvalidValueForCounter.Error(), wantStatus: http.StatusBadRequest},
		{name: "gauge value is invalid", err: metrics.ErrInvalidValueForGauge, wantMessage: metrics.ErrInvalidValueForGauge.Error(), wantStatus: http.StatusBadRequest},
		{name: "gauge value is nil", err: metrics.ErrGaugeValueIsNil, wantMessage: metrics.ErrGaugeValueIsNil.Error(), wantStatus: http.StatusBadRequest},
		{name: "counter value is nil", err: metrics.ErrCounterValueIsNil, wantMessage: metrics.ErrCounterValueIsNil.Error(), wantStatus: http.StatusBadRequest},
		{name: "unexpected error", err: errors.New("database connection lost"), wantMessage: "unexpected error", wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewErrorFromService(tt.err)

			require.Equal(t, tt.wantMessage, got.Message)
			require.Equal(t, tt.wantStatus, got.Status)
		})
	}
}
