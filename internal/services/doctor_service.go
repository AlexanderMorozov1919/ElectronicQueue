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

// GetDoctorScreenState находит расписание врача и полную очередь к его кабинету.
// Если расписание на сегодня не найдено, возвращает nil для schedule и пустую очередь, но без ошибки.
func (s *DoctorService) GetDoctorScreenState(cabinetNumber int) (*models.Schedule, []models.DoctorQueueTicketResponse, error) {
	schedule, err := s.scheduleRepo.FindFirstScheduleForCabinetByDay(cabinetNumber)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// Если произошла реальная ошибка БД, а не просто "не найдено", возвращаем её.
		logger.Default().WithError(err).Error("Ошибка получения расписания для кабинета")
		return nil, nil, err
	}

	// Если расписание найдено, ищем очередь.
	if schedule != nil {
		queue, err := s.ticketRepo.FindTicketsForCabinetQueue(cabinetNumber)
		if err != nil {
			logger.Default().WithError(err).WithField("cabinet", cabinetNumber).Error("Ошибка получения очереди к кабинету")
			// В случае ошибки получения очереди, возвращаем пустую очередь, но с данными о враче.
			return schedule, []models.DoctorQueueTicketResponse{}, nil
		}
		return schedule, queue, nil
	}

	// Если расписание не найдено (gorm.ErrRecordNotFound), возвращаем nil и пустую очередь.
	return nil, []models.DoctorQueueTicketResponse{}, nil
}

// GetAllUniqueCabinets возвращает список всех уникальных кабинетов.
func (s *DoctorService) GetAllUniqueCabinets() ([]int, error) {
	cabinets, err := s.scheduleRepo.GetAllUniqueCabinets()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка всех кабинетов: %w", err)
	}
	return cabinets, nil
}
