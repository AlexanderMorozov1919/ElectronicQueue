package handlers

import (
	"ElectronicQueue/internal/config"
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AudioHandler struct {
	cfg *config.Config // Добавлено поле для хранения конфига
}

func NewAudioHandler(cfg *config.Config) *AudioHandler {
	return &AudioHandler{cfg: cfg} // Обновлен конструктор
}

// GenerateAnnouncement создает и отдает WAV файл с озвучкой талона.
// @Summary      Сгенерировать звуковое оповещение
// @Description  Создает и возвращает WAV файл с озвучкой номера талона и окна.
// @Tags         audio
// @Produce      audio/wav
// @Param        ticket query string true "Номер талона (например, A007 или C21)"
// @Param        window query string true "Номер окна (например, 5)"
// @Success      200 {file} file "WAV файл оповещения"
// @Failure      400 {object} map[string]string "Ошибка: неверные параметры"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/audio/announce [get]
func (h *AudioHandler) GenerateAnnouncement(c *gin.Context) {
	log := logger.Default()
	ticketNumber := c.Query("ticket")
	windowNumber := c.Query("window")

	if ticketNumber == "" || windowNumber == "" {
		log.Warn("Audio handler: ticket or window parameter is missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметры 'ticket' и 'window' обязательны"})
		return
	}
	wavBytes, err := utils.GenerateAnnouncementWav(ticketNumber, windowNumber, "assets/audio", h.cfg.AudioBackgroundMusicEnabled)
	if err != nil {
		log.WithError(err).Error("Audio handler: failed to generate WAV file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сгенерировать аудиофайл: " + err.Error()})
		return
	}

	c.Header("Content-Type", "audio/wav")
	c.Header("Content-Disposition", `inline; filename="announcement.wav"`)
	c.Data(http.StatusOK, "audio/wav", wavBytes)
}
