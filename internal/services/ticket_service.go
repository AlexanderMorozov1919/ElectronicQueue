package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"fmt"
	"time"
)

const maxTicketNumber = 1000

var (
	serviceLetterMap = map[string]string{
		"make_appointment":    "A",
		"confirm_appointment": "B",
		"lab_tests":           "C",
		"documents":           "D",
	}
	serviceNameMap = map[string]string{
		"make_appointment":    "Записаться к врачу",
		"confirm_appointment": "Прием по записи",
		"lab_tests":           "Сдать анализы",
		"documents":           "Другой вопрос",
	}
)

// TicketService предоставляет методы для работы с талонами.
type TicketService struct {
	repo repository.TicketRepository
}

// NewTicketService создает новый экземпляр TicketService.
func NewTicketService(repo repository.TicketRepository) *TicketService {
	return &TicketService{repo: repo}
}

// CreateTicket создает новый талон для выбранной услуги.
func (s *TicketService) CreateTicket(serviceID string) (*models.Ticket, error) {
	if serviceID == "" {
		return nil, fmt.Errorf("serviceID is required")
	}
	ticketNumber, err := s.generateTicketNumber(serviceID)
	if err != nil {
		return nil, err
	}
	ticket := &models.Ticket{
		TicketNumber: ticketNumber,
		Status:       models.StatusWaiting,
		CreatedAt:    time.Now(),
	}
	if err := s.repo.Create(ticket); err != nil {
		return nil, err
	}
	return ticket, nil
}

// generateTicketNumber генерирует уникальный номер талона для услуги.
func (s *TicketService) generateTicketNumber(serviceID string) (string, error) {
	letter := s.getServiceLetter(serviceID)
	maxNum, err := s.repo.GetMaxTicketNumber()
	if err != nil {
		return "", err
	}
	num := maxNum + 1
	if num > maxTicketNumber {
		num = 1
	}
	return fmt.Sprintf("%s%03d", letter, num), nil
}

// getServiceLetter возвращает букву для услуги по её идентификатору.
func (s *TicketService) getServiceLetter(serviceID string) string {
	if letter, ok := serviceLetterMap[serviceID]; ok {
		return letter
	}
	return "Z"
}

// MapServiceIDToName возвращает название услуги по её идентификатору.
func (s *TicketService) MapServiceIDToName(serviceID string) string {
	if name, ok := serviceNameMap[serviceID]; ok {
		return name
	}
	return "Неизвестно"
}

// GetAllServices возвращает все доступные услуги (id и name)
func (s *TicketService) GetAllServices() []struct{ ID, Name string } {
	services := make([]struct{ ID, Name string }, 0, len(serviceNameMap))
	for id, name := range serviceNameMap {
		services = append(services, struct{ ID, Name string }{ID: id, Name: name})
	}
	return services
}
