package handlers

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ExportHandler обрабатывает запросы на экспорт данных.
type ExportHandler struct {
	service *services.ExportService
}

// NewExportHandler создает новый экземпляр ExportHandler.
func NewExportHandler(service *services.ExportService) *ExportHandler {
	return &ExportHandler{service: service}
}

// GetData обрабатывает запрос на получение данных из таблицы.
// @Summary      Экспорт данных из таблицы
// @Description  Позволяет получить данные из указанной таблицы с фильтрацией и пагинацией.
// @Tags         export
// @Accept       json
// @Produce      json
// @Param        table path string true "Имя таблицы для экспорта (e.g., tickets, doctors)"
// @Param        X-API-KEY header string true "Ключ API для доступа"
// @Param        request body models.ExportRequest true "Фильтры и параметры пагинации"
// @Success      200 {object} map[string]interface{} "Успешный ответ с данными"
// @Failure      400 {object} map[string]string "Ошибка в запросе (неверная таблица, поле или оператор)"
// @Failure      401 {object} map[string]string "Отсутствует ключ API"
// @Failure      403 {object} map[string]string "Неверный ключ API"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/export/{table} [post]
func (h *ExportHandler) GetData(c *gin.Context) {
	tableName := c.Param("table")

	var req models.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Default().WithError(err).Warn("Export handler: failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	data, total, err := h.service.ExportData(tableName, req)
	if err != nil {
		logger.Default().WithError(err).Error("Export handler: service returned an error")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":  req.Page,
		"limit": req.Limit,
		"total": total,
		"data":  data,
	})
}
