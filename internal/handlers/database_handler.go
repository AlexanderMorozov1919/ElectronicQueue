package handlers

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DatabaseHandler struct {
	service *services.DatabaseService
}

func NewDatabaseHandler(service *services.DatabaseService) *DatabaseHandler {
	return &DatabaseHandler{service: service}
}

// GetData обрабатывает запрос на получение данных из таблицы.
// @Summary      Получение данных из таблицы
// @Description  Позволяет получить данные из указанной таблицы с фильтрацией и пагинацией.
// @Tags         database
// @Accept       json
// @Produce      json
// @Param        table path string true "Имя таблицы для получения данных (e.g., tickets, doctors)"
// @Param        request body models.GetDataRequest true "Фильтры и параметры пагинации"
// @Success      200 {object} map[string]interface{} "Успешный ответ с данными"
// @Failure      400 {object} map[string]string "Ошибка в запросе (неверная таблица, поле или оператор)"
// @Failure      401 {object} map[string]string "Отсутствует ключ API"
// @Failure      403 {object} map[string]string "Неверный ключ API"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/database/{table}/select [post]
func (h *DatabaseHandler) GetData(c *gin.Context) {
	tableName := c.Param("table")

	var req models.GetDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Default().WithError(err).Warn("Database handler (GetData): failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	data, total, err := h.service.GetData(tableName, req)
	if err != nil {
		logger.Default().WithError(err).Error("Database handler (GetData): service returned an error")
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

// InsertData обрабатывает запрос на вставку данных в таблицу.
// @Summary      Вставка данных в таблицу
// @Description  Позволяет вставить одну или несколько записей в указанную таблицу.
// @Tags         database
// @Accept       json
// @Produce      json
// @Param        table path string true "Имя таблицы для вставки (e.g., services, doctors)"
// @Param        request body models.InsertRequest true "Данные для вставки"
// @Success      201 {object} map[string]interface{} "Данные успешно вставлены"
// @Failure      400 {object} map[string]string "Ошибка в запросе"
// @Failure      401 {object} map[string]string "Отсутствует ключ API"
// @Failure      403 {object} map[string]string "Неверный ключ API"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/database/{table}/insert [post]
func (h *DatabaseHandler) InsertData(c *gin.Context) {
	tableName := c.Param("table")

	var req models.InsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Default().WithError(err).Warn("Database handler (InsertData): failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	rowsAffected, err := h.service.InsertData(tableName, req)
	if err != nil {
		logger.Default().WithError(err).Error("Database handler (InsertData): service returned an error")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Data inserted successfully",
		"rows_affected": rowsAffected,
	})
}

// UpdateData обрабатывает запрос на обновление данных в таблице.
// @Summary      Обновление данных в таблице
// @Description  Позволяет обновить записи в указанной таблице по заданным фильтрам.
// @Tags         database
// @Accept       json
// @Produce      json
// @Param        table path string true "Имя таблицы для обновления (e.g., tickets, doctors)"
// @Param        request body models.UpdateRequest true "Данные и фильтры для обновления"
// @Success      200 {object} map[string]interface{} "Данные успешно обновлены"
// @Failure      400 {object} map[string]string "Ошибка в запросе (например, обновление без фильтров)"
// @Failure      401 {object} map[string]string "Отсутствует ключ API"
// @Failure      403 {object} map[string]string "Неверный ключ API"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/database/{table}/update [patch]
func (h *DatabaseHandler) UpdateData(c *gin.Context) {
	tableName := c.Param("table")

	var req models.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Default().WithError(err).Warn("Database handler (UpdateData): failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	rowsAffected, err := h.service.UpdateData(tableName, req)
	if err != nil {
		logger.Default().WithError(err).Error("Database handler (UpdateData): service returned an error")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Data updated successfully",
		"rows_affected": rowsAffected,
	})
}

// DeleteData обрабатывает запрос на удаление данных из таблицы.
// @Summary      Удаление данных из таблицы
// @Description  Позволяет удалить записи из указанной таблицы по заданным фильтрам.
// @Tags         database
// @Accept       json
// @Produce      json
// @Param        table path string true "Имя таблицы для удаления (e.g., tickets, doctors)"
// @Param        request body models.DeleteRequest true "Фильтры для удаления"
// @Success      200 {object} map[string]interface{} "Данные успешно удалены"
// @Failure      400 {object} map[string]string "Ошибка в запросе (например, удаление без фильтров)"
// @Failure      401 {object} map[string]string "Отсутствует ключ API"
// @Failure      403 {object} map[string]string "Неверный ключ API"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/database/{table}/delete [delete]
func (h *DatabaseHandler) DeleteData(c *gin.Context) {
	tableName := c.Param("table")

	var req models.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Default().WithError(err).Warn("Database handler (DeleteData): failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	rowsAffected, err := h.service.DeleteData(tableName, req)
	if err != nil {
		logger.Default().WithError(err).Error("Database handler (DeleteData): service returned an error")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Data deleted successfully",
		"rows_affected": rowsAffected,
	})
}
