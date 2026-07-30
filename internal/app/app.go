package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/76Parker/metrico/internal/config"
	"github.com/76Parker/metrico/pkg/logger"
)

// Run создает и запускает приложение (блокирующая операция)
func Run(ctx context.Context, cfg config.Config, log logger.Logger) (err error) {
	application, err := newApplication(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("build application: %w", err)
	}

	defer func() {
		err = errors.Join(err, application.close())
	}()

	if err := application.run(ctx); err != nil {
		return fmt.Errorf("run application: %w", err)
	}

	return nil
}
