package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"fmt"
	"time"
)

// DoctorService предоставляет методы для работы врача с талонами
type DoctorService struct {
	ticketRepo repository.TicketRepository
}

func NewDoctorService(ticketRepo repository.TicketRepository) *DoctorService {
	return &DoctorService{
		ticketRepo: ticketRepo,
	}
}

// StartAppointment начинает прием пациента
// Изменяет статус талона на "на_приеме" и фиксирует время начала
func (s *DoctorService) StartAppointment(ticketID uint) (*models.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ticketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	if ticket.Status != models.StatusRegistered {
		return nil, fmt.Errorf("ticket must be in 'зарегистрирован' status to start appointment")
	}

	// Обновляем статус и время начала приема
	now := time.Now()
	ticket.Status = models.StatusInProgress
	ticket.StartedAt = &now

	// Сохраняем изменения в базе данных
	if err := s.ticketRepo.Update(ticket); err != nil {
		return nil, fmt.Errorf("failed to update ticket: %w", err)
	}

	return ticket, nil
}

// CompleteAppointment завершает прием пациента
// Изменяет статус талона на "завершен" и фиксирует время завершения
func (s *DoctorService) CompleteAppointment(ticketID uint) (*models.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ticketID)
	if err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	// Проверяем, что талон в статусе "на_приеме"
	if ticket.Status != models.StatusInProgress {
		return nil, fmt.Errorf("ticket must be in 'на_приеме' status to complete appointment")
	}

	now := time.Now()
	ticket.Status = models.StatusCompleted
	ticket.CompletedAt = &now

	if err := s.ticketRepo.Update(ticket); err != nil {
		return nil, fmt.Errorf("failed to update ticket: %w", err)
	}

	return ticket, nil
}
