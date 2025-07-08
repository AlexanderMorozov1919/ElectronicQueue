package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// RequestLogger логирует все HTTP-запросы
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("[GIN] %s %s\n", c.Request.Method, c.Request.URL.Path)
		c.Next()
	}
}
