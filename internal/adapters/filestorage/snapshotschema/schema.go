package snapshotschema

import _ "embed"

// MetricsSnapshotV1SchemaJSON contains the JSON Schema embedded into the binary.
//
//go:embed metrics-snapshot-v1.schema.json
var MetricsSnapshotV1SchemaJSON []byte
