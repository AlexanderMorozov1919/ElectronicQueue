package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"fmt"
	"time"
)

type TicketService struct {
	Repo repository.TicketRepository
}

func NewTicketService(repo repository.TicketRepository) *TicketService {
	return &TicketService{Repo: repo}
}

// CreateTicket создает новый тикет по услуге
func (s *TicketService) CreateTicket(service string) (*models.Ticket, error) {
	ticketNumber, err := s.generateTicketNumber(service)
	if err != nil {
		return nil, err
	}
	ticket := &models.Ticket{
		TicketNumber: ticketNumber,
		Status:       models.StatusWaiting,
		CreatedAt:    time.Now(),
	}
	err = s.Repo.Create(ticket)
	if err != nil {
		return nil, err
	}
	return ticket, nil
}

func (s *TicketService) generateTicketNumber(service string) (string, error) {
	letter := serviceToLetter(service)
	maxNum, err := s.Repo.GetMaxTicketNumber()
	if err != nil {
		return "", err
	}
	num := maxNum + 1
	if num > 1000 {
		num = 1
	}
	return letter + fmt.Sprintf("%03d", num), nil
}

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

// MapServiceIDToName возвращает русское название услуги по id
func (s *TicketService) MapServiceIDToName(id string) string {
	switch id {
	case "make_appointment":
		return "Записаться к врачу"
	case "confirm_appointment":
		return "Прием по записи"
	case "lab_tests":
		return "Сдать анализы"
	case "documents":
		return "Другой вопрос"
	default:
		return "Неизвестно"
	}
}
