package filestorage_test

import (
	"os"
	"sort"
	"testing"

	"github.com/76Parker/metrico/internal/adapters/filestorage"
	"github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/stretchr/testify/assert"
)

func TestStorage_Save(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		metrics []metrics.Metrics
		wantErr bool
	}{
		{
			name: "valid metrics",
			metrics: []metrics.Metrics{
				{ID: "test", Type: metrics.Counter, Delta: new(int64(100))},
				{ID: "test2", Type: metrics.Gauge, Value: new(100.00)},
				{ID: "test3", Type: metrics.Counter, Delta: new(int64(100)), Value: new(float64(100.00))},
			},
			wantErr: false,
		},
	}

	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "metrics.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	fileStoragePath := tmpFile.Name()
	schemaPath := "/Users/parkersec/go-projects/go-musthave-metrics-tpl/internal/adapters/filestorage/snapshotschema/metrics-snapshot-v1.schema.json"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := filestorage.NewStorage(fileStoragePath, schemaPath)
			if err != nil {
				t.Fatalf("NewStorage() failed: %v", err)
			}
			gotErr := s.Save(t.Context(), tt.metrics)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Save() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Save() succeeded unexpectedly")
			}
		})
	}
}

func TestStorage_Restore(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		fileStoragePath string
		want            []metrics.Metrics
		wantErr         bool
	}{
		{
			name:            "valid metrics",
			fileStoragePath: "/Users/parkersec/go-projects/go-musthave-metrics-tpl/internal/adapters/filestorage/tests/valid_metrics.json",
			want: []metrics.Metrics{
				{ID: "requests_total", Type: "counter", Delta: new(int64(120))},
				{ID: "errors_total", Type: "counter", Delta: new(int64(7))},
				{ID: "temperature", Type: "gauge", Value: new(23.7)},
				{ID: "memory_usage", Type: "gauge", Value: new(68.4)},
			},
			wantErr: false,
		},
		{
			name:            "invalid metrics",
			fileStoragePath: "/Users/parkersec/go-projects/go-musthave-metrics-tpl/internal/adapters/filestorage/tests/invalid_metrics.json",
			want:            nil,
			wantErr:         true,
		},
	}
	schemaPath := "/Users/parkersec/go-projects/go-musthave-metrics-tpl/internal/adapters/filestorage/snapshotschema/metrics-snapshot-v1.schema.json"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := filestorage.NewStorage(tt.fileStoragePath, schemaPath)
			if err != nil {
				t.Fatalf("NewStorage() failed: %v", err)
			}
			got, gotErr := s.Restore(t.Context())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Restore() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Restore() succeeded unexpectedly")
			}
			assert.Equal(t, len(tt.want), len(got), "Restore object length mismatch")
			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].ID < tt.want[j].ID
			})
			sort.Slice(got, func(i, j int) bool {
				return got[i].ID < got[j].ID
			})
			for i, metric := range tt.want {
				assert.Equal(t, metric.ID, got[i].ID, "Restore object ID mismatch")
				assert.Equal(t, metric.Type, got[i].Type, "Restore object Type mismatch")
				switch metric.Type {
				case "counter":
					assert.Equal(t, *metric.Delta, *got[i].Delta, "Restore object Delta mismatch")
				case "gauge":
					assert.Equal(t, *metric.Value, *got[i].Value, "Restore object Value mismatch")
				}
			}
		})
	}
}
