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

func NewRouter(
	metricsHandler *handlers.MetricsHandler,
	healthHandler *handlers.HealthHandler,
	httpCfg config.HTTP,
	log logger.Logger,
) *http.Server {
	router := registerHttpRoutes(metricsHandler, healthHandler, log)
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

func registerHttpRoutes(
	metricsHandler *handlers.MetricsHandler,
	healthHandler *handlers.HealthHandler,
	log logger.Logger,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	registerMetricsRoutes(router, metricsHandler, log)
	registerHealthRoutes(router, healthHandler)
	return router
}

func registerMetricsRoutes(router *gin.Engine, handler *handlers.MetricsHandler, log logger.Logger) {
	logMW := middleware.WithLogging(log)
	compressMW := middleware.Compress()
	router.LoadHTMLGlob("templates/*")
	router.POST("/update/:metricType/:metricName/:metricValue", logMW, compressMW, handler.Update)
	router.POST("/value", logMW, compressMW, handler.GetFromJSON)
	router.POST("/update", logMW, compressMW, handler.UpdateFromJSON)
	router.POST("/updates", logMW, compressMW, handler.BatchUpdateFromJSON)
	router.GET("/value/:metricType/:metricName", logMW, compressMW, handler.GetByName)
	router.GET("/", logMW, compressMW, handler.GetAll)
}

func registerHealthRoutes(router *gin.Engine, handler *handlers.HealthHandler) {
	router.GET("/ping", handler.CheckAvailability)
}
