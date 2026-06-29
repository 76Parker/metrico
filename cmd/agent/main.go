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
	pollSeconds    int
	reportSeconds  int
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}
	if addr != "" {
		if !strings.Contains(addr, "http://") {
			addr = "http://" + addr
		}
	} else {
		addr = defaultAddr
	}

	provider := provider.NewMetricProvider(pollInterval)
	reporter := reporter.NewMetricReporter(addr, httpClient, provider, reportInterval)
	if err := reporter.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}

func init() {
	flag.StringVar(&addr, "a", defaultAddr, "Listener address")
	flag.IntVar(&pollSeconds, "p", int(defaultPollInterval/time.Second), "Poll interval for metric provider in seconds")
	flag.IntVar(&reportSeconds, "r", int(defaultReportInterval/time.Second), "Report interval for metric reporter in seconds")
	flag.Parse()

	pollInterval = time.Duration(pollSeconds) * time.Second
	reportInterval = time.Duration(reportSeconds) * time.Second
}
