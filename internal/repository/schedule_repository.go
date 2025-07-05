package repository

import (
	"ElectronicQueue/internal/models"
	"time"

	"gorm.io/gorm"
)

type scheduleRepo struct {
	db *gorm.DB
}

// NewScheduleRepository - конструктор для scheduleRepo.
func NewScheduleRepository(db *gorm.DB) ScheduleRepository {
	return &scheduleRepo{db: db}
}

func (r *scheduleRepo) Create(schedule *models.Schedule) error {
	return r.db.Create(schedule).Error
}

func (r *scheduleRepo) Update(schedule *models.Schedule) error {
	return r.db.Save(schedule).Error
}

func (r *scheduleRepo) GetByID(id uint) (*models.Schedule, error) {
	var schedule models.Schedule
	if err := r.db.First(&schedule, id).Error; err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (r *scheduleRepo) FindByDoctorAndDate(doctorID uint, date time.Time) ([]models.Schedule, error) {
	var schedules []models.Schedule
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := r.db.Where("doctor_id = ? AND date >= ? AND date < ?", doctorID, startOfDay, endOfDay).Order("start_time asc").Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}
