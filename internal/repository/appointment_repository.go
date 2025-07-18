package repository

import (
	"ElectronicQueue/internal/models"

	"gorm.io/gorm"
)

type appointmentRepo struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return &appointmentRepo{db: db}
}

func (r *appointmentRepo) Create(appointment *models.Appointment) error {
	return r.db.Create(appointment).Error
}

func (r *appointmentRepo) Update(appointment *models.Appointment) error {
	return r.db.Save(appointment).Error
}

func (r *appointmentRepo) GetByID(id uint) (*models.Appointment, error) {
	var appointment models.Appointment
	err := r.db.Preload("Patient").
		Preload("Schedule.Doctor").
		First(&appointment, id).Error
	if err != nil {
		return nil, err
	}
	return &appointment, nil
}
