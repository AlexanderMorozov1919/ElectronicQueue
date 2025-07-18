package logger

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GinLogger возвращает middleware для GIN
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log := Default().WithFields(map[string]interface{}{
			"module":  "GIN",
			"status":  status,
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
			"ip":      c.ClientIP(),
			"latency": latency,
		})

		var logMessage string
		if len(c.Errors) > 0 {
			// Объединяем все ошибки Gin в одну строку
			var errorStrings []string
			for _, e := range c.Errors {
				errorStrings = append(errorStrings, e.Error())
			}
			log = log.WithError(c.Errors.Last().Err)
			logMessage = strings.Join(errorStrings, " | ")
		} else {
			logMessage = "Request handled"
		}

		switch {
		case status >= 500:
			log.Error(logMessage)
		case status >= 400:
			log.Warn(logMessage)
		default:
			log.Info(logMessage)
		}
	}
}
