package services

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DoctorService предоставляет методы для работы врача с талонами
type DoctorService struct {
	ticketRepo repository.TicketRepository
	doctorRepo repository.DoctorRepository
}

func NewDoctorService(ticketRepo repository.TicketRepository, doctorRepo repository.DoctorRepository) *DoctorService {
	return &DoctorService{
		ticketRepo: ticketRepo,
		doctorRepo: doctorRepo,
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

// GetCurrentAppointmentScreenState находит талон "на приеме" и врача для табло.
func (s *DoctorService) GetCurrentAppointmentScreenState() (*models.Doctor, *models.Ticket, error) {
	doctor, err := s.doctorRepo.GetAnyDoctor()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Default().WithError(err).Error("Error fetching a default doctor")
		}
		return nil, nil, fmt.Errorf("no active doctors found in the database: %w", err)
	}

	ticket, err := s.ticketRepo.FindFirstByStatus(models.StatusInProgress)
	if err != nil {
		// "запись не найдена" - нет талона на приеме.
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Default().WithError(err).Error("Error fetching current in-progress ticket")
		}
		return doctor, nil, nil
	}

	return doctor, ticket, nil
}
