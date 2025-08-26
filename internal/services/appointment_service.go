package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"fmt"
	"time"
)

// AppointmentDetailsResponse определяет детальную информацию о записи для истории.
type AppointmentDetailsResponse struct {
	AppointmentID uint    `json:"appointment_id"`
	Date          string  `json:"date"`
	StartTime     string  `json:"start_time"`
	Cabinet       *int    `json:"cabinet"`
	DoctorName    string  `json:"doctor_name"`
	DoctorSpec    string  `json:"doctor_specialization"`
	PatientName   string  `json:"patient_name"`
	TicketNumber  *string `json:"ticket_number"`
	IsFuture      bool    `json:"is_future"`
}

// AppointmentService предоставляет методы для управления записями на прием.
type AppointmentService struct {
	repo       repository.AppointmentRepository
	ticketRepo repository.TicketRepository
}

// NewAppointmentService создает новый экземпляр AppointmentService.
func NewAppointmentService(repo repository.AppointmentRepository, ticketRepo repository.TicketRepository) *AppointmentService {
	return &AppointmentService{repo: repo, ticketRepo: ticketRepo}
}

// GetDoctorScheduleWithAppointments получает расписание врача вместе с информацией о существующих записях.
func (s *AppointmentService) GetDoctorScheduleWithAppointments(doctorID uint, date time.Time) ([]models.ScheduleWithAppointmentInfo, error) {
	schedule, err := s.repo.FindScheduleAndAppointmentsByDoctorAndDate(doctorID, date)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить расписание из репозитория: %w", err)
	}
	return schedule, nil
}

// CreateAppointment обрабатывает логику создания новой записи.
// Основная работа (транзакция) выполняется в репозитории.
func (s *AppointmentService) CreateAppointment(req *models.CreateAppointmentRequest) (*models.Appointment, error) {
	if req.ScheduleID == 0 {
		return nil, fmt.Errorf("ScheduleID является обязательным полем")
	}

	appointment, err := s.repo.CreateAppointmentInTransaction(req)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запись на прием: %w", err)
	}
	return appointment, nil
}

// GetAppointmentsByPatient получает историю записей и преобразует в DTO.
func (s *AppointmentService) GetAppointmentsByPatient(patientID uint) ([]AppointmentDetailsResponse, error) {
	appointments, err := s.repo.FindByPatientID(patientID)
	if err != nil {
		return nil, err
	}

	var response []AppointmentDetailsResponse
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for _, app := range appointments {
		var ticketNum *string
		if app.TicketID != nil {
			ticketNum = &app.Ticket.TicketNumber
		}

		isFuture := false
		appointmentDate := app.Schedule.Date
		if appointmentDate.After(today) || appointmentDate.Equal(today) {
			isFuture = true
		}

		details := AppointmentDetailsResponse{
			AppointmentID: app.ID,
			Date:          app.Schedule.Date.Format("2006-01-02"),
			StartTime:     app.Schedule.StartTime,
			Cabinet:       app.Schedule.Cabinet,
			DoctorName:    app.Schedule.Doctor.FullName,
			DoctorSpec:    app.Schedule.Doctor.Specialization,
			PatientName:   app.Patient.FullName,
			TicketNumber:  ticketNum,
			IsFuture:      isFuture,
		}
		response = append(response, details)
	}
	return response, nil
}

// DeleteAppointment удаляет запись.
func (s *AppointmentService) DeleteAppointment(appointmentID uint) error {
	return s.repo.DeleteAppointmentAndFreeSlot(appointmentID)
}

// ConfirmAppointment подтверждает явку по записи.
func (s *AppointmentService) ConfirmAppointment(appointmentID, ticketID uint) (*models.Appointment, error) {
	appointment, err := s.repo.FindByID(appointmentID)
	if err != nil {
		return nil, fmt.Errorf("запись не найдена: %w", err)
	}

	if appointment.TicketID != nil {
		return nil, fmt.Errorf("запись уже подтверждена и привязана к талону")
	}

	ticket, err := s.ticketRepo.GetByID(ticketID)
	if err != nil {
		return nil, fmt.Errorf("талон не найден: %w", err)
	}

	if ticket.Status == models.StatusRegistered || ticket.Status == models.StatusInProgress {
		return nil, fmt.Errorf("этот талон уже используется")
	}

	appointment.TicketID = &ticketID
	if err := s.repo.Update(appointment); err != nil {
		return nil, fmt.Errorf("не удалось обновить запись: %w", err)
	}

	ticket.Status = models.StatusRegistered
	if err := s.ticketRepo.Update(ticket); err != nil {
		appointment.TicketID = nil
		s.repo.Update(appointment)
		return nil, fmt.Errorf("не удалось обновить статус талона: %w", err)
	}

	return appointment, nil
}
