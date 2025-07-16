package handlers

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/pubsub"
	"ElectronicQueue/internal/services"
	"io"
	"net/http"

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
	DoctorName      string `json:"doctor_name"`
	DoctorSpecialty string `json:"doctor_specialty"`
	OfficeNumber    int    `json:"office_number"`
	TicketNumber    string `json:"ticket_number,omitempty"`
	IsWaiting       bool   `json:"is_waiting"`
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
// @Description  Отправляет начальное состояние и последующие обновления статуса приема через Server-Sent Events.
// @Tags         doctor
// @Produce      text/event-stream
// @Success      200 {object} DoctorScreenResponse "Поток событий"
// @Router       /api/doctor/screen-updates [get]
func (h *DoctorHandler) DoctorScreenUpdates(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	log := logger.Default().WithField("module", "SSE_DOCTOR")

	clientChan := h.broker.Subscribe()
	defer h.broker.Unsubscribe(clientChan)

	// Функция для получения и отправки текущего состояния экрана врача
	sendCurrentState := func() bool {
		doctor, ticket, err := h.doctorService.GetCurrentAppointmentScreenState()
		if err != nil || doctor == nil {
			log.WithError(err).Error("Cannot get doctor screen state, no active doctor found in DB.")
			c.SSEvent("error", gin.H{"error": "No active doctor configured."})
			return false // Останавливаем стрим
		}

		response := DoctorScreenResponse{
			DoctorName:      doctor.FullName,
			DoctorSpecialty: doctor.Specialization,
			OfficeNumber:    1, // TODO: Получить номер кабинета из данных врача
			IsWaiting:       ticket == nil,
		}
		if ticket != nil {
			// Если есть талон на приеме, обновляем данные в ответе
			response.TicketNumber = ticket.TicketNumber
			log.WithFields(logrus.Fields{"ticket": ticket.TicketNumber, "doctor": doctor.FullName}).Info("Sending state: patient is being seen")
		} else {
			// Если талона на приеме нет
			log.WithField("doctor", doctor.FullName).Info("Sending state: waiting for patient")
		}

		c.SSEvent("state_update", response)
		if f, ok := c.Writer.(http.Flusher); ok {
			f.Flush()
			return c.Writer.Status() != http.StatusNotFound
		}
		// Проверяем, не отсоединился ли клиент во время отправки
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
		// При любом уведомлении от брокера перепроверяем состояние и отправляем его
		case _, ok := <-clientChan:
			if !ok {
				log.Info("Notification channel closed for doctor screen.")
				return false
			}
			log.Info("Received ticket update notification, refreshing doctor screen state.")
			return sendCurrentState()

		// Если клиент отключается
		case <-c.Request.Context().Done():
			log.Info("Client disconnected from doctor screen.")
			return false
		}
	})
}
