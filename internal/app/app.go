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
	"github.com/76Parker/metrico/pkg/logger"
)

type LifecycleManager struct {
	closers []io.Closer
	cfg     config.Config
	server  *http.Server
	logger  logger.Logger
}

func NewLifecycleManager(cfg config.Config) (*LifecycleManager, error) {
	manager := &LifecycleManager{cfg: cfg}
	metricHandler := manager.createMetricHandler()

	lvl := "info"
	logger, err := logger.NewZapLogger(lvl)
	if err != nil {
		return nil, err
	}

	server := api.NewRouter(metricHandler, cfg.HttpConfig, logger)
	manager.server = server
	manager.logger = logger
	return manager, nil
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
	lm.logger.Info("Starting HTTP server", "address", lm.cfg.HttpConfig.Address)
	if err := lm.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (lm *LifecycleManager) Stop(ctx context.Context) error {
	lm.logger.Info("Shutting down HTTP server", "address", lm.cfg.HttpConfig.Address)

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
