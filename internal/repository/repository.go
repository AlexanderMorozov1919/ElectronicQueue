package repository

import (
	"ElectronicQueue/internal/models"
	"time"

	"gorm.io/gorm"
)

// ReceptionLogRepository определяет методы для работы с логами обслуживания.
type ReceptionLogRepository interface {
	Create(log *models.ReceptionLog) error
	Update(log *models.ReceptionLog) error
	FindActiveLogByTicketID(ticketID uint) (*models.ReceptionLog, error)
}

// DoctorRepository определяет методы для взаимодействия с данными врачей.
type DoctorRepository interface {
	Create(doctor *models.Doctor) error
	Update(doctor *models.Doctor) error
	Delete(id uint) error
	GetByID(id uint) (*models.Doctor, error)
	GetAll(onlyActive bool) ([]models.Doctor, error)
	GetAnyDoctor() (*models.Doctor, error)
	FindByLogin(login string) (*models.Doctor, error)
	UpdateStatus(doctorID uint, status models.DoctorStatus) error
}

// PatientRepository определяет методы для взаимодействия с данными пациентов.
type PatientRepository interface {
	Create(patient *models.Patient) (*models.Patient, error)
	Search(query string) ([]models.Patient, error)
	FindByPassport(series, number string) (*models.Patient, error)
	FindByPhone(phone string) (*models.Patient, error)
}

// TicketRepository определяет методы для взаимодействия с талонами.
type TicketRepository interface {
	Create(ticket *models.Ticket) error
	Update(ticket *models.Ticket) error
	GetByID(id uint) (*models.Ticket, error)
	FindByStatuses(statuses []models.TicketStatus) ([]models.Ticket, error)
	FindByStatus(status models.TicketStatus) ([]models.Ticket, error)
	GetNextWaitingTicket(categoryPrefixes []string) (*models.Ticket, error)
	GetMaxTicketNumberForPrefix(prefix string) (int, error)
	Delete(id uint) error
	FindInvitedByWindowNumber(windowNumber int) (*models.Ticket, error)
	FindInProgressTicketForCabinet(cabinetNumber int) (*models.Ticket, error)
	FindTicketsForCabinetQueue(cabinetNumber int) ([]models.DoctorQueueTicketResponse, error)
	FindByStatusAndDoctor(status models.TicketStatus, doctorID uint) ([]models.Ticket, error)
	GetDailyReport(date time.Time) ([]models.DailyReportRow, error)
	FindForRegistrar(statuses []models.TicketStatus, categoryPrefixes []string) ([]models.RegistrarTicketResponse, error)
}

// ScheduleRepository определяет методы для взаимодействия с расписанием.
type ScheduleRepository interface {
	Create(schedule *models.Schedule) error
	Update(schedule *models.Schedule) error
	Delete(id uint) error
	GetByID(id uint) (*models.Schedule, error)
	FindByDoctorAndDate(doctorID uint, date time.Time) ([]models.Schedule, error)
	FindByCabinetAndCurrentTime(cabinetNumber int) (*models.Schedule, error)
	GetAllUniqueCabinets() ([]int, error)
	FindFirstScheduleForCabinetByDay(cabinetNumber int) (*models.Schedule, error)
	FindAllSchedulesForDate(date time.Time) ([]models.Schedule, error)
	FindMinMaxTimesForDate(date time.Time) (time.Time, time.Time, error)
}

// AppointmentRepository определяет методы для взаимодействия с записями на прием.
type AppointmentRepository interface {
	CreateAppointmentInTransaction(req *models.CreateAppointmentRequest) (*models.Appointment, error)
	FindScheduleAndAppointmentsByDoctorAndDate(doctorID uint, date time.Time) ([]models.ScheduleWithAppointmentInfo, error)
	FindByID(id uint) (*models.Appointment, error)
	FindByPatientID(patientID uint) ([]models.Appointment, error)
	Update(appointment *models.Appointment) error
	DeleteAppointmentAndFreeSlot(appointmentID uint) error
	FindUpcomingByPatientID(patientID uint, now time.Time) (*models.Appointment, error)
	AssignTicketToAppointment(appointment *models.Appointment, ticket *models.Ticket) error
}

// RegistrarRepository определяет методы для аутентификации регистраторов.
type RegistrarRepository interface {
	FindByLogin(login string) (*models.Registrar, error)
	Create(registrar *models.Registrar) error
}

// AdministratorRepository определяет методы для аутентификации администраторов.
type AdministratorRepository interface {
	FindByLogin(login string) (*models.Administrator, error)
	Create(admin *models.Administrator) error
}

// ServiceRepository определяет методы для работы с услугами терминала.
type ServiceRepository interface {
	GetAll() ([]models.Service, error)
	GetByID(id uint) (*models.Service, error)
	GetByServiceID(serviceID string) (*models.Service, error)
	Create(service *models.Service) error
	Update(service *models.Service) error
	Delete(id uint) error
}

// CleanupRepository определяет методы для очистки данных.
type CleanupRepository interface {
	TruncateTickets() error
	GetTicketsCount() (int64, error)
	GetOrphanedAppointmentsCount() (int64, error)
}

// BusinessProcessRepository определяет методы для управления бизнес-процессами.
type BusinessProcessRepository interface {
	GetAll() ([]models.BusinessProcess, error)
	Update(process *models.BusinessProcess) error
	FindByName(name string) (*models.BusinessProcess, error)
}

// AdRepository определяет методы для работы с рекламными материалами.
type AdRepository interface {
	Create(ad *models.Ad) error
	GetAll() ([]models.Ad, error)
	GetEnabledFor(screen string) ([]models.Ad, error)
	GetByID(id uint) (*models.Ad, error)
	Update(ad *models.Ad) error
	Delete(id uint) error
}

// RegistrarPriorityRepository определяет методы для работы с приоритетами регистратора.
type RegistrarPriorityRepository interface {
	GetPriorities(registrarID uint) ([]models.Service, error)
	SetPriorities(registrarID uint, serviceIDs []uint) error
}

// Repository содержит все репозитории приложения.
type Repository struct {
	Doctor            DoctorRepository
	Patient           PatientRepository
	Ticket            TicketRepository
	Schedule          ScheduleRepository
	Appointment       AppointmentRepository
	Service           ServiceRepository
	Registrar         RegistrarRepository
	Administrator     AdministratorRepository
	Cleanup           CleanupRepository
	BusinessProcess   BusinessProcessRepository
	ReceptionLog      ReceptionLogRepository
	Ad                AdRepository
	RegistrarPriority RegistrarPriorityRepository
}

// NewRepository создает новый экземпляр главного репозитория.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Doctor:            NewDoctorRepository(db),
		Patient:           NewPatientRepository(db),
		Ticket:            NewTicketRepository(db),
		Schedule:          NewScheduleRepository(db),
		Appointment:       NewAppointmentRepository(db),
		Service:           NewServiceRepository(db),
		Registrar:         NewRegistrarRepository(db),
		Administrator:     NewAdministratorRepository(db),
		Cleanup:           NewCleanupRepository(db),
		BusinessProcess:   NewBusinessProcessRepository(db),
		ReceptionLog:      NewReceptionLogRepository(db),
		Ad:                NewAdRepository(db),
		RegistrarPriority: NewRegistrarPriorityRepository(db),
	}
}
