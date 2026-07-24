package filestorage_test

import (
	_ "embed"
)

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
