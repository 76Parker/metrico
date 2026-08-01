package middleware

import "github.com/gin-gonic/gin"

const handlerNameKey = "handler_name"

const unknownHandlerName = "unknown"

func HandlerName(op string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(handlerNameKey, op)
		c.Next()
	}
}
