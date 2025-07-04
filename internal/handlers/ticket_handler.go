package handlers

import (
	"ElectronicQueue/internal/services"
	"path/filepath"

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

// GetServicePage - /terminal/service (GET)
func (h *TicketHandler) GetServicePage(c *gin.Context) {
	c.File(filepath.Join("frontend", "index.html"))
}

// GetSelectServicePage - /terminal/service/select (GET)
func (h *TicketHandler) GetSelectServicePage(c *gin.Context) {
	c.File(filepath.Join("frontend", "select.html"))
}

// HandleService - обработчик для создания талона
func (h *TicketHandler) HandleService(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := h.Service.CreateTicket(service)
		if err != nil {
			c.String(500, "Ошибка создания талона")
			return
		}
		c.File(filepath.Join("frontend", "success.html"))
	}
}
