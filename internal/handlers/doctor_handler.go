package handlers

import (
	"ElectronicQueue/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DoctorHandler содержит обработчики HTTP-запросов для работы врача
type DoctorHandler struct {
	service *services.DoctorService
}

// NewDoctorHandler создает новый DoctorHandler
func NewDoctorHandler(service *services.DoctorService) *DoctorHandler {
	return &DoctorHandler{service: service}
}

// StartAppointmentRequest описывает запрос на начало приема
type StartAppointmentRequest struct {
	TicketID uint `json:"ticket_id" binding:"required"`
}

// CompleteAppointmentRequest описывает запрос на завершение приема
type CompleteAppointmentRequest struct {
	TicketID uint `json:"ticket_id" binding:"required"`
}

// StartAppointment обрабатывает запрос на начало приема пациента
func (h *DoctorHandler) StartAppointment(c *gin.Context) {
	var req StartAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ticket_id is required"})
		return
	}

	// Вызываем сервис для начала приема
	ticket, err := h.service.StartAppointment(req.TicketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем обновленный талон
	c.JSON(http.StatusOK, gin.H{
		"message": "Appointment started successfully",
		"ticket": gin.H{
			"id":            ticket.ID,
			"ticket_number": ticket.TicketNumber,
			"status":        ticket.Status,
			"started_at":    ticket.StartedAt,
		},
	})
}

// CompleteAppointment обрабатывает запрос на завершение приема пациента
func (h *DoctorHandler) CompleteAppointment(c *gin.Context) {
	var req CompleteAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ticket_id is required"})
		return
	}

	// Вызываем сервис для завершения приема
	ticket, err := h.service.CompleteAppointment(req.TicketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем обновленный талон
	c.JSON(http.StatusOK, gin.H{
		"message": "Appointment completed successfully",
		"ticket": gin.H{
			"id":            ticket.ID,
			"ticket_number": ticket.TicketNumber,
			"status":        ticket.Status,
			"completed_at":  ticket.CompletedAt,
		},
	})
}
