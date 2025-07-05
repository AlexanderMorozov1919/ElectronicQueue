package handlers

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TicketHandler содержит зависимости для работы с талонами
type TicketHandler struct {
	Service *services.TicketService
}

// NewTicketHandler создает новый TicketHandler
func NewTicketHandler(service *services.TicketService) *TicketHandler {
	return &TicketHandler{Service: service}
}

// CreateTicketHandler обрабатывает создание нового талона
func (h *TicketHandler) CreateTicketHandler(c *gin.Context) {
	type request struct {
		Service string `json:"service" binding:"required"`
	}
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service is required"})
		return
	}
	ticket, err := h.Service.CreateTicket(req.Service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := models.TicketResponse{
		ID:           ticket.ID,
		TicketNumber: ticket.TicketNumber,
		Status:       ticket.Status,
		CreatedAt:    ticket.CreatedAt,
	}
	c.JSON(http.StatusOK, resp)
}

// GetAvailableServices возвращает список услуг
func (h *TicketHandler) GetAvailableServices(c *gin.Context) {
	services := []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{
		{ID: "make_appointment", Name: "Записаться к врачу"},
		{ID: "confirm_appointment", Name: "Прием по записи"},
		{ID: "lab_tests", Name: "Сдать анализы"},
		{ID: "documents", Name: "Другой вопрос"},
	}
	c.JSON(http.StatusOK, gin.H{"services": services})
}

// GetNextStep определяет следующий шаг после выбора услуги
func (h *TicketHandler) GetNextStep(c *gin.Context) {
	type reqBody struct {
		ServiceID string `json:"service_id" binding:"required"`
	}
	var req reqBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_id is required"})
		return
	}
	serviceName := h.Service.MapServiceIDToName(req.ServiceID)
	resp := struct {
		Action      string `json:"action"`
		ServiceName string `json:"service_name"`
	}{
		Action:      "confirm_print",
		ServiceName: serviceName,
	}
	c.JSON(http.StatusOK, resp)
}

// ConfirmAction обрабатывает подтверждение действия (печать талона и т.д.)
func (h *TicketHandler) ConfirmAction(c *gin.Context) {
	type reqBody struct {
		ServiceID string `json:"service_id" binding:"required"`
		Action    string `json:"action" binding:"required"`
	}
	var req reqBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_id and action are required"})
		return
	}
	ticket, err := h.Service.CreateTicket(req.ServiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	serviceName := h.Service.MapServiceIDToName(req.ServiceID)
	resp := struct {
		ServiceName  string `json:"service_name"`
		TicketNumber string `json:"ticket_number"`
		Print        bool   `json:"print"`
		Message      string `json:"message"`
	}{
		ServiceName:  serviceName,
		TicketNumber: ticket.TicketNumber,
		Print:        req.Action == "print_ticket",
		Message:      "Талон успешно создан",
	}
	c.JSON(http.StatusOK, resp)
}
