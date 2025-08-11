package middleware

import (
	"ElectronicQueue/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CheckBusinessProcess — это middleware, которое проверяет, включен ли хотя бы один из требуемых бизнес-процессов.
// Это позволяет обрабатывать эндпоинты, используемые несколькими сервисами.
func CheckBusinessProcess(processService *services.BusinessProcessService, requiredProcesses ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, processName := range requiredProcesses {
			if processService.IsEnabled(processName) {
				// Если хотя бы один из требуемых процессов включен, разрешаем доступ.
				c.Next()
				return
			}
		}

		// Если ни один из требуемых процессов не включен, возвращаем ошибку.
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "This service is temporarily disabled by the administrator."})
	}
}
