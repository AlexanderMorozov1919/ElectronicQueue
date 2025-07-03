package handlers

import (
	"net/http"

	"ElectronicQueue/internal/models/ticket_model"
	"ElectronicQueue/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketsHandler struct {
	service *services.TicketsService
}

func NewTicketsHandler(service *services.TicketsService) *TicketsHandler {
	return &TicketsHandler{service: service}
}

// GetQueueHandler обрабатывает GET /api/v1/registrar/queue
func (h *TicketsHandler) GetQueueHandler(c *gin.Context) {
	tickets, err := h.service.GetWaitingQueue()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	responses := make([]ticket_model.TicketResponse, len(tickets))
	for i, t := range tickets {
		responses[i] = ticket_model.TicketResponse{
			ID:           t.ID,
			TicketNumber: t.TicketNumber,
			Status:       t.Status,
			CreatedAt:    t.CreatedAt,
			CalledAt:     t.CalledAt,
			StartedAt:    t.StartedAt,
			CompletedAt:  t.CompletedAt,
		}
	}
	c.JSON(http.StatusOK, responses)
}

// CallNextTicketHandler обрабатывает POST /api/v1/registrar/tickets/call-next
func (h *TicketsHandler) CallNextTicketHandler(c *gin.Context) {
	ticket, err := h.service.CallNextTicket()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Очередь пуста"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := ticket_model.TicketResponse{
		ID:           ticket.ID,
		TicketNumber: ticket.TicketNumber,
		Status:       ticket.Status,
		CreatedAt:    ticket.CreatedAt,
		CalledAt:     ticket.CalledAt,
		StartedAt:    ticket.StartedAt,
		CompletedAt:  ticket.CompletedAt,
	}
	c.JSON(http.StatusOK, resp)
}
