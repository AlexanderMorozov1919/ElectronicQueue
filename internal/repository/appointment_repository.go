package repository

import (
	"ElectronicQueue/internal/models/appointment_model"

	"gorm.io/gorm"
)

type appointmentRepo struct {
	db *gorm.DB
}

// NewAppointmentRepository - конструктор для appointmentRepo.
func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return &appointmentRepo{db: db}
}

func (r *appointmentRepo) Create(appointment *appointment_model.Appointment) error {
	return r.db.Create(appointment).Error
}

func (r *appointmentRepo) Update(appointment *appointment_model.Appointment) error {
	return r.db.Save(appointment).Error
}

func (r *appointmentRepo) GetByID(id uint) (*appointment_model.Appointment, error) {
	var appointment appointment_model.Appointment
	err := r.db.Preload("Patient").
		Preload("Schedule.Doctor").
		Preload("Ticket").
		First(&appointment, id).Error
	if err != nil {
		return nil, err
	}
	return &appointment, nil
}

func (r *appointmentRepo) FindByTicketID(ticketID uint) (*appointment_model.Appointment, error) {
	var appointment appointment_model.Appointment
	err := r.db.Preload("Patient").
		Preload("Schedule.Doctor").
		Preload("Ticket").
		Where("ticket_id = ?", ticketID).
		First(&appointment).Error
	if err != nil {
		return nil, err
	}
	return &appointment, nil
}
