package handlers

import (
	"ElectronicQueue/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// OneCHandler handles HTTP requests for 1C integration endpoints.
type OneCHandler struct {
	service *services.OneCService
}

// NewOneCHandler creates a new instance of OneCHandler.
func NewOneCHandler(service *services.OneCService) *OneCHandler {
	return &OneCHandler{service: service}
}

// GetSchedule proxies the request to the 1C service to get the general schedule.
// @Summary      Получить общее расписание из 1С
// @Description  Обращается к сервису-прокси 1С для получения общего расписания.
// @Tags         1C
// @Produce      json
// @Success      200 {object} object "Расписание из 1С"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/1c/getschedule [get]
func (h *OneCHandler) GetSchedule(c *gin.Context) {
	data, err := h.service.GetSchedule()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get schedule from 1C service: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetDoctorSchedule proxies the request to the 1C service to get a patient's appointment time.
// @Summary      Получить время записи пациента из 1С
// @Description  По номеру телефона пациента получает из 1С время, на которое он записан к врачу.
// @Tags         1C
// @Produce      json
// @Param        phone query string true "Номер телефона пациента"
// @Success      200 {object} object "Время записи пациента из 1С"
// @Failure      400 {object} map[string]string "Ошибка: отсутствует параметр 'phone'"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/1c/getdoctorschedule [get]
func (h *OneCHandler) GetDoctorSchedule(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'phone' is required"})
		return
	}

	data, err := h.service.GetDoctorSchedule(phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get doctor schedule from 1C service: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}
