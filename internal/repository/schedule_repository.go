package repository

import (
	"time"

	"ElectronicQueue/internal/models/schedule_model"

	"gorm.io/gorm"
)

type scheduleRepo struct {
	db *gorm.DB
}

// NewScheduleRepository - конструктор для scheduleRepo.
func NewScheduleRepository(db *gorm.DB) ScheduleRepository {
	return &scheduleRepo{db: db}
}

func (r *scheduleRepo) Create(schedule *schedule_model.Schedule) error {
	return r.db.Create(schedule).Error
}

func (r *scheduleRepo) Update(schedule *schedule_model.Schedule) error {
	return r.db.Save(schedule).Error
}

func (r *scheduleRepo) GetByID(id uint) (*schedule_model.Schedule, error) {
	var schedule schedule_model.Schedule
	if err := r.db.First(&schedule, id).Error; err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (r *scheduleRepo) FindByDoctorAndDate(doctorID uint, date time.Time) ([]schedule_model.Schedule, error) {
	var schedules []schedule_model.Schedule
	// Ищем по началу дня, чтобы игнорировать время :)
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := r.db.Where("doctor_id = ? AND date >= ? AND date < ?", doctorID, startOfDay, endOfDay).Order("start_time asc").Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}
