package handlers

import (
	"ElectronicQueue/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TicketHandler содержит реализует обработчики HTTP-запросов, связанных с талонами.
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
	Print        bool   `json:"print"`
	Message      string `json:"message"`
	RedirectURL  string `json:"redirect_url"`
	Timeout      int    `json:"timeout"`
}

// StartPage возвращает стартовую информацию для начальной страницы
func (h *TicketHandler) StartPage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"button_text": "встать в очередь",
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
			Print:       req.Action == "print_ticket",
			Message:     "Возьмите талон",
			RedirectURL: "/api/tickets/start",
			Timeout:     10,
		}
		c.JSON(http.StatusOK, resp)
	} else {
		resp := ConfirmationResponse{
			ServiceName:  serviceName,
			TicketNumber: ticket.TicketNumber,
			Print:        req.Action == "print_ticket",
			Message:      "Ваш жлектронный талон",
			RedirectURL:  "/api/tickets/start",
			Timeout:      15,
		}
		c.JSON(http.StatusOK, resp)
	}
}
