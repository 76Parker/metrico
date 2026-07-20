clean:
	rm -f coverage.out

SUPRESS_OUTPUT = > /dev/null

METRICS_SCHEMA = internal/adapters/filestorage/snapshotschema/metrics-snapshot-v1.schema.json
METRICS_PACKAGE = internal/adapters/filestorage/snapshotschema
METRICS_GEN_OUTPUT_FILE = internal/adapters/filestorage/snapshotschema/metrics-snapshot-v1.gen.go

test-coverage:
	@go test -coverprofile=coverage.out ./... $(SUPRESS_OUTPUT)
	@go tool cover -html=coverage.out $(SUPRESS_OUTPUT)
	@make clean

test-total-coverage:
	@go test ./... -coverprofile=coverage.out $(SUPRESS_OUTPUT)
	@go tool cover -func=coverage.out | grep total | awk '{print $$1, $$3}'
	@make clean $(SUPRESS_OUTPUT)

metrics_schema_codegen:
	go-jsonschema -p $(METRICS_PACKAGE) $(METRICS_SCHEMA) > $(METRICS_GEN_OUTPUT_FILE)
