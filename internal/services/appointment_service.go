package services

import (
	"ElectronicQueue/internal/models"
	"ElectronicQueue/internal/repository"
	"fmt"
	"time"
)

// AppointmentService предоставляет методы для управления записями на прием.
type AppointmentService struct {
	repo repository.AppointmentRepository
}

// NewAppointmentService создает новый экземпляр AppointmentService.
func NewAppointmentService(repo repository.AppointmentRepository) *AppointmentService {
	return &AppointmentService{repo: repo}
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
	if req.ScheduleID == 0 || req.PatientID == 0 {
		return nil, fmt.Errorf("ScheduleID и PatientID являются обязательными полями")
	}

	appointment, err := s.repo.CreateAppointmentInTransaction(req)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запись на прием: %w", err)
	}
	return appointment, nil
}
