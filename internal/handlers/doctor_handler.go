package handlers

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/pubsub"
	"ElectronicQueue/internal/services"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// DoctorHandler содержит обработчики HTTP-запросов для работы врача
type DoctorHandler struct {
	doctorService *services.DoctorService
	broker        *pubsub.Broker
}

// NewDoctorHandler создает новый DoctorHandler
func NewDoctorHandler(service *services.DoctorService, broker *pubsub.Broker) *DoctorHandler {
	return &DoctorHandler{
		doctorService: service,
		broker:        broker,
	}
}

// StartAppointmentRequest описывает запрос на начало приема
// swagger:model StartAppointmentRequest
type StartAppointmentRequest struct {
	TicketID uint `json:"ticket_id" binding:"required" example:"1"`
}

// CompleteAppointmentRequest описывает запрос на завершение приема
// swagger:model CompleteAppointmentRequest
type CompleteAppointmentRequest struct {
	TicketID uint `json:"ticket_id" binding:"required" example:"1"`
}

// DoctorScreenResponse определяет структуру данных для экрана у кабинета врача.
type DoctorScreenResponse struct {
	DoctorName      string `json:"doctor_name,omitempty"`
	DoctorSpecialty string `json:"doctor_specialty,omitempty"`
	CabinetNumber   int    `json:"cabinet_number"`
	TicketNumber    string `json:"ticket_number,omitempty"`
	IsWaiting       bool   `json:"is_waiting"`
	Message         string `json:"message,omitempty"` // Поле для сообщений, например, "нет приема"
}

// GetAllActiveDoctors возвращает список всех активных врачей.
// @Summary      Получить список активных врачей
// @Description  Возвращает список всех врачей, у которых is_active = true. Используется для заполнения выпадающих списков на клиенте.
// @Tags         doctor
// @Produce      json
// @Success      200 {array} models.Doctor "Массив моделей врачей"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/doctor/active [get]
func (h *DoctorHandler) GetAllActiveDoctors(c *gin.Context) {
	doctors, err := h.doctorService.GetAllActiveDoctors()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список врачей: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, doctors)
}

// GetActiveCabinets godoc
// @Summary      Получить список всех существующих кабинетов
// @Description  Возвращает список всех уникальных номеров кабинетов, когда-либо существовавших в расписании.
// @Tags         doctor
// @Produce      json
// @Success      200 {array} integer "Массив номеров кабинетов"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/doctor/cabinets/active [get]
func (h *DoctorHandler) GetActiveCabinets(c *gin.Context) {
	cabinets, err := h.doctorService.GetAllUniqueCabinets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список кабинетов: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, cabinets)
}

// GetRegisteredTickets возвращает талоны со статусом "зарегистрирован"
// @Summary      Получить очередь к врачу
// @Description  Возвращает список талонов со статусом "зарегистрирован", т.е. очередь непосредственно к врачу.
// @Tags         doctor
// @Produce      json
// @Success      200 {object} []models.TicketResponse "Список талонов"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/doctor/tickets/registered [get]
func (h *DoctorHandler) GetRegisteredTickets(c *gin.Context) {
	tickets, err := h.doctorService.GetRegisteredTickets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tickets)
}

// GetInProgressTickets возвращает талоны со статусом "на приеме"
// @Summary      Получить талоны на приеме
// @Description  Возвращает список талонов со статусом "на_приеме". Обычно это один талон.
// @Tags         doctor
// @Produce      json
// @Success      200 {object} []models.TicketResponse "Список талонов"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/doctor/tickets/in-progress [get]
func (h *DoctorHandler) GetInProgressTickets(c *gin.Context) {
	tickets, err := h.doctorService.GetInProgressTickets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tickets)
}

// StartAppointment обрабатывает запрос на начало приема пациента
// @Summary      Начать прием пациента
// @Description  Начинает прием пациента по талону. Статус талона должен быть 'зарегистрирован'.
// @Tags         doctor
// @Accept       json
// @Produce      json
// @Param        request body StartAppointmentRequest true "Данные для начала приема"
// @Success      200 {object} map[string]interface{} "Appointment started successfully"
// @Failure      400 {object} map[string]string "Неверный запрос или статус талона"
// @Router       /api/doctor/start-appointment [post]
func (h *DoctorHandler) StartAppointment(c *gin.Context) {
	var req StartAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ticket_id is required"})
		return
	}

	ticket, err := h.doctorService.StartAppointment(req.TicketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Appointment started successfully",
		"ticket":  ticket.ToResponse(),
	})
}

// CompleteAppointment обрабатывает запрос на завершение приема пациента
// @Summary      Завершить прием пациента
// @Description  Завершает прием пациента по талону. Статус талона должен быть 'на_приеме'.
// @Tags         doctor
// @Accept       json
// @Produce      json
// @Param        request body CompleteAppointmentRequest true "Данные для завершения приема"
// @Success      200 {object} map[string]interface{} "Appointment completed successfully"
// @Failure      400 {object} map[string]string "Неверный запрос или статус талона"
// @Router       /api/doctor/complete-appointment [post]
func (h *DoctorHandler) CompleteAppointment(c *gin.Context) {
	var req CompleteAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ticket_id is required"})
		return
	}

	ticket, err := h.doctorService.CompleteAppointment(req.TicketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Appointment completed successfully",
		"ticket":  ticket.ToResponse(),
	})
}

// DoctorScreenUpdates - SSE эндпоинт для табло у кабинета врача.
// @Summary      Получить обновления для табло врача
// @Description  Отправляет начальное состояние и последующие обновления статуса приема через Server-Sent Events для конкретного кабинета.
// @Tags         doctor
// @Produce      text/event-stream
// @Param        cabinet_number path int true "Номер кабинета"
// @Success      200 {object} DoctorScreenResponse "Поток событий"
// @Failure      400 {object} map[string]string "Неверный формат номера кабинета"
// @Router       /api/doctor/screen-updates/{cabinet_number} [get]
func (h *DoctorHandler) DoctorScreenUpdates(c *gin.Context) {
	cabinetNumberStr := c.Param("cabinet_number")
	cabinetNumber, err := strconv.Atoi(cabinetNumberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный номер кабинета"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	log := logger.Default().WithField("module", "SSE_DOCTOR").WithField("cabinet", cabinetNumber)

	clientChan := h.broker.Subscribe()
	defer h.broker.Unsubscribe(clientChan)

	// Функция для получения и отправки текущего состояния экрана врача
	sendCurrentState := func() bool {
		schedule, ticket, err := h.doctorService.GetCurrentAppointmentScreenState(cabinetNumber)

		// Если нет расписания на данный момент
		if err != nil || schedule == nil || schedule.Doctor.FullName == "" {
			log.Warn("No active schedule found for this cabinet. Sending 'no reception' message.")
			response := DoctorScreenResponse{
				CabinetNumber: cabinetNumber,
				Message:       fmt.Sprintf("В кабинете %d нет приёма", cabinetNumber),
				IsWaiting:     true, // Дефолтное значение
			}
			c.SSEvent("state_update", response)
			c.Writer.Flush()
			return true // Продолжаем слушать, вдруг прием начнется
		}

		response := DoctorScreenResponse{
			DoctorName:      schedule.Doctor.FullName,
			DoctorSpecialty: schedule.Doctor.Specialization,
			CabinetNumber:   cabinetNumber,
			IsWaiting:       ticket == nil,
		}
		if ticket != nil {
			response.TicketNumber = ticket.TicketNumber
			log.WithFields(logrus.Fields{"ticket": ticket.TicketNumber, "doctor": schedule.Doctor.FullName}).Info("Sending state: patient is being seen")
		} else {
			log.WithField("doctor", schedule.Doctor.FullName).Info("Sending state: waiting for patient")
		}

		c.SSEvent("state_update", response)
		if f, ok := c.Writer.(http.Flusher); ok {
			f.Flush()
			return c.Writer.Status() != http.StatusNotFound
		}
		return c.Writer.Status() < http.StatusInternalServerError
	}

	// Отправляем начальное состояние сразу после подключения
	if !sendCurrentState() {
		log.Info("Client disconnected immediately after initial state send.")
		return
	}

	// Запускаем стрим для отправки обновлений
	c.Stream(func(w io.Writer) bool {
		select {
		case _, ok := <-clientChan:
			if !ok {
				log.Info("Notification channel closed for doctor screen.")
				return false
			}
			log.Info("Received ticket update notification, refreshing doctor screen state.")
			return sendCurrentState()

		case <-c.Request.Context().Done():
			log.Info("Client disconnected from doctor screen.")
			return false
		}
	})
}
