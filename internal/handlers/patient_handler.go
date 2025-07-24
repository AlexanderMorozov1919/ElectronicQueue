package handlers

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PatientHandler struct {
	service *services.PatientService
}

func NewPatientHandler(service *services.PatientService) *PatientHandler {
	return &PatientHandler{service: service}
}

// SearchPatients godoc
// @Summary      Поиск пациентов по ФИО, ОМС или паспорту
// @Description  Ищет пациентов по частичному совпадению в ФИО, номере полиса ОМС или полному номеру паспорта (серия + номер без пробелов). Возвращает до 10 совпадений.
// @Tags         registrar
// @Produce      json
// @Param        query query string true "Строка для поиска (минимум 2 символа)"
// @Success      200 {array} models.Patient "Массив найденных пациентов"
// @Failure      400 {object} map[string]string "Ошибка: отсутствует или слишком короткий параметр поиска"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/registrar/patients/search [get]
func (h *PatientHandler) SearchPatients(c *gin.Context) {
	log := logger.Default()
	query := c.Query("query")
	if query == "" {
		log.Warn("SearchPatients: Query parameter 'query' is missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр 'query' для поиска обязателен"})
		return
	}

	// Добавим проверку на минимальную длину запроса для снижения нагрузки на БД
	if len(query) < 2 {
		c.JSON(http.StatusOK, []models.Patient{})
		return
	}

	patients, err := h.service.SearchPatients(query)
	if err != nil {
		log.WithError(err).Error("SearchPatients: Failed to search patients in service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось выполнить поиск пациентов"})
		return
	}

	// Всегда возвращаем JSON-массив, даже если он пустой, это лучшая практика для API.
	if patients == nil {
		c.JSON(http.StatusOK, []models.Patient{})
		return
	}

	c.JSON(http.StatusOK, patients)
}

// CreatePatient godoc
// @Summary      Создать нового пациента
// @Description  Создает новую запись о пациенте в базе данных. Используется, когда пациент не найден через поиск.
// @Tags         registrar
// @Accept       json
// @Produce      json
// @Param        patient body models.CreatePatientRequest true "Данные нового пациента"
// @Success      201 {object} models.Patient "Успешно созданный пациент"
// @Failure      400 {object} map[string]string "Ошибка: неверный формат запроса"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /api/registrar/patients [post]
func (h *PatientHandler) CreatePatient(c *gin.Context) {
	log := logger.Default()

	var req models.CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Warn("CreatePatient: Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса: " + err.Error()})
		return
	}

	patient, err := h.service.CreatePatient(&req)
	if err != nil {
		log.WithError(err).Error("CreatePatient: Failed to create patient in service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать пациента"})
		return
	}

	c.JSON(http.StatusCreated, patient)
}
