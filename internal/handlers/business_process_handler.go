package handlers

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BusinessProcessHandler struct {
	service *services.BusinessProcessService
}

func NewBusinessProcessHandler(service *services.BusinessProcessService) *BusinessProcessHandler {
	return &BusinessProcessHandler{service: service}
}

type UpdateProcessRequest struct {
	IsEnabled bool `json:"is_enabled"`
}

// GetAllProcesses godoc
// @Summary      Получить статусы всех бизнес-процессов (Админ)
// @Description  Возвращает список всех бизнес-процессов и их текущее состояние (включен/отключен).
// @Tags         admin
// @Produce      json
// @Success      200 {array} models.BusinessProcess "Список процессов"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/admin/processes [get]
func (h *BusinessProcessHandler) GetAllProcesses(c *gin.Context) {
	processes, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get processes status"})
		return
	}
	c.JSON(http.StatusOK, processes)
}

// UpdateProcess godoc
// @Summary      Обновить статус бизнес-процесса (Админ)
// @Description  Включает или отключает указанный бизнес-процесс.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        name path string true "Имя процесса"
// @Param        request body UpdateProcessRequest true "Новое состояние"
// @Success      200 {object} models.BusinessProcess "Обновленный процесс"
// @Failure      400 {object} map[string]string "Ошибка в запросе"
// @Failure      404 {object} map[string]string "Процесс не найден"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/admin/processes/{name} [patch]
func (h *BusinessProcessHandler) UpdateProcess(c *gin.Context) {
	log := logger.Default()
	processName := c.Param("name")

	var req UpdateProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Warn("UpdateProcess: Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	process, err := h.service.UpdateStatus(processName, req.IsEnabled)
	if err != nil {
		log.WithError(err).Error("UpdateProcess: service returned an error")
		if err.Error() == "process '"+processName+"' not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, process)
}

// GetProcessStatusByName godoc
// @Summary      Получить статус конкретного бизнес-процесса
// @Description  Возвращает текущее состояние (включен/отключен) для указанного бизнес-процесса.
// @Tags         processes
// @Produce      json
// @Param        name path string true "Имя процесса"
// @Success      200 {object} map[string]bool "Статус процесса"
// @Router       /api/processes/{name} [get]
func (h *BusinessProcessHandler) GetProcessStatusByName(c *gin.Context) {
	processName := c.Param("name")
	isEnabled := h.service.IsEnabled(processName)
	c.JSON(http.StatusOK, gin.H{
		"process_name": processName,
		"is_enabled":   isEnabled,
	})
}
