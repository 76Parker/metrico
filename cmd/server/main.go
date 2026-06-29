package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/76Parker/metrico/internal/app"
	"github.com/76Parker/metrico/internal/config"
)

const (
	defaultAddr = "http://localhost:8080"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	)
	defer stop()
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal("error config load:", err)
	}

	addr := flag.String("a", defaultAddr, "Listener address")
	flag.Parse()
	if *addr != "" && addr != nil {
		if !strings.HasPrefix(*addr, "http://") {
			*addr = "http://" + *addr
		}
		cfg.HttpConfig.Address = *addr
	} else {
		cfg.HttpConfig.Address = defaultAddr
	}

	appManager := app.NewLifecycleManager(*cfg)

	errCh := make(chan error, 1)

	go func() {
		if err := appManager.Start(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("shutdown signal received")

	case err := <-errCh:
		log.Fatal("error app start:", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := appManager.Stop(shutdownCtx); err != nil {
		log.Fatal("error app shutdown:", err)
	}
}
