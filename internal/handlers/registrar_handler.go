package handlers

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegistrarHandler обрабатывает запросы от рабочего места регистратора
type RegistrarHandler struct {
	ticketService *services.TicketService
}

// NewRegistrarHandler создает новый RegistrarHandler
func NewRegistrarHandler(ts *services.TicketService) *RegistrarHandler {
	return &RegistrarHandler{ticketService: ts}
}

// CallNextRequest описывает запрос на вызов следующего пациента
type CallNextRequest struct {
	WindowNumber int `json:"window_number" binding:"required,gt=0"`
}

// CallNext вызывает следующего пациента в очереди
// @Summary      Вызвать следующего пациента
// @Description  Находит первого пациента в очереди, меняет его статус на "приглашен" и присваивает номер окна
// @Tags         registrar
// @Accept       json
// @Produce      json
// @Param        request body CallNextRequest true "Номер окна, которое вызывает пациента"
// @Success      200 {object} models.TicketResponse "Данные вызванного талона"
// @Failure      400 {object} map[string]string "Ошибка: неверный номер окна"
// @Failure      404 {object} map[string]string "Ошибка: очередь пуста"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/registrar/call-next [post]
func (h *RegistrarHandler) CallNext(c *gin.Context) {
	var req CallNextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос: 'window_number' является обязательным положительным числом."})
		return
	}

	ticket, err := h.ticketService.CallNextTicket(req.WindowNumber)
	if err != nil {
		if err.Error() == "очередь пуста" {
			logger.Default().Info("CallNext handler: queue is empty")
			c.JSON(http.StatusNotFound, gin.H{"message": "Очередь пуста"})
			return
		}
		// Проверка на gorm.ErrRecordNotFound
		if err == gorm.ErrRecordNotFound {
			logger.Default().Info("CallNext handler: queue is empty (gorm)")
			c.JSON(http.StatusNotFound, gin.H{"message": "Очередь пуста"})
			return
		}

		logger.Default().Error(fmt.Sprintf("CallNext: failed to call ticket: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось вызвать талон"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

// UpdateStatusRequest описывает запрос для смены статуса тикета
// @Description Запрос для смены статуса тикета
// @Example {"status": "подойти_к_окну"}
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateStatus меняет статус тикета
// @Summary      Сменить статус тикета
// @Description  Изменяет статус тикета по ID
// @Tags         registrar
// @Accept       json
// @Produce      json
// @Param        id path int true "ID тикета"
// @Param        request body UpdateStatusRequest true "Новый статус"
// @Success      200 {object} map[string]string "Статус обновлен"
// @Failure      400 {object} map[string]string "Ошибка запроса"
// @Failure      404 {object} map[string]string "Тикет не найден"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/registrar/tickets/{id}/status [patch]
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

// DeleteTicket удаляет тикет
// @Summary      Удалить тикет
// @Description  Удаляет тикет по ID
// @Tags         registrar
// @Accept       json
// @Produce      json
// @Param        id path int true "ID тикета"
// @Success      200 {object} map[string]string "Тикет удален"
// @Failure      400 {object} map[string]string "Ошибка запроса"
// @Failure      404 {object} map[string]string "Тикет не найден"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/registrar/tickets/{id} [delete]
func (h *RegistrarHandler) DeleteTicket(c *gin.Context) {
	id := c.Param("id")
	if err := h.ticketService.DeleteTicket(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ticket deleted"})
}
