package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"fmt"
	"time"
)

const maxTicketNumber = 1000

// Service описывает услугу
// ID — уникальный идентификатор, Name — русское название
// Letter — буква для талона
type Service struct {
	ID     string
	Name   string
	Letter string
}

// TicketService предоставляет методы для работы с талонами
type TicketService struct {
	repo     repository.TicketRepository
	services []Service
}

// NewTicketService создает новый экземпляр TicketService
func NewTicketService(repo repository.TicketRepository) *TicketService {
	serviceList := []Service{
		{ID: "make_appointment", Name: "Записаться к врачу"},
		{ID: "confirm_appointment", Name: "Прием по записи"},
		{ID: "lab_tests", Name: "Сдать анализы"},
		{ID: "documents", Name: "Другой вопрос"},
	}
	alphabet := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	for i := range serviceList {
		if i < len(alphabet) {
			serviceList[i].Letter = string(alphabet[i])
		} else {
			serviceList[i].Letter = "Z"
		}
	}
	return &TicketService{repo: repo, services: serviceList}
}

// CreateTicket создает новый талон для выбранной услуги
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

// generateTicketNumber генерирует уникальный номер талона для услуги
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

// getServiceLetter возвращает букву для услуги по её идентификатору
func (s *TicketService) getServiceLetter(serviceID string) string {
	for _, svc := range s.services {
		if svc.ID == serviceID {
			return svc.Letter
		}
	}
	return "Z"
}

// MapServiceIDToName возвращает название услуги по её идентификатору
func (s *TicketService) MapServiceIDToName(serviceID string) string {
	for _, svc := range s.services {
		if svc.ID == serviceID {
			return svc.Name
		}
	}
	return "Неизвестно"
}

// GetAllServices возвращает все доступные услуги (id, name, letter)
func (s *TicketService) GetAllServices() []Service {
	return s.services
}
