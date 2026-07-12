package middleware

import (
	"time"

	"github.com/76Parker/metrico/pkg/logger"
	"github.com/gin-gonic/gin"
)

func WithLogging(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()
		status := c.Writer.Status()
		err := c.Errors.Last()
		fields := []any{
			"method", c.Request.Method,
			"uri", c.Request.URL.RequestURI(),
			"status", status,
			"response_size", c.Writer.Size(),
			"duration", time.Since(startedAt),
		}

		if err != nil {
			if status < 500 && status >= 400 {
				log.Warn("request failed", append(fields, "error", err.Err)...)
			} else {
				log.Error("request not executed", append(fields, "error", err.Err)...)
			}
			return
		}

		log.Info("request completed", fields...)
	}
}
