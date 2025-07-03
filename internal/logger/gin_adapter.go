package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GinLogger возвращает middleware для GIN
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if path == "" {
			path = "/"
		}

		instance.WithFields(logrus.Fields{
			"module": "gin",
			"method": c.Request.Method,
			"path":   path,
		}).Info("HTTP request")

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		entry := instance.WithFields(logrus.Fields{
			"module":  "gin",
			"status":  status,
			"latency": latency.String(),
			"ip":      c.ClientIP(),
		})

		if status >= 500 {
			entry.Error("HTTP error")
		} else if status >= 400 {
			entry.Warn("HTTP client error")
		} else {
			entry.Info("HTTP response")
		}
	}
}
