package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"ElectronicQueue/internal/utils"
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

// GetAllServices возвращает все доступные услуги (id, name, letter)
func (s *TicketService) GetAllServices() []Service {
	return s.services
}

// GetByID возвращает тикет по строковому id
func (s *TicketService) GetByID(idStr string) (*models.Ticket, error) {
	var id uint
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		return nil, fmt.Errorf("invalid id")
	}
	return s.repo.GetByID(id)
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

// UpdateTicket обновляет тикет
func (s *TicketService) UpdateTicket(ticket *models.Ticket) error {
	return s.repo.Update(ticket)
}

// DeleteTicket удаляет тикет по строковому id
func (s *TicketService) DeleteTicket(idStr string) error {
	var id uint
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		return fmt.Errorf("invalid id")
	}
	return s.repo.Delete(id)
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

// Модификация существующего метода для использования нового генератора
func (s *TicketService) GenerateTicketImage(baseSize int, ticket *models.Ticket, serviceName string) ([]byte, error) {
	// Подготавливаем данные для талона
	data := utils.TicketData{
		ServiceName:  serviceName,
		TicketNumber: ticket.TicketNumber,
		DateTime:     ticket.CreatedAt,
	}

	// Данные для QR-кода
	qrData := []byte(fmt.Sprintf("Талон: %s\nВремя: %s\nУслуга: %s",
		ticket.TicketNumber,
		ticket.CreatedAt.Format("02.01.2006 15:04:05"),
		serviceName))

	// Генерируем изображение талона с заданным размером
	return utils.GenerateTicketImageWithSizes(baseSize, qrData, data)
}
