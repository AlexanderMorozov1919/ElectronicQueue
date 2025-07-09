package repository

import (
	"ElectronicQueue/internal/models"
	"time"

	"gorm.io/gorm"
)

// DoctorRepository определяет методы для взаимодействия с данными врачей.
type DoctorRepository interface {
	Create(doctor *models.Doctor) error
	Update(doctor *models.Doctor) error
	Delete(id uint) error
	GetByID(id uint) (*models.Doctor, error)
	GetAll(onlyActive bool) ([]models.Doctor, error)
}

// PatientRepository определяет методы для взаимодействия с данными пациентов.
type PatientRepository interface {
	Create(patient *models.Patient) error
	Update(patient *models.Patient) error
	GetByID(id uint) (*models.Patient, error)
	FindByPassport(series, number string) (*models.Patient, error)
}

// TicketRepository определяет методы для взаимодействия с талонами.
type TicketRepository interface {
	Create(ticket *models.Ticket) error
	Update(ticket *models.Ticket) error
	GetByID(id uint) (*models.Ticket, error)
	FindByStatuses(statuses []models.TicketStatus) ([]models.Ticket, error)
	GetNextWaitingTicket() (*models.Ticket, error)
	GetMaxTicketNumber() (int, error)
	Delete(id uint) error
}

// ScheduleRepository определяет методы для взаимодействия с расписанием.
type ScheduleRepository interface {
	Create(schedule *models.Schedule) error
	Update(schedule *models.Schedule) error
	GetByID(id uint) (*models.Schedule, error)
	FindByDoctorAndDate(doctorID uint, date time.Time) ([]models.Schedule, error)
}

// AppointmentRepository определяет методы для взаимодействия с записями на прием.
type AppointmentRepository interface {
	Create(appointment *models.Appointment) error
	Update(appointment *models.Appointment) error
	GetByID(id uint) (*models.Appointment, error)
}

// Repository содержит все репозитории приложения.
type Repository struct {
	Doctor      DoctorRepository
	Patient     PatientRepository
	Ticket      TicketRepository
	Schedule    ScheduleRepository
	Appointment AppointmentRepository
	Service     ServiceRepository // добавлен репозиторий услуг
}

// NewRepository создает новый экземпляр главного репозитория.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Doctor:      NewDoctorRepository(db),
		Patient:     NewPatientRepository(db),
		Ticket:      NewTicketRepository(db),
		Schedule:    NewScheduleRepository(db),
		Appointment: NewAppointmentRepository(db),
		Service:     NewServiceRepository(db), // добавлен репозиторий услуг
	}
}
