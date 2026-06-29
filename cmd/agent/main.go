package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/76Parker/metrico/internal/agent/provider"
	"github.com/76Parker/metrico/internal/agent/reporter"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}
	serverAddr := "http://localhost:8080"
	provider := provider.NewMetricProvider(pollInterval)
	reporter := reporter.NewMetricReporter(serverAddr, httpClient, provider, reportInterval)
	if err := reporter.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
