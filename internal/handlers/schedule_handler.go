// D:\vs\go\ElectronicQueue\internal\handlers\schedule_handler.go

package handlers

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/pubsub"
	"ElectronicQueue/internal/services"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	// Удален неиспользуемый импорт
	// "strings"

	"github.com/gin-gonic/gin"
)

type ScheduleHandler struct {
	service *services.ScheduleService
	broker  *pubsub.Broker
}

func NewScheduleHandler(service *services.ScheduleService, broker *pubsub.Broker) *ScheduleHandler {
	return &ScheduleHandler{service: service, broker: broker}
}

// ... (методы CreateSchedule и DeleteSchedule остаются без изменений) ...
func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	log := logger.Default()
	var req models.CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Warn("CreateSchedule: Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса: " + err.Error()})
		return
	}

	schedule, err := h.service.CreateSchedule(&req)
	if err != nil {
		log.WithError(err).Error("CreateSchedule: Failed to create schedule in service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, schedule)
}

func (h *ScheduleHandler) DeleteSchedule(c *gin.Context) {
	log := logger.Default()
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		log.WithError(err).Warn("DeleteSchedule: Invalid ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID"})
		return
	}

	err = h.service.DeleteSchedule(uint(id))
	if err != nil {
		log.WithError(err).Error("DeleteSchedule: Failed to delete schedule from service")
		if err.Error() == "слот расписания с ID "+idStr+" не найден" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Слот расписания успешно удален"})
}

// --- ЛОГИКА МЕТОДА ПОЛНОСТЬЮ ИЗМЕНЕНА ---
// GetTodayScheduleUpdates godoc
// @Summary      Получить обновления расписания на сегодня
// @Description  Отправляет начальное состояние расписания (`event: schedule_initial`) и последующие полные обновления (`event: schedule_initial`) через Server-Sent Events.
// @Tags         schedule
// @Produce      text/event-stream
// @Success      200 {object} services.TodayScheduleResponse "Поток событий с состоянием расписания"
// @Router       /api/schedules/today/updates [get]
func (h *ScheduleHandler) GetTodayScheduleUpdates(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	log := logger.Default().WithField("module", "SSE_SCHEDULE")

	clientChan := h.broker.Subscribe()
	defer h.broker.Unsubscribe(clientChan)

	// --- 1. Отправка начального состояния из кэша ---
	initialState, err := h.service.GetTodayScheduleState()
	if err != nil {
		log.WithError(err).Error("Критическая ошибка в GetTodayScheduleState")
		c.SSEvent("error", gin.H{"error": err.Error()})
		if f, ok := c.Writer.(http.Flusher); ok {
			f.Flush()
		}
		return
	}

	log.Info("Отправка начального состояния расписания из кэша.")
	// Отправляем как json.RawMessage, чтобы избежать двойной сериализации
	c.SSEvent("schedule_initial", json.RawMessage(initialState))
	if f, ok := c.Writer.(http.Flusher); ok {
		f.Flush()
	}

	// --- 2. Ожидание и отправка обновлений ---
	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-clientChan:
			if !ok {
				log.Info("Канал уведомлений закрыт для расписания.")
				return false
			}

			// Простая проверка, что сообщение похоже на JSON объект
			if len(msg) < 2 || msg[0] != '{' {
				return true
			}

			log.Info("Получено новое расписание, отправка полного обновления клиенту.")
			// Каждое обновление от поллера - это полный новый JSON расписания.
			// Мы отправляем его как 'schedule_initial', чтобы frontend просто заменил все данные.
			c.SSEvent("schedule_initial", json.RawMessage(msg))

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			return true

		case <-c.Request.Context().Done():
			log.Info("Клиент отключился от расписания.")
			return false
		}
	})
}
