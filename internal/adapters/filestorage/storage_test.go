package filestorage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/76Parker/metrico/internal/adapters/filestorage"
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
	// Positive-тест кейсы в `Save` должны сохранять метрики и создавать fileStorage-файл
	for _, tc := range validSaveTestCases {
		t.Run(tc.testName, func(t *testing.T) {
			s, snapshotPath := newTestStorage(t)
			require.NoError(t, s.Save(t.Context(), tc.inputMetricsToSave))

			actualSnapshot, err := os.ReadFile(snapshotPath)
			require.NoError(t, err)
			assert.JSONEq(t, string(tc.expectedSnapshotAfterSave), string(actualSnapshot))
		})
	}
	// Negative-тест кейсы в `Save` должны падать из-за валидации по JSON-схеме и не создавать fileStorage-файл
	for _, tc := range invalidSaveTestCases {
		t.Run(tc.testName, func(t *testing.T) {
			s, snapshotPath := newTestStorage(t)
			err := s.Save(t.Context(), tc.inputMetricsToSave)
			var jsonschemaValidationError *jsonschema.ValidationError
			require.ErrorAs(t, err, &jsonschemaValidationError)

			_, statErr := os.Stat(snapshotPath)
			require.True(t, os.IsNotExist(statErr), "invalid Save must not create a snapshot")
		})
	}
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
	// Positive-тест кейсы в `Restore` должны восстанавливать метрики из fileStorage-файла
	for _, tc := range validRestoreTestCases {
		t.Run(tc.testName, func(t *testing.T) {
			s := newTestStorage(t, tc.inputRestoreFile)
			actualMetrics, err := s.Restore(t.Context())
			require.NoError(t, err)
			assert.Equal(t, tc.expectedMetricsAfterRestore, actualMetrics)
		})
	}
	// Negative-тест кейсы в `Restore` должны падать из-за валидации JSON
	for _, tc := range invalidRestoreTestCases {
		t.Run(tc.testName, func(t *testing.T) {
			s := newTestStorage(t, tc.inputRestoreFile)
			metrics, err := s.Restore(t.Context())
			require.Error(t, err)
			require.Nil(t, metrics)
		})
	}
}
