package middleware

import (
	"github.com/76Parker/metrico/internal/api/errmap"
	"github.com/gin-gonic/gin"
)

// ErrorHandler middleware для обработки ошибок которые произошли в handler'ах
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		err := c.Errors.Last()
		if err == nil || c.Writer.Written() {
			return
		}

		e, ok := errmap.Registry.Resolve(err)
		// Если не нашли ошибку в Registry, то используем UnexpectedError
		if !ok {
			e = errmap.UnexpectedError
		}
		c.AbortWithStatusJSON(e.Status, e)
	}
}
