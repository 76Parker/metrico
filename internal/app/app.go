// `app/app.go`
// пакет для сборки и wiring'a реализаций с внутренними модулями
package app

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/76Parker/metrico/internal/adapters/memstorage"
	"github.com/76Parker/metrico/internal/api"
	"github.com/76Parker/metrico/internal/api/handlers"
	"github.com/76Parker/metrico/internal/config"
	"github.com/76Parker/metrico/internal/usecase/metrics"
)

type LifecycleManager struct {
	closers []io.Closer
	cfg     config.Config
	server  *http.Server
}

func NewLifecycleManager(cfg config.Config) *LifecycleManager {
	manager := &LifecycleManager{cfg: cfg}
	metricHandler := manager.createMetricHandler()

	server := api.NewRouter(metricHandler, cfg.HttpConfig)
	manager.server = server

	return manager
}

// createMetricHandler Создает HTTP-обработчик для взаимодействия с метриками
// Иницализируем все зависимости для обработчика с нижних слоев до верхнего
func (lm *LifecycleManager) createMetricHandler() *handlers.MetricsHandler {
	metricStorage := memstorage.NewMemStorage()
	metricService := metrics.NewService(metricStorage)
	metricHandler := handlers.NewMetricsHandler(metricService)
	return metricHandler
}

func (lm *LifecycleManager) Start() error {
	if err := lm.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (lm *LifecycleManager) Stop(ctx context.Context) error {

	var err error
	if lm.server != nil {
		err = lm.server.Shutdown(ctx)
	}
	for _, closer := range lm.closers {
		if closeErr := closer.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}
	return err
}
