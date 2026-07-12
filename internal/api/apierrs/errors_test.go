package apierrs

import (
	"net/http"
	"testing"

	"github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/stretchr/testify/require"
)

func TestNewErrorServiceDomainErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "metric not found", err: metrics.ErrMetricNotFound, wantStatus: http.StatusNotFound},
		{name: "metric name is empty", err: metrics.ErrMetricNameIsEmpty, wantStatus: http.StatusBadRequest},
		{name: "metric type is invalid", err: metrics.ErrInvalidMetricType, wantStatus: http.StatusBadRequest},
		{name: "counter value is invalid", err: metrics.ErrInvalidValueForCounter, wantStatus: http.StatusBadRequest},
		{name: "gauge value is invalid", err: metrics.ErrInvalidValueForGauge, wantStatus: http.StatusBadRequest},
		{name: "gauge value is nil", err: metrics.ErrGaugeValueIsNil, wantStatus: http.StatusBadRequest},
		{name: "counter value is nil", err: metrics.ErrCounterValueIsNil, wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewErrorFromService(tt.err)

			require.Equal(t, tt.err.Error(), got.Message)
			require.Equal(t, tt.wantStatus, got.Status)
		})
	}
}
