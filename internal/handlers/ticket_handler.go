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
