package services

import (
	"ElectronicQueue/internal/logger"
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/pubsub"
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
	broker       *pubsub.Broker
}

// NewDoctorService создает новый экземпляр DoctorService.
func NewDoctorService(ticketRepo repository.TicketRepository, doctorRepo repository.DoctorRepository, scheduleRepo repository.ScheduleRepository, broker *pubsub.Broker) *DoctorService {
	return &DoctorService{
		ticketRepo:   ticketRepo,
		doctorRepo:   doctorRepo,
		scheduleRepo: scheduleRepo,
		broker:       broker,
	}
}

// GetAllActiveDoctors возвращает всех врачей для использования в выпадающих списках.
func (s *DoctorService) GetAllActiveDoctors() ([]models.Doctor, error) {
	// Установлено в false, чтобы получать всех врачей, а не только активных.
	doctors, err := s.doctorRepo.GetAll(false)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка врачей из репозитория: %w", err)
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

// Получить очередь к врачу (только его талоны)
func (s *DoctorService) GetRegisteredTicketsForDoctor(doctorID uint) ([]models.TicketResponse, error) {
	tickets, err := s.ticketRepo.FindByStatusAndDoctor(models.StatusRegistered, doctorID)
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

// Получить талоны на приеме у врача
func (s *DoctorService) GetInProgressTicketsForDoctor(doctorID uint) ([]models.TicketResponse, error) {
	tickets, err := s.ticketRepo.FindByStatusAndDoctor(models.StatusInProgress, doctorID)
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

// StartBreak начинает перерыв врача
func (s *DoctorService) StartBreak(doctorID uint) error {
	log := logger.Default().WithField("service", "StartBreak").WithField("doctor_id", doctorID)

	doctor, err := s.doctorRepo.GetByID(doctorID)
	if err != nil {
		log.WithError(err).Error("Врач не найден")
		return fmt.Errorf("врач не найден: %w", err)
	}

	if doctor.Status != models.DoctorStatusActive {
		log.WithField("current_status", doctor.Status).Error("Врач должен быть активен для начала перерыва")
		return fmt.Errorf("врач должен быть активен для начала перерыва")
	}

	if err := s.doctorRepo.UpdateStatus(doctorID, models.DoctorStatusOnBreak); err != nil {
		log.WithError(err).Error("Не удалось обновить статус врача")
		return fmt.Errorf("не удалось обновить статус врача: %w", err)
	}

	s.broker.Publish("doctor_status_update")
	log.Info("Перерыв начат успешно")
	return nil
}

// EndBreak завершает перерыв врача
func (s *DoctorService) EndBreak(doctorID uint) error {
	log := logger.Default().WithField("service", "EndBreak").WithField("doctor_id", doctorID)

	doctor, err := s.doctorRepo.GetByID(doctorID)
	if err != nil {
		log.WithError(err).Error("Врач не найден")
		return fmt.Errorf("врач не найден: %w", err)
	}

	if doctor.Status != models.DoctorStatusOnBreak {
		log.WithField("current_status", doctor.Status).Error("Врач должен быть на перерыве для его завершения")
		return fmt.Errorf("врач должен быть на перерыве для его завершения")
	}

	if err := s.doctorRepo.UpdateStatus(doctorID, models.DoctorStatusActive); err != nil {
		log.WithError(err).Error("Не удалось обновить статус врача")
		return fmt.Errorf("не удалось обновить статус врача: %w", err)
	}

	s.broker.Publish("doctor_status_update")
	log.Info("Перерыв завершен успешно")
	return nil
}

// SetDoctorActive устанавливает статус врача как активный (при входе в систему)
func (s *DoctorService) SetDoctorActive(doctorID uint) error {
	log := logger.Default().WithField("service", "SetDoctorActive").WithField("doctor_id", doctorID)

	if err := s.doctorRepo.UpdateStatus(doctorID, models.DoctorStatusActive); err != nil {
		log.WithError(err).Error("Не удалось установить статус активен")
		return fmt.Errorf("не удалось установить статус активен: %w", err)
	}

	s.broker.Publish("doctor_status_update")
	log.Info("Статус врача установлен как активен")
	return nil
}

// SetDoctorInactive устанавливает статус врача как неактивный (при выходе из системы)
func (s *DoctorService) SetDoctorInactive(doctorID uint) error {
	log := logger.Default().WithField("service", "SetDoctorInactive").WithField("doctor_id", doctorID)

	if err := s.doctorRepo.UpdateStatus(doctorID, models.DoctorStatusInactive); err != nil {
		log.WithError(err).Error("Не удалось установить статус неактивен")
		return fmt.Errorf("не удалось установить статус неактивен: %w", err)
	}

	s.broker.Publish("doctor_status_update")
	log.Info("Статус врача установлен как неактивен")
	return nil
}
