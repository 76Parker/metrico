package filestorage_test

import (
	_ "embed"

	"github.com/76Parker/metrico/internal/domain/metrics"
)

/*
 * Fixtures and test cases for method `Restore`
 */

type invalidRestoreTestCase struct {
	testName         string
	inputRestoreFile []byte
}

type validRestoreTestCase struct {
	testName                    string
	inputRestoreFile            []byte
	expectedMetricsAfterRestore []metrics.Metrics
}

var (
	//go:embed testdata/valid-gauge-with-value.json
	validGaugeWithValue []byte

	//go:embed testdata/valid-counter-with-delta.json
	validCounterWithDelta []byte

	//go:embed testdata/valid-metrics-with-value-and-delta.json
	validMetricsWithValueAndDelta []byte

	//go:embed testdata/invalid-gauge-without-value.json
	invalidGaugeWithoutValue []byte

	//go:embed testdata/invalid-counter-without-delta.json
	invalidCounterWithoutDelta []byte

	//go:embed testdata/invalid-empty-id.json
	invalidEmptyID []byte

	//go:embed testdata/invalid-empty-type.json
	invalidEmptyType []byte

	//go:embed testdata/invalid-unknown-type.json
	invalidUnknownType []byte

	//go:embed testdata/invalid-json.json
	invalidJSON []byte
)

var invalidRestoreTestCases = []invalidRestoreTestCase{
	{
		testName:         "invalid/gauge_without_value",
		inputRestoreFile: invalidGaugeWithoutValue,
	},
	{
		testName:         "invalid/counter_without_delta",
		inputRestoreFile: invalidCounterWithoutDelta,
	},
	{
		testName:         "invalid/empty_id",
		inputRestoreFile: invalidEmptyID,
	},
	{
		testName:         "invalid/empty_type",
		inputRestoreFile: invalidEmptyType,
	},
	{
		testName:         "invalid/unknown_type",
		inputRestoreFile: invalidUnknownType,
	},
	{
		testName:         "invalid/invalid_json",
		inputRestoreFile: invalidJSON,
	},
}

var validRestoreTestCases = []validRestoreTestCase{
	{
		testName:         "valid/gauge_with_value",
		inputRestoreFile: validGaugeWithValue,
		expectedMetricsAfterRestore: []metrics.Metrics{
			{ID: "test_gauge", Type: metrics.Gauge, Value: new(10.5)},
		},
	},
	{
		testName:         "valid/counter_with_delta",
		inputRestoreFile: validCounterWithDelta,
		expectedMetricsAfterRestore: []metrics.Metrics{
			{ID: "test_counter", Type: metrics.Counter, Delta: new(int64(10))},
		},
	},
	{
		testName:         "valid/metrics_with_value_and_delta",
		inputRestoreFile: validMetricsWithValueAndDelta,
		expectedMetricsAfterRestore: []metrics.Metrics{
			{ID: "test_gauge", Type: metrics.Gauge, Delta: new(int64(10)), Value: new(10.5)},
			{ID: "test_counter", Type: metrics.Counter, Delta: new(int64(10)), Value: new(10.0)},
		},
	},
}

/*
 * Fixtures and test cases for method `Save`
 */

type invalidSaveTestCase struct {
	testName           string
	inputMetricsToSave []metrics.Metrics
}

type validSaveTestCase struct {
	testName                  string
	inputMetricsToSave        []metrics.Metrics
	expectedSnapshotAfterSave []byte
}

var invalidSaveTestCases = []invalidSaveTestCase{
	{
		testName: "invalid/gauge_without_value",
		inputMetricsToSave: []metrics.Metrics{
			{ID: "temperature", Type: metrics.Gauge},
		},
	},
	{
		testName: "invalid/counter_without_delta",
		inputMetricsToSave: []metrics.Metrics{
			{ID: "errors_total", Type: metrics.Counter},
		},
	},
	{
		testName: "invalid/unknown_type",
		inputMetricsToSave: []metrics.Metrics{
			{ID: "temperature", Type: metrics.MetricType("unknown"), Value: new(23.7)},
		},
	},
	{
		testName: "invalid/empty_id",
		inputMetricsToSave: []metrics.Metrics{
			{ID: "", Type: metrics.Gauge, Value: new(23.7)},
		},
	},
}

var validSaveTestCases = []validSaveTestCase{
	{
		testName: "valid/counter_with_delta",
		inputMetricsToSave: []metrics.Metrics{
			{ID: "requests_total", Type: metrics.Counter, Delta: new(int64(120))},
		},
		expectedSnapshotAfterSave: []byte(`[
			{"id":"requests_total","type":"counter","delta":120}
		]`),
	},
	{
		testName: "valid/gauge_with_value",
		inputMetricsToSave: []metrics.Metrics{
			{ID: "temperature", Type: metrics.Gauge, Value: new(23.7)},
		},
		expectedSnapshotAfterSave: []byte(`[
			{"id":"temperature","type":"gauge","value":23.7}
		]`),
	},
}
