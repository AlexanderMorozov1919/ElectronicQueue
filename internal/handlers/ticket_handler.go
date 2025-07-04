package handlers

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// TicketHandler содержит зависимости для работы с талонами
// (можно расширить при необходимости)
type TicketHandler struct {
	Repo repository.TicketRepository
}

// NewTicketHandler создает новый TicketHandler
func NewTicketHandler(repo repository.TicketRepository) *TicketHandler {
	return &TicketHandler{Repo: repo}
}

// GetServicePage - /terminal/service (GET)
func (h *TicketHandler) GetServicePage(c *gin.Context) {
	c.File(filepath.Join("frontend", "index.html"))
}

// GetSelectServicePage - /terminal/service/select (GET)
func (h *TicketHandler) GetSelectServicePage(c *gin.Context) {
	c.File(filepath.Join("frontend", "select.html"))
}

// HandleService - экспортируемый обработчик для создания талона
func (h *TicketHandler) HandleService(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ticketNumber, err := generateTicketNumber(h.Repo, service)
		if err != nil {
			c.String(500, "Ошибка генерации номера талона")
			return
		}
		ticket := &models.Ticket{
			TicketNumber: ticketNumber,
			Status:       models.StatusWaiting,
			CreatedAt:    time.Now(),
		}
		h.Repo.Create(ticket)
		c.File(filepath.Join("frontend", "success.html"))
	}
}

// generateTicketNumber - генерация номера талона: буква (по услуге) + число (1-1000, глобально)
func generateTicketNumber(repo repository.TicketRepository, service string) (string, error) {
	letter := serviceToLetter(service)
	maxNum, err := repo.GetMaxTicketNumber()
	if err != nil {
		return "", err
	}
	num := maxNum + 1
	if num > 1000 {
		num = 1
	}
	return letter + fmt.Sprintf("%03d", num), nil
}

// serviceToLetter возвращает букву по типу услуги
func serviceToLetter(service string) string {
	switch service {
	case "make_appointment":
		return "A"
	case "confirm_appointment":
		return "B"
	case "lab_tests":
		return "C"
	case "documents":
		return "D"
	default:
		return "Z"
	}
}
