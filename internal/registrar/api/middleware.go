package api

import (
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func LoggingMiddleware(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Infof("HTTP Request: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	}
}
