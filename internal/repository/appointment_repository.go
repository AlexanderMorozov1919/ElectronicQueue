package repository

import (
	"ElectronicQueue/internal/models"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type appointmentRepo struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return &appointmentRepo{db: db}
}

// CreateAppointmentInTransaction создает запись и блокирует слот в рамках одной транзакции.
func (r *appointmentRepo) CreateAppointmentInTransaction(req *models.CreateAppointmentRequest) (*models.Appointment, error) {
	var appointment models.Appointment
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var schedule models.Schedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&schedule, req.ScheduleID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("указанный слот в расписании не найден")
			}
			return err
		}

		if !schedule.IsAvailable {
			return errors.New("выбранное время уже занято")
		}

		appointment = models.Appointment{
			ScheduleID: req.ScheduleID,
			PatientID:  req.PatientID,
			TicketID:   req.TicketID,
		}
		if err := tx.Create(&appointment).Error; err != nil {
			return err
		}

		schedule.IsAvailable = false
		if err := tx.Save(&schedule).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if err := r.db.Preload("Patient").Preload("Schedule.Doctor").First(&appointment, appointment.ID).Error; err != nil {
		return nil, err
	}

	return &appointment, nil
}

// FindScheduleAndAppointmentsByDoctorAndDate находит расписание и связанные с ним записи.
func (r *appointmentRepo) FindScheduleAndAppointmentsByDoctorAndDate(doctorID uint, date time.Time) ([]models.ScheduleWithAppointmentInfo, error) {
	var schedules []models.Schedule
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := r.db.Where("doctor_id = ? AND date >= ? AND date < ?", doctorID, startOfDay, endOfDay).Order("start_time asc").Find(&schedules).Error; err != nil {
		return nil, err
	}

	if len(schedules) == 0 {
		return []models.ScheduleWithAppointmentInfo{}, nil
	}

	var result []models.ScheduleWithAppointmentInfo
	for _, s := range schedules {
		info := models.ScheduleWithAppointmentInfo{Schedule: s}
		if !s.IsAvailable {
			var app models.Appointment
			err := r.db.Preload("Patient").Preload("Ticket").Where("schedule_id = ?", s.ID).First(&app).Error
			if err == nil {
				info.Appointment = &app
				if app.TicketID != nil {
					info.TicketNumber = &app.Ticket.TicketNumber
				}
			}
		}
		result = append(result, info)
	}

	return result, nil
}

// FindByID находит запись по ID со всеми связанными данными.
func (r *appointmentRepo) FindByID(id uint) (*models.Appointment, error) {
	var appointment models.Appointment
	err := r.db.Preload("Patient").Preload("Schedule.Doctor").Preload("Ticket").First(&appointment, id).Error
	return &appointment, err
}

// FindByPatientID находит все записи пациента.
func (r *appointmentRepo) FindByPatientID(patientID uint) ([]models.Appointment, error) {
	var appointments []models.Appointment
	err := r.db.Preload("Schedule.Doctor").Preload("Ticket").
		Where("patient_id = ?", patientID).
		Joins("JOIN schedules ON schedules.schedule_id = appointments.schedule_id").
		Order("schedules.date DESC, schedules.start_time DESC").
		Find(&appointments).Error
	return appointments, err
}

// Update обновляет запись.
func (r *appointmentRepo) Update(appointment *models.Appointment) error {
	return r.db.Save(appointment).Error
}

// DeleteAppointmentAndFreeSlot удаляет запись и освобождает слот в рамках одной транзакции.
func (r *appointmentRepo) DeleteAppointmentAndFreeSlot(appointmentID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var app models.Appointment
		if err := tx.First(&app, appointmentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("запись с ID %d не найдена", appointmentID)
			}
			return err
		}
		if err := tx.Delete(&app).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Schedule{}).Where("schedule_id = ?", app.ScheduleID).Update("is_available", true).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *appointmentRepo) FindUpcomingByPatientID(patientID uint, now time.Time) (*models.Appointment, error) {
	var appointment models.Appointment
	today := now.Format("2006-01-02")

	err := r.db.Joins("JOIN schedules ON schedules.schedule_id = appointments.schedule_id").
		Preload("Schedule.Doctor").
		Where("appointments.patient_id = ? AND appointments.ticket_id IS NULL AND schedules.date = ?", patientID, today).
		Order("schedules.start_time asc").
		First(&appointment).Error

	return &appointment, err
}

func (r *appointmentRepo) AssignTicketToAppointment(appointment *models.Appointment, ticket *models.Ticket) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ticket).Error; err != nil {
			return err
		}
		if err := tx.Model(appointment).Update("ticket_id", ticket.ID).Error; err != nil {
			return err
		}
		return nil
	})
}
