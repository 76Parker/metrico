package filestorage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/76Parker/metrico/internal/adapters/filestorage"
	"github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage_Save(t *testing.T) {
	newTestStorage := func(t *testing.T) (*filestorage.Storage, string) {
		t.Helper()

		path := filepath.Join(t.TempDir(), "metrics.json")

		storage, err := filestorage.NewStorage(path)
		require.NoError(t, err)

		return storage, path
	}
	t.Run("valid/counter_with_delta", func(t *testing.T) {
		s, snapshotPath := newTestStorage(t)
		testData := []metrics.Metrics{{ID: "requests_total", Type: metrics.MetricTypeCounter, Delta: new(int64(120))}}
		require.NoError(t, s.Save(t.Context(), testData))

		actualSnapshot, err := os.ReadFile(snapshotPath)
		require.NoError(t, err)
		assert.JSONEq(t, `[{"id":"requests_total","type":"counter","delta":120}]`, string(actualSnapshot))
	})
	t.Run("valid/gauge_with_value", func(t *testing.T) {
		s, snapshotPath := newTestStorage(t)
		testData := []metrics.Metrics{{ID: "temperature", Type: metrics.MetricTypeGauge, Value: new(23.7)}}
		require.NoError(t, s.Save(t.Context(), testData))

		actualSnapshot, err := os.ReadFile(snapshotPath)
		require.NoError(t, err)
		assert.JSONEq(t, `[{"id":"temperature","type":"gauge","value":23.7}]`, string(actualSnapshot))
	})
	t.Run("invalid/gauge_without_value", func(t *testing.T) {
		s, snapshotPath := newTestStorage(t)
		testData := []metrics.Metrics{{ID: "temperature", Type: metrics.MetricTypeGauge}}
		err := s.Save(t.Context(), testData)
		var jsonschemaValidationError *jsonschema.ValidationError
		require.ErrorAs(t, err, &jsonschemaValidationError)

		_, statErr := os.Stat(snapshotPath)
		require.True(t, os.IsNotExist(statErr), "invalid Save must not create a snapshot")
	})
	t.Run("invalid/counter_without_delta", func(t *testing.T) {
		s, snapshotPath := newTestStorage(t)
		testData := []metrics.Metrics{{ID: "errors_total", Type: metrics.MetricTypeCounter}}
		err := s.Save(t.Context(), testData)
		var jsonschemaValidationError *jsonschema.ValidationError
		require.ErrorAs(t, err, &jsonschemaValidationError)

		_, statErr := os.Stat(snapshotPath)
		require.True(t, os.IsNotExist(statErr), "invalid Save must not create a snapshot")
	})
	t.Run("invalid/unknown_type", func(t *testing.T) {
		s, snapshotPath := newTestStorage(t)
		testData := []metrics.Metrics{{ID: "temperature", Type: metrics.MetricType("unknown"), Value: new(23.7)}}
		err := s.Save(t.Context(), testData)
		var jsonschemaValidationError *jsonschema.ValidationError
		require.ErrorAs(t, err, &jsonschemaValidationError)

		_, statErr := os.Stat(snapshotPath)
		require.True(t, os.IsNotExist(statErr), "invalid Save must not create a snapshot")
	})
	t.Run("invalid/empty_id", func(t *testing.T) {
		s, snapshotPath := newTestStorage(t)
		testData := []metrics.Metrics{{ID: "", Type: metrics.MetricTypeGauge, Value: new(23.7)}}
		err := s.Save(t.Context(), testData)
		var jsonschemaValidationError *jsonschema.ValidationError
		require.ErrorAs(t, err, &jsonschemaValidationError)

		_, statErr := os.Stat(snapshotPath)
		require.True(t, os.IsNotExist(statErr), "invalid Save must not create a snapshot")
	})
}

func TestStorage_Restore(t *testing.T) {
	newTestStorage := func(t *testing.T, fixtureBytes []byte) *filestorage.Storage {
		t.Helper()
		tempDir := t.TempDir()
		testFilePath := filepath.Join(tempDir, "metrics.json")
		require.NoError(t, os.WriteFile(testFilePath, fixtureBytes, 0644))
		s, err := filestorage.NewStorage(testFilePath)
		require.NoError(t, err)
		return s
	}
	t.Run("valid/gauge_with_value", func(t *testing.T) {
		s := newTestStorage(t, validGaugeWithValue)
		actualMetrics, err := s.Restore(t.Context())
		require.NoError(t, err)

		expectedMetrics := []metrics.Metrics{
			{ID: "test_gauge", Type: metrics.MetricTypeGauge, Value: new(10.5)},
		}
		assert.Equal(t, expectedMetrics, actualMetrics)
	})
	t.Run("valid/counter_with_delta", func(t *testing.T) {
		s := newTestStorage(t, validCounterWithDelta)
		actualMetrics, err := s.Restore(t.Context())
		require.NoError(t, err)

		expectedMetrics := []metrics.Metrics{
			{ID: "test_counter", Type: metrics.MetricTypeCounter, Delta: new(int64(10))},
		}
		assert.Equal(t, expectedMetrics, actualMetrics)
	})
	t.Run("valid/metrics_with_value_and_delta", func(t *testing.T) {
		s := newTestStorage(t, validMetricsWithValueAndDelta)
		actualMetrics, err := s.Restore(t.Context())
		require.NoError(t, err)

		expectedMetrics := []metrics.Metrics{
			{ID: "test_gauge", Type: metrics.MetricTypeGauge, Delta: new(int64(10)), Value: new(10.5)},
			{ID: "test_counter", Type: metrics.MetricTypeCounter, Delta: new(int64(10)), Value: new(10.0)},
		}
		assert.Equal(t, expectedMetrics, actualMetrics)
	})
	t.Run("invalid/gauge_without_value", func(t *testing.T) {
		s := newTestStorage(t, invalidGaugeWithoutValue)
		metrics, err := s.Restore(t.Context())
		require.Error(t, err)
		require.Nil(t, metrics)
	})
	t.Run("invalid/counter_without_delta", func(t *testing.T) {
		s := newTestStorage(t, invalidCounterWithoutDelta)
		metrics, err := s.Restore(t.Context())
		require.Error(t, err)
		require.Nil(t, metrics)
	})
	t.Run("invalid/empty_id", func(t *testing.T) {
		s := newTestStorage(t, invalidEmptyID)
		metrics, err := s.Restore(t.Context())
		require.Error(t, err)
		require.Nil(t, metrics)
	})
	t.Run("invalid/empty_type", func(t *testing.T) {
		s := newTestStorage(t, invalidEmptyType)
		metrics, err := s.Restore(t.Context())
		require.Error(t, err)
		require.Nil(t, metrics)
	})
	t.Run("invalid/unknown_type", func(t *testing.T) {
		s := newTestStorage(t, invalidUnknownType)
		metrics, err := s.Restore(t.Context())
		require.Error(t, err)
		require.Nil(t, metrics)
	})
	t.Run("invalid/invalid_json", func(t *testing.T) {
		s := newTestStorage(t, invalidJSON)
		metrics, err := s.Restore(t.Context())
		require.Error(t, err)
		require.Nil(t, metrics)
	})
}
