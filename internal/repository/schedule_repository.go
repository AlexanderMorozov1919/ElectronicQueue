package repository

import (
	"ElectronicQueue/internal/models"
	"time"

	"gorm.io/gorm"
)

type scheduleRepo struct {
	db *gorm.DB
}

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

// FindByCabinetAndCurrentTime находит активное расписание для кабинета в данный момент времени.
func (r *scheduleRepo) FindByCabinetAndCurrentTime(cabinetNumber int) (*models.Schedule, error) {
	var schedule models.Schedule
	now := time.Now()
	currentTime := now.Format("15:04:05")

	err := r.db.Preload("Doctor").
		Where("cabinet = ? AND date = ? AND start_time <= ? AND end_time > ?",
			cabinetNumber,
			now.Format("2006-01-02"),
			currentTime,
			currentTime,
		).
		First(&schedule).Error

	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

// FindFirstScheduleForCabinetByDay находит первое расписание для кабинета на сегодня.
// Это нужно, чтобы получить информацию о враче, даже если его смена еще не началась.
func (r *scheduleRepo) FindFirstScheduleForCabinetByDay(cabinetNumber int) (*models.Schedule, error) {
	var schedule models.Schedule
	today := time.Now().Format("2006-01-02")

	err := r.db.Preload("Doctor").
		Where("cabinet = ? AND date = ?", cabinetNumber, today).
		Order("start_time asc").
		First(&schedule).Error

	return &schedule, err
}

// GetAllUniqueCabinets возвращает отсортированный список всех уникальных номеров кабинетов.
func (r *scheduleRepo) GetAllUniqueCabinets() ([]int, error) {
	var cabinets []int

	err := r.db.Model(&models.Schedule{}).
		Distinct().
		Where("cabinet IS NOT NULL").
		Order("cabinet asc").
		Pluck("cabinet", &cabinets).Error

	if err != nil {
		return nil, err
	}
	return cabinets, nil
}

func (r *scheduleRepo) Delete(id uint) error {
	return r.db.Delete(&models.Schedule{}, id).Error
}