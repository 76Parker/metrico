package errmap

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/76Parker/metrico/internal/api/handlers"
	"github.com/stretchr/testify/require"
)

func TestRegistryResolveHandlerErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"valid/long_metric_value", handlers.ErrFieldTooLong, http.StatusBadRequest, CodeInvalidRequest},
		{"valid/invalid_json", handlers.ErrInvalidJSON, http.StatusBadRequest, CodeInvalidJSON},
		{"valid/missing_metric_type", fmt.Errorf("get metric: %w", handlers.ErrMetricTypeNotFound), http.StatusNotFound, CodeNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Registry.Resolve(tt.err)

			require.True(t, ok)
			require.Equal(t, tt.wantStatus, got.Status)
			require.Equal(t, tt.wantCode, got.Code)
		})
	}
}
