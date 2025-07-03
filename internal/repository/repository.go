package repository

import (
	"time"

	"github.com/AlexanderMorozov1919/ElectronicQueue/internal/models/appointment_model"
	"github.com/AlexanderMorozov1919/ElectronicQueue/internal/models/doctor_model"
	"github.com/AlexanderMorozov1919/ElectronicQueue/internal/models/patient_model"
	"github.com/AlexanderMorozov1919/ElectronicQueue/internal/models/schedule_model"
	"github.com/AlexanderMorozov1919/ElectronicQueue/internal/models/ticket_model"
	"gorm.io/gorm"
)

// DoctorRepository определяет методы для взаимодействия с данными врачей.
type DoctorRepository interface {
	Create(doctor *doctor_model.Doctor) error
	Update(doctor *doctor_model.Doctor) error
	Delete(id uint) error
	GetByID(id uint) (*doctor_model.Doctor, error)
	GetAll(onlyActive bool) ([]doctor_model.Doctor, error)
}

// PatientRepository определяет методы для взаимодействия с данными пациентов.
type PatientRepository interface {
	Create(patient *patient_model.Patient) error
	Update(patient *patient_model.Patient) error
	GetByID(id uint) (*patient_model.Patient, error)
	FindByPassport(series, number string) (*patient_model.Patient, error)
}

// TicketRepository определяет методы для взаимодействия с талонами.
type TicketRepository interface {
	Create(ticket *ticket_model.Ticket) error
	Update(ticket *ticket_model.Ticket) error
	GetByID(id uint) (*ticket_model.Ticket, error)
	FindByStatuses(statuses []ticket_model.TicketStatus) ([]ticket_model.Ticket, error)
}

// ScheduleRepository определяет методы для взаимодействия с расписанием.
type ScheduleRepository interface {
	Create(schedule *schedule_model.Schedule) error
	Update(schedule *schedule_model.Schedule) error
	GetByID(id uint) (*schedule_model.Schedule, error)
	FindByDoctorAndDate(doctorID uint, date time.Time) ([]schedule_model.Schedule, error)
}

// AppointmentRepository определяет методы для взаимодействия с записями на прием.
type AppointmentRepository interface {
	Create(appointment *appointment_model.Appointment) error
	Update(appointment *appointment_model.Appointment) error
	GetByID(id uint) (*appointment_model.Appointment, error)
	FindByTicketID(ticketID uint) (*appointment_model.Appointment, error)
}

// Repository содержит все репозитории приложения.
type Repository struct {
	Doctor      DoctorRepository
	Patient     PatientRepository
	Ticket      TicketRepository
	Schedule    ScheduleRepository
	Appointment AppointmentRepository
}

// NewRepository создает новый экземпляр главного репозитория.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Doctor:      NewDoctorRepository(db),
		Patient:     NewPatientRepository(db),
		Ticket:      NewTicketRepository(db),
		Schedule:    NewScheduleRepository(db),
		Appointment: NewAppointmentRepository(db),
	}
}
