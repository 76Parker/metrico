package metrics

import (
	"context"

	"github.com/76Parker/metrico/internal/domain/metrics"
)

type core interface {
	update(ctx context.Context, cmd UpdateCommand) error
	batchUpdate(ctx context.Context, cmd []UpdateCommand) error
	getAll(ctx context.Context) ([]metrics.Metrics, error)
	getByName(ctx context.Context, name string) (metrics.Metrics, error)
}

type noopSnapshotter struct{}

func (n *noopSnapshotter) onMetricsChanged(ctx context.Context) error {
	return nil
}

type snapshotter interface {
	onMetricsChanged(ctx context.Context) error
}

// Application представляет собой изолированное приложение сервиса метрик,
// и предоставляет сценарии использования для HTTP API
type Application struct {
	metricsCore core
	snapshotter snapshotter
}

func NewApplication(metricsCore core, opts ...ApplicationOption) *Application {
	app := &Application{
		metricsCore: metricsCore,
		snapshotter: &noopSnapshotter{}, // По умолчанию без применения опций - snapshotter не используется
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

// Update сценарий обновления/создания метрики
func (s *Application) Update(ctx context.Context, cmd UpdateCommand) error {
	// обновляем метрику в физическом хранилище метрик
	err := s.metricsCore.update(ctx, cmd)
	if err != nil {
		return err
	}
	// уведомляем snapshotter об изменении метрик
	return s.snapshotter.onMetricsChanged(ctx)
}

// BatchUpdate сценарий обновления/создания метрик в BATCH-режиме
func (s *Application) BatchUpdate(ctx context.Context, cmd []UpdateCommand) error {
	// обновляем метрики в физическом хранилище метрик
	err := s.metricsCore.batchUpdate(ctx, cmd)
	if err != nil {
		return err
	}
	// уведомляем snapshotter об изменении метрик
	return s.snapshotter.onMetricsChanged(ctx)
}

// GetByName сценарий получения метрики по имени
func (s *Application) GetByName(ctx context.Context, name string) (metrics.Metrics, error) {
	// получаем метрику по имени из физического хранилища метрик
	return s.metricsCore.getByName(ctx, name)
}

// GetAll сценарий получения всех метрик
func (s *Application) GetAll(ctx context.Context) ([]metrics.Metrics, error) {
	// получаем все метрики из физического хранилища метрик
	return s.metricsCore.getAll(ctx)
}
