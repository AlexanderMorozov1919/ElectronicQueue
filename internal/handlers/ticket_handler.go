package handlers

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/services"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// TicketHandler содержит реализует обработчики HTTP-запросов, связанных с талонами
// @Description Обработчики HTTP-запросов для работы с талонами электронной очереди
// @Tags tickets
// @Accept json
// @Produce json
type TicketHandler struct {
	service *services.TicketService
}

// NewTicketHandler создает новый TicketHandler
func NewTicketHandler(service *services.TicketService) *TicketHandler {
	return &TicketHandler{service: service}
}

// ServiceSelectionRequest описывает запрос выбора услуги
// @Description Запрос для выбора услуги
// @Example {"service_id": "make_appointment"}
type ServiceSelectionRequest struct {
	ServiceID string `json:"service_id" binding:"required" example:"make_appointment"`
}

// ServiceSelectionResponse описывает ответ после выбора услуги
// @Description Ответ после выбора услуги
// @Example {"action": "confirm_print", "service_name": "Записаться к врачу"}
type ServiceSelectionResponse struct {
	Action      string `json:"action" example:"confirm_print"`
	ServiceName string `json:"service_name" example:"Записаться к врачу"`
}

// ConfirmationRequest описывает запрос подтверждения действия
// @Description Запрос подтверждения действия (печать талона или получение электронного)
// @Example {"service_id": "make_appointment", "action": "print_ticket"}
type ConfirmationRequest struct {
	ServiceID string `json:"service_id" binding:"required" example:"make_appointment"`
	Action    string `json:"action" binding:"required" example:"print_ticket"`
}

// ConfirmationResponse описывает ответ после подтверждения действия
// @Description Ответ после подтверждения действия
// @Example {"service_name": "Записаться к врачу", "ticket_number": "A001", "message": "Ваш электронный талон", "timeout": 10}
type ConfirmationResponse struct {
	ServiceName  string `json:"service_name" example:"Записаться к врачу"`
	TicketNumber string `json:"ticket_number,omitempty" example:"A001"`
	Message      string `json:"message" example:"Ваш электронный талон"`
	Timeout      int    `json:"timeout" example:"10"`
}

// TicketStatusRequest описывает запрос для смены статуса тикета
type TicketStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// StartPage возвращает стартовую информацию для начальной страницы
// StartPage godoc
// @Summary      Получить стартовую информацию
// @Description  Возвращает стартовую информацию для клиента (например, текст кнопки)
// @Tags         tickets
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]string "Успешный ответ: текст кнопки"
// @Router       /api/tickets/start [get]

func (h *TicketHandler) StartPage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"button_text": "Встать в очередь",
	})
}

// Services godoc
// @Summary      Получить список услуг
// @Description  Возвращает список доступных услуг
// @Tags         tickets
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string][]services.Service "Список услуг"
// @Router       /api/tickets/services [get]
func (h *TicketHandler) Services(c *gin.Context) {
	services := h.service.GetAllServices()
	c.JSON(http.StatusOK, gin.H{"services": services})
}

// Selection godoc
// @Summary      Выбор услуги
// @Description  Определяет следующий шаг после выбора услуги
// @Tags         tickets
// @Accept       json
// @Produce      json
// @Param        request body ServiceSelectionRequest true "Данные для выбора услуги"
// @Success      200 {object} ServiceSelectionResponse "Следующий шаг после выбора услуги"
// @Failure      400 {object} map[string]string "Ошибка: не передан service_id"
// @Router       /api/tickets/selection [post]
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

// Confirmation godoc
// @Summary      Подтверждение действия
// @Description  Обрабатывает подтверждение действия (печать талона или получение электронного)
// @Tags         tickets
// @Accept       json
// @Produce      json
// @Param        request body ConfirmationRequest true "Данные для подтверждения действия"
// @Success      200 {object} ConfirmationResponse "Ответ после подтверждения действия"
// @Failure      400 {object} map[string]string "Ошибка: не передан service_id или action"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/tickets/confirmation [post]
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
		pdfBytes, err := h.service.GenerateTicketPDF(ticket, serviceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("PDF generation failed: %v", err)})
			return
		}
		// Сохраняем PDF на диск
		dir := "tickets"
		if err := os.MkdirAll(dir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tickets directory"})
			return
		}
		filePath := filepath.Join(dir, ticket.TicketNumber+".pdf")
		if err := os.WriteFile(filePath, pdfBytes, 0644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save PDF"})
			return
		}
		resp := ConfirmationResponse{
			ServiceName:  serviceName,
			TicketNumber: ticket.TicketNumber,
			Message:      "Ваш талон напечатан и сохранён как PDF",
			Timeout:      5,
		}
		c.JSON(http.StatusOK, resp)
		return
	}
	resp := ConfirmationResponse{
		ServiceName:  serviceName,
		TicketNumber: ticket.TicketNumber,
		Message:      "Ваш электронный талон",
		Timeout:      10,
	}
	c.JSON(http.StatusOK, resp)
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
