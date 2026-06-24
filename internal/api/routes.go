// Пакет содержит в себе регистрацию HTTP маршрутов для API и инициализацию HTTP сервера
package api

import (
	"net/http"

	"github.com/76Parker/metrico/internal/api/handlers"
	"github.com/76Parker/metrico/internal/config"
)

func NewRouter(handler *handlers.MetricsHandler, httpCfg config.HTTP) *http.Server {
	router := registerHttpRoutes(handler)
	return newHttpServer(httpCfg, router)
}

func newHttpServer(cfg config.HTTP, router *http.ServeMux) *http.Server {
	return &http.Server{
		Addr:              cfg.Address,
		Handler:           router,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		IdleTimeout:       cfg.IdleTimeout,
	}
}

func registerHttpRoutes(service *handlers.MetricsHandler) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("POST /update/{metricType}/{metricName}/{metricValue}", service.UpdateMetric)
	return router
}
