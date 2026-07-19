package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/76Parker/metrico/internal/app"
	"github.com/76Parker/metrico/internal/config"
	"github.com/caarlos0/env/v11"
)

// Default values for flags
const (
	defaultAddr            = "localhost:8080"
	defaultFileStoragePath = "/var/lib/metrico/metrics.json"
	defaultStoreInterval   = 300
	defaultRestore         = false
)

// Application configuration flags
var (
	hostPort        string // -a (or ADDRESS env var)
	fileStoragePath string // -f (or FILE_STORAGE_PATH env var)
	storeInterval   int    // -i (or STORE_INTERVAL env var)
	restore         bool   // -r (or RESTORE env var)
)

type envConfig struct {
	Addr            string `env:"ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	Restore         bool   `env:"RESTORE"`
}

func (c *envConfig) applyOverrides() {
	if _, ok := os.LookupEnv("ADDRESS"); ok {
		hostPort = c.Addr
	}
	if _, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		fileStoragePath = c.FileStoragePath
	}
	if _, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		storeInterval = c.StoreInterval
	}
	if _, ok := os.LookupEnv("RESTORE"); ok {
		restore = c.Restore
	}
}

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
	// Flag parsing
	flag.StringVar(&hostPort, "a", defaultAddr, "HTTP listener address")               // -a
	flag.StringVar(&fileStoragePath, "f", defaultFileStoragePath, "file storage path") // -f
	flag.IntVar(&storeInterval, "i", defaultStoreInterval, "store interval")           // -i
	flag.BoolVar(&restore, "r", defaultRestore, "restore from file")                   // -r

	flag.Parse()
	configFromEnv, err := env.ParseAs[envConfig]()
	if err != nil {
		log.Fatal("error parsing env:", err)
	}
	configFromEnv.applyOverrides()

	if err := validateHostPort(hostPort); err != nil {
		log.Fatalf("invalid server address: %v", err)
	}

	cfg.HttpConfig.Address = hostPort
	cfg.SnapshotServiceConfig = config.SnapshotService{
		StoreInterval:   time.Duration(storeInterval) * time.Second,
		FileStoragePath: fileStoragePath,
		Restore:         restore,
		SchemaPath:      "/Users/parkersec/go-projects/go-musthave-metrics-tpl/internal/adapters/filestorage/snapshotschema/metrics-snapshot-v1.schema.json",
	}
	appManager, err := app.NewLifecycleManager(ctx, *cfg)
	if err != nil {
		log.Fatal("error creating app manager:", err)
	}

	errCh := make(chan error, 1)

	go func() {
		if err := appManager.Start(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		// log.Println("shutdown signal received")

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
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf(`address must have the format "host:port"`)
	}
	if port == "" {
		return fmt.Errorf("port cannot be empty")
	}
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("port must be a number")
	}
	if host == "" {
		return fmt.Errorf("address cannot be empty")
	}
	return nil
}
