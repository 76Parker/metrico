package snapshot

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/76Parker/metrico/internal/domain/metrics"
)

type updateMode int

const (
	asyncUpdate updateMode = iota // asynchronous update mode
	syncUpdate                    // synchronous update mode
)

type Service struct {
	metricsStore  metricsStore
	snapshotStore snapshotStore
	storeInterval time.Duration
	updateMode    updateMode
	restoreOnce   sync.Once
	runOnce       sync.Once

	cancellation context.CancelFunc

	isRunning bool
}

func NewService(metricsStore metricsStore, snapshotStore snapshotStore, storeInterval time.Duration) *Service {
	// На уровне создания сервиса определяем режим обновления
	var mode updateMode
	switch storeInterval {
	case 0:
		mode = syncUpdate
	default:
		mode = asyncUpdate
	}
	return &Service{
		metricsStore:  metricsStore,
		snapshotStore: snapshotStore,
		storeInterval: storeInterval,
		updateMode:    mode,
		restoreOnce:   sync.Once{},
		runOnce:       sync.Once{},
		isRunning:     false,
	}
}

// Restore restores metrics from the snapshot store into the metrics store with once-semantic.
func (s *Service) Restore(ctx context.Context) error {
	var restoreErr error
	s.restoreOnce.Do(func() {
		metrics, err := s.snapshotStore.Restore(ctx)
		if err != nil {
			restoreErr = err
			return
		}
		if err := s.metricsStore.Load(ctx, metrics); err != nil {
			restoreErr = err
			return
		}
	})
	return restoreErr
}

// Run runs the snapshot service, starting the update loop based on the update mode.
func (s *Service) Run(ctx context.Context) {
	s.runOnce.Do(func() {
		ctxWithCancel, cancel := context.WithCancel(ctx)
		s.cancellation = cancel
		if s.updateMode == asyncUpdate {
			go s.asyncUpdate(ctxWithCancel)
		}
		s.isRunning = true
	})
}

func (s *Service) Stop() error {
	if !s.isRunning {
		return errors.New("Stop(): snapshot service is not running")
	}
	s.cancellation()
	return nil
}

func (s *Service) UpdateSnapshot(ctx context.Context) error {
	// При асинхронном режиме ничего не делаем (так как обновление происходит асинхронно)
	if s.updateMode == asyncUpdate {
		return nil
	}
	metrics, err := s.metricsStore.GetAll(ctx)
	if err != nil {
		return err
	}
	// Синхронное обновление
	return s.syncUpdate(ctx, metrics)
}

func (s *Service) syncUpdate(ctx context.Context, metrics []metrics.Metrics) error {
	if err := s.snapshotStore.Save(ctx, metrics); err != nil {
		return err
	}
	return nil
}

func (s *Service) asyncUpdate(ctx context.Context) {
	ticker := time.NewTicker(s.storeInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			metrics, err := s.metricsStore.GetAll(ctx)
			if err != nil {
				return
			}
			if err := s.snapshotStore.Save(ctx, metrics); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
