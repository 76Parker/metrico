package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/76Parker/metrico/internal/agent/provider"
	"github.com/76Parker/metrico/internal/agent/reporter"
	"github.com/caarlos0/env/v11"
)

const (
	defaultPollInterval   = 2 * time.Second
	defaultReportInterval = 10 * time.Second
	defaultAddr           = "http://localhost:8080"
)

var (
	pollInterval   time.Duration
	reportInterval time.Duration
	addr           string
)

type envConfig struct {
	ServerAddress  string        `env:"ADDRESS"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
}

func (e *envConfig) redefineConfigFromEnv() {
	if _, ok := os.LookupEnv("ADDRESS"); ok {
		addr = e.ServerAddress
	}
	if _, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		reportInterval = e.ReportInterval
	}
	if _, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		pollInterval = e.PollInterval
	}
}

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var pollSeconds, reportSeconds int
	flag.StringVar(&addr, "a", defaultAddr, "Listener address")
	flag.IntVar(&pollSeconds, "p", int(defaultPollInterval/time.Second), "Poll interval for metric provider in seconds")
	flag.IntVar(&reportSeconds, "r", int(defaultReportInterval/time.Second), "Report interval for metric reporter in seconds")
	flag.Parse()

	pollInterval = time.Duration(pollSeconds) * time.Second
	reportInterval = time.Duration(reportSeconds) * time.Second

	envCfg, err := applyEnvOverrides()
	if err != nil {
		log.Fatalf("parse config from env error: %v", err)
	}
	envCfg.redefineConfigFromEnv()

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	provider := provider.NewMetricProvider(pollInterval)
	reporter := reporter.NewMetricReporter(addr, httpClient, provider, reportInterval)
	if err := reporter.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}

func applyEnvOverrides() (envConfig, error) {
	return env.ParseAs[envConfig]()
}
