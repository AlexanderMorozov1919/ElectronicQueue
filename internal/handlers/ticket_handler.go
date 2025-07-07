package handlers

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TicketHandler содержит реализует обработчики HTTP-запросов, связанных с талонами
type TicketHandler struct {
	service *services.TicketService
}

// NewTicketHandler создает новый TicketHandler
func NewTicketHandler(service *services.TicketService) *TicketHandler {
	return &TicketHandler{service: service}
}

// ServiceSelectionRequest описывает запрос выбора услуги
type ServiceSelectionRequest struct {
	ServiceID string `json:"service_id" binding:"required"`
}

// ServiceSelectionResponse описывает ответ после выбора услуги
type ServiceSelectionResponse struct {
	Action      string `json:"action"`
	ServiceName string `json:"service_name"`
}

// ConfirmationRequest описывает запрос подтверждения действия
type ConfirmationRequest struct {
	ServiceID string `json:"service_id" binding:"required"`
	Action    string `json:"action" binding:"required"`
}

// ConfirmationResponse описывает ответ после подтверждения действия
type ConfirmationResponse struct {
	ServiceName  string `json:"service_name"`
	TicketNumber string `json:"ticket_number"`
	Message      string `json:"message"`
	Timeout      int    `json:"timeout"`
}

// TicketStatusRequest описывает запрос для смены статуса тикета
type TicketStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// StartPage возвращает стартовую информацию для начальной страницы
func (h *TicketHandler) StartPage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"button_text": "Встать в очередь",
	})
}

// Services возвращает список доступных услуг
func (h *TicketHandler) Services(c *gin.Context) {
	services := h.service.GetAllServices()
	c.JSON(http.StatusOK, gin.H{"services": services})
}

// Selection определяет следующий шаг после выбора услуги
func (h *TicketHandler) Selection(c *gin.Context) {
	var req ServiceSelectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_id is required"})
		return
	}
	serviceName := h.service.MapServiceIDToName(req.ServiceID)
	resp := ServiceSelectionResponse{
		Action:      "confirm_print",
		ServiceName: serviceName,
	}
	c.JSON(http.StatusOK, resp)
}

// Confirmation обрабатывает подтверждение действия (печать талона или нет)
func (h *TicketHandler) Confirmation(c *gin.Context) {
	var req ConfirmationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_id and action are required"})
		return
	}
	ticket, err := h.service.CreateTicket(req.ServiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	serviceName := h.service.MapServiceIDToName(req.ServiceID)
	if req.Action == "print_ticket" {
		resp := ConfirmationResponse{
			ServiceName: serviceName,
			Message:     "Возьмите талон",
			Timeout:     5,
		}
		c.JSON(http.StatusOK, resp)
	} else {
		resp := ConfirmationResponse{
			ServiceName:  serviceName,
			TicketNumber: ticket.TicketNumber,
			Message:      "Ваш электронный талон",
			Timeout:      10,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// UpdateStatus Сменить статус тикета (регистратор)
func (h *TicketHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req TicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}
	ticket, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	var newStatus string
	switch req.Status {
	case "подойти_к_окну":
		ticket.Status = models.StatusToWindow
		newStatus = "подойти_к_окну"
	case "зарегистрирован":
		ticket.Status = models.StatusRegistered
		newStatus = "зарегистрирован"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}
	if err := h.service.UpdateTicket(ticket); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated", "status": newStatus})
}

// DeleteTicket Удалить тикет (регистратор)
func (h *TicketHandler) DeleteTicket(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteTicket(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ticket deleted"})
}
