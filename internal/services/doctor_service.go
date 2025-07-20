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

// NewDoctorService создает новый экземпляр DoctorService.
func NewDoctorService(ticketRepo repository.TicketRepository, doctorRepo repository.DoctorRepository) *DoctorService {
	return &DoctorService{
		ticketRepo: ticketRepo,
		doctorRepo: doctorRepo,
	}
}

// GetAllActiveDoctors возвращает всех врачей со статусом is_active = true.
func (s *DoctorService) GetAllActiveDoctors() ([]models.Doctor, error) {
	doctors, err := s.doctorRepo.GetAll(true)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения активных врачей из репозитория: %w", err)
	}
	return doctors, nil
}

// GetRegisteredTickets возвращает талоны со статусом "зарегистрирован"
func (s *DoctorService) GetRegisteredTickets() ([]models.TicketResponse, error) {
	tickets, err := s.ticketRepo.FindByStatus(models.StatusRegistered)
	if err != nil {
		return nil, err
	}

	var response []models.TicketResponse
	for _, ticket := range tickets {
		response = append(response, ticket.ToResponse())
	}

	return response, nil
}

// GetInProgressTickets возвращает талоны со статусом "на_приеме"
func (s *DoctorService) GetInProgressTickets() ([]models.TicketResponse, error) {
	tickets, err := s.ticketRepo.FindByStatus(models.StatusInProgress)
	if err != nil {
		return nil, err
	}

	var response []models.TicketResponse
	for _, ticket := range tickets {
		response = append(response, ticket.ToResponse())
	}

	return response, nil
}

// StartAppointment начинает прием пациента
func (s *DoctorService) StartAppointment(ticketID uint) (*models.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ticketID)
	if err != nil {
		return nil, fmt.Errorf("талон не найден: %w", err)
	}

	if ticket.Status != models.StatusRegistered {
		return nil, fmt.Errorf("для начала приема талон должен иметь статус 'зарегистрирован'")
	}

	now := time.Now()
	ticket.Status = models.StatusInProgress
	ticket.StartedAt = &now

	if err := s.ticketRepo.Update(ticket); err != nil {
		return nil, fmt.Errorf("не удалось обновить талон: %w", err)
	}

	return ticket, nil
}

// CompleteAppointment завершает прием пациента
func (s *DoctorService) CompleteAppointment(ticketID uint) (*models.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ticketID)
	if err != nil {
		return nil, fmt.Errorf("талон не найден: %w", err)
	}

	if ticket.Status != models.StatusInProgress {
		return nil, fmt.Errorf("для завершения приема талон должен иметь статус 'на_приеме'")
	}

	now := time.Now()
	ticket.Status = models.StatusCompleted
	ticket.CompletedAt = &now

	if err := s.ticketRepo.Update(ticket); err != nil {
		return nil, fmt.Errorf("не удалось обновить талон: %w", err)
	}

	return ticket, nil
}

// GetCurrentAppointmentScreenState находит талон "на приеме" и врача для табло.
func (s *DoctorService) GetCurrentAppointmentScreenState() (*models.Doctor, *models.Ticket, error) {
	doctor, err := s.doctorRepo.GetAnyDoctor()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Default().WithError(err).Error("Ошибка получения врача по умолчанию")
		}
		return nil, nil, fmt.Errorf("в базе данных не найдены активные врачи: %w", err)
	}

	ticket, err := s.ticketRepo.FindFirstByStatus(models.StatusInProgress)
	if err != nil {
		// "запись не найдена" - это нормальная ситуация, когда нет пациента на приеме
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Default().WithError(err).Error("Ошибка получения текущего талона в статусе 'на приеме'")
		}
		// Возвращаем врача без талона
		return doctor, nil, nil
	}

	return doctor, ticket, nil
}
