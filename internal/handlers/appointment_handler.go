package handlers

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AppointmentHandler обрабатывает HTTP-запросы для записей на прием.
type AppointmentHandler struct {
	service *services.AppointmentService
}

// NewAppointmentHandler создает новый экземпляр AppointmentHandler.
func NewAppointmentHandler(service *services.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{service: service}
}

// GetDoctorSchedule godoc
// @Summary      Получить расписание врача с информацией о записях
// @Description  Возвращает все временные слоты врача на указанную дату, включая информацию о том, кто записан в занятые слоты.
// @Tags         registrar
// @Produce      json
// @Param        doctor_id path int true "ID Врача"
// @Param        date query string true "Дата в формате YYYY-MM-DD"
// @Success      200 {array} models.ScheduleWithAppointmentInfo "Массив слотов расписания с информацией о записях"
// @Failure      400 {object} map[string]string "Ошибка: неверный ID или формат даты"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/registrar/schedules/doctor/{doctor_id} [get]
func (h *AppointmentHandler) GetDoctorSchedule(c *gin.Context) {
	log := logger.Default()

	doctorIDStr := c.Param("doctor_id")
	doctorID, err := strconv.ParseUint(doctorIDStr, 10, 64)
	if err != nil {
		log.WithError(err).Warn("GetDoctorSchedule: Invalid doctor ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID врача"})
		return
	}

	dateStr := c.Query("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.WithError(err).Warn("GetDoctorSchedule: Invalid date format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты, используйте YYYY-MM-DD"})
		return
	}

	schedule, err := h.service.GetDoctorScheduleWithAppointments(uint(doctorID), date)
	if err != nil {
		log.WithError(err).Error("GetDoctorSchedule: Failed to get doctor schedule from service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить расписание врача"})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// CreateAppointment godoc
// @Summary      Создать новую запись на прием
// @Description  Создает новую запись на прием для пациента, связывая ее со слотом в расписании и исходным талоном. Обновляет слот как занятый.
// @Tags         registrar
// @Accept       json
// @Produce      json
// @Param        request body models.CreateAppointmentRequest true "Данные для создания записи"
// @Success      201 {object} models.Appointment "Успешно созданная запись"
// @Failure      400 {object} map[string]string "Ошибка: неверный формат запроса"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера (например, слот уже занят)"
// @Security     ApiKeyAuth
// @Router       /api/registrar/appointments [post]
func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	log := logger.Default()

	var req models.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Warn("CreateAppointment: Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса: " + err.Error()})
		return
	}

	appointment, err := h.service.CreateAppointment(&req)
	if err != nil {
		log.WithError(err).Error("CreateAppointment: Failed to create appointment in service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, appointment)
}
