// `metrics/service.go`
// сервисный слой для работы с метриками.
// Описывает пользовательские сценарии взаимодействия с метриками.
package metrics

import (
	"context"
	"strconv"

	"github.com/76Parker/metrico/internal/domain/metrics"
)

// Repository интерфейс для работы c физическим хранилищем метрик
type Repository interface {
	// Apply применяет список изменений к хранилищу метрик (создает либо обновляет),
	// является единой логикой обновления для single-update и batch-update.
	Apply(ctx context.Context, mutations []Change) error
	Get(ctx context.Context, metricName string) (metrics.Metrics, error)
	GetAll(ctx context.Context) ([]metrics.Metrics, error)
}

// CoreService представляет собой сервисный слой для работы с метриками,
// он работает с физическим хранилищем метрик и не допускает невалидные операции (см. buildChangeOperation)
type CoreService struct {
	storage Repository
}

// NewCoreService Конструктор для CoreService
func NewCoreService(storage Repository) *CoreService {
	return &CoreService{
		storage: storage,
	}
}

// update обновляет/создает метрику в физическом хранилище
func (s *CoreService) update(ctx context.Context, cmd UpdateCommand) error {
	mutation, err := buildChangeOperation(cmd)
	if err != nil {
		return err
	}
	if err := s.storage.Apply(ctx, []Change{mutation}); err != nil {
		return err
	}
	return nil
}

// batchUpdate обновляет/создает метрики в физическом хранилище через BATCH
func (s *CoreService) batchUpdate(ctx context.Context, cmd []UpdateCommand) error {
	changes := make([]Change, 0, len(cmd))
	for _, c := range cmd {
		mutation, err := buildChangeOperation(c)
		if err != nil {
			return err
		}
		changes = append(changes, mutation)
	}
	if err := s.storage.Apply(ctx, changes); err != nil {
		return err
	}
	return nil
}

// getByName возвращает метрику по имени из физическом хранилища
func (s *CoreService) getByName(ctx context.Context, name string) (metrics.Metrics, error) {
	metric, err := s.storage.Get(ctx, name)
	if err != nil {
		return metrics.Metrics{}, err
	}
	return metric, nil
}

// getAll возвращает все метрики из физического хранилища
func (s *CoreService) getAll(ctx context.Context) ([]metrics.Metrics, error) {
	return s.storage.GetAll(ctx)
}

// buildChangeOperation строит операцию изменения для физического хранилища на основе входящей команды
// и не допускает невалидные операции и параметры
func buildChangeOperation(cmd UpdateCommand) (Change, error) {
	if cmd.Name == "" {
		return Change{}, metrics.ErrMetricNameIsEmpty
	}
	if cmd.MetricType == "" {
		return Change{}, metrics.ErrMetricTypeIsEmpty
	}
	switch metrics.MetricType(cmd.MetricType) {
	case metrics.MetricTypeGauge:
		value, err := strconv.ParseFloat(cmd.Value, 64)
		if err != nil {
			return Change{}, metrics.ErrInvalidValueForGauge
		}
		return Change{
			name:  cmd.Name,
			kind:  ChangeKindSetGauge,
			value: value,
		}, nil
	case metrics.MetricTypeCounter:
		delta, err := strconv.ParseInt(cmd.Value, 10, 64)
		if err != nil {
			return Change{}, metrics.ErrInvalidValueForCounter
		}
		return Change{
			name:  cmd.Name,
			kind:  ChangeKindAddCounter,
			delta: delta,
		}, nil
	default:
		return Change{}, metrics.ErrInvalidMetricType
	}
}
