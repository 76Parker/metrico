// Пакет содержит в себе регистрацию HTTP маршрутов для API и инициализацию HTTP сервера
package api

import (
	"net/http"

	"github.com/76Parker/metrico/internal/api/handlers"
	"github.com/76Parker/metrico/internal/config"
	"github.com/gin-gonic/gin"
)

func NewRouter(handler *handlers.MetricsHandler, httpCfg config.HTTP) *http.Server {
	router := registerHttpRoutes(handler)
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

func registerHttpRoutes(handler *handlers.MetricsHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.LoadHTMLGlob("templates/*")
	router.POST("/update/:metricType/:metricName/:metricValue", handler.UpdateMetric)
	router.GET("/update/:metricType/:metricName", handler.GetMetricByName)
	router.GET("/", handler.GetAllMetrics)
	return router
}
