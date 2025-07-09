package handlers

import (
	"ElectronicQueue/internal/config"
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/services"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TicketHandler содержит реализует обработчики HTTP-запросов, связанных с талонами
// @Description Обработчики HTTP-запросов для работы с талонами электронной очереди
// @Tags tickets
// @Accept json
// @Produce json
type TicketHandler struct {
	service *services.TicketService
	config  *config.Config
}

// NewTicketHandler создает новый TicketHandler
func NewTicketHandler(service *services.TicketService, cfg *config.Config) *TicketHandler {
	return &TicketHandler{service: service, config: cfg}
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
// @Router       /api/tickets/print/selection [post]
func (h *TicketHandler) Selection(c *gin.Context) {
	var req ServiceSelectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Default().Error(fmt.Sprintf("Selection: failed to bind JSON: %v", err))
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
// @Router       /api/tickets/print/confirmation [post]
func (h *TicketHandler) Confirmation(c *gin.Context) {
	var req ConfirmationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Default().Error(fmt.Sprintf("Confirmation: failed to bind JSON: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_id and action are required"})
		return
	}

	ticket, err := h.service.CreateTicket(req.ServiceID)
	if err != nil {
		logger.Default().Error(fmt.Sprintf("Confirmation: failed to create ticket: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	serviceName := h.service.MapServiceIDToName(req.ServiceID)

	if req.Action == "print_ticket" {
		height := 800
		if h.config != nil && h.config.TicketHeight != "" {
			if parsed, err := strconv.Atoi(h.config.TicketHeight); err == nil {
				height = parsed
			}
		}
		imageBytes, err := h.service.GenerateTicketImage(height, ticket, serviceName, h.config.TicketMode)
		if err != nil {
			logger.Default().Error(fmt.Sprintf("Confirmation: image generation failed: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Image generation failed: %v", err)})
			return
		}

		// Сохраняем изображение на диск
		dir := h.config.TicketDir
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Default().Error(fmt.Sprintf("Confirmation: failed to create tickets directory: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tickets directory"})
			return
		}

		filePath := filepath.Join(dir, ticket.TicketNumber+".png")
		if err := os.WriteFile(filePath, imageBytes, 0644); err != nil {
			logger.Default().Error(fmt.Sprintf("Confirmation: failed to save image: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save image"})
			return
		}

		resp := ConfirmationResponse{
			ServiceName:  serviceName,
			TicketNumber: ticket.TicketNumber,
			Message:      "Ваш талон напечатан и сохранён как изображение",
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

// Эндпоинт для скачивания изображения талона
func (h *TicketHandler) DownloadTicket(c *gin.Context) {
	ticketNumber := c.Param("ticket_number")
	if ticketNumber == "" {
		logger.Default().Error("DownloadTicket: ticket_number is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "ticket_number is required"})
		return
	}

	filePath := filepath.Join(h.config.TicketDir, ticketNumber+".png")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Default().Error(fmt.Sprintf("DownloadTicket: ticket not found: %s", filePath))
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.png", ticketNumber))
	c.Header("Content-Type", "image/png")

	c.File(filePath)
}

// Эндпоинт для просмотра талона в браузере
func (h *TicketHandler) ViewTicket(c *gin.Context) {
	ticketNumber := c.Param("ticket_number")
	if ticketNumber == "" {
		logger.Default().Error("ViewTicket: ticket_number is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "ticket_number is required"})
		return
	}

	filePath := filepath.Join(h.config.TicketDir, ticketNumber+".png")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Default().Error(fmt.Sprintf("ViewTicket: ticket not found: %s", filePath))
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	c.Header("Content-Type", "image/png")

	c.File(filePath)
}
