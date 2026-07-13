// Пакет содержит в себе регистрацию HTTP маршрутов для API и инициализацию HTTP сервера
package api

import (
	"net/http"

	"github.com/76Parker/metrico/internal/api/handlers"
	"github.com/76Parker/metrico/internal/api/middleware"
	"github.com/76Parker/metrico/internal/config"
	"github.com/76Parker/metrico/pkg/logger"
	"github.com/gin-gonic/gin"
)

func NewRouter(handler *handlers.MetricsHandler, httpCfg config.HTTP, log logger.Logger) *http.Server {
	router := registerHttpRoutes(handler, log)
	return newHttpServer(httpCfg, router)
}

func newHttpServer(cfg config.HTTP, router *gin.Engine) *http.Server {
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

func registerHttpRoutes(handler *handlers.MetricsHandler, log logger.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(middleware.WithLogging(log))
	router.Use(middleware.Compress())
	router.LoadHTMLGlob("templates/*")
	router.POST("/update/:metricType/:metricName/:metricValue", handler.UpdateMetric)
	router.POST("/value", handler.GetMetricByNameJSON)
	router.POST("/update", handler.UpdateMetricJSON)
	router.GET("/value/:metricType/:metricName", handler.GetMetricByName)
	router.GET("/", handler.GetAllMetrics)
	return router
}
