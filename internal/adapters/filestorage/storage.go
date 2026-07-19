package filestorage

import (
	"bytes"
	"context"
	"os"
	"sync"

	"github.com/76Parker/metrico/internal/adapters/filestorage/snapshotschema"
	"github.com/76Parker/metrico/internal/domain/metrics"
	"github.com/76Parker/metrico/internal/usecase/snapshot"
	goccyjson "github.com/goccy/go-json"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

type Storage struct {
	fileStoragePath string
	schema          *jsonschema.Schema
	mu              sync.Mutex
}

func NewStorage(fileStoragePath, schemaPath string) (*Storage, error) {

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(schemaPath)
	if err != nil {
		return nil, err
	}

	return &Storage{
		fileStoragePath: fileStoragePath,
		schema:          schema,
		mu:              sync.Mutex{},
	}, nil
}

func (s *Storage) Save(ctx context.Context, metricsSlice []metrics.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	metricSnapshot := make(snapshotschema.MetricsSnapshotV1, 0, len(metricsSlice))
	for _, metric := range metricsSlice {
		m := snapshotschema.Metric{
			Id:   metric.ID,
			Type: snapshotschema.MetricType(metric.Type),
		}
		switch metric.Type {
		case metrics.Gauge:
			m.Value = metric.Value
		case metrics.Counter:
			m.Delta = metric.Delta
		}
		metricSnapshot = append(metricSnapshot, m)
	}
	data, err := goccyjson.MarshalIndent(metricSnapshot, "", "  ")
	if err != nil {
		return err
	}
	if err := s.eraseFileContent(); err != nil {
		if os.IsNotExist(err) {
			return snapshot.ErrStorageFileNotFound
		}
		return err
	}
	if err := s.writeSnapshot(data); err != nil {
		if os.IsNotExist(err) {
			return snapshot.ErrStorageFileNotFound
		}
		return err
	}
	return nil
}

func (s *Storage) Restore(ctx context.Context) ([]metrics.Metrics, error) {
	rawFile, err := os.ReadFile(s.fileStoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, snapshot.ErrStorageFileNotFound
		}
		return nil, err
	}

	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(rawFile))
	if err != nil {
		return nil, err
	}
	if err := s.schema.Validate(doc); err != nil {
		return nil, err
	}
	var snapshotFromFile snapshotschema.MetricsSnapshotV1
	if err := goccyjson.Unmarshal(rawFile, &snapshotFromFile); err != nil {
		return nil, err
	}
	snapshotLen := len(snapshotFromFile)

	metricsSlice := make([]metrics.Metrics, 0, snapshotLen)
	for _, metric := range snapshotFromFile {
		m := metrics.Metrics{
			ID:   metric.Id,
			Type: metrics.MetricType(metric.Type),
		}
		switch metric.Type {
		case snapshotschema.MetricTypeGauge:
			m.Value = metric.Value
		case snapshotschema.MetricTypeCounter:
			m.Delta = metric.Delta
		}
		metricsSlice = append(metricsSlice, m)
	}
	return metricsSlice, nil
}

func (s *Storage) eraseFileContent() error {
	if err := os.Truncate(s.fileStoragePath, 0); err != nil {
		return err
	}
	return nil
}

func (s *Storage) writeSnapshot(snapshotData []byte) error {
	if err := os.WriteFile(s.fileStoragePath, snapshotData, 0o644); err != nil {
		return err
	}
	return nil
}
