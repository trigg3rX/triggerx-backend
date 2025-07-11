package api

import (
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

func LoggerMiddleware(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("logger", logger)
		c.Next()
	}
}