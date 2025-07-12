package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireAPIKey проверяет наличие и правильность API-ключа в заголовке.
func RequireAPIKey(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		providedKey := c.GetHeader("X-API-KEY")
		if providedKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key is missing"})
			return
		}
		if providedKey != apiKey {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid API key"})
			return
		}
		c.Next()
	}
}
