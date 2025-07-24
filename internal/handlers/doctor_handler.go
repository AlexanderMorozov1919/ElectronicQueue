package handlers

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/pubsub"
	"ElectronicQueue/internal/services"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

// StartBreakRequest описывает запрос на начало перерыва
// swagger:model StartBreakRequest
type StartBreakRequest struct {
	DoctorID uint `json:"doctor_id" binding:"required" example:"1"`
}

// EndBreakRequest описывает запрос на завершение перерыва
// swagger:model EndBreakRequest
type EndBreakRequest struct {
	DoctorID uint `json:"doctor_id" binding:"required" example:"1"`
}

// SetActiveRequest описывает запрос на установку статуса активный
// swagger:model SetActiveRequest
type SetActiveRequest struct {
	DoctorID uint `json:"doctor_id" binding:"required" example:"1"`
}

// SetInactiveRequest описывает запрос на установку статуса неактивный
// swagger:model SetInactiveRequest
type SetInactiveRequest struct {
	DoctorID uint `json:"doctor_id" binding:"required" example:"1"`
}

// DoctorScreenResponse определяет структуру данных для экрана у кабинета врача.
// @swagger:response DoctorScreenResponse
type DoctorScreenResponse struct {
	DoctorName      string                             `json:"doctor_name,omitempty"`
	DoctorSpecialty string                             `json:"doctor_specialty,omitempty"`
	CabinetNumber   int                                `json:"cabinet_number"`
	Queue           []models.DoctorQueueTicketResponse `json:"queue,omitempty"`
	Message         string                             `json:"message,omitempty"`
}

// GetAllActiveDoctors возвращает список всех врачей.
// @Summary      Получить список всех врачей
// @Description  Возвращает список всех врачей в системе. Используется для заполнения выпадающих списков на клиенте.
// @Tags         doctor
// @Produce      json
// @Success      200 {array} models.Doctor "Массив моделей врачей"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
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
	// Получаем ID врача из JWT токена
	doctorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID врача не найден в токене"})
		return
	}

	doctorIDUint, ok := doctorID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Неверный формат ID врача"})
		return
	}

	// Получить только талоны этого врача
	tickets, err := h.doctorService.GetRegisteredTicketsForDoctor(doctorIDUint)
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
	// ID врача из JWT токена
	doctorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID врача не найден в токене"})
		return
	}

	doctorIDUint, ok := doctorID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Неверный формат ID врача"})
		return
	}

	// Получить только талоны этого врача
	tickets, err := h.doctorService.GetInProgressTicketsForDoctor(doctorIDUint)
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
// @Security     ApiKeyAuth
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
// @Security     ApiKeyAuth
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

// StartBreak обрабатывает запрос на начало перерыва врача
// @Summary      Начать перерыв врача
// @Description  Начинает перерыв врача. Статус врача должен быть 'активен'.
// @Tags         doctor
// @Accept       json
// @Produce      json
// @Param        request body StartBreakRequest true "Данные для начала перерыва"
// @Success      200 {object} map[string]string "Break started successfully"
// @Failure      400 {object} map[string]string "Неверный запрос или статус врача"
// @Security     ApiKeyAuth
// @Router       /api/doctor/start-break [post]
func (h *DoctorHandler) StartBreak(c *gin.Context) {
	log := logger.Default().WithField("handler", "StartBreak")

	var req StartBreakRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, gin.H{"error": "doctor_id обязателен"})
		return
	}

	log.WithField("doctor_id", req.DoctorID).Info("Начало перерыва для врача")

	if err := h.doctorService.StartBreak(req.DoctorID); err != nil {
		log.WithError(err).WithField("doctor_id", req.DoctorID).Error("Ошибка начала перерыва")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("doctor_id", req.DoctorID).Info("Перерыв начат успешно")
	c.JSON(http.StatusOK, gin.H{"message": "Перерыв начат успешно"})
}

// EndBreak обрабатывает запрос на завершение перерыва врача
// @Summary      Завершить перерыв врача
// @Description  Завершает перерыв врача. Статус врача должен быть 'перерыв'.
// @Tags         doctor
// @Accept       json
// @Produce      json
// @Param        request body EndBreakRequest true "Данные для завершения перерыва"
// @Success      200 {object} map[string]string "Break ended successfully"
// @Failure      400 {object} map[string]string "Неверный запрос или статус врача"
// @Security     ApiKeyAuth
// @Router       /api/doctor/end-break [post]
func (h *DoctorHandler) EndBreak(c *gin.Context) {
	log := logger.Default().WithField("handler", "EndBreak")

	var req EndBreakRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, gin.H{"error": "doctor_id обязателен"})
		return
	}

	log.WithField("doctor_id", req.DoctorID).Info("Завершение перерыва для врача")

	if err := h.doctorService.EndBreak(req.DoctorID); err != nil {
		log.WithError(err).WithField("doctor_id", req.DoctorID).Error("Ошибка завершения перерыва")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("doctor_id", req.DoctorID).Info("Перерыв завершен успешно")
	c.JSON(http.StatusOK, gin.H{"message": "Перерыв завершен успешно"})
}

// SetDoctorActive обрабатывает запрос на установку статуса врача как активный
// @Summary      Установить статус врача как активный
// @Description  Устанавливает статус врача как активный (при входе в систему).
// @Tags         doctor
// @Accept       json
// @Produce      json
// @Param        request body SetActiveRequest true "Данные для установки статуса"
// @Success      200 {object} map[string]string "Doctor status set to active"
// @Failure      400 {object} map[string]string "Неверный запрос"
// @Router       /api/doctor/set-active [post]
func (h *DoctorHandler) SetDoctorActive(c *gin.Context) {
	log := logger.Default().WithField("handler", "SetDoctorActive")

	var req SetActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, gin.H{"error": "doctor_id обязателен"})
		return
	}

	log.WithField("doctor_id", req.DoctorID).Info("Установка статуса активен для врача")

	if err := h.doctorService.SetDoctorActive(req.DoctorID); err != nil {
		log.WithError(err).WithField("doctor_id", req.DoctorID).Error("Ошибка установки статуса активен")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("doctor_id", req.DoctorID).Info("Статус активен установлен успешно")
	c.JSON(http.StatusOK, gin.H{"message": "Статус активен установлен успешно"})
}

// SetDoctorInactive обрабатывает запрос на установку статуса врача как неактивный
// @Summary      Установить статус врача как неактивный
// @Description  Устанавливает статус врача как неактивный (при выходе из системы).
// @Tags         doctor
// @Accept       json
// @Produce      json
// @Param        request body SetInactiveRequest true "Данные для установки статуса"
// @Success      200 {object} map[string]string "Doctor status set to inactive"
// @Failure      400 {object} map[string]string "Неверный запрос"
// @Router       /api/doctor/set-inactive [post]
func (h *DoctorHandler) SetDoctorInactive(c *gin.Context) {
	log := logger.Default().WithField("handler", "SetDoctorInactive")

	var req SetInactiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, gin.H{"error": "doctor_id обязателен"})
		return
	}

	log.WithField("doctor_id", req.DoctorID).Info("Установка статуса неактивен для врача")

	if err := h.doctorService.SetDoctorInactive(req.DoctorID); err != nil {
		log.WithError(err).WithField("doctor_id", req.DoctorID).Error("Ошибка установки статуса неактивен")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("doctor_id", req.DoctorID).Info("Статус неактивен установлен успешно")
	c.JSON(http.StatusOK, gin.H{"message": "Статус неактивен установлен успешно"})
}

// DoctorScreenUpdates - SSE эндпоинт для табло у кабинета врача.
// @Summary      Получить обновления для табло врача
// @Description  Отправляет начальное состояние и последующие обновления статуса приема через Server-Sent Events для конкретного кабинета.
// @Tags         doctor
// @Produce      text/event-stream
// @Param        cabinet_number path int true "Номер кабинета"
// @Success      200 {object} DoctorScreenResponse "Поток событий (см. реальную структуру ответа в коде)"
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
		schedule, queue, err := h.doctorService.GetDoctorScreenState(cabinetNumber)
		if err != nil {
			// Если произошла критическая ошибка в сервисе, логируем и прекращаем.
			log.WithError(err).Error("Критическая ошибка в GetDoctorScreenState")
			return false
		}

		doctorName := ""
		doctorSpecialty := ""
		doctorStatus := models.DoctorStatusInactive
		if schedule != nil {
			doctorName = schedule.Doctor.FullName
			doctorSpecialty = schedule.Doctor.Specialization
			doctorStatus = schedule.Doctor.Status
		}

		response := gin.H{
			"doctor_name":      doctorName,
			"doctor_specialty": doctorSpecialty,
			"cabinet_number":   cabinetNumber,
			"queue":            queue,
			"message":          "",
			"doctor_status":    doctorStatus,
		}

		log.WithField("queue_size", len(queue)).Info("Отправка обновления состояния экрана врача")
		c.SSEvent("state_update", response)

		// Проверяем, жив ли клиент, и сбрасываем буфер.
		if f, ok := c.Writer.(http.Flusher); ok {
			f.Flush()
			return c.Writer.Status() != http.StatusNotFound
		}
		return c.Writer.Status() < http.StatusInternalServerError
	}

	// Отправляем начальное состояние сразу после подключения
	if !sendCurrentState() {
		log.Info("Клиент отключился сразу после отправки начального состояния.")
		return
	}

	// Запускаем стрим для отправки обновлений
	c.Stream(func(w io.Writer) bool {
		select {
		case _, ok := <-clientChan:
			if !ok {
				log.Info("Канал уведомления закрыт для экрана врача.")
				return false
			}
			log.Info("Получено уведомление об обновлении талона, обновление состояния экрана врача.")
			return sendCurrentState()

		case <-c.Request.Context().Done():
			log.Info("Клиент отключился от экрана врача.")
			return false
		}
	})
}
