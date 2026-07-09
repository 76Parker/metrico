package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/76Parker/metrico/internal/app"
	"github.com/76Parker/metrico/internal/config"
)

const (
	defaultAddr = "localhost:8080"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	)
	var hostPort string
	defer stop()
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal("error config load:", err)
	}
	addrFromEnv := os.Getenv("ADDRESS")
	addrFromFlag := flag.String("a", defaultAddr, "HTTP listener address")
	flag.Parse()

	if addrFromEnv != "" {
		hostPort = addrFromEnv
	} else {
		hostPort = *addrFromFlag
	}

	if err := validateHostPort(hostPort); err != nil {
		log.Fatalf("invalid server address: %v", err)
	}

	cfg.HttpConfig.Address = hostPort
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

func validateHostPort(address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf(`address must have the format "host:port"`)
	}
	if host == "" {
		return fmt.Errorf("address cannot be empty")
	}
	return nil
}
