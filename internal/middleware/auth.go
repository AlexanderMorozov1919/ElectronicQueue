package middleware

import (
	"ElectronicQueue/internal/utils"
	"net/http"
	"strings"

	"ElectronicQueue/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RequireRole middleware проверяет JWT токен и роль пользователя
func RequireRole(jwtManager *utils.JWTManager, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.Default().WithField("middleware", "RequireRole")

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("Отсутствует заголовок авторизации")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "отсутствует токен"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Warn("Неверный формат заголовка авторизации")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "неверный формат токена"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtManager.ValidateJWT(tokenString)
		if err != nil {
			log.WithError(err).Warn("Ошибка валидации JWT токена")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "неверный токен"})
			c.Abort()
			return
		}

		if claims.Role != requiredRole {
			log.WithFields(logrus.Fields{
				"required_role": requiredRole,
				"actual_role":   claims.Role,
			}).Warn("Несоответствие роли")
			c.JSON(http.StatusForbidden, gin.H{"error": "недостаточно прав"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}
