package handlers

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RegistrarHandler struct {
	ticketService    *services.TicketService
	registrarService *services.RegistrarService
}

func NewRegistrarHandler(ts *services.TicketService, rs *services.RegistrarService) *RegistrarHandler {
	return &RegistrarHandler{
		ticketService:    ts,
		registrarService: rs,
	}
}

func (h *RegistrarHandler) GetTickets(c *gin.Context) {
	log := logger.Default()
	categoryPrefix := c.Query("category")

	registrarID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID регистратора не найден в токене"})
		return
	}
	registrarIDUint, _ := registrarID.(uint)

	tickets, err := h.ticketService.GetTicketsForRegistrar(categoryPrefix, registrarIDUint)
	if err != nil {
		log.WithError(err).Error("GetTickets: failed to get tickets from service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список талонов"})
		return
	}

	c.JSON(http.StatusOK, tickets)
}

func (h *RegistrarHandler) GetCurrentTicket(c *gin.Context) {
	windowNumberStr := c.Query("window_number")
	if windowNumberStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'window_number' is required"})
		return
	}

	windowNumber, err := strconv.Atoi(windowNumberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'window_number' format"})
		return
	}

	ticket, err := h.ticketService.GetInvitedTicketForWindow(windowNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current ticket"})
		return
	}

	if ticket == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No active ticket for this window"})
		return
	}

	c.JSON(http.StatusOK, ticket.ToResponse())
}

type CallNextRequest struct {
	WindowNumber   int    `json:"window_number" binding:"required,gt=0"`
	CategoryPrefix string `json:"category_prefix,omitempty"`
}

type CallSpecificRequest struct {
	TicketID     uint `json:"ticket_id" binding:"required"`
	WindowNumber int  `json:"window_number" binding:"required,gt=0"`
}

func (h *RegistrarHandler) CallNext(c *gin.Context) {
	var req CallNextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос: 'window_number' должен быть числом."})
		return
	}

	registrarID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID регистратора не найден в токене"})
		return
	}
	registrarIDUint, _ := registrarID.(uint)

	ticket, err := h.ticketService.CallNextTicket(req.WindowNumber, req.CategoryPrefix, registrarIDUint)
	if err != nil {
		if err.Error() == "очередь пуста" {
			c.JSON(http.StatusNotFound, gin.H{"message": "Очередь пуста"})
			return
		}
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Очередь пуста"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось вызвать талон"})
		return
	}

	c.JSON(http.StatusOK, ticket.ToResponse())
}

func (h *RegistrarHandler) CallSpecific(c *gin.Context) {
	var req CallSpecificRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос: " + err.Error()})
		return
	}

	ticket, err := h.ticketService.CallSpecificTicket(req.TicketID, req.WindowNumber)
	if err != nil {
		if strings.Contains(err.Error(), "не найден") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "имеет неверный статус") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось вызвать талон"})
		return
	}

	c.JSON(http.StatusOK, ticket.ToResponse())
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *RegistrarHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}
	ticket, err := h.ticketService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}
	ticket.Status = models.TicketStatus(req.Status)
	if err := h.ticketService.UpdateTicket(ticket); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h *RegistrarHandler) DeleteTicket(c *gin.Context) {
	id := c.Param("id")
	if err := h.ticketService.DeleteTicket(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ticket deleted"})
}

func (h *RegistrarHandler) GetDailyReport(c *gin.Context) {
	log := logger.Default()
	reportData, err := h.ticketService.GetDailyReport()
	if err != nil {
		log.WithError(err).Error("GetDailyReport: Failed to get daily report from service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить дневной отчет"})
		return
	}

	c.JSON(http.StatusOK, reportData)
}

func (h *RegistrarHandler) GetAllServices(c *gin.Context) {
	services, err := h.registrarService.GetAllServices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get services"})
		return
	}
	c.JSON(http.StatusOK, services)
}

func (h *RegistrarHandler) GetPriorities(c *gin.Context) {
	registrarID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID регистратора не найден в токене"})
		return
	}
	registrarIDUint, _ := registrarID.(uint)

	priorities, err := h.registrarService.GetPriorities(registrarIDUint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get priorities"})
		return
	}
	c.JSON(http.StatusOK, priorities)
}

type SetPrioritiesRequest struct {
	ServiceIDs []uint `json:"service_ids"`
}

func (h *RegistrarHandler) SetPriorities(c *gin.Context) {
	registrarID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID регистратора не найден в токене"})
		return
	}
	registrarIDUint, _ := registrarID.(uint)

	var req SetPrioritiesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if err := h.registrarService.SetPriorities(registrarIDUint, req.ServiceIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set priorities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Priorities updated successfully"})
}
