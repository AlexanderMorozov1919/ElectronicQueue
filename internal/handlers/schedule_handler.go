package handlers

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ScheduleHandler struct {
	service *services.ScheduleService
}

func NewScheduleHandler(service *services.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{service: service}
}

// CreateSchedule godoc
// @Summary      Создать слот в расписании (Админ)
// @Description  Создает новый временной слот для врача. Требует INTERNAL_API_KEY.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request body models.CreateScheduleRequest true "Данные для создания слота"
// @Success      201 {object} models.Schedule "Успешно созданный слот"
// @Failure      400 {object} map[string]string "Ошибка: неверный формат запроса"
// @Failure      401 {object} map[string]string "Отсутствует ключ API"
// @Failure      403 {object} map[string]string "Неверный ключ API"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/admin/schedules [post]
func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	log := logger.Default()
	var req models.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Warn("CreateSchedule: Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса: " + err.Error()})
		return
	}

	schedule, err := h.service.CreateSchedule(&req)
	if err != nil {
		log.WithError(err).Error("CreateSchedule: Failed to create schedule in service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, schedule)
}

// DeleteSchedule godoc
// @Summary      Удалить слот из расписания (Админ)
// @Description  Удаляет временной слот из расписания по его ID. Требует INTERNAL_API_KEY.
// @Tags         admin
// @Produce      json
// @Param        id path int true "ID слота расписания"
// @Success      200 {object} map[string]string "Слот успешно удален"
// @Failure      400 {object} map[string]string "Ошибка: неверный ID"
// @Failure      401 {object} map[string]string "Отсутствует ключ API"
// @Failure      403 {object} map[string]string "Неверный ключ API"
// @Failure      404 {object} map[string]string "Слот не найден"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/admin/schedules/{id} [delete]
func (h *ScheduleHandler) DeleteSchedule(c *gin.Context) {
	log := logger.Default()
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		log.WithError(err).Warn("DeleteSchedule: Invalid ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID"})
		return
	}

	err = h.service.DeleteSchedule(uint(id))
	if err != nil {
		log.WithError(err).Error("DeleteSchedule: Failed to delete schedule from service")
		if err.Error() == "слот расписания с ID "+idStr+" не найден" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Слот расписания успешно удален"})
}
