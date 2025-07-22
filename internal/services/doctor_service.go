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
	ticketRepo   repository.TicketRepository
	doctorRepo   repository.DoctorRepository
	scheduleRepo repository.ScheduleRepository
}

// NewDoctorService создает новый экземпляр DoctorService.
func NewDoctorService(ticketRepo repository.TicketRepository, doctorRepo repository.DoctorRepository, scheduleRepo repository.ScheduleRepository) *DoctorService {
	return &DoctorService{
		ticketRepo:   ticketRepo,
		doctorRepo:   doctorRepo,
		scheduleRepo: scheduleRepo,
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

// GetCurrentAppointmentScreenState находит талон "на приеме" и врача для табло конкретного кабинета.
func (s *DoctorService) GetCurrentAppointmentScreenState(cabinetNumber int) (*models.Schedule, *models.Ticket, error) {
	// Ищем расписание (и врача) по номеру кабинета и текущему времени
	schedule, err := s.scheduleRepo.FindByCabinetAndCurrentTime(cabinetNumber)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Default().WithError(err).Error("Ошибка получения расписания для кабинета")
		}
		// Это не ошибка, просто в данный момент в кабинете нет приема по расписанию
		return nil, nil, fmt.Errorf("в кабинете %d в данный момент нет приема по расписанию", cabinetNumber)
	}

	// Ищем талон со статусом "на приеме" для конкретного кабинета через таблицу appointments
	ticket, err := s.ticketRepo.FindInProgressTicketForCabinet(cabinetNumber)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Default().WithError(err).WithField("cabinet", cabinetNumber).Error("Ошибка получения текущего талона в статусе 'на приеме' для кабинета")
		}
		// Возвращаем расписание (с врачом) без талона, т.к. в данный момент никого не принимают
		return schedule, nil, nil
	}

	return schedule, ticket, nil
}

// GetAllUniqueCabinets возвращает список всех уникальных кабинетов.
func (s *DoctorService) GetAllUniqueCabinets() ([]int, error) {
	cabinets, err := s.scheduleRepo.GetAllUniqueCabinets()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка всех кабинетов: %w", err)
	}
	return cabinets, nil
}
