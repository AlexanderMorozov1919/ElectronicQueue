package handlers

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdHandler struct {
	service *services.AdService
}

func NewAdHandler(service *services.AdService) *AdHandler {
	return &AdHandler{service: service}
}

// GetAllAds godoc
// @Summary      Получить список всех рекламных материалов (Админ)
// @Description  Возвращает список всех рекламных материалов без самих изображений.
// @Tags         admin
// @Produce      json
// @Success      200 {array} models.AdResponse "Список рекламных материалов"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/admin/ads [get]
func (h *AdHandler) GetAllAds(c *gin.Context) {
	ads, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ads"})
		return
	}

	var response []models.AdResponse
	for _, ad := range ads {
		respAd := ad.ToResponse()
		respAd.Picture = ""
		respAd.Video = ""
		response = append(response, respAd)
	}
	c.JSON(http.StatusOK, response)
}

// GetAdByID godoc
// @Summary      Получить рекламный материал по ID (Админ)
// @Description  Возвращает полную информацию о рекламном материале, включая изображение.
// @Tags         admin
// @Produce      json
// @Param        id path int true "ID рекламного материала"
// @Success      200 {object} models.AdResponse "Рекламный материал"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      404 {object} map[string]string "Не найдено"
// @Security     ApiKeyAuth
// @Router       /api/admin/ads/{id} [get]
func (h *AdHandler) GetAdByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	ad, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ad not found"})
		return
	}
	c.JSON(http.StatusOK, ad.ToResponse())
}

// GetEnabledAds godoc
// @Summary      Получить активные рекламные материалы (Табло)
// @Description  Возвращает список всех включенных рекламных материалов с изображениями.
// @Tags         ads
// @Produce      json
// @Param        screen query string true "Тип экрана ('reception' или 'schedule')"
// @Success      200 {array} models.AdResponse "Список активных рекламных материалов"
// @Failure      400 {object} map[string]string "Неверный тип экрана"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/ads/enabled [get]
func (h *AdHandler) GetEnabledAds(c *gin.Context) {
	screenType := c.Query("screen")
	if screenType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'screen' is required"})
		return
	}

	ads, err := h.service.GetEnabled(screenType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get enabled ads"})
		return
	}
	var response []models.AdResponse
	for _, ad := range ads {
		response = append(response, ad.ToResponse())
	}
	c.JSON(http.StatusOK, response)
}

// CreateAd godoc
// @Summary      Создать рекламный материал (Админ)
// @Description  Загружает новое рекламное объявление.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request body models.CreateAdRequest true "Данные для создания"
// @Success      201 {object} models.AdResponse "Созданный материал"
// @Failure      400 {object} map[string]string "Ошибка в запросе"
// @Security     ApiKeyAuth
// @Router       /api/admin/ads [post]
func (h *AdHandler) CreateAd(c *gin.Context) {
	var req models.CreateAdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	ad, err := h.service.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ad.ToResponse())
}

// UpdateAd godoc
// @Summary      Обновить рекламный материал (Админ)
// @Description  Обновляет существующее рекламное объявление.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id path int true "ID рекламного материала"
// @Param        request body models.UpdateAdRequest true "Данные для обновления"
// @Success      200 {object} models.AdResponse "Обновленный материал"
// @Failure      400 {object} map[string]string "Ошибка в запросе"
// @Failure      404 {object} map[string]string "Не найдено"
// @Security     ApiKeyAuth
// @Router       /api/admin/ads/{id} [patch]
func (h *AdHandler) UpdateAd(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req models.UpdateAdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	ad, err := h.service.Update(uint(id), &req)
	if err != nil {
		if err.Error() == "ad with id "+c.Param("id")+" not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, ad.ToResponse())
}

// DeleteAd godoc
// @Summary      Удалить рекламный материал (Админ)
// @Description  Удаляет рекламное объявление по ID.
// @Tags         admin
// @Produce      json
// @Param        id path int true "ID рекламного материала"
// @Success      200 {object} map[string]string "Удалено"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      500 {object} map[string]string "Ошибка удаления"
// @Security     ApiKeyAuth
// @Router       /api/admin/ads/{id} [delete]
func (h *AdHandler) DeleteAd(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		logger.Default().WithError(err).Error("Failed to delete ad")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete ad"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ad deleted successfully"})
}
