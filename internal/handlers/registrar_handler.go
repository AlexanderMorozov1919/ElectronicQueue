package handlers

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RegistrarHandler struct {
	ticketService *services.TicketService
}

func NewRegistrarHandler(ts *services.TicketService) *RegistrarHandler {
	return &RegistrarHandler{ticketService: ts}
}

// GetTickets godoc
// @Summary      Получить список талонов для регистратора
// @Description  Возвращает список талонов по нужным статусам, с возможностью фильтрации по категории.
// @Tags         registrar
// @Produce      json
// @Param        category query string false "Префикс категории для фильтрации (например, 'A', 'B')"
// @Success      200 {array} models.RegistrarTicketResponse "Массив талонов"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/registrar/tickets [get]
func (h *RegistrarHandler) GetTickets(c *gin.Context) {
	log := logger.Default()
	categoryPrefix := c.Query("category")

	tickets, err := h.ticketService.GetTicketsForRegistrar(categoryPrefix)
	if err != nil {
		log.WithError(err).Error("GetTickets: failed to get tickets from service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список талонов"})
		return
	}

	c.JSON(http.StatusOK, tickets)
}

type CallNextRequest struct {
	WindowNumber   int    `json:"window_number" binding:"required,gt=0"`
	CategoryPrefix string `json:"category_prefix,omitempty"`
}

type CallSpecificRequest struct {
	TicketID     uint `json:"ticket_id" binding:"required"`
	WindowNumber int  `json:"window_number" binding:"required,gt=0"`
}

// CallNext вызывает следующего пациента в очереди
// @Summary      Вызвать следующего пациента
// @Description  Находит первого пациента в очереди (опционально по категории), меняет его статус на "приглашен" и присваивает номер окна
// @Tags         registrar
// @Accept       json
// @Produce      json
// @Param        request body CallNextRequest true "Номер окна и опциональный префикс категории талона"
// @Success      200 {object} models.TicketResponse "Данные вызванного талона"
// @Failure      400 {object} map[string]string "Ошибка: неверный номер окна"
// @Failure      404 {object} map[string]string "Ошибка: очередь пуста"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/registrar/call-next [post]
func (h *RegistrarHandler) CallNext(c *gin.Context) {
	var req CallNextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос: 'window_number' должен быть числом."})
		return
	}

	if req.WindowNumber <= 0 {
		req.WindowNumber = 1
	}

	ticket, err := h.ticketService.CallNextTicket(req.WindowNumber, req.CategoryPrefix)
	if err != nil {
		if err.Error() == "очередь пуста" {
			logger.Default().Info("CallNext handler: queue is empty")
			c.JSON(http.StatusNotFound, gin.H{"message": "Очередь пуста"})
			return
		}
		if err == gorm.ErrRecordNotFound {
			logger.Default().Info("CallNext handler: queue is empty (gorm)")
			c.JSON(http.StatusNotFound, gin.H{"message": "Очередь пуста"})
			return
		}

		logger.Default().WithError(err).Error("CallNext: failed to call ticket")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось вызвать талон"})
		return
	}

	c.JSON(http.StatusOK, ticket.ToResponse())
}

// CallSpecific вызывает конкретного пациента по ID талона
// @Summary      Вызвать конкретного пациента
// @Description  Находит пациента по ID талона, меняет его статус на "приглашен" и присваивает номер окна. Доступно только для талонов в статусе 'ожидает'.
// @Tags         registrar
// @Accept       json
// @Produce      json
// @Param        request body CallSpecificRequest true "ID талона и номер окна"
// @Success      200 {object} models.TicketResponse "Данные вызванного талона"
// @Failure      400 {object} map[string]string "Ошибка: неверный ID, номер окна или неверный статус талона"
// @Failure      404 {object} map[string]string "Ошибка: талон не найден"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/registrar/call-specific [post]
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
		logger.Default().WithError(err).Error("CallSpecific: failed to call ticket")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось вызвать талон"})
		return
	}

	c.JSON(http.StatusOK, ticket.ToResponse())
}

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
// @Security     ApiKeyAuth
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
// @Summary      Удалить тикет (Админ)
// @Description  Удаляет тикет по ID. Требует INTERNAL_API_KEY.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id path int true "ID тикета"
// @Success      200 {object} map[string]string "Тикет удален"
// @Failure      400 {object} map[string]string "Ошибка запроса"
// @Failure      401 {object} map[string]string "Отсутствует ключ API"
// @Failure      403 {object} map[string]string "Неверный ключ API"
// @Failure      404 {object} map[string]string "Тикет не найден"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/admin/tickets/{id} [delete]
func (h *RegistrarHandler) DeleteTicket(c *gin.Context) {
	id := c.Param("id")
	if err := h.ticketService.DeleteTicket(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ticket deleted"})
}

// GetDailyReport godoc
// @Summary      Получить отчет по талонам за текущий день
// @Description  Возвращает список всех талонов, созданных сегодня, с детальной информацией.
// @Tags         registrar
// @Produce      json
// @Success      200 {array} models.DailyReportRow "Массив строк отчета"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/registrar/reports/daily [get]
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
