package handlers

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/services"
	"net/http"
	"strconv"
	"strings"
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

// GetPatientAppointments godoc
// @Summary      Получить историю записей пациента
// @Description  Возвращает все прошлые и будущие записи для указанного пациента.
// @Tags         registrar
// @Produce      json
// @Param        patient_id path int true "ID Пациента"
// @Success      200 {array} services.AppointmentDetailsResponse "Массив записей пациента"
// @Failure      400 {object} map[string]string "Ошибка: неверный ID пациента"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/registrar/patients/{patient_id}/appointments [get]
func (h *AppointmentHandler) GetPatientAppointments(c *gin.Context) {
	patientIDStr := c.Param("patient_id")
	patientID, err := strconv.ParseUint(patientIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID пациента"})
		return
	}

	appointments, err := h.service.GetAppointmentsByPatient(uint(patientID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить записи пациента"})
		return
	}

	c.JSON(http.StatusOK, appointments)
}

// DeleteAppointment godoc
// @Summary      Удалить будущую запись
// @Description  Удаляет запись на прием и освобождает связанный с ней слот в расписании.
// @Tags         registrar
// @Produce      json
// @Param        id path int true "ID Записи"
// @Success      200 {object} map[string]string "Запись успешно удалена"
// @Failure      400 {object} map[string]string "Ошибка: неверный ID"
// @Failure      404 {object} map[string]string "Запись не найдена"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/registrar/appointments/{id} [delete]
func (h *AppointmentHandler) DeleteAppointment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID"})
		return
	}

	if err := h.service.DeleteAppointment(uint(id)); err != nil {
		if strings.Contains(err.Error(), "не найдена") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении записи"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Запись успешно удалена"})
}

type ConfirmAppointmentRequest struct {
	TicketID uint `json:"ticket_id" binding:"required"`
}

// ConfirmAppointment godoc
// @Summary      Подтвердить явку по записи
// @Description  Привязывает существующую запись к новому талону, который пациент получил сегодня, и меняет статус талона на 'зарегистрирован', отправляя пациента в очередь к врачу.
// @Tags         registrar
// @Accept       json
// @Produce      json
// @Param        id path int true "ID Записи для подтверждения"
// @Param        request body ConfirmAppointmentRequest true "ID нового талона"
// @Success      200 {object} models.Appointment "Обновленная запись с привязанным талоном"
// @Failure      400 {object} map[string]string "Ошибка: неверный формат запроса"
// @Failure      500 {object} map[string]string "Ошибка сервера (запись или талон не найдены, или запись уже подтверждена)"
// @Security     ApiKeyAuth
// @Router       /api/registrar/appointments/{id}/confirm [patch]
func (h *AppointmentHandler) ConfirmAppointment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID записи"})
		return
	}

	var req ConfirmAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса: " + err.Error()})
		return
	}

	appointment, err := h.service.ConfirmAppointment(uint(id), req.TicketID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось подтвердить запись: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, appointment)
}
